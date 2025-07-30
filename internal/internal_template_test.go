package internal

import (
	"testing"
)

func TestRenderReadme(t *testing.T) {
	action := &ActionYML{
		Name:        "MyAction",
		Description: "desc",
		Inputs: map[string]ActionInput{
			"foo": {Description: "Foo input", Required: true},
		},
	}
	tmpl := "../templates/readme.tmpl"
	opts := TemplateOptions{TemplatePath: tmpl, Format: "md"}
	out, err := RenderReadme(action, opts)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if len(out) < 10 || out[0:1] != "#" {
		t.Error("unexpected output content")
	}
}
