package wizard

import (
	"strings"
	"testing"
)

// assertValidationError checks if validation result contains expected error.
// This helper reduces cognitive complexity in validation tests by centralizing
// the error-finding logic that was repeated across multiple test functions.
func assertValidationError(t *testing.T, result *ValidationResult, field string, expectError bool, expectedMsg string) {
	t.Helper()

	if !expectError {
		if len(result.Errors) > 0 {
			t.Errorf("expected no errors for field %s, got: %v", field, result.Errors)
		}

		return
	}

	// Find error matching expected message
	found := false
	for _, err := range result.Errors {
		if strings.Contains(err.Message, expectedMsg) {
			found = true

			break
		}
	}

	if !found {
		t.Errorf("expected error containing %q for field %s, got errors: %v", expectedMsg, field, result.Errors)
	}
}
