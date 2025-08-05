package internal

import (
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

func TestParseActionYML_Valid(t *testing.T) {
	// Create temporary action file using fixture
	actionPath := testutil.CreateTemporaryAction(t, "actions/javascript/simple.yml")
	action, err := ParseActionYML(actionPath)
	if err != nil {
		t.Fatalf("failed to parse action.yml: %v", err)
	}
	if action.Name != "Simple JavaScript Action" {
		t.Errorf("expected name 'Simple JavaScript Action', got '%s'", action.Name)
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
