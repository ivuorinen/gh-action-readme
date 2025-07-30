package internal

import "testing"

func TestFillMissing(t *testing.T) {

	a := &ActionYML{}
	defs := DefaultValues{
		Name:        "Default Name",
		Description: "Default Desc",
		Runs:        map[string]any{"using": "node20"},
		Branding:    Branding{Icon: "zap", Color: "yellow"},
	}
	FillMissing(a, defs)
	if a.Name != "Default Name" || a.Description != "Default Desc" {
		t.Error("defaults not filled correctly")
	}
	if a.Branding == nil || a.Branding.Icon != "zap" {
		t.Error("branding default not set")
	}
	if a.Runs["using"] != "node20" {
		t.Error("runs default not set")
	}
}
