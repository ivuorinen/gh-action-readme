package internal

import (
	"strings"
	"testing"
)

// TestGetActionVersionStripsLeadingAt verifies that a Config.Version pasted with
// a leading '@' (e.g. "@v2", copied from a uses: line) is accepted and normalized
// rather than silently dropped to the default fallback.
func TestGetActionVersionStripsLeadingAt(t *testing.T) {
	t.Parallel()

	td := &TemplateData{Config: &AppConfig{Version: "@v2"}}
	if got := getActionVersion(td); got != "v2" {
		t.Errorf("getActionVersion(@v2) = %q, want v2", got)
	}

	tdPlain := &TemplateData{Config: &AppConfig{Version: "v3"}}
	if got := getActionVersion(tdPlain); got != "v3" {
		t.Errorf("getActionVersion(v3) = %q, want v3", got)
	}
}

// TestJSONWriterReferenceResolution verifies the JSON writer uses the resolved git
// uses-statement and repository URL when present, and falls back to placeholders
// only when they are unavailable.
func TestJSONWriterReferenceResolution(t *testing.T) {
	t.Parallel()

	action := &ActionYML{Name: "Foo"}
	jw := &JSONWriter{Config: &AppConfig{}}

	if got := jw.usesReference(action); !strings.HasPrefix(got, "your-org/") || !strings.HasSuffix(got, "@v1") {
		t.Errorf("usesReference fallback = %q, want your-org/...@v1", got)
	}
	if got := jw.repositoryLink(action); !strings.HasPrefix(got, "https://github.com/your-org/") {
		t.Errorf("repositoryLink fallback = %q, want https://github.com/your-org/...", got)
	}

	jw.usesStatement = "acme/foo@main"
	jw.repoURL = "https://github.com/acme/foo"
	if got := jw.usesReference(action); got != "acme/foo@main" {
		t.Errorf("usesReference resolved = %q, want acme/foo@main", got)
	}
	if got := jw.repositoryLink(action); got != "https://github.com/acme/foo" {
		t.Errorf("repositoryLink resolved = %q, want https://github.com/acme/foo", got)
	}
}
