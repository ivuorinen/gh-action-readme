package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

// newTestJSONWriter creates a JSONWriter with a default config for testing.
func newTestJSONWriter() *JSONWriter {
	return NewJSONWriter(DefaultAppConfig())
}

// parseFixtureAction loads and parses an action YAML fixture by its fixture path.
func parseFixtureAction(t *testing.T, fixturePath string) *ActionYML {
	t.Helper()

	tmpFile := filepath.Join(t.TempDir(), "action.yml")
	testutil.WriteTestFile(t, tmpFile, testutil.MustReadFixture(fixturePath))

	action, err := ParseActionYML(tmpFile)
	if err != nil {
		t.Fatalf("parseFixtureAction(%q): %v", fixturePath, err)
	}

	return action
}

// actionWithNoInputsOutputs returns an ActionYML with no inputs or outputs.
func actionWithNoInputsOutputs(t *testing.T) *ActionYML {
	t.Helper()

	return parseFixtureAction(t, testutil.TestFixtureJSONWriterNoInputsOutputs)
}

// actionWithInputs returns an ActionYML with one input, optionally with a default value.
func actionWithInputs(t *testing.T, withDefault bool) *ActionYML {
	t.Helper()
	if withDefault {
		return parseFixtureAction(t, testutil.TestFixtureJSONWriterWithInputsDefault)
	}

	return parseFixtureAction(t, testutil.TestFixtureJSONWriterWithInputsNoDefault)
}

// actionWithOutputs returns an ActionYML with one output.
func actionWithOutputs(t *testing.T) *ActionYML {
	t.Helper()

	return parseFixtureAction(t, testutil.TestFixtureJSONWriterWithOutputs)
}

// TestConvertToJSONOutput_NoInputs verifies that an action with no inputs
// produces a JSON output with no "inputs" section.
func TestConvertToJSONOutput_NoInputs(t *testing.T) {
	jw := newTestJSONWriter()
	action := actionWithNoInputsOutputs(t)

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
	action := actionWithInputs(t, false)

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
	action := actionWithNoInputsOutputs(t)

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
	action := actionWithOutputs(t)

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
	action := actionWithNoInputsOutputs(t)

	example := jw.generateBasicExample(action)
	if strings.Contains(example, "with:") {
		t.Errorf("expected no 'with:' block for action with no inputs, got:\n%s", example)
	}
}

// TestGenerateBasicExample_WithInputs verifies that generateBasicExample with inputs
// includes a "with:" block in the YAML output.
func TestGenerateBasicExample_WithInputs(t *testing.T) {
	jw := newTestJSONWriter()
	action := actionWithInputs(t, false)

	example := jw.generateBasicExample(action)
	if !strings.Contains(example, "with:") {
		t.Errorf("expected 'with:' block for action with inputs, got:\n%s", example)
	}
}

// TestGenerateBasicExample_InputWithDefault verifies that when an input has a Default
// value, generateBasicExample uses that value instead of the placeholder "value".
func TestGenerateBasicExample_InputWithDefault(t *testing.T) {
	jw := newTestJSONWriter()
	action := actionWithInputs(t, true)

	example := jw.generateBasicExample(action)
	if !strings.Contains(example, "my-default") {
		t.Errorf("expected default value 'my-default' in example output, got:\n%s", example)
	}
}

// TestGenerateBasicExample_InputWithNilDefault verifies that when an input has
// Default == nil, generateBasicExample uses the placeholder "value".
func TestGenerateBasicExample_InputWithNilDefault(t *testing.T) {
	jw := newTestJSONWriter()
	action := actionWithInputs(t, false)

	example := jw.generateBasicExample(action)
	if !strings.Contains(example, `"value"`) {
		t.Errorf("expected placeholder value 'value' in example output when Default is nil, got:\n%s", example)
	}
}

type jsonWriterWriteCase struct {
	name          string
	buildAction   func(t *testing.T) *ActionYML
	wantInputKey  string
	wantOutputKey string
}

// TestJSONWriter_Write verifies end-to-end JSON file generation and validates
// that inputs/outputs sections are correctly included or excluded in the raw JSON.
func TestJSONWriter_Write(t *testing.T) {
	tests := []jsonWriterWriteCase{
		{
			name:        "action with no inputs or outputs",
			buildAction: actionWithNoInputsOutputs,
		},
		{
			name: "action with inputs only",
			buildAction: func(t *testing.T) *ActionYML {
				t.Helper()

				return actionWithInputs(t, true)
			},
			wantInputKey: testGenTokenKey,
		},
		{
			name:          "action with outputs only",
			buildAction:   actionWithOutputs,
			wantOutputKey: "result",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			outputPath := filepath.Join(dir, "output.json")

			jw := newTestJSONWriter()
			if err := jw.Write(tt.buildAction(t), outputPath); err != nil {
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

			rawAction := unmarshalRawAction(t, data)
			assertRawJSONKey(t, rawAction, "inputs", tt.wantInputKey != "")
			assertRawJSONKey(t, rawAction, "outputs", tt.wantOutputKey != "")
			assertInputKey(t, out.Action.Inputs, tt.wantInputKey)
			assertOutputKey(t, out.Action.Outputs, tt.wantOutputKey)
		})
	}
}

func unmarshalRawAction(t *testing.T, data []byte) map[string]json.RawMessage {
	t.Helper()

	var raw struct {
		Action map[string]json.RawMessage `json:"action"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal raw JSON output: %v", err)
	}

	return raw.Action
}

func assertRawJSONKey(t *testing.T, obj map[string]json.RawMessage, key string, wantPresent bool) {
	t.Helper()

	_, exists := obj[key]
	if exists != wantPresent {
		t.Errorf("expected raw JSON key %q presence=%v, got %v", key, wantPresent, exists)
	}
}

func assertInputKey(t *testing.T, inputs map[string]ActionInputForJSON, key string) {
	t.Helper()

	if key == "" {
		return
	}
	if _, exists := inputs[key]; !exists {
		t.Errorf("expected input key %q in JSON output", key)
	}
}

func assertOutputKey(t *testing.T, outputs map[string]ActionOutputForJSON, key string) {
	t.Helper()

	if key == "" {
		return
	}
	if _, exists := outputs[key]; !exists {
		t.Errorf("expected output key %q in JSON output", key)
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
