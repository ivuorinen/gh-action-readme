package internal

import (
	"path/filepath"
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
		Name:        "MyAction",
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
}
