package internal

import (
	"fmt"
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

	// Validation feedback
	if len(result.MissingFields) == 0 {
		fmt.Println("Validation passed.")
	} else {
		fmt.Printf("Missing required fields: %v\n", result.MissingFields)
	}

	return result
}
