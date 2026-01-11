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
	m.mu.Lock()
	defer m.mu.Unlock()
	m.InfoCalls = append(m.InfoCalls, fmt.Sprintf(format, args...))
}

// Success captures success message calls.
func (m *MessageLoggerMock) Success(format string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SuccessCalls = append(m.SuccessCalls, fmt.Sprintf(format, args...))
}

// Warning captures warning message calls.
func (m *MessageLoggerMock) Warning(format string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.WarningCalls = append(m.WarningCalls, fmt.Sprintf(format, args...))
}

// Bold captures bold message calls.
func (m *MessageLoggerMock) Bold(format string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BoldCalls = append(m.BoldCalls, fmt.Sprintf(format, args...))
}

// Printf captures printf calls.
func (m *MessageLoggerMock) Printf(format string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PrintfCalls = append(m.PrintfCalls, fmt.Sprintf(format, args...))
}

// Fprintf captures fprintf calls.
func (m *MessageLoggerMock) Fprintf(_ *os.File, format string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FprintfCalls = append(m.FprintfCalls, fmt.Sprintf(format, args...))
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
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ErrorCalls = append(m.ErrorCalls, fmt.Sprintf(format, args...))
}

// ErrorWithSuggestions captures error with suggestions calls.
func (m *ErrorReporterMock) ErrorWithSuggestions(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err != nil {
		m.ErrorWithSuggestionsCalls = append(m.ErrorWithSuggestionsCalls, err.Error())
	}
}

// ErrorWithContext captures error with context calls.
func (m *ErrorReporterMock) ErrorWithContext(_ appconstants.ErrorCode, message string, _ map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ErrorWithContextCalls = append(m.ErrorWithContextCalls, message)
}

// ErrorWithSimpleFix captures error with simple fix calls.
func (m *ErrorReporterMock) ErrorWithSimpleFix(message, suggestion string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ErrorWithSimpleFixCalls = append(m.ErrorWithSimpleFixCalls, message+": "+suggestion)
}

// ProgressReporterMock tracks progress reporter calls for testing.
type ProgressReporterMock struct {
	mu            sync.Mutex
	ProgressCalls []string
}

// Progress captures progress calls.
func (m *ProgressReporterMock) Progress(format string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ProgressCalls = append(m.ProgressCalls, fmt.Sprintf(format, args...))
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
