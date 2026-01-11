package internal

import (
	"errors"
	"os"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/apperrors"
	"github.com/ivuorinen/gh-action-readme/testutil"
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
			err:      errors.New(testutil.UnknownErrorMsg),
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
			s:      testutil.HelloWorldStr,
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
			s:      testutil.HelloWorldStr,
			substr: "goodbye",
			want:   false,
		},
		{
			name:   "empty substring",
			s:      testutil.HelloWorldStr,
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

// NOTE: HandleSimpleError testing is covered by TestDetermineErrorCode
// since HandleSimpleError calls determineErrorCode and then os.Exit().
// Testing os.Exit() directly is not practical in unit tests.

// TestFatalErrorComponents tests the components used in fatal error handling.
// NOTE: We cannot test HandleFatalError directly as it calls os.Exit().
// This test verifies that error construction components work correctly.
func TestFatalErrorComponents(t *testing.T) {
	// Test the logic that HandleFatalError uses before calling os.Exit

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
