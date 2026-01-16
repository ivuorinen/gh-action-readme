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

// validationSummaryParams holds parameters for creating validation summary test cases.
type validationSummaryParams struct {
	name                                                         string
	totalFiles, validFiles, totalIssues, resultCount, errorCount int
	wantWarning, wantError, wantInfo                             int
}

// createValidationSummaryTest creates a validation summary test case with defaults.
// Default values: wantBold=1, wantSuccess=1, wantWarning=0, wantError=0, wantInfo=0
// Only provide the fields that differ from defaults.
func createValidationSummaryTest(params validationSummaryParams) validationSummaryTestCase {
	return validationSummaryTestCase{
		name:        params.name,
		totalFiles:  params.totalFiles,
		validFiles:  params.validFiles,
		totalIssues: params.totalIssues,
		resultCount: params.resultCount,
		errorCount:  params.errorCount,
		wantBold:    1, // Always 1
		wantSuccess: 1, // Always 1
		wantWarning: params.wantWarning,
		wantError:   params.wantError,
		wantInfo:    params.wantInfo,
	}
}
