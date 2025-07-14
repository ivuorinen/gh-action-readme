package internal

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ivuorinen/gh-action-readme/schemas"
)

func TestValidateActionYML(t *testing.T) {
	tests := []struct {
		name   string
		action ActionYML
		want   []string
	}{
		{
			name: "all fields present",
			action: ActionYML{
				Name:        "foo",
				Description: "desc",
				Runs:        map[string]any{"using": "node20"},
			},
			want: nil,
		},
		{
			name:   "missing name",
			action: ActionYML{Description: "desc", Runs: map[string]any{"using": "node20"}},
			want:   []string{"name"},
		},
		{
			name:   "missing description",
			action: ActionYML{Name: "foo", Runs: map[string]any{"using": "node20"}},
			want:   []string{"description"},
		},
		{
			name:   "missing runs",
			action: ActionYML{Name: "foo", Description: "desc"},
			want:   []string{"runs"},
		},
		{
			name:   "all missing",
			action: ActionYML{},
			want:   []string{"name", "description", "runs"},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := ValidateActionYML(&tt.action)
				if !reflect.DeepEqual(got.MissingFields, tt.want) {
					t.Errorf("ValidateActionYML() = %v, want %v", got.MissingFields, tt.want)
				}
			},
		)
	}
}

func TestYamlToJSON_Simple(t *testing.T) {
	yamlData := []byte(`
name: Example
description: Test
runs:
  using: node20
`)
	jsonBytes, err := yamlToJSON(yamlData)
	if err != nil {
		t.Fatalf("yamlToJSON failed: %v", err)
	}
	// Keys should be strings in the resulting JSON
	if !containsString(jsonBytes, `"name":"Example"`) {
		t.Errorf("json output missing name: %s", string(jsonBytes))
	}
	if !containsString(jsonBytes, `"using":"node20"`) {
		t.Errorf("json output missing runs.using: %s", string(jsonBytes))
	}
}

func TestYamlToJSON_Nested(t *testing.T) {
	yamlData := []byte(`
inputs:
  foo:
    description: Foo input
    required: true
`)
	jsonBytes, err := yamlToJSON(yamlData)
	if err != nil {
		t.Fatalf("yamlToJSON failed: %v", err)
	}
	if !containsString(jsonBytes, `"inputs"`) || !containsString(jsonBytes, `"foo"`) {
		t.Errorf("json output missing nested keys: %s", string(jsonBytes))
	}
}

func TestValidateActionYMLSchema_InvalidFile(t *testing.T) {
	_, err := ValidateActionYMLSchema("nonexistent.yml", schemas.RelPath)
	if err == nil {
		t.Error("expected error for missing action.yml file")
	}
}

func TestValidateActionYMLSchema_ValidMinimal(t *testing.T) {
	// Minimal valid action.yml and schema file in a temp dir
	tmpDir := t.TempDir()
	actionPath := filepath.Join(tmpDir, "action.yml")
	schemaPath := filepath.Join(tmpDir, "action.schema.json")

	actionContent := `
name: foo
description: bar
runs:
  using: node20
`
	schemaContent := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["name", "description", "runs"],
  "properties": {
    "name": {"type": "string"},
    "description": {"type": "string"},
    "runs": {"type": "object"}
  }
}`

	if err := os.WriteFile(actionPath, []byte(actionContent), 0o600); err != nil {
		t.Fatalf("failed to write action.yml: %v", err)
	}
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0o600); err != nil {
		t.Fatalf("failed to write schema: %v", err)
	}

	errs, err := ValidateActionYMLSchema(actionPath, schemaPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) != 0 {
		t.Errorf("expected no schema errors, got: %v", errs)
	}
}

// --- Advanced YAML features: anchors/aliases ---

func TestYamlToJSON_AnchorsAliases(t *testing.T) {
	yamlData := []byte(`
defaults: &defaults
  description: Default description
  required: false

inputs:
  foo:
    <<: *defaults
    description: Overridden description
    required: true
  bar:
    <<: *defaults
`)
	jsonBytes, err := yamlToJSON(yamlData)
	if err != nil {
		t.Fatalf("yamlToJSON failed: %v", err)
	}
	if !containsString(jsonBytes, `"foo"`) || !containsString(jsonBytes, `"bar"`) {
		t.Errorf("json output missing keys: %s", string(jsonBytes))
	}
	if !containsString(jsonBytes, `"description":"Overridden description"`) {
		t.Errorf("anchor/alias merge did not override description: %s", string(jsonBytes))
	}
	if !containsString(jsonBytes, `"required":true`) {
		t.Errorf("anchor/alias merge did not override required: %s", string(jsonBytes))
	}
	if !containsString(jsonBytes, `"required":false`) {
		t.Errorf("anchor/alias merge did not preserve default required: %s", string(jsonBytes))
	}
}

// ValidateActionYMLSchema should accept anchors/aliases if schema is otherwise valid
func TestValidateActionYMLSchema_AnchorsAliases(t *testing.T) {
	projectRoot := findProjectRoot(t)
	actionPath := filepath.Join(projectRoot, "testdata", "example-action-anchors", "action.yml")
	schemaPath := filepath.Join(projectRoot, "schemas", "action.schema.json")

	errs, err := ValidateActionYMLSchema(actionPath, schemaPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) != 0 {
		t.Errorf("expected no schema errors, got: %v", errs)
	}
}
