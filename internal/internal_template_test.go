package internal

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

func TestRenderReadme(t *testing.T) {
	t.Parallel()
	// Set up test templates
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()
	testutil.SetupTestTemplates(t, tmpDir)

	action := &ActionYML{
		Name:        testutil.TestActionNameMyAction,
		Description: testGenShortDesc,
		Inputs: map[string]ActionInput{
			"foo": {Description: "Foo input", Required: true},
		},
	}
	tmpl := filepath.Join(tmpDir, "templates", testutil.TestTemplateReadme)
	opts := TemplateOptions{TemplatePath: tmpl, Format: appconstants.OutputFormatMarkdown}
	out, err := RenderReadme(action, opts)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if len(out) < 10 || out[0:1] != "#" {
		t.Error("unexpected output content")
	}
	// Verify the action data actually flowed into the output, not just that the
	// template's literal "#" header rendered. A regressed data pipeline (empty
	// .Name / dropped Inputs) would still satisfy the length+prefix check above.
	if !strings.Contains(out, testutil.TestActionNameMyAction) {
		t.Errorf("rendered output missing action name %q; got:\n%s", testutil.TestActionNameMyAction, out)
	}
}
