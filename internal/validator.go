package internal

import (
	"fmt"
	"strings"
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
		result.MissingFields = append(result.MissingFields, "name")
		result.Suggestions = append(result.Suggestions, "Add 'name: Your Action Name' to describe your action")
	}
	if action.Description == "" {
		result.MissingFields = append(result.MissingFields, "description")
		result.Suggestions = append(
			result.Suggestions,
			"Add 'description: Brief description of what your action does' for better documentation",
		)
	}
	if len(action.Runs) == 0 {
		result.MissingFields = append(result.MissingFields, "runs")
		result.Suggestions = append(
			result.Suggestions,
			"Add 'runs:' section with 'using: node20' or 'using: docker' and specify the main file",
		)
	} else {
		// Validate the runs section content
		if using, ok := action.Runs["using"].(string); ok {
			if !isValidRuntime(using) {
				result.MissingFields = append(result.MissingFields, "runs.using")
				result.Suggestions = append(
					result.Suggestions,
					fmt.Sprintf("Invalid runtime '%s'. Valid runtimes: node12, node16, node20, docker, composite", using),
				)
			}
		} else {
			result.MissingFields = append(result.MissingFields, "runs.using")
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

// isValidRuntime checks if the given runtime is valid for GitHub Actions.
func isValidRuntime(runtime string) bool {
	validRuntimes := []string{
		"node12",    // Legacy Node.js runtime (deprecated)
		"node16",    // Legacy Node.js runtime (deprecated)
		"node20",    // Current Node.js runtime
		"docker",    // Docker container runtime
		"composite", // Composite action runtime
	}

	runtime = strings.TrimSpace(strings.ToLower(runtime))
	for _, valid := range validRuntimes {
		if runtime == valid {
			return true
		}
	}

	return false
}
