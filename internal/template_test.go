package internal

import (
	"os"
	"path/filepath"
	"strings"
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
		Dependencies: []ActionDependency{
			{Name: "actions/checkout", Version: "v4", Ref: "v4", Pinned: false},
		},
	}
	opts := TemplateOptions{
		TemplateContent: filepath.Join(root, "templates/readme"),
		HeaderBase:      filepath.Join(root, "templates/header"),
		FooterBase:      filepath.Join(root, "templates/footer"),
		Format:          "md",
		Org:             "testorg",
		Repo:            "monorepo/someaction",
		Version:         "release-tag",
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
	if !contains(out, "Dependencies") || !contains(out, "actions/checkout") {
		t.Errorf("expected dependencies table in output: %s", out)
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
		Dependencies: []ActionDependency{
			{Name: "actions/setup-node", Version: "v4", Ref: "v4", Pinned: false},
		},
	}
	opts := TemplateOptions{
		TemplateContent: filepath.Join(root, "templates/readme"),
		HeaderBase:      filepath.Join(root, "templates/header"),
		FooterBase:      filepath.Join(root, "templates/footer"),
		Format:          "html",
		Org:             "testorghtml",
		Repo:            "foo/bar/action",
		Version:         "main",
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
	if !contains(out, "<table>") || !contains(out, "<thead>") {
		t.Errorf("expected tables for inputs/outputs: %s", out)
	}
	if !contains(out, "Dependencies") || !contains(out, "actions/setup-node") {
		t.Errorf("expected dependencies table in HTML output: %s", out)
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
		TemplateContent: filepath.Join(root, "templates/readme"),
		HeaderBase:      filepath.Join(root, "templates/header"),
		FooterBase:      filepath.Join(root, "templates/footer"),
		Format:          "md",
		Org:             "edgeorg",
		Repo:            "edge/repo",
		Version:         "edge",
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

func TestRenderReadme_VersionPlaceholder(t *testing.T) {
	tmp := t.TempDir()
	tmplPath := filepath.Join(tmp, "readme.md.tmpl")
	err := os.WriteFile(tmplPath, []byte("Version: {version}"), 0o600)
	if err != nil {
		t.Fatalf("failed to write temp template: %v", err)
	}
	action := &ActionYML{Name: "x", Description: "y", Runs: map[string]any{"using": "node20"}}
	opts := TemplateOptions{
		TemplateContent: strings.TrimSuffix(tmplPath, ".md.tmpl"),
		Format:          "md",
		Version:         "v1.2.3",
	}
	out, err := RenderReadme(action, opts)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if !contains(out, "v1.2.3") {
		t.Errorf("version placeholder not replaced: %s", out)
	}
}

func TestParseActionYML_DocsBlock(t *testing.T) {
	root := findProjectRoot(t)
	path := filepath.Join(root, "testdata/example-action-docs/action.yml")
	action, err := ParseActionYML(path)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	want := "This is a longer description for testing.\n\nIt spans multiple lines."
	if action.LongDescription != want {
		t.Errorf("unexpected long description: %q", action.LongDescription)
	}
}

func TestParseActionYML_Dependencies(t *testing.T) {
	root := findProjectRoot(t)
	path := filepath.Join(root, "testdata/example-action-complex/action.yml")
	action, err := ParseActionYML(path)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(action.Dependencies) != 3 {
		t.Fatalf("expected 3 dependencies, got %d", len(action.Dependencies))
	}
	if action.Dependencies[0].Name != "actions/checkout" || !action.Dependencies[0].Pinned {
		t.Errorf("unexpected dependency parsing: %+v", action.Dependencies[0])
	}
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
