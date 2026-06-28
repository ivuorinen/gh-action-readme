package internal

import (
	"fmt"
	"strings"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

// ValidationResult holds the results of action.yml validation.
type ValidationResult struct {
	MissingFields []string
	Warnings      []string
	Suggestions   []string
}

// ValidateActionYML checks if required fields are present and valid.
func ValidateActionYML(action *ActionYML) ValidationResult {
	result := ValidationResult{}

	// Validate required fields with helpful suggestions
	if action.Name == "" {
		result.MissingFields = append(result.MissingFields, appconstants.FieldName)
		result.Suggestions = append(result.Suggestions, "Add 'name: Your Action Name' to describe your action")
	}
	if action.Description == "" {
		result.MissingFields = append(result.MissingFields, appconstants.FieldDescription)
		result.Suggestions = append(
			result.Suggestions,
			"Add 'description: Brief description of what your action does' for better documentation",
		)
	}
	if len(action.Runs) == 0 {
		result.MissingFields = append(result.MissingFields, appconstants.FieldRuns)
		result.Suggestions = append(
			result.Suggestions,
			"Add 'runs:' section with 'using: node20' or 'using: docker' and specify the main file",
		)
	} else {
		// Validate the runs section content
		if using, ok := action.Runs["using"].(string); ok {
			if !isValidRuntime(using) {
				result.MissingFields = append(result.MissingFields, appconstants.FieldRunsUsing)
				result.Suggestions = append(
					result.Suggestions,
					fmt.Sprintf(
						"Invalid runtime '%s'. Valid runtimes: node20, node24, docker, composite",
						using,
					),
				)
			} else {
				if isDeprecatedRuntime(using) {
					result.Warnings = append(result.Warnings, appconstants.FieldRunsUsing)
					result.Suggestions = append(
						result.Suggestions,
						fmt.Sprintf(
							"Runtime '%s' is deprecated and no longer supported by GitHub Actions; migrate to %s",
							using, appconstants.NodeRuntimeNode24,
						),
					)
				}
				validateRunsRequiredField(using, action.Runs, &result)
			}
		} else {
			result.MissingFields = append(result.MissingFields, appconstants.FieldRunsUsing)
			result.Suggestions = append(
				result.Suggestions,
				"Missing 'using' field in runs section. Specify 'using: node20', 'using: docker', or 'using: composite'",
			)
		}
	}

	// Add warnings for optional but recommended fields
	if action.Branding == nil {
		result.Warnings = append(result.Warnings, "branding")
		result.Suggestions = append(
			result.Suggestions,
			"Consider adding 'branding:' with 'icon' and 'color' for better marketplace appearance",
		)
	}
	if len(action.Inputs) == 0 {
		result.Warnings = append(result.Warnings, "inputs")
		result.Suggestions = append(result.Suggestions, "Consider adding 'inputs:' if your action accepts parameters")
	}
	if len(action.Outputs) == 0 {
		result.Warnings = append(result.Warnings, "outputs")
		result.Suggestions = append(result.Suggestions, "Consider adding 'outputs:' if your action produces results")
	}

	return result
}

// validateRunsRequiredField appends a MissingField when the runs section lacks
// the entry point GitHub requires for the given runtime: main for node*, image
// for docker, steps for composite. Without this, ValidateActionYML reports
// "all validations passed" for an action GitHub would reject (e.g. a node action
// with no main).
func validateRunsRequiredField(using string, runs map[string]any, result *ValidationResult) {
	normalized := strings.TrimSpace(strings.ToLower(using))

	var key, field, suggestion string
	switch {
	case strings.HasPrefix(normalized, "node"):
		key, field = appconstants.RunsKeyMain, appconstants.FieldRunsMain
		suggestion = "Add 'main:' (the JavaScript entry point, e.g. index.js) to the runs section"
	case normalized == appconstants.ActionTypeDocker:
		key, field = appconstants.RunsKeyImage, appconstants.FieldRunsImage
		suggestion = "Add 'image:' (a Dockerfile path or a docker:// image) to the runs section"
	case normalized == appconstants.ActionTypeComposite:
		key, field = appconstants.RunsKeySteps, appconstants.FieldRunsSteps
		suggestion = "Add 'steps:' to the runs section for a composite action"
	default:
		return
	}

	if val, ok := runs[key]; !ok || isEmptyRuntimeValue(val) {
		result.MissingFields = append(result.MissingFields, field)
		result.Suggestions = append(result.Suggestions, suggestion)
	}
}

// isEmptyRuntimeValue reports whether a present runs.* entry carries no usable
// value (runs.main:, runs.image: "", runs.steps: with no items). GitHub rejects
// such actions even though the key exists, so key presence alone is insufficient.
func isEmptyRuntimeValue(v any) bool {
	switch val := v.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(val) == ""
	case []any:
		return len(val) == 0
	case map[string]any:
		return len(val) == 0
	default:
		return false
	}
}

// isValidRuntime checks if the given runtime is valid for GitHub Actions.
func isValidRuntime(runtime string) bool {
	validRuntimes := []string{
		appconstants.NodeRuntimeNode12,   // Deprecated Node.js runtime
		appconstants.NodeRuntimeNode16,   // Deprecated Node.js runtime
		appconstants.NodeRuntimeNode20,   // Current Node.js runtime
		appconstants.NodeRuntimeNode24,   // Current Node.js runtime
		appconstants.ActionTypeDocker,    // Docker container runtime
		appconstants.ActionTypeComposite, // Composite action runtime
	}

	runtime = strings.TrimSpace(strings.ToLower(runtime))
	for _, valid := range validRuntimes {
		if runtime == valid {
			return true
		}
	}

	return false
}

// isDeprecatedRuntime checks if the given runtime is deprecated.
func isDeprecatedRuntime(runtime string) bool {
	deprecatedRuntimes := []string{
		appconstants.NodeRuntimeNode12,
		appconstants.NodeRuntimeNode16,
	}

	runtime = strings.TrimSpace(strings.ToLower(runtime))
	for _, deprecated := range deprecatedRuntimes {
		if runtime == deprecated {
			return true
		}
	}

	return false
}
