// Package apperrors provides enhanced error types with contextual information and suggestions.
package apperrors

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

// Sentinel errors for typed error checking.
var (
	// ErrFileNotFound indicates a file was not found.
	ErrFileNotFound = errors.New("file not found")
	// ErrPermissionDenied indicates a permission error.
	ErrPermissionDenied = errors.New("permission denied")
	// ErrInvalidYAML indicates YAML parsing failed.
	ErrInvalidYAML = errors.New("invalid YAML")
	// ErrGitHubAPI indicates a GitHub API error.
	ErrGitHubAPI = errors.New("GitHub API error")
	// ErrConfiguration indicates a configuration error.
	ErrConfiguration = errors.New("configuration error")
)

// ContextualError provides enhanced error information with actionable suggestions.
type ContextualError struct {
	Code        appconstants.ErrorCode
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
		fmt.Fprintf(&b, "%s: %v", ce.Context, ce.Err)
	} else {
		b.WriteString(ce.Err.Error())
	}

	// Add error code for reference
	fmt.Fprintf(&b, " [%s]", ce.Code)

	// Add details if available (sorted for deterministic, stable output —
	// Go map iteration order is randomized).
	if len(ce.Details) > 0 {
		b.WriteString("\n\nDetails:")
		for _, key := range slices.Sorted(maps.Keys(ce.Details)) {
			fmt.Fprintf(&b, "\n  %s: %s", key, ce.Details[key])
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
func New(code appconstants.ErrorCode, message string) *ContextualError {
	return &ContextualError{
		Code: code,
		Err:  errors.New(message),
	}
}

// Wrap wraps an existing error with contextual information.
func Wrap(err error, code appconstants.ErrorCode, context string) *ContextualError {
	if err == nil {
		return nil
	}

	// If already a ContextualError, preserve existing info by creating a copy
	if ce, ok := err.(*ContextualError); ok {
		// Create a copy to avoid mutating the original
		errCopy := &ContextualError{
			Code:    ce.Code,
			Err:     ce.Err,
			Context: ce.Context,
			// Clone the slice so a later WithSuggestions append on the copy cannot
			// write into the original's backing array (the comment above promises
			// the original is not mutated).
			Suggestions: slices.Clone(ce.Suggestions),
			HelpURL:     ce.HelpURL,
			Details:     maps.Clone(ce.Details),
		}

		// Only update if not already set
		if errCopy.Code == appconstants.ErrCodeUnknown {
			errCopy.Code = code
		}
		if errCopy.Context == "" {
			errCopy.Context = context
		}

		return errCopy
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
	maps.Copy(ce.Details, details)

	return ce
}

// WithHelpURL adds a help URL to a ContextualError.
func (ce *ContextualError) WithHelpURL(url string) *ContextualError {
	ce.HelpURL = url

	return ce
}

// GetHelpURL returns a help URL for the given error code.
func GetHelpURL(code appconstants.ErrorCode) string {
	baseURL := "https://github.com/ivuorinen/gh-action-readme/blob/main/docs/troubleshooting.md"

	anchors := map[appconstants.ErrorCode]string{
		appconstants.ErrCodeFileNotFound:       "#file-not-found",
		appconstants.ErrCodePermission:         "#permission-denied",
		appconstants.ErrCodeInvalidYAML:        "#invalid-yaml",
		appconstants.ErrCodeInvalidAction:      "#invalid-action-file",
		appconstants.ErrCodeNoActionFiles:      "#no-action-files",
		appconstants.ErrCodeGitHubAPI:          "#github-api-errors",
		appconstants.ErrCodeGitHubRateLimit:    "#rate-limit-exceeded",
		appconstants.ErrCodeGitHubAuth:         "#authentication-errors",
		appconstants.ErrCodeConfiguration:      "#configuration-errors",
		appconstants.ErrCodeValidation:         "#validation-errors",
		appconstants.ErrCodeTemplateRender:     "#template-errors",
		appconstants.ErrCodeFileWrite:          "#file-write-errors",
		appconstants.ErrCodeDependencyAnalysis: "#dependency-analysis",
		appconstants.ErrCodeCacheAccess:        "#cache-errors",
	}

	if anchor, ok := anchors[code]; ok {
		return baseURL + anchor
	}

	return baseURL
}
