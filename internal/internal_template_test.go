package internal

import (
	"path/filepath"
	"testing"

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
		Description: "desc",
		Inputs: map[string]ActionInput{
			"foo": {Description: "Foo input", Required: true},
		},
	}
	tmpl := filepath.Join(tmpDir, "templates", "readme.tmpl")
	opts := TemplateOptions{TemplatePath: tmpl, Format: "md"}
	out, err := RenderReadme(action, opts)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if len(out) < 10 || out[0:1] != "#" {
		t.Error("unexpected output content")
	}
}
