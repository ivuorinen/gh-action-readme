package internal

import (
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
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
		Description: testGenShortDesc,
		Runs:        map[string]any{testGenRunsUsing: appconstants.NodeRuntimeNode20},
	}
	res := ValidateActionYML(a)
	if len(res.MissingFields) != 0 {
		t.Errorf("expected no missing fields, got %v", res.MissingFields)
	}
}

// TestValidateActionYML_DeprecatedRuntime verifies that node12 and node16 produce
// a deprecation warning but are still considered valid (not in MissingFields).
func TestValidateActionYML_DeprecatedRuntime(t *testing.T) {
	t.Parallel()

	for _, runtime := range []string{appconstants.NodeRuntimeNode12, appconstants.NodeRuntimeNode16} {
		runtime := runtime
		t.Run(runtime, func(t *testing.T) {
			t.Parallel()

			a := &ActionYML{
				Name:        testGenActionName,
				Description: testGenShortDesc,
				Runs:        map[string]any{testGenRunsUsing: runtime},
			}
			res := ValidateActionYML(a)

			if len(res.MissingFields) != 0 {
				t.Errorf("deprecated runtime %q should not produce missing fields, got: %v", runtime, res.MissingFields)
			}

			if !containsWarning(t, res.Warnings, appconstants.FieldRunsUsing) {
				t.Errorf("expected deprecation warning for runtime %q, got warnings: %v", runtime, res.Warnings)
			}

			found := false
			for _, s := range res.Suggestions {
				if strings.Contains(s, runtime) && strings.Contains(s, "deprecated") {
					found = true

					break
				}
			}
			if !found {
				t.Errorf("expected deprecation suggestion mentioning %q, got: %v", runtime, res.Suggestions)
			}
		})
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
			Name:        testGenActionName,
			Description: testGenShortDesc,
			Runs:        map[string]any{testGenRunsUsing: appconstants.NodeRuntimeNode20},
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
		a.Branding = &Branding{Icon: appconstants.ActivityWorkflowType, Color: "blue"}
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
			Name:        testGenActionName,
			Description: testGenShortDesc,
			Runs:        map[string]any{testGenRunsUsing: appconstants.NodeRuntimeNode20},
			Branding:    &Branding{Icon: appconstants.ActivityWorkflowType},
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
		a.Inputs = map[string]ActionInput{testGenTokenKey: {Description: "A token"}}
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
			Name:        testGenActionName,
			Description: testGenShortDesc,
			Runs:        map[string]any{testGenRunsUsing: appconstants.NodeRuntimeNode20},
			Branding:    &Branding{Icon: appconstants.ActivityWorkflowType},
			Inputs:      map[string]ActionInput{testGenTokenKey: {Description: "A token"}},
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
