// Package internal provides centralized error handling utilities.
package internal

import (
	"os"
	"strings"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/apperrors"
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
func (eh *ErrorHandler) HandleError(err *apperrors.ContextualError) {
	eh.output.ErrorWithSuggestions(err)
	os.Exit(appconstants.ExitCodeError)
}

// HandleFatalError handles fatal errors with contextual information.
func (eh *ErrorHandler) HandleFatalError(code appconstants.ErrorCode, message string, context map[string]string) {
	suggestions := apperrors.GetSuggestions(code, context)
	helpURL := apperrors.GetHelpURL(code)

	contextualErr := apperrors.New(code, message).
		WithSuggestions(suggestions...).
		WithHelpURL(helpURL)

	if len(context) > 0 {
		contextualErr = contextualErr.WithDetails(context)
	}

	eh.HandleError(contextualErr)
}

// HandleSimpleError handles simple errors with automatic context detection.
func (eh *ErrorHandler) HandleSimpleError(message string, err error) {
	code := appconstants.ErrCodeUnknown
	context := make(map[string]string)

	// Try to determine appropriate error code based on error content
	if err != nil {
		context[appconstants.ContextKeyError] = err.Error()
		code = eh.determineErrorCode(err)
	}

	eh.HandleFatalError(code, message, context)
}

// determineErrorCode attempts to determine appropriate error code from error content.
func (eh *ErrorHandler) determineErrorCode(err error) appconstants.ErrorCode {
	errStr := err.Error()

	switch {
	case contains(errStr, appconstants.ErrorPatternFileNotFound):
		return appconstants.ErrCodeFileNotFound
	case contains(errStr, appconstants.ErrorPatternPermission):
		return appconstants.ErrCodePermission
	case contains(errStr, appconstants.ErrorPatternYAML):
		return appconstants.ErrCodeInvalidYAML
	case contains(errStr, appconstants.ErrorPatternGitHub):
		return appconstants.ErrCodeGitHubAPI
	case contains(errStr, appconstants.ErrorPatternConfig):
		return appconstants.ErrCodeConfiguration
	default:
		return appconstants.ErrCodeUnknown
	}
}

// contains checks if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
