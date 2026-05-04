package internal

import (
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

func TestFillMissing(t *testing.T) {
	t.Parallel()

	a := &ActionYML{}
	defs := DefaultValues{
		Name:        "Default Name",
		Description: "Default Desc",
		Runs:        map[string]any{testGenRunsUsing: appconstants.NodeRuntimeNode20},
		Branding:    Branding{Icon: "zap", Color: "yellow"},
	}
	FillMissing(a, defs)
	if a.Name != "Default Name" || a.Description != "Default Desc" {
		t.Error("defaults not filled correctly")
	}
	if a.Branding == nil || a.Branding.Icon != "zap" {
		t.Error("branding default not set")
	}
	if a.Runs["using"] != appconstants.NodeRuntimeNode20 {
		t.Error("runs default not set")
	}
}

// TestFillMissing_RunsNotOverwritten kills the mutation on config.go:136
// (len(action.Runs) == 0 && len(defs.Runs) > 0).
// When action.Runs is already populated it must NOT be replaced by defs.Runs.
func TestFillMissing_RunsNotOverwritten(t *testing.T) {
	t.Parallel()

	a := &ActionYML{
		Runs: map[string]any{testGenRunsUsing: appconstants.ActionTypeComposite},
	}
	defs := DefaultValues{
		Runs: map[string]any{testGenRunsUsing: appconstants.NodeRuntimeNode20},
	}

	FillMissing(a, defs)

	if a.Runs["using"] != appconstants.ActionTypeComposite {
		t.Errorf("FillMissing() must not overwrite non-empty Runs; got %q, want %q",
			a.Runs["using"], appconstants.ActionTypeComposite)
	}
}

// TestFillMissing_RunsEmptyDefsEmpty verifies that when both action.Runs and defs.Runs
// are empty, Runs stays empty (second half of the && condition: len(defs.Runs) > 0).
func TestFillMissing_RunsEmptyDefsEmpty(t *testing.T) {
	t.Parallel()

	a := &ActionYML{}
	defs := DefaultValues{
		Runs: map[string]any{}, // empty defs
	}

	FillMissing(a, defs)

	if len(a.Runs) != 0 {
		t.Errorf("FillMissing() should not set Runs when defs.Runs is empty; got %v", a.Runs)
	}
}
