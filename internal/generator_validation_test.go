package internal

import (
	"os"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/apperrors"
)

// capturedOutput captures output for testing validation reporting functions.
// Implements CompleteOutput interface.
type capturedOutput struct {
	boldMessages               []string
	successMessages            []string
	errorMessages              []string
	warningMessages            []string
	infoMessages               []string
	printfMessages             []string
	progressMessages           []string
	errorWithSuggestionsCalls  []string
	errorWithContextCalls      []string
	errorWithSimpleFixCalls    []string
	formatContextualErrorCalls []string
}

// MessageLogger implementation.
func (c *capturedOutput) Bold(format string, args ...any) {
	c.boldMessages = append(c.boldMessages, formatMessage(format, args...))
}

func (c *capturedOutput) Success(format string, args ...any) {
	c.successMessages = append(c.successMessages, formatMessage(format, args...))
}

func (c *capturedOutput) Error(format string, args ...any) {
	c.errorMessages = append(c.errorMessages, formatMessage(format, args...))
}

func (c *capturedOutput) Warning(format string, args ...any) {
	c.warningMessages = append(c.warningMessages, formatMessage(format, args...))
}

func (c *capturedOutput) Info(format string, args ...any) {
	c.infoMessages = append(c.infoMessages, formatMessage(format, args...))
}

func (c *capturedOutput) Printf(format string, args ...any) {
	c.printfMessages = append(c.printfMessages, formatMessage(format, args...))
}

func (c *capturedOutput) Fprintf(_ *os.File, format string, args ...any) {
	c.printfMessages = append(c.printfMessages, formatMessage(format, args...))
}

// ErrorReporter implementation.
func (c *capturedOutput) ErrorWithSuggestions(err *apperrors.ContextualError) {
	if err != nil {
		c.errorWithSuggestionsCalls = append(c.errorWithSuggestionsCalls, err.Error())
	}
}

func (c *capturedOutput) ErrorWithContext(_ appconstants.ErrorCode, message string, _ map[string]string) {
	c.errorWithContextCalls = append(c.errorWithContextCalls, message)
}

func (c *capturedOutput) ErrorWithSimpleFix(message, suggestion string) {
	c.errorWithSimpleFixCalls = append(c.errorWithSimpleFixCalls, message+": "+suggestion)
}

// ErrorFormatter implementation.
func (c *capturedOutput) FormatContextualError(err *apperrors.ContextualError) string {
	if err != nil {
		formatted := err.Error()
		c.formatContextualErrorCalls = append(c.formatContextualErrorCalls, formatted)

		return formatted
	}

	return ""
}

// ProgressReporter implementation.
func (c *capturedOutput) Progress(format string, args ...any) {
	c.progressMessages = append(c.progressMessages, formatMessage(format, args...))
}

// OutputConfig implementation.
func (c *capturedOutput) IsQuiet() bool {
	return false
}

// allMessages consolidates all message slices into a single slice.
func (c *capturedOutput) allMessages() []string {
	total := len(c.infoMessages) + len(c.errorMessages) + len(c.warningMessages) + len(c.printfMessages)
	messages := make([]string, 0, total)
	messages = append(messages, c.infoMessages...)
	messages = append(messages, c.errorMessages...)
	messages = append(messages, c.warningMessages...)
	messages = append(messages, c.printfMessages...)

	return messages
}

// containsAnyMessage checks if any message in the consolidated list contains the needle.
func (c *capturedOutput) containsAnyMessage(needle string) bool {
	for _, msg := range c.allMessages() {
		if strings.Contains(msg, needle) {
			return true
		}
	}

	return false
}

// containsErrorMessage checks if any error message contains the needle.
func (c *capturedOutput) containsErrorMessage(needle string) bool {
	for _, msg := range c.errorMessages {
		if strings.Contains(msg, needle) {
			return true
		}
	}

	return false
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
				{MissingFields: []string{appconstants.ValidationTestFile1}},
				{MissingFields: []string{appconstants.ValidationTestFile2}},
			},
			wantValidFiles:  2,
			wantTotalIssues: 0,
		},
		{
			name: "all invalid files",
			results: []ValidationResult{
				{MissingFields: []string{appconstants.ValidationTestFile1, "name", "description"}},
				{MissingFields: []string{appconstants.ValidationTestFile2, "runs"}},
			},
			wantValidFiles:  0,
			wantTotalIssues: 3, // 2 issues in first file + 1 in second
		},
		{
			name: "mixed valid and invalid",
			results: []ValidationResult{
				{MissingFields: []string{appconstants.ValidationTestFile1}},                        // Valid
				{MissingFields: []string{appconstants.ValidationTestFile2, "name", "description"}}, // 2 issues
				{MissingFields: []string{"file: action3.yml"}},                                     // Valid
				{MissingFields: []string{"file: action4.yml", "runs"}},                             // 1 issue
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
				{MissingFields: []string{appconstants.ValidationTestFile3}},
			},
			wantValidFiles:  1,
			wantTotalIssues: 0,
		},
		{
			name: "single invalid file with multiple issues",
			results: []ValidationResult{
				{MissingFields: []string{appconstants.ValidationTestFile3, "name", "description", "runs"}},
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
		{"bold messages", len(output.boldMessages), want.bold},
		{"success messages", len(output.successMessages), want.success},
		{"warning messages", len(output.warningMessages), want.warning},
		{"error messages", len(output.errorMessages), want.error},
		{"info messages", len(output.infoMessages), want.info},
	}

	for _, check := range checks {
		if check.got != check.expected {
			t.Errorf("showValidationSummary() %s = %d, want %d", check.name, check.got, check.expected)
		}
	}
}

// TestShowValidationSummary tests the validation summary display function.
func TestShowValidationSummary(t *testing.T) {
	tests := []struct {
		name        string
		totalFiles  int
		validFiles  int
		totalIssues int
		resultCount int
		errorCount  int
		wantBold    int // Number of bold messages expected
		wantSuccess int // Number of success messages expected
		wantWarning int // Number of warning messages expected
		wantError   int // Number of error messages expected
		wantInfo    int // Number of info messages expected
	}{
		{
			name:        "all valid files",
			totalFiles:  3,
			validFiles:  3,
			totalIssues: 0,
			resultCount: 3,
			errorCount:  0,
			wantBold:    1,
			wantSuccess: 1,
			wantWarning: 0,
			wantError:   0,
			wantInfo:    0,
		},
		{
			name:        "some files with issues",
			totalFiles:  3,
			validFiles:  1,
			totalIssues: 5,
			resultCount: 3,
			errorCount:  0,
			wantBold:    1,
			wantSuccess: 1,
			wantWarning: 1, // Files with issues
			wantError:   0,
			wantInfo:    1, // Total issues
		},
		{
			name:        "parse errors present",
			totalFiles:  5,
			validFiles:  2,
			totalIssues: 3,
			resultCount: 3,
			errorCount:  2,
			wantBold:    1,
			wantSuccess: 1,
			wantWarning: 1, // Files with issues
			wantError:   1, // Parse errors
			wantInfo:    1, // Total issues
		},
		{
			name:        "only parse errors",
			totalFiles:  2,
			validFiles:  0,
			totalIssues: 0,
			resultCount: 0,
			errorCount:  2,
			wantBold:    1,
			wantSuccess: 1,
			wantWarning: 0,
			wantError:   1, // Parse errors
			wantInfo:    0,
		},
		{
			name:        "zero files",
			totalFiles:  0,
			validFiles:  0,
			totalIssues: 0,
			resultCount: 0,
			errorCount:  0,
			wantBold:    1,
			wantSuccess: 1,
			wantWarning: 0,
			wantError:   0,
			wantInfo:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &capturedOutput{}
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
			output := &capturedOutput{}
			gen := &Generator{Output: output}

			gen.showParseErrors(tt.errors)

			if len(output.boldMessages) != tt.wantBold {
				t.Errorf("showParseErrors() bold messages = %d, want %d", len(output.boldMessages), tt.wantBold)
			}
			if len(output.errorMessages) != tt.wantError {
				t.Errorf("showParseErrors() error messages = %d, want %d", len(output.errorMessages), tt.wantError)
			}

			if tt.wantContains != "" && !output.containsErrorMessage(tt.wantContains) {
				t.Errorf(
					"showParseErrors() error messages should contain %q, got %v",
					tt.wantContains,
					output.errorMessages,
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
				MissingFields: []string{appconstants.ValidationTestFile3, "name", "description"},
			},
			wantInfo:     1, // File name only (no suggestions)
			wantError:    2, // 2 missing fields
			wantWarning:  0,
			wantContains: "name",
		},
		{
			name: "file with warnings only",
			result: ValidationResult{
				MissingFields: []string{appconstants.ValidationTestFile3},
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
				MissingFields: []string{appconstants.ValidationTestFile3, "name"},
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
				MissingFields: []string{appconstants.ValidationTestFile3, "name"},
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
				MissingFields: []string{appconstants.ValidationTestFile3},
			},
			wantInfo:     1, // Just file name
			wantError:    0,
			wantWarning:  0,
			wantContains: "action.yml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &capturedOutput{}
			gen := &Generator{Output: output}

			gen.showFileIssues(tt.result)

			if len(output.infoMessages) < tt.wantInfo {
				t.Errorf("showFileIssues() info messages = %d, want at least %d", len(output.infoMessages), tt.wantInfo)
			}
			if len(output.errorMessages) != tt.wantError {
				t.Errorf("showFileIssues() error messages = %d, want %d", len(output.errorMessages), tt.wantError)
			}
			if len(output.warningMessages) != tt.wantWarning {
				t.Errorf("showFileIssues() warning messages = %d, want %d", len(output.warningMessages), tt.wantWarning)
			}

			// Check if expected content appears somewhere in the output
			if tt.wantContains != "" && !output.containsAnyMessage(tt.wantContains) {
				t.Errorf("showFileIssues() output should contain %q, got info=%v, error=%v, warning=%v",
					tt.wantContains, output.infoMessages, output.errorMessages, output.warningMessages)
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
				{MissingFields: []string{appconstants.ValidationTestFile1}},
				{MissingFields: []string{appconstants.ValidationTestFile2}},
			},
			totalIssues: 0,
			verbose:     true,
			wantBold:    1, // Should show header even with no issues
		},
		{
			name: "some issues",
			results: []ValidationResult{
				{MissingFields: []string{appconstants.ValidationTestFile1, "name"}},
				{MissingFields: []string{appconstants.ValidationTestFile2}},
			},
			totalIssues: 1,
			verbose:     false,
			wantBold:    1, // Should show details
		},
		{
			name: "files with warnings",
			results: []ValidationResult{
				{MissingFields: []string{appconstants.ValidationTestFile1}, Warnings: []string{"author recommended"}},
			},
			totalIssues: 0,
			verbose:     false,
			wantBold:    0, // No bold output (warnings don't count as issues, early return)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &capturedOutput{}
			gen := &Generator{
				Output: output,
				Config: &AppConfig{Verbose: tt.verbose},
			}

			gen.showDetailedIssues(tt.results, tt.totalIssues)

			if len(output.boldMessages) != tt.wantBold {
				t.Errorf("showDetailedIssues() bold messages = %d, want %d", len(output.boldMessages), tt.wantBold)
			}
		})
	}
}

// TestReportValidationResults tests the main validation reporting function.
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
				{MissingFields: []string{appconstants.ValidationTestFile1}},
				{MissingFields: []string{appconstants.ValidationTestFile2}},
			},
			errors:      []string{},
			wantBold:    1,
			wantSuccess: true,
			wantError:   false,
		},
		{
			name: "some invalid files",
			results: []ValidationResult{
				{MissingFields: []string{appconstants.ValidationTestFile1, "name"}},
				{MissingFields: []string{appconstants.ValidationTestFile2}},
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
				{MissingFields: []string{appconstants.ValidationTestFile1, "name", "description"}},
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
			output := &capturedOutput{}
			gen := &Generator{
				Output: output,
				Config: &AppConfig{Verbose: false},
			}

			gen.reportValidationResults(tt.results, tt.errors)

			gotBoldCount := len(output.boldMessages)
			gotSuccessCount := len(output.successMessages)
			gotErrorCount := len(output.errorMessages)

			if gotBoldCount < tt.wantBold {
				t.Errorf(
					"reportValidationResults() bold messages = %d, want at least %d",
					gotBoldCount,
					tt.wantBold,
				)
			}
			if tt.wantSuccess && gotSuccessCount == 0 {
				t.Error("reportValidationResults() expected success messages, got none")
			}
			if tt.wantError && gotErrorCount == 0 {
				t.Error("reportValidationResults() expected error messages, got none")
			}
			if !tt.wantError && gotErrorCount > 0 {
				t.Errorf("reportValidationResults() expected no error messages, got %d", gotErrorCount)
			}
		})
	}
}
