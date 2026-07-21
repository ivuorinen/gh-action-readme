package internal

import (
	"slices"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

func TestValidateActionYMLRequired(t *testing.T) {
	t.Parallel()

	a := &ActionYML{
		Name:        "",
		Description: "",
		Runs:        ActionRuns{},
	}
	res := ValidateActionYML(a)
	if len(res.MissingFields) == 0 {
		t.Error("should detect missing fields")
	}
}

func TestValidateActionYMLValid(t *testing.T) {
	t.Parallel()
	a := &ActionYML{
		Name:        testutil.TestActionNameMyAction,
		Description: testGenShortDesc,
		Runs:        ActionRuns{Using: appconstants.NodeRuntimeNode20, Main: "index.js"},
	}
	res := ValidateActionYML(a)
	if len(res.MissingFields) != 0 {
		t.Errorf("expected no missing fields, got %v", res.MissingFields)
	}
}

// assertHasDeprecationSuggestion fails if no suggestion mentions runtime and "deprecated".
func assertHasDeprecationSuggestion(t *testing.T, suggestions []string, runtime string) {
	t.Helper()

	for _, s := range suggestions {
		if strings.Contains(s, runtime) && strings.Contains(s, "deprecated") {
			return
		}
	}

	t.Errorf("expected deprecation suggestion mentioning %q, got: %v", runtime, suggestions)
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
				Runs:        ActionRuns{Using: runtime, Main: "index.js"},
			}
			res := ValidateActionYML(a)

			if len(res.MissingFields) != 0 {
				t.Errorf("deprecated runtime %q should not produce missing fields, got: %v", runtime, res.MissingFields)
			}

			if !containsWarning(t, res.Warnings, appconstants.FieldRunsUsing) {
				t.Errorf("expected deprecation warning for runtime %q, got warnings: %v", runtime, res.Warnings)
			}

			assertHasDeprecationSuggestion(t, res.Suggestions, runtime)
		})
	}
}

// TestValidateActionYMLMissingRuntimeFields verifies N154: an action whose runtime
// lacks its required entry point (node→main, composite→steps, docker→image) is
// reported as structurally invalid instead of passing validation.
func TestValidateActionYMLMissingRuntimeFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		using     string
		wantField string
	}{
		{"node without main", appconstants.NodeRuntimeNode20, "runs.main"},
		{"composite without steps", appconstants.ActionTypeComposite, "runs.steps"},
		{"docker without image", "docker", "runs.image"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a := &ActionYML{
				Name:        testGenActionName,
				Description: testGenShortDesc,
				Runs:        ActionRuns{Using: tt.using},
			}
			res := ValidateActionYML(a)

			if !slices.Contains(res.MissingFields, tt.wantField) {
				t.Errorf("expected missing field %q for runtime %q, got: %v", tt.wantField, tt.using, res.MissingFields)
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
			Runs:        ActionRuns{Using: appconstants.NodeRuntimeNode20},
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
			Runs:        ActionRuns{Using: appconstants.NodeRuntimeNode20},
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
			Runs:        ActionRuns{Using: appconstants.NodeRuntimeNode20},
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
