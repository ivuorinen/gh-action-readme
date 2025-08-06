package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestContextualError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *ContextualError
		contains []string
	}{
		{
			name: "basic error",
			err: &ContextualError{
				Code: ErrCodeFileNotFound,
				Err:  errors.New("file not found"),
			},
			contains: []string{"file not found", "[FILE_NOT_FOUND]"},
		},
		{
			name: "error with context",
			err: &ContextualError{
				Code:    ErrCodeInvalidYAML,
				Err:     errors.New("invalid syntax"),
				Context: "parsing action.yml",
			},
			contains: []string{"parsing action.yml: invalid syntax", "[INVALID_YAML]"},
		},
		{
			name: "error with suggestions",
			err: &ContextualError{
				Code: ErrCodeNoActionFiles,
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
				Code: ErrCodeConfiguration,
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
				Code:    ErrCodeGitHubAPI,
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
				Code:    ErrCodeValidation,
				Err:     errors.New("validation failed"),
				Context: "validating action.yml",
				Details: map[string]string{"file": "action.yml"},
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

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf(
						"Error() result missing expected content:\nExpected to contain: %q\nActual result:\n%s",
						expected,
						result,
					)
				}
			}
		})
	}
}

func TestContextualError_Unwrap(t *testing.T) {
	t.Parallel()

	originalErr := errors.New("original error")
	contextualErr := &ContextualError{
		Code: ErrCodeFileNotFound,
		Err:  originalErr,
	}

	if unwrapped := contextualErr.Unwrap(); unwrapped != originalErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, originalErr)
	}
}

func TestContextualError_Is(t *testing.T) {
	t.Parallel()

	originalErr := errors.New("original error")
	contextualErr := &ContextualError{
		Code: ErrCodeFileNotFound,
		Err:  originalErr,
	}

	// Test Is with same error code
	sameCodeErr := &ContextualError{Code: ErrCodeFileNotFound}
	if !contextualErr.Is(sameCodeErr) {
		t.Error("Is() should return true for same error code")
	}

	// Test Is with different error code
	differentCodeErr := &ContextualError{Code: ErrCodeInvalidYAML}
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

	err := New(ErrCodeFileNotFound, "test message")

	if err.Code != ErrCodeFileNotFound {
		t.Errorf("New() code = %v, want %v", err.Code, ErrCodeFileNotFound)
	}

	if err.Err.Error() != "test message" {
		t.Errorf("New() message = %v, want %v", err.Err.Error(), "test message")
	}
}

func TestWrap(t *testing.T) {
	t.Parallel()

	originalErr := errors.New("original error")

	// Test wrapping normal error
	wrapped := Wrap(originalErr, ErrCodeFileNotFound, "test context")
	if wrapped.Code != ErrCodeFileNotFound {
		t.Errorf("Wrap() code = %v, want %v", wrapped.Code, ErrCodeFileNotFound)
	}
	if wrapped.Context != "test context" {
		t.Errorf("Wrap() context = %v, want %v", wrapped.Context, "test context")
	}
	if wrapped.Err != originalErr {
		t.Errorf("Wrap() err = %v, want %v", wrapped.Err, originalErr)
	}

	// Test wrapping nil error
	nilWrapped := Wrap(nil, ErrCodeFileNotFound, "test context")
	if nilWrapped != nil {
		t.Error("Wrap(nil) should return nil")
	}

	// Test wrapping already contextual error
	contextualErr := &ContextualError{
		Code:    ErrCodeUnknown,
		Err:     originalErr,
		Context: "",
	}
	rewrapped := Wrap(contextualErr, ErrCodeFileNotFound, "new context")
	if rewrapped.Code != ErrCodeFileNotFound {
		t.Error("Wrap() should update code if it was ErrCodeUnknown")
	}
	if rewrapped.Context != "new context" {
		t.Error("Wrap() should update context if it was empty")
	}
}

func TestContextualError_WithMethods(t *testing.T) {
	t.Parallel()

	err := New(ErrCodeFileNotFound, "test error")

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
		code     ErrorCode
		contains string
	}{
		{ErrCodeFileNotFound, "#file-not-found"},
		{ErrCodeInvalidYAML, "#invalid-yaml"},
		{ErrCodeGitHubAPI, "#github-api-errors"},
		{ErrCodeUnknown, "troubleshooting.md"}, // Should return base URL
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
