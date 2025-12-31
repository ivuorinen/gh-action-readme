// Package internal provides centralized error handling utilities.
package internal

import (
	"os"
	"strings"

	"github.com/ivuorinen/gh-action-readme/internal/apperrors"
)

// Error detection constants for automatic error code determination.
const (
	// File system error patterns.
	errorPatternFileNotFound = "no such file or directory"
	errorPatternPermission   = "permission denied"

	// Content format error patterns.
	errorPatternYAML = "yaml"

	// Service-specific error patterns.
	errorPatternGitHub = "github"
	errorPatternConfig = "config"

	// Exit code constants.
	exitCodeError = 1
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
	os.Exit(exitCodeError)
}

// HandleFatalError handles fatal errors with contextual information.
func (eh *ErrorHandler) HandleFatalError(code apperrors.ErrorCode, message string, context map[string]string) {
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
	code := apperrors.ErrCodeUnknown
	context := make(map[string]string)

	// Try to determine appropriate error code based on error content
	if err != nil {
		context[ContextKeyError] = err.Error()
		code = eh.determineErrorCode(err)
	}

	eh.HandleFatalError(code, message, context)
}

// determineErrorCode attempts to determine appropriate error code from error content.
func (eh *ErrorHandler) determineErrorCode(err error) apperrors.ErrorCode {
	errStr := err.Error()

	switch {
	case contains(errStr, errorPatternFileNotFound):
		return apperrors.ErrCodeFileNotFound
	case contains(errStr, errorPatternPermission):
		return apperrors.ErrCodePermission
	case contains(errStr, errorPatternYAML):
		return apperrors.ErrCodeInvalidYAML
	case contains(errStr, errorPatternGitHub):
		return apperrors.ErrCodeGitHubAPI
	case contains(errStr, errorPatternConfig):
		return apperrors.ErrCodeConfiguration
	default:
		return apperrors.ErrCodeUnknown
	}
}

// contains checks if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
