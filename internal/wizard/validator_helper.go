package wizard

import (
	"fmt"
	"strings"
)

// validateFieldWithEmptyCheck is a generic helper for fields that:
// - Allow empty values (with optional warning)
// - Validate non-empty values with a custom validator function
// - Add error and optional suggestion if validation fails.
func (v *ConfigValidator) validateFieldWithEmptyCheck(
	field, fieldValue string,
	isValid func(string) bool,
	emptyWarning, errorMsg, suggestion string,
	result *ValidationResult,
) {
	if fieldValue == "" {
		if emptyWarning != "" {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:   field,
				Message: emptyWarning,
				Value:   fieldValue,
			})
		}

		return
	}

	if !isValid(fieldValue) {
		result.Errors = append(result.Errors, ValidationError{
			Field:   field,
			Message: errorMsg,
			Value:   fieldValue,
		})

		if suggestion != "" {
			result.Suggestions = append(result.Suggestions, suggestion)
		}
	}
}

// validateFieldInList is a generic helper for fields that must be
// one of a predefined list of valid values.
func (v *ConfigValidator) validateFieldInList(
	field, fieldValue string,
	validValues []string,
	errorMsg string,
	result *ValidationResult,
) {
	if !v.isValueInList(fieldValue, validValues) {
		result.Errors = append(result.Errors, ValidationError{
			Field:   field,
			Message: errorMsg,
			Value:   fieldValue,
		})
		result.Suggestions = append(result.Suggestions,
			fmt.Sprintf("Valid %ss: %s", field, strings.Join(validValues, ", ")))
	}
}
