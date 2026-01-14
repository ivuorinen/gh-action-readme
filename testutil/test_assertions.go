package testutil

import "testing"

// AssertMessageCounts verifies that output has expected message counts.
// Reduces duplication in validation tests (8+ occurrences).
//
// Example:
//
//	output := captureOutput(...)
//	testutil.AssertMessageCounts(t, "test case", output, 2, 1, 0, 1)
func AssertMessageCounts(t *testing.T, testName string, output *CapturedOutput,
	wantInfo, wantError, wantWarning, wantBold int) {
	t.Helper()

	if len(output.InfoMessages) != wantInfo {
		t.Errorf("%s: info messages = %d, want %d",
			testName, len(output.InfoMessages), wantInfo)
	}

	if len(output.ErrorMessages) != wantError {
		t.Errorf("%s: error messages = %d, want %d",
			testName, len(output.ErrorMessages), wantError)
	}

	if len(output.WarningMessages) != wantWarning {
		t.Errorf("%s: warning messages = %d, want %d",
			testName, len(output.WarningMessages), wantWarning)
	}

	if len(output.BoldMessages) != wantBold {
		t.Errorf("%s: bold messages = %d, want %d",
			testName, len(output.BoldMessages), wantBold)
	}
}
