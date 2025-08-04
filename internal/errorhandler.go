// Package internal provides centralized error handling utilities.
package internal

import (
	"os"

	"github.com/ivuorinen/gh-action-readme/internal/errors"
)

// ErrorHandler provides centralized error handling and exit management.
type ErrorHandler struct {
	output *ColoredOutput
}

// NewErrorHandler creates a new error handler.
func NewErrorHandler(output *ColoredOutput) *ErrorHandler {
	return &ErrorHandler{
		output: output,
	}
}

// HandleError handles contextual errors and exits with appropriate code.
func (eh *ErrorHandler) HandleError(err *errors.ContextualError) {
	eh.output.ErrorWithSuggestions(err)
	os.Exit(1)
}

// HandleFatalError handles fatal errors with contextual information.
func (eh *ErrorHandler) HandleFatalError(code errors.ErrorCode, message string, context map[string]string) {
	suggestions := errors.GetSuggestions(code, context)
	helpURL := errors.GetHelpURL(code)

	contextualErr := errors.New(code, message).
		WithSuggestions(suggestions...).
		WithHelpURL(helpURL)

	if len(context) > 0 {
		contextualErr = contextualErr.WithDetails(context)
	}

	eh.HandleError(contextualErr)
}

// HandleSimpleError handles simple errors with automatic context detection.
func (eh *ErrorHandler) HandleSimpleError(message string, err error) {
	code := errors.ErrCodeUnknown
	context := make(map[string]string)

	// Try to determine appropriate error code based on error content
	if err != nil {
		context["error"] = err.Error()
		code = eh.determineErrorCode(err)
	}

	eh.HandleFatalError(code, message, context)
}

// determineErrorCode attempts to determine appropriate error code from error content.
func (eh *ErrorHandler) determineErrorCode(err error) errors.ErrorCode {
	errStr := err.Error()

	switch {
	case contains(errStr, "no such file or directory"):
		return errors.ErrCodeFileNotFound
	case contains(errStr, "permission denied"):
		return errors.ErrCodePermission
	case contains(errStr, "yaml"):
		return errors.ErrCodeInvalidYAML
	case contains(errStr, "github"):
		return errors.ErrCodeGitHubAPI
	case contains(errStr, "config"):
		return errors.ErrCodeConfiguration
	default:
		return errors.ErrCodeUnknown
	}
}

// contains checks if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	// Simple implementation - could use strings.Contains with strings.ToLower
	// but avoiding extra imports for now
	sLen := len(s)
	substrLen := len(substr)

	if substrLen > sLen {
		return false
	}

	for i := 0; i <= sLen-substrLen; i++ {
		match := true
		for j := 0; j < substrLen; j++ {
			if toLower(s[i+j]) != toLower(substr[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}

	return false
}

// toLower converts a byte to lowercase.
func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}
