package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// renderThemeForHygiene generates a README for the given theme into a temp dir and
// returns its content. Uses a minimal action so every optional section (inputs,
// outputs, permissions, dependencies) is absent — the case whose collapsed guards
// used to leave long blank-line runs behind.
func renderThemeForHygiene(t *testing.T, theme string) string {
	t.Helper()

	tmpDir := t.TempDir()
	actionPath := testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureMinimalComposite)

	cfg := DefaultAppConfig()
	cfg.Theme = theme
	cfg.OutputDir = tmpDir
	cfg.Quiet = true

	gen := NewGenerator(cfg)
	if err := gen.GenerateFromFile(actionPath); err != nil {
		t.Fatalf("generate %s: %v", theme, err)
	}

	// #nosec G304 -- path is t.TempDir() joined with a constant file name
	data, err := os.ReadFile(filepath.Join(tmpDir, appconstants.ReadmeMarkdown))
	if err != nil {
		t.Fatalf("read generated README for %s: %v", theme, err)
	}

	return string(data)
}

// TestGeneratedMarkdownHasNoBlankLineRuns is the regression guard for
// audit-46c56415: absent optional sections used to leave 7-8 consecutive blank
// lines, producing 11 markdownlint MD012 violations in a single generated file.
// "\n\n\n" is a dependency-free proxy for MD012 that runs in `go test` without
// requiring node.
func TestGeneratedMarkdownHasNoBlankLineRuns(t *testing.T) {
	t.Parallel()

	for _, theme := range appconstants.GetSupportedThemes() {
		t.Run(theme, func(t *testing.T) {
			t.Parallel()

			content := renderThemeForHygiene(t, theme)
			if strings.Contains(content, "\n\n\n") {
				t.Errorf(
					"%s theme output contains a run of blank lines (markdownlint MD012); "+
						"guard optional sections with {{- if}} / {{- end}}",
					theme,
				)
			}
		})
	}
}

// TestGeneratedMarkdownHeadingsHaveBlankLineAbove guards MD022: trimming with
// {{- if}} removes the separator before a section heading, so the blank line must
// live inside the conditional. Anything else renders "text\n## Heading".
func TestGeneratedMarkdownHeadingsHaveBlankLineAbove(t *testing.T) {
	t.Parallel()

	for _, theme := range appconstants.GetSupportedThemes() {
		t.Run(theme, func(t *testing.T) {
			t.Parallel()

			lines := strings.Split(renderThemeForHygiene(t, theme), "\n")
			for i, line := range lines {
				if i == 0 || !strings.HasPrefix(line, "#") {
					continue
				}
				if strings.TrimSpace(lines[i-1]) != "" {
					t.Errorf("%s theme: heading %q on line %d has no blank line above it (MD022)",
						theme, line, i+1)
				}
			}
		})
	}
}
