package internal

import (
	"encoding/json"
	"testing"

	"github.com/ivuorinen/gh-action-readme/internal/dependencies"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// marshalJSONOutput renders a writer's output document and decodes it back into a
// generic map, so assertions can check key presence exactly as a consumer would.
func marshalJSONOutput(t *testing.T, jw *JSONWriter, action *ActionYML) map[string]any {
	t.Helper()

	raw, err := json.Marshal(jw.convertToJSONOutput(action))
	if err != nil {
		t.Fatalf("marshal JSON output: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal JSON output: %v", err)
	}

	return decoded
}

// TestJSONOutputCarriesDependencies is the regression guard for unwired-8b5b59c2:
// the JSON path used to run the full dependency analysis and then discard it, so
// `-f json` silently omitted a section the Markdown themes rendered.
func TestJSONOutputCarriesDependencies(t *testing.T) {
	t.Parallel()

	action := &ActionYML{Name: "Dep Action", Description: "has dependencies"}
	jw := NewJSONWriter(DefaultAppConfig())
	jw.dependencies = []dependencies.Dependency{
		{Name: "actions/checkout", Uses: "actions/checkout@v4", Version: "v4"},
	}

	decoded := marshalJSONOutput(t, jw, action)

	deps, ok := decoded["dependencies"].([]any)
	if !ok {
		t.Fatalf("expected a dependencies array in JSON output, got keys %v", mapKeys(decoded))
	}
	testutil.AssertEqual(t, 1, len(deps))

	first, ok := deps[0].(map[string]any)
	if !ok {
		t.Fatalf("dependency entry is not an object: %T", deps[0])
	}
	testutil.AssertEqual(t, "actions/checkout", first["name"])
	// unwired-e2ba31f9: the write-only ScriptURL field was removed; SourceURL is the
	// one that carries the link.
	if _, present := first["script_url"]; present {
		t.Error("script_url must not appear in JSON output; the field was removed")
	}
}

// TestJSONOutputOmitsDependenciesWhenAbsent pins the back-compat half: with analysis
// disabled the key must be absent entirely, not an empty array, so consumers written
// against the previous output keep working.
func TestJSONOutputOmitsDependenciesWhenAbsent(t *testing.T) {
	t.Parallel()

	action := &ActionYML{Name: "Plain Action", Description: "no dependencies"}
	jw := NewJSONWriter(DefaultAppConfig())

	decoded := marshalJSONOutput(t, jw, action)

	if _, present := decoded["dependencies"]; present {
		t.Errorf("dependencies key must be omitted when analysis is off, got keys %v", mapKeys(decoded))
	}
}

// mapKeys returns a map's keys, for readable failure messages.
func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}
