package internal

import (
	"testing"
)

func TestParseActionYML_Valid(t *testing.T) {
	path := "../testdata/example-action/action.yml"
	action, err := ParseActionYML(path)
	if err != nil {
		t.Fatalf("failed to parse action.yml: %v", err)
	}
	if action.Name != "Example Action" {
		t.Errorf("expected name 'Example Action', got '%s'", action.Name)
	}
	if action.Description == "" {
		t.Error("expected non-empty description")
	}
	if len(action.Inputs) != 2 {
		t.Errorf("expected 2 inputs, got %d", len(action.Inputs))
	}
}

func TestParseActionYML_MissingFile(t *testing.T) {
	_, err := ParseActionYML("notfound/action.yml")
	if err == nil {
		t.Error("expected error on missing file")
	}
}
