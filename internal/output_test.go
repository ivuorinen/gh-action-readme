package internal

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/apperrors"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// captureStdout captures stdout output for testing.
func captureStdout(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	_ = w.Close() // Ignore error in test helper
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Ignore error in test helper

	return buf.String()
}

// captureStderr captures stderr output for testing.
func captureStderr(f func()) string {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	f()

	_ = w.Close() // Ignore error in test helper
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Ignore error in test helper

	return buf.String()
}

// TestNewColoredOutput tests colored output creation.
func TestNewColoredOutput(t *testing.T) {
	tests := []struct {
		name      string
		quiet     bool
		wantQuiet bool
	}{
		{
			name:      testutil.TestScenarioQuietEnabled,
			quiet:     true,
			wantQuiet: true,
		},
		{
			name:      testutil.TestScenarioQuietDisabled,
			quiet:     false,
			wantQuiet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewColoredOutput(tt.quiet)

			if output == nil {
				t.Fatal("NewColoredOutput() returned nil")
			}

			if output.Quiet != tt.wantQuiet {
				t.Errorf("Quiet = %v, want %v", output.Quiet, tt.wantQuiet)
			}
		})
	}
}

// TestIsQuiet tests quiet mode detection.
func TestIsQuiet(t *testing.T) {
	tests := []struct {
		name  string
		quiet bool
		want  bool
	}{
		{
			name:  testutil.TestScenarioQuietEnabled,
			quiet: true,
			want:  true,
		},
		{
			name:  testutil.TestScenarioQuietDisabled,
			quiet: false,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &ColoredOutput{Quiet: tt.quiet, NoColor: true}
			got := output.IsQuiet()

			if got != tt.want {
				t.Errorf("IsQuiet() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSuccess tests success message output.
func TestSuccess(t *testing.T) {
	tests := []struct {
		name         string
		quiet        bool
		message      string
		wantContains string
		wantEmpty    bool
	}{
		{
			name:         "success message displayed",
			quiet:        false,
			message:      testutil.TestMsgOperationCompleted,
			wantContains: "✅ Operation completed",
			wantEmpty:    false,
		},
		{
			name:      testutil.TestMsgQuietSuppressOutput,
			quiet:     true,
			message:   testutil.TestMsgOperationCompleted,
			wantEmpty: true,
		},
		{
			name:         "success with formatting",
			quiet:        false,
			message:      "Processed %d files",
			wantContains: "✅ Processed %!d(MISSING) files",
			wantEmpty:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &ColoredOutput{Quiet: tt.quiet, NoColor: true}

			captured := captureStdout(func() {
				output.Success(tt.message)
			})

			if tt.wantEmpty && captured != "" {
				t.Errorf(testutil.TestMsgNoOutputInQuiet, captured)
			}

			if !tt.wantEmpty && !strings.Contains(captured, "✅") {
				t.Errorf("Output missing success emoji: %q", captured)
			}
		})
	}
}

// TestError tests error message output.
func TestError(t *testing.T) {
	tests := []struct {
		name         string
		message      string
		wantContains string
	}{
		{
			name:         "error message displayed",
			message:      testutil.TestMsgFileNotFound,
			wantContains: "❌ File not found",
		},
		{
			name:         "error with formatting",
			message:      "Failed to process %s",
			wantContains: "❌ Failed to process %!s(MISSING)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &ColoredOutput{NoColor: true}

			captured := captureStderr(func() {
				output.Error(tt.message)
			})

			if !strings.Contains(captured, "❌") {
				t.Errorf(testutil.TestMsgOutputMissingEmoji, captured)
			}

			if !strings.Contains(captured, strings.TrimPrefix(tt.wantContains, "❌ ")) {
				t.Errorf("Output doesn't contain expected message. Got: %q", captured)
			}
		})
	}
}

// TestWarning tests warning message output.
func TestWarning(t *testing.T) {
	tests := []struct {
		name      string
		quiet     bool
		message   string
		wantEmpty bool
	}{
		{
			name:      "warning message displayed",
			quiet:     false,
			message:   "Deprecated feature",
			wantEmpty: false,
		},
		{
			name:      testutil.TestMsgQuietSuppressOutput,
			quiet:     true,
			message:   "Deprecated feature",
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &ColoredOutput{Quiet: tt.quiet, NoColor: true}

			captured := captureStdout(func() {
				output.Warning(tt.message)
			})

			if tt.wantEmpty && captured != "" {
				t.Errorf(testutil.TestMsgNoOutputInQuiet, captured)
			}

			if !tt.wantEmpty && !strings.Contains(captured, "⚠️") {
				t.Errorf("Output missing warning emoji: %q", captured)
			}
		})
	}
}

// TestInfo tests info message output.
func TestInfo(t *testing.T) {
	tests := []struct {
		name      string
		quiet     bool
		message   string
		wantEmpty bool
	}{
		{
			name:      "info message displayed",
			quiet:     false,
			message:   testutil.TestMsgProcessingStarted,
			wantEmpty: false,
		},
		{
			name:      testutil.TestMsgQuietSuppressOutput,
			quiet:     true,
			message:   testutil.TestMsgProcessingStarted,
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &ColoredOutput{Quiet: tt.quiet, NoColor: true}

			captured := captureStdout(func() {
				output.Info(tt.message)
			})

			if tt.wantEmpty && captured != "" {
				t.Errorf(testutil.TestMsgNoOutputInQuiet, captured)
			}

			if !tt.wantEmpty && !strings.Contains(captured, "ℹ️") {
				t.Errorf("Output missing info emoji: %q", captured)
			}
		})
	}
}

// TestProgress tests progress message output.
func TestProgress(t *testing.T) {
	tests := []struct {
		name      string
		quiet     bool
		message   string
		wantEmpty bool
	}{
		{
			name:      "progress message displayed",
			quiet:     false,
			message:   "Loading data...",
			wantEmpty: false,
		},
		{
			name:      testutil.TestMsgQuietSuppressOutput,
			quiet:     true,
			message:   "Loading data...",
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &ColoredOutput{Quiet: tt.quiet, NoColor: true}

			captured := captureStdout(func() {
				output.Progress(tt.message)
			})

			if tt.wantEmpty && captured != "" {
				t.Errorf(testutil.TestMsgNoOutputInQuiet, captured)
			}

			if !tt.wantEmpty && !strings.Contains(captured, "🔄") {
				t.Errorf("Output missing progress emoji: %q", captured)
			}
		})
	}
}

// TestBold tests bold text output.
func TestBold(t *testing.T) {
	tests := []struct {
		name      string
		quiet     bool
		message   string
		wantEmpty bool
	}{
		{
			name:      "bold message displayed",
			quiet:     false,
			message:   "Important Notice",
			wantEmpty: false,
		},
		{
			name:      testutil.TestMsgQuietSuppressOutput,
			quiet:     true,
			message:   "Important Notice",
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &ColoredOutput{Quiet: tt.quiet, NoColor: true}

			captured := captureStdout(func() {
				output.Bold(tt.message)
			})

			if tt.wantEmpty && captured != "" {
				t.Errorf(testutil.TestMsgNoOutputInQuiet, captured)
			}

			if !tt.wantEmpty && !strings.Contains(captured, tt.message) {
				t.Errorf("Output doesn't contain message. Got: %q", captured)
			}
		})
	}
}

// TestPrintf tests formatted print output.
func TestPrintf(t *testing.T) {
	tests := []struct {
		name      string
		quiet     bool
		wantEmpty bool
	}{
		{
			name:      "printf output displayed",
			quiet:     false,
			wantEmpty: false,
		},
		{
			name:      testutil.TestMsgQuietSuppressOutput,
			quiet:     true,
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &ColoredOutput{Quiet: tt.quiet, NoColor: true}

			captured := captureStdout(func() {
				output.Printf("Test message\n")
			})

			if tt.wantEmpty && captured != "" {
				t.Errorf(testutil.TestMsgNoOutputInQuiet, captured)
			}

			if !tt.wantEmpty && captured == "" {
				t.Error("Expected output, got empty string")
			}
		})
	}
}

// TestFprintf tests file output.
func TestFprintf(t *testing.T) {
	// Create temporary file for testing
	tmpfile, err := os.CreateTemp(t.TempDir(), "test-fprintf-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpfile.Name()) }() // Ignore error
	defer func() { _ = tmpfile.Close() }()           // Ignore error

	output := &ColoredOutput{NoColor: true}
	output.Fprintf(tmpfile, "Test message: %s\n", "hello")

	// Read back the content
	_, _ = tmpfile.Seek(0, 0) // Ignore error in test
	content := make([]byte, 100)
	n, _ := tmpfile.Read(content)

	got := string(content[:n])
	want := "Test message: hello\n"

	if got != want {
		t.Errorf("Fprintf() wrote %q, want %q", got, want)
	}
}

// TestErrorWithSuggestions tests contextual error output.
func TestErrorWithSuggestions(t *testing.T) {
	tests := []struct {
		name         string
		err          *apperrors.ContextualError
		wantContains string
	}{
		{
			name:         "nil error does nothing",
			err:          nil,
			wantContains: "",
		},
		{
			name: "error with suggestions",
			err: apperrors.New(appconstants.ErrCodeFileNotFound, testutil.TestMsgFileNotFound).
				WithSuggestions(testutil.TestMsgCheckFilePath),
			wantContains: "❌",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &ColoredOutput{NoColor: true}

			captured := captureStderr(func() {
				output.ErrorWithSuggestions(tt.err)
			})

			if tt.wantContains == "" && captured != "" {
				t.Errorf("Expected no output for nil error, got %q", captured)
			}

			if tt.wantContains != "" && !strings.Contains(captured, tt.wantContains) {
				t.Errorf("Output doesn't contain %q. Got: %q", tt.wantContains, captured)
			}
		})
	}
}

// TestErrorWithContext tests contextual error creation and output.
func TestErrorWithContext(t *testing.T) {
	tests := []struct {
		name    string
		code    appconstants.ErrorCode
		message string
		context map[string]string
	}{
		{
			name:    "error with context",
			code:    appconstants.ErrCodeFileNotFound,
			message: testutil.TestMsgFileNotFound,
			context: map[string]string{testutil.TestKeyFile: appconstants.ActionFileNameYML},
		},
		{
			name:    "error without context",
			code:    appconstants.ErrCodeInvalidYAML,
			message: testutil.TestMsgInvalidYAML,
			context: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &ColoredOutput{NoColor: true}

			captured := captureStderr(func() {
				output.ErrorWithContext(tt.code, tt.message, tt.context)
			})

			if !strings.Contains(captured, "❌") {
				t.Errorf(testutil.TestMsgOutputMissingEmoji, captured)
			}
		})
	}
}

// TestErrorWithSimpleFix tests simple error with fix output.
func TestErrorWithSimpleFix(t *testing.T) {
	output := &ColoredOutput{NoColor: true}

	captured := captureStderr(func() {
		output.ErrorWithSimpleFix("Something went wrong", "Try running it again")
	})

	if !strings.Contains(captured, "❌") {
		t.Errorf(testutil.TestMsgOutputMissingEmoji, captured)
	}
}

// TestFormatContextualError tests contextual error formatting.
func TestFormatContextualError(t *testing.T) {
	tests := []struct {
		name         string
		err          *apperrors.ContextualError
		wantContains []string
	}{
		{
			name:         "nil error returns empty string",
			err:          nil,
			wantContains: nil,
		},
		{
			name: "error with all sections",
			err: apperrors.New(appconstants.ErrCodeFileNotFound, testutil.TestMsgFileNotFound).
				WithSuggestions(testutil.TestMsgCheckFilePath, testutil.TestMsgVerifyPermissions).
				WithDetails(map[string]string{testutil.TestKeyFile: appconstants.ActionFileNameYML}).
				WithHelpURL(testutil.TestURLHelp),
			wantContains: []string{
				"❌",
				testutil.TestMsgFileNotFound,
				testutil.TestMsgCheckFilePath,
				testutil.TestURLHelp,
			},
		},
		{
			name:         "error without suggestions",
			err:          apperrors.New(appconstants.ErrCodeInvalidYAML, testutil.TestMsgInvalidYAML),
			wantContains: []string{"❌", testutil.TestMsgInvalidYAML},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &ColoredOutput{NoColor: true}
			got := output.FormatContextualError(tt.err)

			if tt.err == nil && got != "" {
				t.Errorf("Expected empty string for nil error, got %q", got)
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("FormatContextualError() missing %q. Got:\n%s", want, got)
				}
			}
		})
	}
}

// TestFormatMainError tests main error message formatting.
func TestFormatMainError(t *testing.T) {
	tests := []struct {
		name         string
		noColor      bool
		err          *apperrors.ContextualError
		wantContains []string
	}{
		{
			name:         testutil.TestScenarioColorDisabled,
			noColor:      true,
			err:          apperrors.New(appconstants.ErrCodeFileNotFound, testutil.TestMsgFileNotFound),
			wantContains: []string{"❌", testutil.TestMsgFileNotFound, string(appconstants.ErrCodeFileNotFound)},
		},
		{
			name:         testutil.TestScenarioColorEnabled,
			noColor:      false,
			err:          apperrors.New(appconstants.ErrCodeInvalidYAML, testutil.TestMsgInvalidYAML),
			wantContains: []string{"❌", testutil.TestMsgInvalidYAML, string(appconstants.ErrCodeInvalidYAML)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &ColoredOutput{NoColor: tt.noColor}
			got := output.formatMainError(tt.err)

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("formatMainError() missing %q. Got: %q", want, got)
				}
			}
		})
	}
}

// TestFormatDetailsSection tests details section formatting.
func TestFormatDetailsSection(t *testing.T) {
	tests := []struct {
		name         string
		noColor      bool
		details      map[string]string
		wantContains []string
	}{
		{
			name:    testutil.TestScenarioColorDisabled,
			noColor: true,
			details: map[string]string{testutil.TestKeyFile: appconstants.ActionFileNameYML, "line": "10"},
			wantContains: []string{
				testutil.TestMsgDetails,
				testutil.TestKeyFile,
				appconstants.ActionFileNameYML,
				"line",
				"10",
			},
		},
		{
			name:         testutil.TestScenarioColorEnabled,
			noColor:      false,
			details:      map[string]string{testutil.TestKeyPath: "/tmp/test"},
			wantContains: []string{testutil.TestMsgDetails, "path", "/tmp/test"},
		},
		{
			name:         "empty details",
			noColor:      true,
			details:      map[string]string{},
			wantContains: []string{testutil.TestMsgDetails},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &ColoredOutput{NoColor: tt.noColor}
			got := output.formatDetailsSection(tt.details)
			gotStr := strings.Join(got, "\n")

			for _, want := range tt.wantContains {
				if !strings.Contains(gotStr, want) {
					t.Errorf("formatDetailsSection() missing %q. Got:\n%s", want, gotStr)
				}
			}
		})
	}
}

// TestFormatSuggestionsSection tests suggestions section formatting.
func TestFormatSuggestionsSection(t *testing.T) {
	tests := []struct {
		name         string
		noColor      bool
		suggestions  []string
		wantContains []string
	}{
		{
			name:        testutil.TestScenarioColorDisabled,
			noColor:     true,
			suggestions: []string{"Check the file", testutil.TestMsgVerifyPermissions},
			wantContains: []string{
				testutil.TestMsgSuggestions,
				"•",
				"Check the file",
				testutil.TestMsgVerifyPermissions,
			},
		},
		{
			name:         testutil.TestScenarioColorEnabled,
			noColor:      false,
			suggestions:  []string{testutil.TestMsgTryAgain},
			wantContains: []string{testutil.TestMsgSuggestions, "•", testutil.TestMsgTryAgain},
		},
		{
			name:         "empty suggestions",
			noColor:      true,
			suggestions:  []string{},
			wantContains: []string{testutil.TestMsgSuggestions},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &ColoredOutput{NoColor: tt.noColor}
			got := output.formatSuggestionsSection(tt.suggestions)
			gotStr := strings.Join(got, "\n")

			for _, want := range tt.wantContains {
				if !strings.Contains(gotStr, want) {
					t.Errorf("formatSuggestionsSection() missing %q. Got:\n%s", want, gotStr)
				}
			}
		})
	}
}

// TestFormatHelpURLSection tests help URL section formatting.
func TestFormatHelpURLSection(t *testing.T) {
	tests := []struct {
		name         string
		noColor      bool
		helpURL      string
		wantContains []string
	}{
		{
			name:         testutil.TestScenarioColorDisabled,
			noColor:      true,
			helpURL:      testutil.TestURLHelp,
			wantContains: []string{"For more help", testutil.TestURLHelp},
		},
		{
			name:         testutil.TestScenarioColorEnabled,
			noColor:      false,
			helpURL:      "https://docs.example.com",
			wantContains: []string{"For more help", "https://docs.example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &ColoredOutput{NoColor: tt.noColor}
			got := output.formatHelpURLSection(tt.helpURL)

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("formatHelpURLSection() missing %q. Got: %q", want, got)
				}
			}
		})
	}
}
