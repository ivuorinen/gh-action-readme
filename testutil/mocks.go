package testutil

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

// CapturedOutput captures all output for testing.
// Implements CompleteOutput interface (all focused interfaces).
type CapturedOutput struct {
	BoldMessages               []string
	SuccessMessages            []string
	ErrorMessages              []string
	WarningMessages            []string
	InfoMessages               []string
	PrintfMessages             []string
	ProgressMessages           []string
	ErrorWithSuggestionsCalls  []string
	ErrorWithContextCalls      []string
	ErrorWithSimpleFixCalls    []string
	FormatContextualErrorCalls []string
	QuietMode                  bool
}

// Bold appends a bold-formatted message to the captured output.
func (c *CapturedOutput) Bold(format string, args ...any) {
	c.recordMessage(&c.BoldMessages, format, args...)
}

// Success appends a success message to the captured output.
func (c *CapturedOutput) Success(format string, args ...any) {
	c.recordMessage(&c.SuccessMessages, format, args...)
}

// Error appends an error message to the captured output.
func (c *CapturedOutput) Error(format string, args ...any) {
	c.recordMessage(&c.ErrorMessages, format, args...)
}

// Warning appends a warning message to the captured output.
func (c *CapturedOutput) Warning(format string, args ...any) {
	c.recordMessage(&c.WarningMessages, format, args...)
}

// Info appends an info message to the captured output.
func (c *CapturedOutput) Info(format string, args ...any) {
	c.recordMessage(&c.InfoMessages, format, args...)
}

// Printf appends a printf-formatted message to the captured output.
func (c *CapturedOutput) Printf(format string, args ...any) {
	c.recordMessage(&c.PrintfMessages, format, args...)
}

// Fprintf appends a fprintf-formatted message to the captured output.
func (c *CapturedOutput) Fprintf(_ *os.File, format string, args ...any) {
	c.recordMessage(&c.PrintfMessages, format, args...)
}

// ErrorWithSuggestions captures error reporting with suggestions.
func (c *CapturedOutput) ErrorWithSuggestions(err error) {
	if err != nil {
		c.ErrorWithSuggestionsCalls = append(c.ErrorWithSuggestionsCalls, err.Error())
	}
}

// ErrorWithContext captures contextual error reporting.
func (c *CapturedOutput) ErrorWithContext(_ appconstants.ErrorCode, message string, _ map[string]string) {
	c.ErrorWithContextCalls = append(c.ErrorWithContextCalls, message)
}

// ErrorWithSimpleFix captures error reporting with a simple fix suggestion.
func (c *CapturedOutput) ErrorWithSimpleFix(message, suggestion string) {
	c.ErrorWithSimpleFixCalls = append(c.ErrorWithSimpleFixCalls, message+": "+suggestion)
}

// FormatContextualError captures and returns formatted contextual error.
func (c *CapturedOutput) FormatContextualError(err error) string {
	if err != nil {
		formatted := err.Error()
		c.FormatContextualErrorCalls = append(c.FormatContextualErrorCalls, formatted)

		return formatted
	}

	return ""
}

// Progress captures progress reporting messages.
func (c *CapturedOutput) Progress(format string, args ...any) {
	c.recordMessage(&c.ProgressMessages, format, args...)
}

// IsQuiet returns whether the output is in quiet mode.
func (c *CapturedOutput) IsQuiet() bool {
	return c.QuietMode
}

// AllMessages consolidates all message slices into a single slice.
func (c *CapturedOutput) AllMessages() []string {
	messages := make([]string, 0,
		len(c.BoldMessages)+len(c.SuccessMessages)+
			len(c.InfoMessages)+len(c.ErrorMessages)+
			len(c.WarningMessages)+len(c.PrintfMessages)+
			len(c.ProgressMessages))
	messages = append(messages, c.BoldMessages...)
	messages = append(messages, c.SuccessMessages...)
	messages = append(messages, c.InfoMessages...)
	messages = append(messages, c.ErrorMessages...)
	messages = append(messages, c.WarningMessages...)
	messages = append(messages, c.PrintfMessages...)
	messages = append(messages, c.ProgressMessages...)

	return messages
}

// ContainsMessage checks if any message in the consolidated list contains the needle.
func (c *CapturedOutput) ContainsMessage(needle string) bool {
	return ContainsInSlice(c.AllMessages(), needle)
}

// ContainsError checks if any error message contains the needle.
func (c *CapturedOutput) ContainsError(needle string) bool {
	return ContainsInSlice(c.ErrorMessages, needle)
}

// ContainsWarning checks if any warning message contains the needle.
func (c *CapturedOutput) ContainsWarning(needle string) bool {
	return ContainsInSlice(c.WarningMessages, needle)
}

// recordMessage is a helper that appends a formatted message to the specified message slice.
// This reduces duplication across Bold, Success, Error, Warning, Info, Printf, and Progress methods.
func (c *CapturedOutput) recordMessage(messageSlice *[]string, format string, args ...any) {
	*messageSlice = append(*messageSlice, fmt.Sprintf(format, args...))
}

// ContainsInSlice checks if any string in the slice contains the substring.
func ContainsInSlice(slice []string, substring string) bool {
	for _, s := range slice {
		if strings.Contains(s, substring) {
			return true
		}
	}

	return false
}

// AssertSliceLength asserts that a slice has the expected length.
func AssertSliceLength(t *testing.T, slice []string, expected int, label string) {
	t.Helper()

	if len(slice) != expected {
		t.Errorf("%s length = %d, want %d", label, len(slice), expected)
	}
}
