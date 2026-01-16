package internal

import (
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/apperrors"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// capturedOutput wraps testutil.CapturedOutput to satisfy CompleteOutput interface.
type capturedOutput struct {
	*testutil.CapturedOutput
}

// ErrorWithSuggestions wraps the testutil version to match interface signature.
func (c *capturedOutput) ErrorWithSuggestions(err *apperrors.ContextualError) {
	c.CapturedOutput.ErrorWithSuggestions(err)
}

// FormatContextualError wraps the testutil version to match interface signature.
func (c *capturedOutput) FormatContextualError(err *apperrors.ContextualError) string {
	return c.CapturedOutput.FormatContextualError(err)
}

// newCapturedOutput creates a new capturedOutput instance.
func newCapturedOutput() *capturedOutput {
	return &capturedOutput{
		CapturedOutput: &testutil.CapturedOutput{},
	}
}

// TestCountValidationStats tests the validation statistics counting function.
func TestCountValidationStats(t *testing.T) {
	tests := []struct {
		name            string
		results         []ValidationResult
		wantValidFiles  int
		wantTotalIssues int
	}{
		{
			name: "all valid files",
			results: []ValidationResult{
				{MissingFields: []string{testutil.ValidationTestFile1}},
				{MissingFields: []string{testutil.ValidationTestFile2}},
			},
			wantValidFiles:  2,
			wantTotalIssues: 0,
		},
		{
			name: "all invalid files",
			results: []ValidationResult{
				{MissingFields: []string{testutil.ValidationTestFile1, "name", "description"}},
				{MissingFields: []string{testutil.ValidationTestFile2, "runs"}},
			},
			wantValidFiles:  0,
			wantTotalIssues: 3, // 2 issues in first file + 1 in second
		},
		{
			name: "mixed valid and invalid",
			results: []ValidationResult{
				{MissingFields: []string{testutil.ValidationTestFile1}},                        // Valid
				{MissingFields: []string{testutil.ValidationTestFile2, "name", "description"}}, // 2 issues
				{MissingFields: []string{"file: action3.yml"}},                                 // Valid
				{MissingFields: []string{"file: action4.yml", "runs"}},                         // 1 issue
			},
			wantValidFiles:  2,
			wantTotalIssues: 3,
		},
		{
			name:            "empty results",
			results:         []ValidationResult{},
			wantValidFiles:  0,
			wantTotalIssues: 0,
		},
		{
			name: "single valid file",
			results: []ValidationResult{
				{MissingFields: []string{testutil.ValidationTestFile3}},
			},
			wantValidFiles:  1,
			wantTotalIssues: 0,
		},
		{
			name: "single invalid file with multiple issues",
			results: []ValidationResult{
				{MissingFields: []string{testutil.ValidationTestFile3, "name", "description", "runs"}},
			},
			wantValidFiles:  0,
			wantTotalIssues: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := &Generator{}
			gotValid, gotIssues := gen.countValidationStats(tt.results)

			if gotValid != tt.wantValidFiles {
				t.Errorf("countValidationStats() validFiles = %d, want %d", gotValid, tt.wantValidFiles)
			}
			if gotIssues != tt.wantTotalIssues {
				t.Errorf("countValidationStats() totalIssues = %d, want %d", gotIssues, tt.wantTotalIssues)
			}
		})
	}
}

// messageCountExpectations defines expected message counts for validation tests.
type messageCountExpectations struct {
	bold    int
	success int
	warning int
	error   int
	info    int
}

// assertMessageCounts checks that message counts match expectations.
func assertMessageCounts(t *testing.T, output *capturedOutput, want messageCountExpectations) {
	t.Helper()

	checks := []struct {
		name     string
		got      int
		expected int
	}{
		{"bold messages", len(output.BoldMessages), want.bold},
		{"success messages", len(output.SuccessMessages), want.success},
		{"warning messages", len(output.WarningMessages), want.warning},
		{"error messages", len(output.ErrorMessages), want.error},
		{"info messages", len(output.InfoMessages), want.info},
	}

	for _, check := range checks {
		if check.got != check.expected {
			t.Errorf("showValidationSummary() %s = %d, want %d", check.name, check.got, check.expected)
		}
	}
}

// TestShowValidationSummary tests the validation summary display function.
func TestShowValidationSummary(t *testing.T) {
	tests := []validationSummaryTestCase{
		createValidationSummaryTest(validationSummaryParams{
			name:        "all valid files",
			totalFiles:  3,
			validFiles:  3,
			totalIssues: 0,
			resultCount: 3,
			errorCount:  0,
			wantWarning: 0,
			wantError:   0,
			wantInfo:    0,
		}),
		createValidationSummaryTest(validationSummaryParams{
			name:        "some files with issues",
			totalFiles:  3,
			validFiles:  1,
			totalIssues: 5,
			resultCount: 3,
			errorCount:  0,
			wantWarning: 1,
			wantError:   0,
			wantInfo:    1,
		}),
		createValidationSummaryTest(validationSummaryParams{
			name:        "parse errors present",
			totalFiles:  5,
			validFiles:  2,
			totalIssues: 3,
			resultCount: 3,
			errorCount:  2,
			wantWarning: 1,
			wantError:   1,
			wantInfo:    1,
		}),
		createValidationSummaryTest(validationSummaryParams{
			name:        "only parse errors",
			totalFiles:  2,
			validFiles:  0,
			totalIssues: 0,
			resultCount: 0,
			errorCount:  2,
			wantWarning: 0,
			wantError:   1,
			wantInfo:    0,
		}),
		createValidationSummaryTest(validationSummaryParams{
			name:        "zero files",
			totalFiles:  0,
			validFiles:  0,
			totalIssues: 0,
			resultCount: 0,
			errorCount:  0,
			wantWarning: 0,
			wantError:   0,
			wantInfo:    0,
		}),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := newCapturedOutput()
			gen := &Generator{Output: output}

			gen.showValidationSummary(tt.totalFiles, tt.validFiles, tt.totalIssues, tt.resultCount, tt.errorCount)

			assertMessageCounts(t, output, messageCountExpectations{
				bold:    tt.wantBold,
				success: tt.wantSuccess,
				warning: tt.wantWarning,
				error:   tt.wantError,
				info:    tt.wantInfo,
			})
		})
	}
}

// TestShowParseErrors tests the parse error display function.
func TestShowParseErrors(t *testing.T) {
	tests := []struct {
		name         string
		errors       []string
		wantBold     int
		wantError    int
		wantContains string
	}{
		{
			name:         "no parse errors",
			errors:       []string{},
			wantBold:     0,
			wantError:    0,
			wantContains: "",
		},
		{
			name:         "single parse error",
			errors:       []string{"Failed to parse action.yml: invalid YAML"},
			wantBold:     1,
			wantError:    1,
			wantContains: "Failed to parse",
		},
		{
			name: "multiple parse errors",
			errors: []string{
				"Failed to parse action1.yml: invalid YAML",
				"Failed to parse action2.yml: file not found",
				"Failed to parse action3.yml: permission denied",
			},
			wantBold:     1,
			wantError:    3,
			wantContains: "Failed to parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := newCapturedOutput()
			gen := &Generator{Output: output}

			gen.showParseErrors(tt.errors)

			testutil.AssertMessageCounts(t, tt.name, output.CapturedOutput, 0, tt.wantError, 0, tt.wantBold)

			if tt.wantContains != "" && !output.ContainsError(tt.wantContains) {
				t.Errorf(
					"showParseErrors() error messages should contain %q, got %v",
					tt.wantContains,
					output.ErrorMessages,
				)
			}
		})
	}
}

// TestShowFileIssues tests the file-specific issue display function.
func TestShowFileIssues(t *testing.T) {
	tests := []struct {
		name         string
		result       ValidationResult
		wantInfo     int
		wantError    int
		wantWarning  int
		wantContains string
	}{
		{
			name: "file with missing fields only",
			result: ValidationResult{
				MissingFields: []string{testutil.ValidationTestFile3, "name", "description"},
			},
			wantInfo:     1, // File name only (no suggestions)
			wantError:    2, // 2 missing fields
			wantWarning:  0,
			wantContains: "name",
		},
		{
			name: "file with warnings only",
			result: ValidationResult{
				MissingFields: []string{testutil.ValidationTestFile3},
				Warnings:      []string{"author field is recommended", "icon field is recommended"},
			},
			wantInfo:     1, // File name
			wantError:    0,
			wantWarning:  2,
			wantContains: "author",
		},
		{
			name: "file with missing fields and warnings",
			result: ValidationResult{
				MissingFields: []string{testutil.ValidationTestFile3, "name"},
				Warnings:      []string{"author field is recommended"},
			},
			wantInfo:     1,
			wantError:    1,
			wantWarning:  1,
			wantContains: "name",
		},
		{
			name: "file with suggestions",
			result: ValidationResult{
				MissingFields: []string{testutil.ValidationTestFile3, "name"},
				Suggestions:   []string{"Add a descriptive name field", "See documentation for examples"},
			},
			wantInfo:     2, // File name + Suggestions header
			wantError:    1,
			wantWarning:  0,
			wantContains: "descriptive name",
		},
		{
			name: "valid file (no issues)",
			result: ValidationResult{
				MissingFields: []string{testutil.ValidationTestFile3},
			},
			wantInfo:     1, // Just file name
			wantError:    0,
			wantWarning:  0,
			wantContains: appconstants.ActionFileNameYML,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := newCapturedOutput()
			gen := &Generator{Output: output}

			gen.showFileIssues(tt.result)

			if len(output.InfoMessages) < tt.wantInfo {
				t.Errorf("showFileIssues() info messages = %d, want at least %d", len(output.InfoMessages), tt.wantInfo)
			}
			if len(output.ErrorMessages) != tt.wantError {
				t.Errorf("showFileIssues() error messages = %d, want %d", len(output.ErrorMessages), tt.wantError)
			}
			if len(output.WarningMessages) != tt.wantWarning {
				t.Errorf("showFileIssues() warning messages = %d, want %d", len(output.WarningMessages), tt.wantWarning)
			}

			// Check if expected content appears somewhere in the output
			if tt.wantContains != "" && !output.ContainsMessage(tt.wantContains) {
				t.Errorf("showFileIssues() output should contain %q, got info=%v, error=%v, warning=%v",
					tt.wantContains, output.InfoMessages, output.ErrorMessages, output.WarningMessages)
			}
		})
	}
}

// TestShowDetailedIssues tests the detailed issues display function.
func TestShowDetailedIssues(t *testing.T) {
	tests := []struct {
		name        string
		results     []ValidationResult
		totalIssues int
		verbose     bool
		wantBold    int // Expected number of bold messages
	}{
		{
			name: "no issues, not verbose",
			results: []ValidationResult{
				{MissingFields: []string{"file: action1.yml"}},
				{MissingFields: []string{"file: action2.yml"}},
			},
			totalIssues: 0,
			verbose:     false,
			wantBold:    0, // Should not show details
		},
		{
			name: "no issues, verbose mode",
			results: []ValidationResult{
				{MissingFields: []string{testutil.ValidationTestFile1}},
				{MissingFields: []string{testutil.ValidationTestFile2}},
			},
			totalIssues: 0,
			verbose:     true,
			wantBold:    1, // Should show header even with no issues
		},
		{
			name: "some issues",
			results: []ValidationResult{
				{MissingFields: []string{testutil.ValidationTestFile1, "name"}},
				{MissingFields: []string{testutil.ValidationTestFile2}},
			},
			totalIssues: 1,
			verbose:     false,
			wantBold:    1, // Should show details
		},
		{
			name: "files with warnings",
			results: []ValidationResult{
				{MissingFields: []string{testutil.ValidationTestFile1}, Warnings: []string{"author recommended"}},
			},
			totalIssues: 0,
			verbose:     false,
			wantBold:    0, // No bold output (warnings don't count as issues, early return)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := newCapturedOutput()
			gen := &Generator{
				Output: output,
				Config: &AppConfig{Verbose: tt.verbose},
			}

			gen.showDetailedIssues(tt.results, tt.totalIssues)

			if len(output.BoldMessages) != tt.wantBold {
				t.Errorf("showDetailedIssues() bold messages = %d, want %d", len(output.BoldMessages), tt.wantBold)
			}
		})
	}
}

// TestReportValidationResults tests the main validation reporting function.
// reportCounts holds the expected counts for validation report output.
type reportCounts struct {
	bold    int
	success bool
	error   bool
}

// validateReportCounts validates that the report output contains expected message counts.
func validateReportCounts(
	t *testing.T,
	gotBold, gotSuccess, gotError int,
	want reportCounts,
	allowUnexpectedErrors bool,
) {
	t.Helper()

	if gotBold < want.bold {
		t.Errorf("Bold messages = %d, want at least %d", gotBold, want.bold)
	}

	if want.success && gotSuccess == 0 {
		t.Error("Expected success messages, got none")
	}

	if want.error && gotError == 0 {
		t.Error("Expected error messages, got none")
	}

	if !allowUnexpectedErrors && gotError > 0 {
		t.Errorf("Expected no error messages, got %d", gotError)
	}
}

func TestReportValidationResults(t *testing.T) {
	tests := []struct {
		name        string
		results     []ValidationResult
		errors      []string
		wantBold    int // Minimum number of bold messages
		wantSuccess bool
		wantError   bool
	}{
		{
			name: "all valid, no errors",
			results: []ValidationResult{
				{MissingFields: []string{testutil.ValidationTestFile1}},
				{MissingFields: []string{testutil.ValidationTestFile2}},
			},
			errors:      []string{},
			wantBold:    1,
			wantSuccess: true,
			wantError:   false,
		},
		{
			name: "some invalid files",
			results: []ValidationResult{
				{MissingFields: []string{testutil.ValidationTestFile1, "name"}},
				{MissingFields: []string{testutil.ValidationTestFile2}},
			},
			errors:      []string{},
			wantBold:    2, // Summary + Details
			wantSuccess: true,
			wantError:   true,
		},
		{
			name:        "parse errors only",
			results:     []ValidationResult{},
			errors:      []string{"Failed to parse action.yml"},
			wantBold:    2, // Summary + Parse Errors
			wantSuccess: true,
			wantError:   true,
		},
		{
			name: "mixed validation issues and parse errors",
			results: []ValidationResult{
				{MissingFields: []string{testutil.ValidationTestFile1, "name", "description"}},
			},
			errors:      []string{"Failed to parse action2.yml"},
			wantBold:    3, // Summary + Details + Parse Errors
			wantSuccess: true,
			wantError:   true,
		},
		{
			name:        "empty results",
			results:     []ValidationResult{},
			errors:      []string{},
			wantBold:    1,
			wantSuccess: true,
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := newCapturedOutput()
			gen := &Generator{
				Output: output,
				Config: &AppConfig{Verbose: false},
			}

			gen.reportValidationResults(tt.results, tt.errors)

			counts := reportCounts{
				bold:    tt.wantBold,
				success: tt.wantSuccess,
				error:   tt.wantError,
			}
			validateReportCounts(
				t,
				len(output.BoldMessages),
				len(output.SuccessMessages),
				len(output.ErrorMessages),
				counts,
				tt.wantError,
			)
		})
	}
}
