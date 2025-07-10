package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRenderReadmeMarkdownOrgRepoVersion(t *testing.T) {
	root := findProjectRoot(t)
	action := &ActionYML{
		Name:        "MyAction",
		Description: "desc",
		Inputs: map[string]ActionInput{
			"foo": {Description: "Foo input", Required: true},
		},
	}
	opts := TemplateOptions{
		TemplateBase: filepath.Join(root, "templates/readme"),
		HeaderBase:   filepath.Join(root, "templates/header"),
		FooterBase:   filepath.Join(root, "templates/footer"),
		Format:       "md",
		Org:          "testorg",
		Repo:         "monorepo/someaction",
		Version:      "release-tag",
	}
	out, err := RenderReadme(action, opts)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if len(out) < 10 || out[0:1] != "#" {
		t.Error("unexpected markdown output")
	}
	if want := "testorg/monorepo/someaction@release-tag"; !contains(out, want) {
		t.Errorf("expected uses block: %s", want)
	}
}

func TestRenderReadmeHTMLOrgRepoVersion(t *testing.T) {
	root := findProjectRoot(t)
	action := &ActionYML{
		Name:        "MyAction",
		Description: "desc",
		Inputs: map[string]ActionInput{
			"foo": {Description: "Foo input", Required: true},
		},
	}
	opts := TemplateOptions{
		TemplateBase: filepath.Join(root, "templates/readme"),
		HeaderBase:   filepath.Join(root, "templates/header"),
		FooterBase:   filepath.Join(root, "templates/footer"),
		Format:       "html",
		Org:          "testorghtml",
		Repo:         "foo/bar/action",
		Version:      "main",
	}
	out, err := RenderReadme(action, opts)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if len(out) < 10 || (out[:4] != "<h1>" && out[:4] != "<!DO") {
		t.Error("unexpected HTML output")
	}
	if want := "testorghtml/foo/bar/action@main"; !contains(out, want) {
		t.Errorf("expected uses block: %s", want)
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && substr != "" && (len(substr) <= len(s)) && stringContains(s, substr)
}

func TestRenderReadme_EdgeCases(t *testing.T) {
	root := findProjectRoot(t)
	// Test with missing optional fields and extra unknown fields
	action := &ActionYML{
		Name:        "Edge Case Action",
		Description: "This action is used to test edge cases in parsing and validation.",
		Inputs: map[string]ActionInput{
			"required_input": {Description: "A required input", Required: true},
			"optional_input": {Description: "An optional input"},
		},
		Outputs:  map[string]ActionOutput{}, // Outputs intentionally empty
		Runs:     map[string]any{"using": "node20", "main": "dist/index.js"},
		Branding: &Branding{Icon: "zap", Color: "yellow"},
		// No author, no custom fields
	}
	opts := TemplateOptions{
		TemplateBase: filepath.Join(root, "templates/readme"),
		HeaderBase:   filepath.Join(root, "templates/header"),
		FooterBase:   filepath.Join(root, "templates/footer"),
		Format:       "md",
		Org:          "edgeorg",
		Repo:         "edge/repo",
		Version:      "edge",
	}
	out, err := RenderReadme(action, opts)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if !contains(out, "Edge Case Action") {
		t.Errorf("expected action name in output")
	}
	if contains(out, "Outputs") {
		t.Errorf("should not render Outputs section when outputs are empty")
	}
	if !contains(out, "required_input") || !contains(out, "optional_input") {
		t.Errorf("inputs missing in output")
	}
	// Should not error or panic on missing author/custom fields
}

// findProjectRoot walks up from the current directory until it finds go.mod.
func findProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working directory: %v", err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find project root (go.mod)")
		}
		dir = parent
	}
}
