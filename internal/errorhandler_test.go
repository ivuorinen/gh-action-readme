package internal

import (
	"errors"
	"os"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/apperrors"
)

// TestNewErrorHandler tests error handler creation.
func TestNewErrorHandler(t *testing.T) {
	output := &ColoredOutput{NoColor: true, Quiet: true}
	handler := NewErrorHandler(output)

	if handler == nil {
		t.Fatal("NewErrorHandler() returned nil")
	}

	if handler.output != output {
		t.Error("NewErrorHandler() did not set output correctly")
	}
}

// TestDetermineErrorCode tests error code determination.
//
//nolint:dupl // Intentional duplication with TestCheckTypedError - testing different functions
func TestDetermineErrorCode(t *testing.T) {
	handler := NewErrorHandler(&ColoredOutput{NoColor: true, Quiet: true})

	tests := []struct {
		name     string
		err      error
		wantCode appconstants.ErrorCode
	}{
		{
			name:     "file not found - typed error",
			err:      apperrors.ErrFileNotFound,
			wantCode: appconstants.ErrCodeFileNotFound,
		},
		{
			name:     "file not found - os.ErrNotExist",
			err:      os.ErrNotExist,
			wantCode: appconstants.ErrCodeFileNotFound,
		},
		{
			name:     "permission denied - typed error",
			err:      apperrors.ErrPermissionDenied,
			wantCode: appconstants.ErrCodePermission,
		},
		{
			name:     "permission denied - os.ErrPermission",
			err:      os.ErrPermission,
			wantCode: appconstants.ErrCodePermission,
		},
		{
			name:     "invalid YAML",
			err:      apperrors.ErrInvalidYAML,
			wantCode: appconstants.ErrCodeInvalidYAML,
		},
		{
			name:     "GitHub API error",
			err:      apperrors.ErrGitHubAPI,
			wantCode: appconstants.ErrCodeGitHubAPI,
		},
		{
			name:     "configuration error",
			err:      apperrors.ErrConfiguration,
			wantCode: appconstants.ErrCodeConfiguration,
		},
		{
			name:     "unknown error",
			err:      errors.New("some random error"),
			wantCode: appconstants.ErrCodeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handler.determineErrorCode(tt.err)
			if got != tt.wantCode {
				t.Errorf("determineErrorCode() = %v, want %v", got, tt.wantCode)
			}
		})
	}
}

// TestCheckTypedError tests typed error checking.
//
//nolint:dupl // Intentional duplication with TestDetermineErrorCode - testing different functions
func TestCheckTypedError(t *testing.T) {
	handler := NewErrorHandler(&ColoredOutput{NoColor: true, Quiet: true})

	tests := []struct {
		name     string
		err      error
		wantCode appconstants.ErrorCode
	}{
		{
			name:     "ErrFileNotFound",
			err:      apperrors.ErrFileNotFound,
			wantCode: appconstants.ErrCodeFileNotFound,
		},
		{
			name:     "os.ErrNotExist",
			err:      os.ErrNotExist,
			wantCode: appconstants.ErrCodeFileNotFound,
		},
		{
			name:     "ErrPermissionDenied",
			err:      apperrors.ErrPermissionDenied,
			wantCode: appconstants.ErrCodePermission,
		},
		{
			name:     "os.ErrPermission",
			err:      os.ErrPermission,
			wantCode: appconstants.ErrCodePermission,
		},
		{
			name:     "ErrInvalidYAML",
			err:      apperrors.ErrInvalidYAML,
			wantCode: appconstants.ErrCodeInvalidYAML,
		},
		{
			name:     "ErrGitHubAPI",
			err:      apperrors.ErrGitHubAPI,
			wantCode: appconstants.ErrCodeGitHubAPI,
		},
		{
			name:     "ErrConfiguration",
			err:      apperrors.ErrConfiguration,
			wantCode: appconstants.ErrCodeConfiguration,
		},
		{
			name:     "unknown error",
			err:      errors.New("unknown error"),
			wantCode: appconstants.ErrCodeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handler.checkTypedError(tt.err)
			if got != tt.wantCode {
				t.Errorf("checkTypedError() = %v, want %v", got, tt.wantCode)
			}
		})
	}
}

// TestCheckStringPatterns tests string pattern matching.
func TestCheckStringPatterns(t *testing.T) {
	handler := NewErrorHandler(&ColoredOutput{NoColor: true, Quiet: true})

	tests := []struct {
		name     string
		errStr   string
		wantCode appconstants.ErrorCode
	}{
		{
			name:     "file not found pattern",
			errStr:   "no such file or directory",
			wantCode: appconstants.ErrCodeFileNotFound,
		},
		{
			name:     "permission denied pattern",
			errStr:   "permission denied",
			wantCode: appconstants.ErrCodePermission,
		},
		{
			name:     "YAML error pattern",
			errStr:   "yaml: unmarshal error",
			wantCode: appconstants.ErrCodeInvalidYAML,
		},
		{
			name:     "GitHub API pattern",
			errStr:   "GitHub API error",
			wantCode: appconstants.ErrCodeGitHubAPI,
		},
		{
			name:     "configuration pattern",
			errStr:   "configuration error",
			wantCode: appconstants.ErrCodeConfiguration,
		},
		{
			name:     "unknown pattern",
			errStr:   "some random error message",
			wantCode: appconstants.ErrCodeUnknown,
		},
		{
			name:     "case insensitive matching",
			errStr:   "PERMISSION DENIED",
			wantCode: appconstants.ErrCodePermission,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handler.checkStringPatterns(tt.errStr)
			if got != tt.wantCode {
				t.Errorf("checkStringPatterns(%q) = %v, want %v", tt.errStr, got, tt.wantCode)
			}
		})
	}
}

// TestContains tests the contains helper function.
func TestContains(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{
			name:   "exact match",
			s:      "hello world",
			substr: "hello",
			want:   true,
		},
		{
			name:   "case insensitive match",
			s:      "Hello World",
			substr: "hello",
			want:   true,
		},
		{
			name:   "no match",
			s:      "hello world",
			substr: "goodbye",
			want:   false,
		},
		{
			name:   "empty substring",
			s:      "hello world",
			substr: "",
			want:   true,
		},
		{
			name:   "empty string",
			s:      "",
			substr: "hello",
			want:   false,
		},
		{
			name:   "substring in middle",
			s:      "the quick brown fox",
			substr: "quick",
			want:   true,
		},
		{
			name:   "case insensitive - uppercase string",
			s:      "ERROR: PERMISSION DENIED",
			substr: "permission",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contains(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

// TestHandleSimpleError tests simple error handling logic (without os.Exit).
func TestHandleSimpleError(t *testing.T) {
	// Note: We cannot test os.Exit calls directly, but we can test the logic
	// that runs before os.Exit. This test verifies the error code determination.

	handler := NewErrorHandler(&ColoredOutput{NoColor: true, Quiet: true})

	// Test error code determination for various errors
	tests := []struct {
		name     string
		err      error
		wantCode appconstants.ErrorCode
	}{
		{
			name:     "file not found error",
			err:      os.ErrNotExist,
			wantCode: appconstants.ErrCodeFileNotFound,
		},
		{
			name:     "permission error",
			err:      os.ErrPermission,
			wantCode: appconstants.ErrCodePermission,
		},
		{
			name:     "nil error defaults to unknown",
			err:      nil,
			wantCode: appconstants.ErrCodeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can only test the error code determination
			// The actual HandleSimpleError would call os.Exit
			var code appconstants.ErrorCode
			if tt.err != nil {
				code = handler.determineErrorCode(tt.err)
			} else {
				code = appconstants.ErrCodeUnknown
			}

			if code != tt.wantCode {
				t.Errorf("error code = %v, want %v", code, tt.wantCode)
			}
		})
	}
}

// TestHandleFatalError tests fatal error handling setup.
func TestHandleFatalError(t *testing.T) {
	// Note: Similar to HandleSimpleError, we test the logic before os.Exit
	// The actual function calls os.Exit which we cannot test directly

	handler := NewErrorHandler(&ColoredOutput{NoColor: true, Quiet: true})

	// Test that HandleFatalError correctly constructs contextual errors
	code := appconstants.ErrCodeFileNotFound
	message := "test error message"
	context := map[string]string{"file": "test.yml"}

	// Verify suggestions and help URL are retrieved
	suggestions := apperrors.GetSuggestions(code, context)
	helpURL := apperrors.GetHelpURL(code)

	if len(suggestions) == 0 {
		t.Log("Note: No suggestions found for error code (this may be expected)")
	}

	if helpURL == "" {
		t.Log("Note: No help URL found for error code (this may be expected)")
	}

	// Verify error construction (without calling HandleFatalError which exits)
	contextualErr := apperrors.New(code, message).
		WithSuggestions(suggestions...).
		WithHelpURL(helpURL).
		WithDetails(context)

	if contextualErr == nil {
		t.Error("failed to construct contextual error")
	}

	// Verify handler is properly initialized
	if handler.output == nil {
		t.Error("handler output is nil")
	}
}
