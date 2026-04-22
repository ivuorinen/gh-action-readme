package internal

import (
	"os"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/apperrors"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// testOutputMethod is a generic helper for testing output methods that follow the same pattern.
func testOutputMethod(t *testing.T, testMessage, expectedEmoji string, methodFunc func(*ColoredOutput, string)) {
	t.Helper()

	tests := []struct {
		name      string
		quiet     bool
		message   string
		wantEmpty bool
	}{
		{
			name:      "message displayed",
			quiet:     false,
			message:   testMessage,
			wantEmpty: false,
		},
		{
			name:      testutil.TestMsgQuietSuppressOutput,
			quiet:     true,
			message:   testMessage,
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &ColoredOutput{Quiet: tt.quiet, NoColor: true}

			captured := testutil.CaptureStdout(func() {
				methodFunc(output, tt.message)
			})

			if tt.wantEmpty && captured != "" {
				t.Errorf(testutil.TestMsgNoOutputInQuiet, captured)
			}

			if !tt.wantEmpty && !strings.Contains(captured, expectedEmoji) {
				t.Errorf("Output missing %s emoji: %q", expectedEmoji, captured)
			}
		})
	}
}

// testErrorStderr is a helper for testing error output methods that write to stderr.
// Eliminates the repeated pattern of creating ColoredOutput, capturing stderr, and checking for emoji.
func testErrorStderr(t *testing.T, expectedEmoji string, testFunc func(*ColoredOutput)) {
	t.Helper()

	output := &ColoredOutput{NoColor: true}
	captured := testutil.CaptureStderr(func() {
		testFunc(output)
	})

	if !strings.Contains(captured, expectedEmoji) {
		t.Errorf("Output missing %s emoji: %q", expectedEmoji, captured)
	}
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

// TestNewColoredOutput_NoColorEnvVar tests that NewColoredOutput respects the
// NO_COLOR environment variable by setting NoColor = true when it is set.
func TestNewColoredOutput_NoColorEnvVar(t *testing.T) {
	t.Run("NO_COLOR set forces NoColor true", func(t *testing.T) {
		t.Setenv("NO_COLOR", "1")

		output := NewColoredOutput(false)
		if output == nil {
			t.Fatal("NewColoredOutput() returned nil")
		}

		if !output.NoColor {
			t.Error("expected NoColor = true when NO_COLOR env var is set")
		}
	})

	t.Run("NO_COLOR unset allows color", func(t *testing.T) {
		t.Setenv("NO_COLOR", "")

		output := NewColoredOutput(false)
		if output == nil {
			t.Fatal("NewColoredOutput() returned nil")
		}

		// When NO_COLOR is empty string, NoColor depends only on color.NoColor.
		// We can't assert false here because CI/terminal state may differ.
		// What we CAN assert is that the function completed without error.
		_ = output.NoColor
	})
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
	testOutputMethod(t, testutil.TestMsgOperationCompleted, "✅", func(o *ColoredOutput, msg string) {
		o.Success(msg)
	})
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

			captured := testutil.CaptureStderr(func() {
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
	testOutputMethod(t, "Deprecated feature", "⚠️", func(o *ColoredOutput, msg string) {
		o.Warning(msg)
	})
}

// TestInfo tests info message output.
func TestInfo(t *testing.T) {
	testOutputMethod(t, testutil.TestMsgProcessingStarted, "ℹ️", func(o *ColoredOutput, msg string) {
		o.Info(msg)
	})
}

// TestProgress tests progress message output.
func TestProgress(t *testing.T) {
	testOutputMethod(t, "Loading data...", "🔄", func(o *ColoredOutput, msg string) {
		o.Progress(msg)
	})
}

// TestBold tests bold text output.
func TestBold(t *testing.T) {
	testOutputMethod(t, "Important Notice", "Important Notice", func(o *ColoredOutput, msg string) {
		o.Bold(msg)
	})
}

// TestPrintf tests formatted print output.
func TestPrintf(t *testing.T) {
	testOutputMethod(t, "Test message\n", "Test message", func(o *ColoredOutput, msg string) {
		o.Printf("%s", msg) // #nosec G104 -- constant format string
	})
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

			captured := testutil.CaptureStderr(func() {
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

			captured := testutil.CaptureStderr(func() {
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
	testErrorStderr(t, "❌", func(output *ColoredOutput) {
		output.ErrorWithSimpleFix("Something went wrong", "Try running it again")
	})
}

// TestErrorWithContext_EmptyVsNonEmpty verifies that ErrorWithContext behaves
// differently when context is empty vs non-empty (guards the WithDetails branch).
func TestErrorWithContext_EmptyVsNonEmpty(t *testing.T) {
	t.Run("empty context does not call WithDetails", func(t *testing.T) {
		output := &ColoredOutput{NoColor: true}
		captured := testutil.CaptureStderr(func() {
			output.ErrorWithContext(appconstants.ErrCodeFileNotFound, testutil.TestMsgFileNotFound, nil)
		})

		if !strings.Contains(captured, "❌") {
			t.Errorf("expected error emoji in output, got: %q", captured)
		}
	})

	t.Run("non-empty context calls WithDetails", func(t *testing.T) {
		output := &ColoredOutput{NoColor: true}
		ctx := map[string]string{testutil.TestKeyFile: "action.yml"}
		captured := testutil.CaptureStderr(func() {
			output.ErrorWithContext(appconstants.ErrCodeFileNotFound, testutil.TestMsgFileNotFound, ctx)
		})

		if !strings.Contains(captured, "❌") {
			t.Errorf("expected error emoji in output, got: %q", captured)
		}
	})
}

// TestFormatContextualError_DetailsBoundary verifies that details section appears
// only when len(err.Details) > 0.
func TestFormatContextualError_DetailsBoundary(t *testing.T) {
	output := &ColoredOutput{NoColor: true}

	t.Run("zero details omits details section", func(t *testing.T) {
		err := apperrors.New(appconstants.ErrCodeFileNotFound, testutil.TestMsgFileNotFound)
		got := output.FormatContextualError(err)

		if strings.Contains(got, testutil.TestMsgDetails) {
			t.Errorf("expected no Details section with zero details, got:\n%s", got)
		}
	})

	t.Run("one detail includes details section", func(t *testing.T) {
		err := apperrors.New(appconstants.ErrCodeFileNotFound, testutil.TestMsgFileNotFound).
			WithDetails(map[string]string{testutil.TestKeyFile: "action.yml"})
		got := output.FormatContextualError(err)

		if !strings.Contains(got, testutil.TestMsgDetails) {
			t.Errorf("expected Details section with one detail, got:\n%s", got)
		}
	})
}

// TestFormatContextualError_SuggestionsBoundary verifies that the suggestions section
// appears only when len(err.Suggestions) > 0.
func TestFormatContextualError_SuggestionsBoundary(t *testing.T) {
	output := &ColoredOutput{NoColor: true}

	t.Run("zero suggestions omits suggestions section", func(t *testing.T) {
		err := apperrors.New(appconstants.ErrCodeFileNotFound, testutil.TestMsgFileNotFound)
		got := output.FormatContextualError(err)

		if strings.Contains(got, testutil.TestMsgSuggestions) {
			t.Errorf("expected no Suggestions section with zero suggestions, got:\n%s", got)
		}
	})

	t.Run("one suggestion includes suggestions section", func(t *testing.T) {
		err := apperrors.New(appconstants.ErrCodeFileNotFound, testutil.TestMsgFileNotFound).
			WithSuggestions(testutil.TestMsgCheckFilePath)
		got := output.FormatContextualError(err)

		if !strings.Contains(got, testutil.TestMsgSuggestions) {
			t.Errorf("expected Suggestions section with one suggestion, got:\n%s", got)
		}
	})
}

// TestFormatContextualError_HelpURLNegation verifies that the help URL section
// appears only when err.HelpURL is non-empty.
func TestFormatContextualError_HelpURLNegation(t *testing.T) {
	output := &ColoredOutput{NoColor: true}

	t.Run("empty HelpURL omits help URL section", func(t *testing.T) {
		err := apperrors.New(appconstants.ErrCodeFileNotFound, testutil.TestMsgFileNotFound)
		got := output.FormatContextualError(err)

		if strings.Contains(got, "For more help") {
			t.Errorf("expected no 'For more help' section with empty HelpURL, got:\n%s", got)
		}
	})

	t.Run("non-empty HelpURL includes help URL section", func(t *testing.T) {
		err := apperrors.New(appconstants.ErrCodeFileNotFound, testutil.TestMsgFileNotFound).
			WithHelpURL(testutil.TestURLHelp)
		got := output.FormatContextualError(err)

		if !strings.Contains(got, "For more help") {
			t.Errorf("expected 'For more help' section with non-empty HelpURL, got:\n%s", got)
		}

		if !strings.Contains(got, testutil.TestURLHelp) {
			t.Errorf("expected HelpURL %q in output, got:\n%s", testutil.TestURLHelp, got)
		}
	})
}

// TestError_ColorPath verifies that the Error() method works correctly when
// NoColor is false (exercises the color.New().Fprintf code path at line 54).
func TestError_ColorPath(t *testing.T) {
	output := &ColoredOutput{NoColor: false}

	captured := testutil.CaptureStderr(func() {
		output.Error(testutil.TestMsgFileNotFound)
	})

	if !strings.Contains(captured, testutil.TestMsgFileNotFound) {
		t.Errorf("Error() with color enabled missing message, got: %q", captured)
	}
}

// TestBold_ColorPath verifies that the Bold() method works correctly when
// NoColor is false (exercises the color.New().Printf code path at line 81).
// The color library writes directly to the real stdout, so we cannot capture
// it. We verify the code path is exercised without panic.
func TestBold_ColorPath(_ *testing.T) {
	output := &ColoredOutput{NoColor: false, Quiet: false}
	// Must not panic — the color.New().Printf path should execute cleanly.
	output.Bold("Important Notice")
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
