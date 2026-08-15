package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// Local constants so repeated literals stay single-sourced (no-constant-duplication).
const (
	licFileLICENSE = "LICENSE"
	licFileCOPYING = "COPYING"
	licBodyMIT     = "MIT License\n"
	// licFileBritish is the British spelling, which findLicenseFile must also match.
	// The nolint is load-bearing: golangci-lint runs misspell with locale: US and is
	// invoked with --fix by both the pre-commit hook and the post-edit hook, so
	// without it the literal is rewritten to "LICENSE" — silently making this a
	// duplicate of licFileLICENSE and deleting the coverage it exists to provide.
	licFileBritish = "LICENCE" //nolint:misspell // deliberate British spelling under test
)

// writeLicenseFile drops a license file with the given name and body into dir.
func writeLicenseFile(t *testing.T, dir, name, body string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), appconstants.FilePermDefault); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

// TestLicenseFileRankOrdering pins the ordering contract licenseFileRank exists to
// provide: LICENSE outranks COPYING, and a bare name outranks a suffixed one. The
// ranks must be strictly ordered, not merely distinct — findLicenseFile sorts on
// them, so an inverted pair silently changes which file a repository advertises.
func TestLicenseFileRankOrdering(t *testing.T) {
	t.Parallel()

	// Ascending order of preference: every entry must rank strictly below the next.
	ordered := []string{
		licFileLICENSE,          // bare LICENSE — the most preferred
		"LICENSE.md",            // suffixed LICENSE
		licFileCOPYING,          // bare COPYING loses to any LICENSE
		licFileCOPYING + ".txt", // suffixed COPYING — least preferred
	}

	for i := 0; i < len(ordered)-1; i++ {
		lo, hi := licenseFileRank(ordered[i]), licenseFileRank(ordered[i+1])
		if lo >= hi {
			t.Errorf("licenseFileRank(%q) = %d must sort before licenseFileRank(%q) = %d",
				ordered[i], lo, ordered[i+1], hi)
		}
	}
}

// TestLicenseFileRankIsCaseInsensitive guards the lowercasing step: a repository
// using lowercase names must rank identically to the uppercase form, or the
// COPYING-loses-to-LICENSE rule silently stops applying to it.
func TestLicenseFileRankIsCaseInsensitive(t *testing.T) {
	t.Parallel()

	pairs := [][2]string{
		{licFileLICENSE, "license"},
		{licFileCOPYING, "copying"},
		{"LICENSE.md", "license.MD"},
	}

	for _, p := range pairs {
		if got, want := licenseFileRank(p[1]), licenseFileRank(p[0]); got != want {
			t.Errorf("licenseFileRank(%q) = %d, want %d (same as %q)", p[1], got, want, p[0])
		}
	}
}

// TestParseLicenseFromHeaderAndYAML pins the two in-file license sources and their
// precedence: the top-level `license:` key beats the `# license:` header comment.
func TestParseLicenseFromHeaderAndYAML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		fixture string
		want    string
	}{
		{"header comment only", testutil.TestFixtureLicenseHeaderComment, appconstants.SPDXApache2},
		{
			"yaml key beats header comment",
			testutil.TestFixtureLicenseYAMLKeyWins, appconstants.SPDXBSD3Clause,
		},
		{"quoted value with inline comment", testutil.TestFixtureLicenseQuoted, appconstants.SPDXMPL2},
		{"no license declared", testutil.TestFixtureLicenseNone, ""},
		{
			// Regression: the license line is what dedents out of the permissions
			// block, so it must be re-offered to the scanner rather than swallowed
			// by the block it closed.
			"license after an indented permissions entry",
			testutil.TestFixtureLicenseAfterPermissions, appconstants.SPDXApache2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			actionPath := testutil.WriteActionFixture(t, tmpDir, tt.fixture)

			action, err := ParseActionYML(actionPath)
			if err != nil {
				t.Fatalf("ParseActionYML: %v", err)
			}
			testutil.AssertEqual(t, tt.want, action.License)
		})
	}
}

// TestParseLicenseDoesNotBreakPermissions guards the shared header scan: adding the
// license field must not disturb the permissions block parsed from the same comments.
func TestParseLicenseDoesNotBreakPermissions(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	actionPath := testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureLicenseHeaderComment)

	action, err := ParseActionYML(actionPath)
	if err != nil {
		t.Fatalf("ParseActionYML: %v", err)
	}
	testutil.AssertEqual(t, "Apache-2.0", action.License)
	testutil.AssertEqual(t, appconstants.PermissionRead, action.Permissions["contents"])
}

// TestDetectLicense covers the LICENSE-file sniffer, including the multi-line titles
// that a newline-naive regex silently fails to match.
func TestDetectLicense(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fileName string
		body     string
		want     string
	}{
		{
			name:     "apache title spans a newline",
			fileName: licFileLICENSE,
			body:     "Apache License\nVersion 2.0, January 2004\nhttp://www.apache.org/licenses/\n",
			want:     "Apache-2.0",
		},
		{"mit", licFileLICENSE, "MIT License\n\nCopyright (c) 2025 Someone\n", appconstants.SPDXMIT},
		{
			"gpl3", licFileCOPYING,
			"GNU GENERAL PUBLIC LICENSE\nVersion 3, 29 June 2007\n", appconstants.SPDXGPL3,
		},
		{
			"gpl2", licFileCOPYING,
			"GNU GENERAL PUBLIC LICENSE\nVersion 2, June 1991\n", appconstants.SPDXGPL2,
		},
		{
			"agpl beats gpl", licFileLICENSE,
			"GNU AFFERO GENERAL PUBLIC LICENSE\nVersion 3\n", appconstants.SPDXAGPL3,
		},
		{"lowercase filename", "license.md", licBodyMIT, appconstants.SPDXMIT},
		{"british spelling variant", licFileBritish, licBodyMIT, appconstants.SPDXMIT},
		{
			name:     "explicit SPDX tag is authoritative",
			fileName: licFileLICENSE,
			body:     "SPDX-License-Identifier: BSD-2-Clause\nApache License\nVersion 2.0\n",
			want:     "BSD-2-Clause",
		},
		{"unrecognized text yields empty", licFileLICENSE, "All rights reserved. Ask nicely.\n", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			writeLicenseFile(t, dir, tt.fileName, tt.body)
			testutil.AssertEqual(t, tt.want, detectLicense(dir))
		})
	}
}

// TestDetectLicenseNoFile confirms a repo with no license file yields "" rather than
// a default — asserting an undeclared license is the bug this whole path exists to fix.
func TestDetectLicenseNoFile(t *testing.T) {
	t.Parallel()

	testutil.AssertEqual(t, "", detectLicense(t.TempDir()))
	testutil.AssertEqual(t, "", detectLicense(""))
}

// TestResolveLicensePrecedence pins the full chain:
// config > action.yml > (header comment, already folded in) > detected file.
func TestResolveLicensePrecedence(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	writeLicenseFile(t, repoRoot, licFileLICENSE, licBodyMIT)

	tests := []struct {
		name          string
		configLicense string
		actionLicense string
		want          string
	}{
		{
			"config wins over everything",
			appconstants.SPDXGPL3, appconstants.SPDXBSD3Clause, appconstants.SPDXGPL3,
		},
		{
			"action wins over detection",
			"", appconstants.SPDXBSD3Clause, appconstants.SPDXBSD3Clause,
		},
		{"falls back to detection", "", "", appconstants.SPDXMIT},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			action := &ActionYML{License: tt.actionLicense}
			config := &AppConfig{License: tt.configLicense}
			testutil.AssertEqual(t, tt.want, resolveLicense(action, config, repoRoot))
		})
	}
}

// TestResolveRepoFilePathRejectsEscapes guards the repoFile contract: the name comes
// from a template, which may be a user-supplied --template, so it must not be able to
// resolve a file outside the repository and emit a link to it.
func TestResolveRepoFilePathRejectsEscapes(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	writeLicenseFile(t, repoRoot, licFileLICENSE, licBodyMIT)
	// A real file outside the repo that an escape would otherwise resolve.
	outside := t.TempDir()
	writeLicenseFile(t, outside, "SECRET.md", "top secret\n")

	escapes := []string{
		"../SECRET.md",
		"../../etc/passwd",
		filepath.Join(outside, "SECRET.md"), // absolute
		"./LICENSE",                         // changes under Clean
		"sub/../../SECRET.md",
	}
	for _, name := range escapes {
		if got := resolveRepoFilePath(repoRoot, name); got != "" {
			t.Errorf("resolveRepoFilePath(%q) = %q, want empty (escapes the repository)", name, got)
		}
	}

	// A legitimate in-repo name still resolves.
	if resolveRepoFilePath(repoRoot, licFileLICENSE) == "" {
		t.Error("resolveRepoFilePath(LICENSE) returned empty for a file that exists")
	}
}

// TestResolveLicenseUnknownStaysEmpty is the regression guard for the original defect:
// with nothing declared and no license file, the result must be empty so templates
// render no license section at all.
func TestResolveLicenseUnknownStaysEmpty(t *testing.T) {
	t.Parallel()

	got := resolveLicense(&ActionYML{}, &AppConfig{}, t.TempDir())
	if got != "" {
		t.Errorf("resolveLicense with nothing declared = %q, want empty", got)
	}
}
