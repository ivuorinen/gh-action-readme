// Package internal provides tests for the focused interfaces and demonstrates improved testability.
package internal

import (
	"fmt"
	"os"
	"testing"

	"github.com/schollz/progressbar/v3"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/apperrors"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// MockMessageLogger implements MessageLogger for testing.
type MockMessageLogger struct {
	InfoCalls    []string
	SuccessCalls []string
	WarningCalls []string
	BoldCalls    []string
	PrintfCalls  []string
}

func (m *MockMessageLogger) Info(format string, args ...any) {
	m.recordCall(&m.InfoCalls, format, args...)
}

func (m *MockMessageLogger) Success(format string, args ...any) {
	m.recordCall(&m.SuccessCalls, format, args...)
}

func (m *MockMessageLogger) Warning(format string, args ...any) {
	m.recordCall(&m.WarningCalls, format, args...)
}

func (m *MockMessageLogger) Bold(format string, args ...any) {
	m.recordCall(&m.BoldCalls, format, args...)
}

func (m *MockMessageLogger) Printf(format string, args ...any) {
	m.recordCall(&m.PrintfCalls, format, args...)
}

func (m *MockMessageLogger) Fprintf(_ *os.File, format string, args ...any) {
	// For testing, just track the formatted message
	m.recordCall(&m.PrintfCalls, format, args...)
}

// recordCall is a helper to reduce duplication in mock methods.
func (m *MockMessageLogger) recordCall(callSlice *[]string, format string, args ...any) {
	*callSlice = append(*callSlice, fmt.Sprintf(format, args...))
}

// MockErrorReporter records error-reporting calls for testing.
type MockErrorReporter struct {
	ErrorCalls                []string
	ErrorWithSuggestionsCalls []string
	ErrorWithContextCalls     []string
	ErrorWithSimpleFixCalls   []string
}

func (m *MockErrorReporter) Error(format string, args ...any) {
	m.recordCall(&m.ErrorCalls, format, args...)
}

func (m *MockErrorReporter) ErrorWithSuggestions(err *apperrors.ContextualError) {
	if err != nil {
		m.ErrorWithSuggestionsCalls = append(m.ErrorWithSuggestionsCalls, err.Error())
	}
}

func (m *MockErrorReporter) ErrorWithContext(_ appconstants.ErrorCode, message string, _ map[string]string) {
	m.ErrorWithContextCalls = append(m.ErrorWithContextCalls, message)
}

func (m *MockErrorReporter) ErrorWithSimpleFix(message, suggestion string) {
	m.ErrorWithSimpleFixCalls = append(m.ErrorWithSimpleFixCalls, message+": "+suggestion)
}

// recordCall is a helper to reduce duplication in mock methods.
func (m *MockErrorReporter) recordCall(callSlice *[]string, format string, args ...any) {
	*callSlice = append(*callSlice, fmt.Sprintf(format, args...))
}

// MockProgressReporter implements ProgressReporter for testing.
type MockProgressReporter struct {
	ProgressCalls []string
}

func (m *MockProgressReporter) Progress(format string, args ...any) {
	m.recordCall(&m.ProgressCalls, format, args...)
}

// recordCall is a helper to reduce duplication in mock methods.
func (m *MockProgressReporter) recordCall(callSlice *[]string, format string, args ...any) {
	*callSlice = append(*callSlice, fmt.Sprintf(format, args...))
}

// MockQuietChecker implements QuietChecker for testing.
type MockQuietChecker struct {
	QuietMode bool
}

func (m *MockQuietChecker) IsQuiet() bool {
	return m.QuietMode
}

// MockProgressManager implements ProgressManager for testing.
type MockProgressManager struct {
	CreateProgressBarCalls            []string
	CreateProgressBarForFilesCalls    []string
	FinishProgressBarCalls            int
	FinishProgressBarWithNewlineCalls int
	UpdateProgressBarCalls            int
	ProcessWithProgressBarCalls       []string
}

func (m *MockProgressManager) CreateProgressBar(description string, total int) *progressbar.ProgressBar {
	m.CreateProgressBarCalls = append(m.CreateProgressBarCalls, fmt.Sprintf("%s (total: %d)", description, total))

	return nil // Return nil for mock to avoid actual progress bar
}

func (m *MockProgressManager) CreateProgressBarForFiles(description string, files []string) *progressbar.ProgressBar {
	m.CreateProgressBarForFilesCalls = append(
		m.CreateProgressBarForFilesCalls,
		fmt.Sprintf("%s (files: %d)", description, len(files)),
	)

	return nil // Return nil for mock to avoid actual progress bar
}

func (m *MockProgressManager) FinishProgressBar(_ *progressbar.ProgressBar) {
	m.FinishProgressBarCalls++
}

func (m *MockProgressManager) FinishProgressBarWithNewline(_ *progressbar.ProgressBar) {
	m.FinishProgressBarWithNewlineCalls++
}

func (m *MockProgressManager) UpdateProgressBar(_ *progressbar.ProgressBar) {
	m.UpdateProgressBarCalls++
}

func (m *MockProgressManager) ProcessWithProgressBar(
	description string,
	items []string,
	processFunc func(item string, bar *progressbar.ProgressBar),
) {
	m.ProcessWithProgressBarCalls = append(
		m.ProcessWithProgressBarCalls,
		fmt.Sprintf("%s (items: %d)", description, len(items)),
	)
	// Execute the process function for each item
	for _, item := range items {
		processFunc(item, nil)
	}
}

func TestFocusedInterfacesGeneratorWithDependencyInjection(t *testing.T) {
	t.Parallel()
	// Create focused mocks
	mockOutput := &mockCompleteOutput{
		logger:    &MockMessageLogger{},
		reporter:  &MockErrorReporter{},
		formatter: &errorFormatterWrapper{&testutil.ErrorFormatterMock{}},
		progress:  &MockProgressReporter{},
		config:    &MockQuietChecker{QuietMode: false},
	}
	mockProgress := &MockProgressManager{}

	// Create generator with dependency injection
	config := &AppConfig{
		Theme:        appconstants.ThemeDefault,
		OutputFormat: appconstants.OutputFormatMarkdown,
		OutputDir:    ".",
		Verbose:      false,
		Quiet:        false,
	}

	generator := NewGeneratorWithDependencies(config, mockOutput, mockProgress)

	// Verify the generator was created with the injected dependencies
	if generator == nil {
		t.Fatal("expected generator to be created")
	}
	if generator.Config != config {
		t.Error("expected generator to have the provided config")
	}
	if generator.Output != mockOutput {
		t.Error("expected generator to have the injected output")
	}
	if generator.Progress != mockProgress {
		t.Error("expected generator to have the injected progress manager")
	}
}

// Composite mock types to implement the composed interfaces

type mockCompleteOutput struct {
	logger    MessageLogger
	reporter  *MockErrorReporter
	formatter *errorFormatterWrapper
	progress  ProgressReporter
	config    QuietChecker
}

func (m *mockCompleteOutput) Info(format string, args ...any)    { m.logger.Info(format, args...) }
func (m *mockCompleteOutput) Success(format string, args ...any) { m.logger.Success(format, args...) }
func (m *mockCompleteOutput) Warning(format string, args ...any) { m.logger.Warning(format, args...) }
func (m *mockCompleteOutput) Bold(format string, args ...any)    { m.logger.Bold(format, args...) }
func (m *mockCompleteOutput) Printf(format string, args ...any)  { m.logger.Printf(format, args...) }
func (m *mockCompleteOutput) Fprintf(w *os.File, format string, args ...any) {
	m.logger.Fprintf(w, format, args...)
}
func (m *mockCompleteOutput) Error(format string, args ...any) { m.reporter.Error(format, args...) }
func (m *mockCompleteOutput) ErrorWithSuggestions(err *apperrors.ContextualError) {
	m.reporter.ErrorWithSuggestions(err)
}
func (m *mockCompleteOutput) ErrorWithContext(code appconstants.ErrorCode, message string, context map[string]string) {
	m.reporter.ErrorWithContext(code, message, context)
}
func (m *mockCompleteOutput) ErrorWithSimpleFix(message, suggestion string) {
	m.reporter.ErrorWithSimpleFix(message, suggestion)
}
func (m *mockCompleteOutput) FormatContextualError(err *apperrors.ContextualError) string {
	return m.formatter.FormatContextualError(err)
}
func (m *mockCompleteOutput) Progress(format string, args ...any) {
	m.progress.Progress(format, args...)
}
func (m *mockCompleteOutput) IsQuiet() bool { return m.config.IsQuiet() }

// errorFormatterWrapper wraps testutil.ErrorFormatterMock to implement ErrorFormatter interface.
type errorFormatterWrapper struct {
	*testutil.ErrorFormatterMock
}

// FormatContextualError adapts the generic error interface to ContextualError.
func (w *errorFormatterWrapper) FormatContextualError(err *apperrors.ContextualError) string {
	return w.ErrorFormatterMock.FormatContextualError(err)
}
