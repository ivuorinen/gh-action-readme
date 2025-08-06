// Package errors provides enhanced error types with contextual information and suggestions.
package errors

import (
	"errors"
	"fmt"
	"strings"
)

// ErrorCode represents a category of error for providing specific help.
type ErrorCode string

// Error code constants for categorizing errors.
const (
	ErrCodeFileNotFound       ErrorCode = "FILE_NOT_FOUND"
	ErrCodePermission         ErrorCode = "PERMISSION_DENIED"
	ErrCodeInvalidYAML        ErrorCode = "INVALID_YAML"
	ErrCodeInvalidAction      ErrorCode = "INVALID_ACTION"
	ErrCodeNoActionFiles      ErrorCode = "NO_ACTION_FILES"
	ErrCodeGitHubAPI          ErrorCode = "GITHUB_API_ERROR"
	ErrCodeGitHubRateLimit    ErrorCode = "GITHUB_RATE_LIMIT"
	ErrCodeGitHubAuth         ErrorCode = "GITHUB_AUTH_ERROR"
	ErrCodeConfiguration      ErrorCode = "CONFIG_ERROR"
	ErrCodeValidation         ErrorCode = "VALIDATION_ERROR"
	ErrCodeTemplateRender     ErrorCode = "TEMPLATE_ERROR"
	ErrCodeFileWrite          ErrorCode = "FILE_WRITE_ERROR"
	ErrCodeDependencyAnalysis ErrorCode = "DEPENDENCY_ERROR"
	ErrCodeCacheAccess        ErrorCode = "CACHE_ERROR"
	ErrCodeUnknown            ErrorCode = "UNKNOWN_ERROR"
)

// ContextualError provides enhanced error information with actionable suggestions.
type ContextualError struct {
	Code        ErrorCode
	Err         error
	Context     string
	Suggestions []string
	HelpURL     string
	Details     map[string]string
}

// Error implements the error interface.
func (ce *ContextualError) Error() string {
	var b strings.Builder

	// Primary error message
	if ce.Context != "" {
		b.WriteString(fmt.Sprintf("%s: %v", ce.Context, ce.Err))
	} else {
		b.WriteString(ce.Err.Error())
	}

	// Add error code for reference
	b.WriteString(fmt.Sprintf(" [%s]", ce.Code))

	// Add details if available
	if len(ce.Details) > 0 {
		b.WriteString("\n\nDetails:")
		for key, value := range ce.Details {
			b.WriteString(fmt.Sprintf("\n  %s: %s", key, value))
		}
	}

	// Add suggestions
	if len(ce.Suggestions) > 0 {
		b.WriteString("\n\nSuggestions:")
		for _, suggestion := range ce.Suggestions {
			b.WriteString("\n  • " + suggestion)
		}
	}

	// Add help URL
	if ce.HelpURL != "" {
		b.WriteString("\n\nFor more help: " + ce.HelpURL)
	}

	return b.String()
}

// Unwrap returns the wrapped error.
func (ce *ContextualError) Unwrap() error {
	return ce.Err
}

// Is implements errors.Is support.
func (ce *ContextualError) Is(target error) bool {
	if target == nil {
		return false
	}

	// Check if target is also a ContextualError with same code
	if targetCE, ok := target.(*ContextualError); ok {
		return ce.Code == targetCE.Code
	}

	// Check wrapped error
	return errors.Is(ce.Err, target)
}

// New creates a new ContextualError with the given code and message.
func New(code ErrorCode, message string) *ContextualError {
	return &ContextualError{
		Code: code,
		Err:  errors.New(message),
	}
}

// Wrap wraps an existing error with contextual information.
func Wrap(err error, code ErrorCode, context string) *ContextualError {
	if err == nil {
		return nil
	}

	// If already a ContextualError, preserve existing info
	if ce, ok := err.(*ContextualError); ok {
		// Only update if not already set
		if ce.Code == ErrCodeUnknown {
			ce.Code = code
		}
		if ce.Context == "" {
			ce.Context = context
		}

		return ce
	}

	return &ContextualError{
		Code:    code,
		Err:     err,
		Context: context,
	}
}

// WithSuggestions adds suggestions to a ContextualError.
func (ce *ContextualError) WithSuggestions(suggestions ...string) *ContextualError {
	ce.Suggestions = append(ce.Suggestions, suggestions...)

	return ce
}

// WithDetails adds detail key-value pairs to a ContextualError.
func (ce *ContextualError) WithDetails(details map[string]string) *ContextualError {
	if ce.Details == nil {
		ce.Details = make(map[string]string)
	}
	for k, v := range details {
		ce.Details[k] = v
	}

	return ce
}

// WithHelpURL adds a help URL to a ContextualError.
func (ce *ContextualError) WithHelpURL(url string) *ContextualError {
	ce.HelpURL = url

	return ce
}

// GetHelpURL returns a help URL for the given error code.
func GetHelpURL(code ErrorCode) string {
	baseURL := "https://github.com/ivuorinen/gh-action-readme/blob/main/docs/troubleshooting.md"

	anchors := map[ErrorCode]string{
		ErrCodeFileNotFound:       "#file-not-found",
		ErrCodePermission:         "#permission-denied",
		ErrCodeInvalidYAML:        "#invalid-yaml",
		ErrCodeInvalidAction:      "#invalid-action-file",
		ErrCodeNoActionFiles:      "#no-action-files",
		ErrCodeGitHubAPI:          "#github-api-errors",
		ErrCodeGitHubRateLimit:    "#rate-limit-exceeded",
		ErrCodeGitHubAuth:         "#authentication-errors",
		ErrCodeConfiguration:      "#configuration-errors",
		ErrCodeValidation:         "#validation-errors",
		ErrCodeTemplateRender:     "#template-errors",
		ErrCodeFileWrite:          "#file-write-errors",
		ErrCodeDependencyAnalysis: "#dependency-analysis",
		ErrCodeCacheAccess:        "#cache-errors",
	}

	if anchor, ok := anchors[code]; ok {
		return baseURL + anchor
	}

	return baseURL
}
