package internal

// validationSummaryTestCase defines a test case for validation summary tests.
// This helper reduces duplication in test case definitions by providing
// a factory function with sensible defaults.
type validationSummaryTestCase struct {
	name        string
	totalFiles  int
	validFiles  int
	totalIssues int
	resultCount int
	errorCount  int
	wantBold    int
	wantSuccess int
	wantWarning int
	wantError   int
	wantInfo    int
}

// createValidationSummaryTest creates a validation summary test case with defaults.
// Default values: wantBold=1, wantSuccess=1, wantWarning=0, wantError=0, wantInfo=0
// Only provide the fields that differ from defaults.
func createValidationSummaryTest(
	name string,
	totalFiles, validFiles, totalIssues, resultCount, errorCount int,
	wantWarning, wantError, wantInfo int,
) validationSummaryTestCase {
	return validationSummaryTestCase{
		name:        name,
		totalFiles:  totalFiles,
		validFiles:  validFiles,
		totalIssues: totalIssues,
		resultCount: resultCount,
		errorCount:  errorCount,
		wantBold:    1, // Always 1
		wantSuccess: 1, // Always 1
		wantWarning: wantWarning,
		wantError:   wantError,
		wantInfo:    wantInfo,
	}
}
