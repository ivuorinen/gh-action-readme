package testutil

import (
	"fmt"
	"os"
	"sync"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

// MessageLoggerMock tracks message logger calls for testing.
type MessageLoggerMock struct {
	mu           sync.Mutex
	InfoCalls    []string
	SuccessCalls []string
	WarningCalls []string
	BoldCalls    []string
	PrintfCalls  []string
	FprintfCalls []string
}

// Info captures info message calls.
func (m *MessageLoggerMock) Info(format string, args ...any) {
	m.recordMessage(&m.InfoCalls, format, args...)
}

// Success captures success message calls.
func (m *MessageLoggerMock) Success(format string, args ...any) {
	m.recordMessage(&m.SuccessCalls, format, args...)
}

// Warning captures warning message calls.
func (m *MessageLoggerMock) Warning(format string, args ...any) {
	m.recordMessage(&m.WarningCalls, format, args...)
}

// Bold captures bold message calls.
func (m *MessageLoggerMock) Bold(format string, args ...any) {
	m.recordMessage(&m.BoldCalls, format, args...)
}

// Printf captures printf calls.
func (m *MessageLoggerMock) Printf(format string, args ...any) {
	m.recordMessage(&m.PrintfCalls, format, args...)
}

// Fprintf captures fprintf calls.
func (m *MessageLoggerMock) Fprintf(_ *os.File, format string, args ...any) {
	m.recordMessage(&m.FprintfCalls, format, args...)
}

// recordMessage is a generic helper for recording formatted messages with thread-safety.
func (m *MessageLoggerMock) recordMessage(callSlice *[]string, format string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	*callSlice = append(*callSlice, fmt.Sprintf(format, args...))
}

// ErrorReporterMock tracks error reporter calls for testing.
type ErrorReporterMock struct {
	mu                        sync.Mutex
	ErrorCalls                []string
	ErrorWithSuggestionsCalls []string
	ErrorWithContextCalls     []string
	ErrorWithSimpleFixCalls   []string
}

// Error captures error calls.
func (m *ErrorReporterMock) Error(format string, args ...any) {
	m.recordError(&m.ErrorCalls, fmt.Sprintf(format, args...))
}

// ErrorWithSuggestions captures error with suggestions calls.
func (m *ErrorReporterMock) ErrorWithSuggestions(err error) {
	if err != nil {
		m.recordError(&m.ErrorWithSuggestionsCalls, err.Error())
	}
}

// ErrorWithContext captures error with context calls.
func (m *ErrorReporterMock) ErrorWithContext(_ appconstants.ErrorCode, message string, _ map[string]string) {
	m.recordError(&m.ErrorWithContextCalls, message)
}

// ErrorWithSimpleFix captures error with simple fix calls.
func (m *ErrorReporterMock) ErrorWithSimpleFix(message, suggestion string) {
	m.recordError(&m.ErrorWithSimpleFixCalls, message+": "+suggestion)
}

// recordError is a generic helper for recording error messages with thread-safety.
func (m *ErrorReporterMock) recordError(callSlice *[]string, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	*callSlice = append(*callSlice, message)
}

// ProgressReporterMock tracks progress reporter calls for testing.
type ProgressReporterMock struct {
	mu            sync.Mutex
	ProgressCalls []string
}

// Progress captures progress calls.
func (m *ProgressReporterMock) Progress(format string, args ...any) {
	m.recordProgress(&m.ProgressCalls, format, args...)
}

// recordProgress is a generic helper for recording progress messages with thread-safety.
func (m *ProgressReporterMock) recordProgress(callSlice *[]string, format string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	*callSlice = append(*callSlice, fmt.Sprintf(format, args...))
}

// ErrorFormatterMock tracks error formatter calls for testing.
type ErrorFormatterMock struct {
	mu                         sync.Mutex
	FormatContextualErrorCalls []string
}

// FormatContextualError captures contextual error formatting calls.
func (m *ErrorFormatterMock) FormatContextualError(err error) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err != nil {
		formatted := err.Error()
		m.FormatContextualErrorCalls = append(m.FormatContextualErrorCalls, formatted)

		return formatted
	}

	return ""
}

// OutputConfigMock implements OutputConfig for testing.
type OutputConfigMock struct {
	QuietMode bool
}

// IsQuiet returns whether quiet mode is enabled.
func (m *OutputConfigMock) IsQuiet() bool {
	return m.QuietMode
}
