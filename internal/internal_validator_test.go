package internal

import (
	"strings"
	"testing"
)

func TestValidateActionYMLRequired(t *testing.T) {
	t.Parallel()

	a := &ActionYML{
		Name:        "",
		Description: "",
		Runs:        map[string]any{},
	}
	res := ValidateActionYML(a)
	if len(res.MissingFields) == 0 {
		t.Error("should detect missing fields")
	}
}

func TestValidateActionYMLValid(t *testing.T) {
	t.Parallel()
	a := &ActionYML{
		Name:        "MyAction",
		Description: "desc",
		Runs:        map[string]any{"using": "node12"},
	}
	res := ValidateActionYML(a)
	if len(res.MissingFields) != 0 {
		t.Errorf("expected no missing fields, got %v", res.MissingFields)
	}
}

// containsWarning is a helper to check whether a warning string appears in the slice.
func containsWarning(t *testing.T, warnings []string, want string) bool {
	t.Helper()

	for _, w := range warnings {
		if strings.Contains(w, want) {
			return true
		}
	}

	return false
}

// TestValidateActionYML_BrandingWarning kills the CONDITIONALS_NEGATION at validator.go:62
// by verifying that a nil Branding generates the "branding" warning and a non-nil Branding
// does NOT.
func TestValidateActionYML_BrandingWarning(t *testing.T) {
	t.Parallel()

	base := func() *ActionYML {
		return &ActionYML{
			Name:        "Action",
			Description: "desc",
			Runs:        map[string]any{"using": "node20"},
		}
	}

	t.Run("nil branding produces warning", func(t *testing.T) {
		t.Parallel()

		a := base()
		a.Branding = nil
		res := ValidateActionYML(a)

		if !containsWarning(t, res.Warnings, "branding") {
			t.Errorf("expected 'branding' warning for nil Branding, got: %v", res.Warnings)
		}
	})

	t.Run("non-nil branding suppresses warning", func(t *testing.T) {
		t.Parallel()

		a := base()
		a.Branding = &Branding{Icon: "activity", Color: "blue"}
		res := ValidateActionYML(a)

		if containsWarning(t, res.Warnings, "branding") {
			t.Errorf("unexpected 'branding' warning for non-nil Branding, got: %v", res.Warnings)
		}
	})
}

// TestValidateActionYML_InputsWarning kills the CONDITIONALS_NEGATION at validator.go:69.
func TestValidateActionYML_InputsWarning(t *testing.T) {
	t.Parallel()

	base := func() *ActionYML {
		return &ActionYML{
			Name:        "Action",
			Description: "desc",
			Runs:        map[string]any{"using": "node20"},
			Branding:    &Branding{Icon: "activity"},
		}
	}

	t.Run("empty inputs produces warning", func(t *testing.T) {
		t.Parallel()

		a := base()
		res := ValidateActionYML(a)

		if !containsWarning(t, res.Warnings, "inputs") {
			t.Errorf("expected 'inputs' warning for empty inputs, got: %v", res.Warnings)
		}
	})

	t.Run("non-empty inputs suppresses warning", func(t *testing.T) {
		t.Parallel()

		a := base()
		a.Inputs = map[string]ActionInput{"token": {Description: "A token"}}
		res := ValidateActionYML(a)

		if containsWarning(t, res.Warnings, "inputs") {
			t.Errorf("unexpected 'inputs' warning when inputs are set, got: %v", res.Warnings)
		}
	})
}

// TestValidateActionYML_OutputsWarning kills the CONDITIONALS_NEGATION at validator.go:73.
func TestValidateActionYML_OutputsWarning(t *testing.T) {
	t.Parallel()

	base := func() *ActionYML {
		return &ActionYML{
			Name:        "Action",
			Description: "desc",
			Runs:        map[string]any{"using": "node20"},
			Branding:    &Branding{Icon: "activity"},
			Inputs:      map[string]ActionInput{"token": {Description: "A token"}},
		}
	}

	t.Run("empty outputs produces warning", func(t *testing.T) {
		t.Parallel()

		a := base()
		res := ValidateActionYML(a)

		if !containsWarning(t, res.Warnings, "outputs") {
			t.Errorf("expected 'outputs' warning for empty outputs, got: %v", res.Warnings)
		}
	})

	t.Run("non-empty outputs suppresses warning", func(t *testing.T) {
		t.Parallel()

		a := base()
		a.Outputs = map[string]ActionOutput{"result": {Description: "The result"}}
		res := ValidateActionYML(a)

		if containsWarning(t, res.Warnings, "outputs") {
			t.Errorf("unexpected 'outputs' warning when outputs are set, got: %v", res.Warnings)
		}
	})
}
