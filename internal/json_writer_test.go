package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newTestJSONWriter creates a JSONWriter with a default config for testing.
func newTestJSONWriter() *JSONWriter {
	return NewJSONWriter(DefaultAppConfig())
}

// actionWithNoInputsOutputs returns an ActionYML with no inputs or outputs.
func actionWithNoInputsOutputs() *ActionYML {
	return &ActionYML{
		Name:        "My Action",
		Description: "A test action",
		Runs:        map[string]any{"using": "composite", "steps": []any{}},
	}
}

// actionWithInputs returns an ActionYML with one input, optionally with a default value.
func actionWithInputs(withDefault bool) *ActionYML {
	var defaultVal any
	if withDefault {
		defaultVal = "my-default"
	}

	return &ActionYML{
		Name:        "My Action",
		Description: "A test action with inputs",
		Inputs: map[string]ActionInput{
			"token": {
				Description: "The token to use",
				Required:    true,
				Default:     defaultVal,
			},
		},
		Runs: map[string]any{"using": "composite", "steps": []any{}},
	}
}

// actionWithOutputs returns an ActionYML with one output.
func actionWithOutputs() *ActionYML {
	return &ActionYML{
		Name:        "My Action",
		Description: "A test action with outputs",
		Outputs: map[string]ActionOutput{
			"result": {
				Description: "The result value",
			},
		},
		Runs: map[string]any{"using": "composite", "steps": []any{}},
	}
}

// TestConvertToJSONOutput_NoInputs verifies that an action with no inputs
// produces a JSON output with no "inputs" section.
func TestConvertToJSONOutput_NoInputs(t *testing.T) {
	jw := newTestJSONWriter()
	action := actionWithNoInputsOutputs()

	out := jw.convertToJSONOutput(action)
	if len(out.Action.Inputs) > 0 {
		t.Errorf("expected no inputs in JSON output, got %d", len(out.Action.Inputs))
	}

	assertJSONSectionsForType(t, out, "inputs", false)
}

// TestConvertToJSONOutput_WithInputs verifies that an action with inputs
// produces a JSON output that includes an "inputs" section.
func TestConvertToJSONOutput_WithInputs(t *testing.T) {
	jw := newTestJSONWriter()
	action := actionWithInputs(false)

	out := jw.convertToJSONOutput(action)
	if len(out.Action.Inputs) == 0 {
		t.Error("expected inputs in JSON output, got none")
	}

	assertJSONSectionsForType(t, out, "inputs", true)
}

// TestConvertToJSONOutput_NoOutputs verifies that an action with no outputs
// produces a JSON output without an "outputs" section.
func TestConvertToJSONOutput_NoOutputs(t *testing.T) {
	jw := newTestJSONWriter()
	action := actionWithNoInputsOutputs()

	out := jw.convertToJSONOutput(action)
	if len(out.Action.Outputs) > 0 {
		t.Errorf("expected no outputs in JSON output, got %d", len(out.Action.Outputs))
	}

	assertJSONSectionsForType(t, out, "outputs", false)
}

// TestConvertToJSONOutput_WithOutputs verifies that an action with outputs
// produces a JSON output that includes an "outputs" section.
func TestConvertToJSONOutput_WithOutputs(t *testing.T) {
	jw := newTestJSONWriter()
	action := actionWithOutputs()

	out := jw.convertToJSONOutput(action)
	if len(out.Action.Outputs) == 0 {
		t.Error("expected outputs in JSON output, got none")
	}

	assertJSONSectionsForType(t, out, "outputs", true)
}

// TestGenerateBasicExample_NoInputs verifies that generateBasicExample with no
// inputs produces YAML that does not include a "with:" block.
func TestGenerateBasicExample_NoInputs(t *testing.T) {
	jw := newTestJSONWriter()
	action := actionWithNoInputsOutputs()

	example := jw.generateBasicExample(action)
	if strings.Contains(example, "with:") {
		t.Errorf("expected no 'with:' block for action with no inputs, got:\n%s", example)
	}
}

// TestGenerateBasicExample_WithInputs verifies that generateBasicExample with inputs
// includes a "with:" block in the YAML output.
func TestGenerateBasicExample_WithInputs(t *testing.T) {
	jw := newTestJSONWriter()
	action := actionWithInputs(false)

	example := jw.generateBasicExample(action)
	if !strings.Contains(example, "with:") {
		t.Errorf("expected 'with:' block for action with inputs, got:\n%s", example)
	}
}

// TestGenerateBasicExample_InputWithDefault verifies that when an input has a Default
// value, generateBasicExample uses that value instead of the placeholder "value".
func TestGenerateBasicExample_InputWithDefault(t *testing.T) {
	jw := newTestJSONWriter()
	action := actionWithInputs(true)

	example := jw.generateBasicExample(action)
	if !strings.Contains(example, "my-default") {
		t.Errorf("expected default value 'my-default' in example output, got:\n%s", example)
	}
}

// TestGenerateBasicExample_InputWithNilDefault verifies that when an input has
// Default == nil, generateBasicExample uses the placeholder "value".
func TestGenerateBasicExample_InputWithNilDefault(t *testing.T) {
	jw := newTestJSONWriter()
	action := actionWithInputs(false)

	example := jw.generateBasicExample(action)
	if !strings.Contains(example, `"value"`) {
		t.Errorf("expected placeholder value 'value' in example output when Default is nil, got:\n%s", example)
	}
}

// TestJSONWriter_Write verifies end-to-end JSON file generation and validates
// that inputs/outputs sections are correctly included or excluded.
func TestJSONWriter_Write(t *testing.T) {
	tests := []struct {
		name          string
		action        *ActionYML
		wantInputKey  string
		wantOutputKey string
	}{
		{
			name:   "action with no inputs or outputs",
			action: actionWithNoInputsOutputs(),
		},
		{
			name:         "action with inputs only",
			action:       actionWithInputs(true),
			wantInputKey: "token",
		},
		{
			name:          "action with outputs only",
			action:        actionWithOutputs(),
			wantOutputKey: "result",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			outputPath := filepath.Join(dir, "output.json")

			jw := newTestJSONWriter()
			if err := jw.Write(tt.action, outputPath); err != nil {
				t.Fatalf("Write() error = %v", err)
			}

			data, err := os.ReadFile(outputPath) // #nosec G304 -- test path from t.TempDir()
			if err != nil {
				t.Fatalf("failed to read output file: %v", err)
			}

			var out JSONOutput
			if err = json.Unmarshal(data, &out); err != nil {
				t.Fatalf("failed to unmarshal JSON output: %v", err)
			}

			if tt.wantInputKey != "" {
				if _, exists := out.Action.Inputs[tt.wantInputKey]; !exists {
					t.Errorf("expected input key %q in JSON output", tt.wantInputKey)
				}
			}

			if tt.wantOutputKey != "" {
				if _, exists := out.Action.Outputs[tt.wantOutputKey]; !exists {
					t.Errorf("expected output key %q in JSON output", tt.wantOutputKey)
				}
			}
		})
	}
}

// assertJSONSectionsForType checks whether a section of the given type exists in the
// documentation sections list.
func assertJSONSectionsForType(t *testing.T, out *JSONOutput, sectionType string, wantPresent bool) {
	t.Helper()

	found := false

	for _, sec := range out.Documentation.Sections {
		if sec.Type == sectionType {
			found = true

			break
		}
	}

	if wantPresent && !found {
		t.Errorf("expected section of type %q to be present in documentation sections", sectionType)
	}

	if !wantPresent && found {
		t.Errorf("expected section of type %q to be absent from documentation sections", sectionType)
	}
}
