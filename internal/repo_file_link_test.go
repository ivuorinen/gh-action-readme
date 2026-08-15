package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

// repoFileContributing is the non-LICENSE repo file used to exercise the exact-name
// branch of resolveRepoFilePath (LICENSE takes the findLicenseFile branch instead).
const repoFileContributing = "CONTRIBUTING.md"

// newLinkRepo builds a repository containing CONTRIBUTING.md and LICENSE at its
// root, plus an actions/foo/ subdirectory, and returns the symlink-resolved root.
// The resolution matters: on macOS t.TempDir() hands back a /var path that is a
// symlink to /private/var, and repoFileLink resolves symlinks internally — an
// unresolved root would make isWithin report a false negative.
func newLinkRepo(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}

	writeLicenseFile(t, root, licFileLICENSE, licBodyMIT)
	writeLicenseFile(t, root, repoFileContributing, "# Contributing\n")

	if err := os.MkdirAll(filepath.Join(root, "actions", "foo"), appconstants.FilePermDir); err != nil {
		t.Fatalf("mkdir actions/foo: %v", err)
	}

	return root
}

// TestRepoFileLinkRelativeToDocument is the core contract: the emitted link is
// relative to the directory the document lands in, so a monorepo action documented
// two levels down reaches the root file by climbing exactly two levels.
func TestRepoFileLinkRelativeToDocument(t *testing.T) {
	t.Parallel()

	root := newLinkRepo(t)

	tests := []struct {
		name      string
		outputDir string
		file      string
		want      string
	}{
		{"root document, root file", root, repoFileContributing, repoFileContributing},
		{
			"nested document climbs out",
			filepath.Join(root, "actions", "foo"),
			repoFileContributing,
			"../../" + repoFileContributing,
		},
		{"LICENSE resolves via findLicenseFile", root, licFileLICENSE, licFileLICENSE},
		{
			"nested document reaches LICENSE",
			filepath.Join(root, "actions", "foo"),
			licFileLICENSE,
			"../../" + licFileLICENSE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			td := &TemplateData{RepoRoot: root, OutputDir: tt.outputDir}
			if got := repoFileLink(td, tt.file); got != tt.want {
				t.Errorf("repoFileLink(%q) = %q, want %q", tt.file, got, tt.want)
			}
		})
	}
}

// TestRepoFileLinkRefusesEscapes is the security guard: a template supplies the
// name, and a user-provided --template is not trusted. Every form that would reach
// outside the repository must yield "" rather than a link that climbs out of the
// tree or points at an absolute system path.
func TestRepoFileLinkRefusesEscapes(t *testing.T) {
	t.Parallel()

	root := newLinkRepo(t)
	td := &TemplateData{RepoRoot: root, OutputDir: root}

	escapes := []string{
		"/etc/passwd",                // absolute
		"../" + repoFileContributing, // walks upward
		"sub/../../SECRET.md",        // walks upward after a segment
		"./" + repoFileContributing,  // changes under Clean
		"..",                         // the parent itself
		"does-not-exist.md",          // in-repo but absent
		"",                           // empty name
	}

	for _, name := range escapes {
		if got := repoFileLink(td, name); got != "" {
			t.Errorf("repoFileLink(%q) = %q, want \"\" (must not link outside the repository)", name, got)
		}
	}
}

// TestRepoFileLinkDocumentOutsideRepo covers the isWithin guard: when --output
// places the document outside the repository, no relative path is sensible and the
// link must be omitted rather than emitted as a chain of "..".
func TestRepoFileLinkDocumentOutsideRepo(t *testing.T) {
	t.Parallel()

	root := newLinkRepo(t)
	outside := t.TempDir()

	td := &TemplateData{RepoRoot: root, OutputDir: outside}
	if got := repoFileLink(td, repoFileContributing); got != "" {
		t.Errorf("repoFileLink from a document outside the repo = %q, want \"\"", got)
	}
}

// TestRepoFileLinkRejectsBadData covers the guard clauses: anything that is not
// *TemplateData, or that carries no repo root, yields "" instead of panicking.
func TestRepoFileLinkRejectsBadData(t *testing.T) {
	t.Parallel()

	if got := repoFileLink("not template data", licFileLICENSE); got != "" {
		t.Errorf("repoFileLink(non-TemplateData) = %q, want \"\"", got)
	}

	if got := repoFileLink(&TemplateData{}, licFileLICENSE); got != "" {
		t.Errorf("repoFileLink with empty RepoRoot = %q, want \"\"", got)
	}
}

// TestDocumentDirPrecedence pins the OutputDir-over-ActionPath fallback. Getting
// this backwards would compute every link against the action's source directory
// rather than where the document is actually written.
func TestDocumentDirPrecedence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		td   *TemplateData
		want string
	}{
		{
			"OutputDir wins",
			&TemplateData{OutputDir: "/out", ActionPath: testTplRepoRoot + "/actions/foo/action.yml"},
			"/out",
		},
		{
			"falls back to the action's directory",
			&TemplateData{ActionPath: testTplRepoRoot + "/actions/foo/action.yml"},
			testTplRepoRoot + "/actions/foo",
		},
		{"neither set", &TemplateData{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := documentDir(tt.td); got != tt.want {
				t.Errorf("documentDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestIsWithin pins containment, including the boundary cases a naive
// strings.HasPrefix check gets wrong: a sibling directory sharing a name prefix
// ("/repo-other" under "/repo") is NOT within, and the root itself IS.
func TestIsWithin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		root string
		path string
		want bool
	}{
		{"root is within itself", testTplRepoRoot, testTplRepoRoot, true},
		{"child is within", testTplRepoRoot, testTplRepoRoot + "/actions/foo", true},
		{"parent is not within", testTplRepoRoot, "/", false},
		{"sibling is not within", testTplRepoRoot, "/other", false},
		{"name-prefix sibling is not within", testTplRepoRoot, testTplRepoRoot + "-other", false},
		{"empty root", "", testTplRepoRoot, false},
		{"empty path", testTplRepoRoot, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := isWithin(tt.root, tt.path); got != tt.want {
				t.Errorf("isWithin(%q, %q) = %v, want %v", tt.root, tt.path, got, tt.want)
			}
		})
	}
}
