package apperrors

import (
	"errors"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

const (
	testOriginalError = "original error"
	testMessage       = "test message"
	testContext       = "test context"
)

func TestContextualErrorError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *ContextualError
		contains []string
	}{
		{
			name: "basic error",
			err: &ContextualError{
				Code: appconstants.ErrCodeFileNotFound,
				Err:  errors.New("file not found"),
			},
			contains: []string{"file not found", "[FILE_NOT_FOUND]"},
		},
		{
			name: "error with context",
			err: &ContextualError{
				Code:    appconstants.ErrCodeInvalidYAML,
				Err:     errors.New("invalid syntax"),
				Context: "parsing action.yml",
			},
			contains: []string{"parsing action.yml: invalid syntax", "[INVALID_YAML]"},
		},
		{
			name: "error with suggestions",
			err: &ContextualError{
				Code: appconstants.ErrCodeNoActionFiles,
				Err:  errors.New("no files found"),
				Suggestions: []string{
					"Check current directory",
					"Use --recursive flag",
				},
			},
			contains: []string{
				"no files found",
				"Suggestions:",
				"• Check current directory",
				"• Use --recursive flag",
			},
		},
		{
			name: "error with details",
			err: &ContextualError{
				Code: appconstants.ErrCodeConfiguration,
				Err:  errors.New("config error"),
				Details: map[string]string{
					"config_path": "/path/to/config",
					"line":        "42",
				},
			},
			contains: []string{
				"config error",
				"Details:",
				"config_path: /path/to/config",
				"line: 42",
			},
		},
		{
			name: "error with help URL",
			err: &ContextualError{
				Code:    appconstants.ErrCodeGitHubAPI,
				Err:     errors.New("API error"),
				HelpURL: "https://docs.github.com/api",
			},
			contains: []string{
				"API error",
				"For more help: https://docs.github.com/api",
			},
		},
		{
			name: "complete error",
			err: &ContextualError{
				Code:    appconstants.ErrCodeValidation,
				Err:     errors.New("validation failed"),
				Context: "validating action.yml",
				Details: map[string]string{"file": appconstants.ActionFileNameYML},
				Suggestions: []string{
					"Check required fields",
					"Validate YAML syntax",
				},
				HelpURL: "https://example.com/help",
			},
			contains: []string{
				"validating action.yml: validation failed",
				"[VALIDATION_ERROR]",
				"Details:",
				"file: action.yml",
				"Suggestions:",
				"• Check required fields",
				"• Validate YAML syntax",
				"For more help: https://example.com/help",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.err.Error()
			testutil.AssertSliceContainsAll(t, []string{result}, tt.contains)
		})
	}
}

func TestContextualErrorUnwrap(t *testing.T) {
	t.Parallel()

	originalErr := errors.New(testOriginalError)
	contextualErr := &ContextualError{
		Code: appconstants.ErrCodeFileNotFound,
		Err:  originalErr,
	}

	if unwrapped := contextualErr.Unwrap(); unwrapped != originalErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, originalErr)
	}
}

func TestContextualErrorIs(t *testing.T) {
	t.Parallel()

	originalErr := errors.New(testOriginalError)
	contextualErr := &ContextualError{
		Code: appconstants.ErrCodeFileNotFound,
		Err:  originalErr,
	}

	// Test Is with same error code
	sameCodeErr := &ContextualError{Code: appconstants.ErrCodeFileNotFound}
	if !contextualErr.Is(sameCodeErr) {
		t.Error("Is() should return true for same error code")
	}

	// Test Is with different error code
	differentCodeErr := &ContextualError{Code: appconstants.ErrCodeInvalidYAML}
	if contextualErr.Is(differentCodeErr) {
		t.Error("Is() should return false for different error code")
	}

	// Test Is with wrapped error
	if !errors.Is(contextualErr, originalErr) {
		t.Error("errors.Is() should work with wrapped error")
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	err := New(appconstants.ErrCodeFileNotFound, testMessage)

	if err.Code != appconstants.ErrCodeFileNotFound {
		t.Errorf("New() code = %v, want %v", err.Code, appconstants.ErrCodeFileNotFound)
	}

	if err.Err.Error() != testMessage {
		t.Errorf("New() message = %v, want %v", err.Err.Error(), testMessage)
	}
}

func TestWrap(t *testing.T) {
	t.Parallel()

	originalErr := errors.New(testOriginalError)

	// Test wrapping normal error
	wrapped := Wrap(originalErr, appconstants.ErrCodeFileNotFound, testContext)
	if wrapped.Code != appconstants.ErrCodeFileNotFound {
		t.Errorf("Wrap() code = %v, want %v", wrapped.Code, appconstants.ErrCodeFileNotFound)
	}
	if wrapped.Context != testContext {
		t.Errorf("Wrap() context = %v, want %v", wrapped.Context, testContext)
	}
	if wrapped.Err != originalErr {
		t.Errorf("Wrap() err = %v, want %v", wrapped.Err, originalErr)
	}

	// Test wrapping nil error
	nilWrapped := Wrap(nil, appconstants.ErrCodeFileNotFound, testContext)
	if nilWrapped != nil {
		t.Error("Wrap(nil) should return nil")
	}

	// Test wrapping already contextual error
	contextualErr := &ContextualError{
		Code:    appconstants.ErrCodeUnknown,
		Err:     originalErr,
		Context: "",
	}
	rewrapped := Wrap(contextualErr, appconstants.ErrCodeFileNotFound, "new context")
	if rewrapped.Code != appconstants.ErrCodeFileNotFound {
		t.Error("Wrap() should update code if it was appconstants.ErrCodeUnknown")
	}
	if rewrapped.Context != "new context" {
		t.Error("Wrap() should update context if it was empty")
	}
}

func TestContextualErrorWithMethods(t *testing.T) {
	t.Parallel()

	err := New(appconstants.ErrCodeFileNotFound, "test error")

	// Test WithSuggestions
	err = err.WithSuggestions("suggestion 1", "suggestion 2")
	if len(err.Suggestions) != 2 {
		t.Errorf("WithSuggestions() length = %d, want 2", len(err.Suggestions))
	}
	if err.Suggestions[0] != "suggestion 1" {
		t.Errorf("WithSuggestions()[0] = %s, want 'suggestion 1'", err.Suggestions[0])
	}

	// Test WithDetails
	details := map[string]string{"key1": "value1", "key2": "value2"}
	err = err.WithDetails(details)
	if len(err.Details) != 2 {
		t.Errorf("WithDetails() length = %d, want 2", len(err.Details))
	}
	if err.Details["key1"] != "value1" {
		t.Errorf("WithDetails()['key1'] = %s, want 'value1'", err.Details["key1"])
	}

	// Test WithHelpURL
	url := "https://example.com/help"
	err = err.WithHelpURL(url)
	if err.HelpURL != url {
		t.Errorf("WithHelpURL() = %s, want %s", err.HelpURL, url)
	}
}

func TestGetHelpURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		code     appconstants.ErrorCode
		contains string
	}{
		{appconstants.ErrCodeFileNotFound, "#file-not-found"},
		{appconstants.ErrCodeInvalidYAML, "#invalid-yaml"},
		{appconstants.ErrCodeGitHubAPI, "#github-api-errors"},
		{appconstants.ErrCodeUnknown, "troubleshooting.md"}, // Should return base URL
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			t.Parallel()

			url := GetHelpURL(tt.code)
			if !strings.Contains(url, tt.contains) {
				t.Errorf("GetHelpURL(%s) = %s, should contain %s", tt.code, url, tt.contains)
			}
		})
	}
}
