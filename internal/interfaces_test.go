// Package internal provides tests for the focused interfaces and demonstrates improved testability.
package internal

import (
	"os"
	"strings"
	"testing"

	"github.com/schollz/progressbar/v3"

	"github.com/ivuorinen/gh-action-readme/internal/errors"
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
	m.InfoCalls = append(m.InfoCalls, formatMessage(format, args...))
}

func (m *MockMessageLogger) Success(format string, args ...any) {
	m.SuccessCalls = append(m.SuccessCalls, formatMessage(format, args...))
}

func (m *MockMessageLogger) Warning(format string, args ...any) {
	m.WarningCalls = append(m.WarningCalls, formatMessage(format, args...))
}

func (m *MockMessageLogger) Bold(format string, args ...any) {
	m.BoldCalls = append(m.BoldCalls, formatMessage(format, args...))
}

func (m *MockMessageLogger) Printf(format string, args ...any) {
	m.PrintfCalls = append(m.PrintfCalls, formatMessage(format, args...))
}

func (m *MockMessageLogger) Fprintf(_ *os.File, format string, args ...any) {
	// For testing, just track the formatted message
	m.PrintfCalls = append(m.PrintfCalls, formatMessage(format, args...))
}

// MockErrorReporter implements ErrorReporter for testing.
type MockErrorReporter struct {
	ErrorCalls                []string
	ErrorWithSuggestionsCalls []string
	ErrorWithContextCalls     []string
	ErrorWithSimpleFixCalls   []string
}

func (m *MockErrorReporter) Error(format string, args ...any) {
	m.ErrorCalls = append(m.ErrorCalls, formatMessage(format, args...))
}

func (m *MockErrorReporter) ErrorWithSuggestions(err *errors.ContextualError) {
	if err != nil {
		m.ErrorWithSuggestionsCalls = append(m.ErrorWithSuggestionsCalls, err.Error())
	}
}

func (m *MockErrorReporter) ErrorWithContext(_ errors.ErrorCode, message string, _ map[string]string) {
	m.ErrorWithContextCalls = append(m.ErrorWithContextCalls, message)
}

func (m *MockErrorReporter) ErrorWithSimpleFix(message, suggestion string) {
	m.ErrorWithSimpleFixCalls = append(m.ErrorWithSimpleFixCalls, message+": "+suggestion)
}

// MockProgressReporter implements ProgressReporter for testing.
type MockProgressReporter struct {
	ProgressCalls []string
}

func (m *MockProgressReporter) Progress(format string, args ...any) {
	m.ProgressCalls = append(m.ProgressCalls, formatMessage(format, args...))
}

// MockOutputConfig implements OutputConfig for testing.
type MockOutputConfig struct {
	QuietMode bool
}

func (m *MockOutputConfig) IsQuiet() bool {
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
	m.CreateProgressBarCalls = append(m.CreateProgressBarCalls, formatMessage("%s (total: %d)", description, total))
	return nil // Return nil for mock to avoid actual progress bar
}

func (m *MockProgressManager) CreateProgressBarForFiles(description string, files []string) *progressbar.ProgressBar {
	m.CreateProgressBarForFilesCalls = append(
		m.CreateProgressBarForFilesCalls,
		formatMessage("%s (files: %d)", description, len(files)),
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
		formatMessage("%s (items: %d)", description, len(items)),
	)
	// Execute the process function for each item
	for _, item := range items {
		processFunc(item, nil)
	}
}

// Helper function to format messages consistently.
func formatMessage(format string, args ...any) string {
	if len(args) == 0 {
		return format
	}
	// Simple formatting for test purposes
	result := format
	for _, arg := range args {
		result = strings.Replace(result, "%s", toString(arg), 1)
		result = strings.Replace(result, "%d", toString(arg), 1)
		result = strings.Replace(result, "%v", toString(arg), 1)
	}
	return result
}

func toString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return formatInt(val)
	default:
		return "unknown"
	}
}

func formatInt(i int) string {
	// Simple int to string conversion for testing
	if i == 0 {
		return "0"
	}
	result := ""
	negative := i < 0
	if negative {
		i = -i
	}
	for i > 0 {
		digit := i % 10
		result = string(rune('0'+digit)) + result
		i /= 10
	}
	if negative {
		result = "-" + result
	}
	return result
}

// Test that demonstrates improved testability with focused interfaces.
func TestFocusedInterfaces_SimpleLogger(t *testing.T) {
	mockLogger := &MockMessageLogger{}
	simpleLogger := NewSimpleLogger(mockLogger)

	// Test successful operation
	simpleLogger.LogOperation("test-operation", true)

	// Verify the expected calls were made
	if len(mockLogger.InfoCalls) != 1 {
		t.Errorf("expected 1 Info call, got %d", len(mockLogger.InfoCalls))
	}
	if len(mockLogger.SuccessCalls) != 1 {
		t.Errorf("expected 1 Success call, got %d", len(mockLogger.SuccessCalls))
	}
	if len(mockLogger.WarningCalls) != 0 {
		t.Errorf("expected 0 Warning calls, got %d", len(mockLogger.WarningCalls))
	}

	// Check message content
	if !strings.Contains(mockLogger.InfoCalls[0], "test-operation") {
		t.Errorf("expected Info call to contain 'test-operation', got: %s", mockLogger.InfoCalls[0])
	}

	if !strings.Contains(mockLogger.SuccessCalls[0], "test-operation") {
		t.Errorf("expected Success call to contain 'test-operation', got: %s", mockLogger.SuccessCalls[0])
	}
}

func TestFocusedInterfaces_SimpleLogger_WithFailure(t *testing.T) {
	mockLogger := &MockMessageLogger{}
	simpleLogger := NewSimpleLogger(mockLogger)

	// Test failed operation
	simpleLogger.LogOperation("failing-operation", false)

	// Verify the expected calls were made
	if len(mockLogger.InfoCalls) != 1 {
		t.Errorf("expected 1 Info call, got %d", len(mockLogger.InfoCalls))
	}
	if len(mockLogger.SuccessCalls) != 0 {
		t.Errorf("expected 0 Success calls, got %d", len(mockLogger.SuccessCalls))
	}
	if len(mockLogger.WarningCalls) != 1 {
		t.Errorf("expected 1 Warning call, got %d", len(mockLogger.WarningCalls))
	}
}

func TestFocusedInterfaces_ErrorManager(t *testing.T) {
	mockReporter := &MockErrorReporter{}
	mockFormatter := &MockErrorFormatter{}
	mockManager := &mockErrorManager{
		reporter:  mockReporter,
		formatter: mockFormatter,
	}
	errorManager := NewFocusedErrorManager(mockManager)

	// Test validation error handling
	errorManager.HandleValidationError("test-file.yml", []string{"name", "description"})

	// Verify the expected calls were made
	if len(mockReporter.ErrorWithContextCalls) != 1 {
		t.Errorf("expected 1 ErrorWithContext call, got %d", len(mockReporter.ErrorWithContextCalls))
	}

	if !strings.Contains(mockReporter.ErrorWithContextCalls[0], "test-file.yml") {
		t.Errorf("expected error message to contain 'test-file.yml', got: %s", mockReporter.ErrorWithContextCalls[0])
	}
}

func TestFocusedInterfaces_TaskProgress(t *testing.T) {
	mockReporter := &MockProgressReporter{}
	taskProgress := NewTaskProgress(mockReporter)

	// Test progress reporting
	taskProgress.ReportProgress("compile", 3, 10)

	// Verify the expected calls were made
	if len(mockReporter.ProgressCalls) != 1 {
		t.Errorf("expected 1 Progress call, got %d", len(mockReporter.ProgressCalls))
	}

	if !strings.Contains(mockReporter.ProgressCalls[0], "compile") {
		t.Errorf("expected progress message to contain 'compile', got: %s", mockReporter.ProgressCalls[0])
	}
}

func TestFocusedInterfaces_ConfigAwareComponent(t *testing.T) {
	tests := []struct {
		name       string
		quietMode  bool
		shouldShow bool
	}{
		{
			name:       "normal mode should output",
			quietMode:  false,
			shouldShow: true,
		},
		{
			name:       "quiet mode should not output",
			quietMode:  true,
			shouldShow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConfig := &MockOutputConfig{QuietMode: tt.quietMode}
			component := NewConfigAwareComponent(mockConfig)

			result := component.ShouldOutput()

			if result != tt.shouldShow {
				t.Errorf("expected ShouldOutput() to return %v, got %v", tt.shouldShow, result)
			}
		})
	}
}

func TestFocusedInterfaces_CompositeOutputWriter(t *testing.T) {
	// Create a composite mock that implements OutputWriter
	mockLogger := &MockMessageLogger{}
	mockProgress := &MockProgressReporter{}
	mockConfig := &MockOutputConfig{QuietMode: false}

	compositeWriter := &CompositeOutputWriter{
		writer: &mockOutputWriter{
			logger:   mockLogger,
			reporter: mockProgress,
			config:   mockConfig,
		},
	}

	items := []string{"item1", "item2", "item3"}
	compositeWriter.ProcessWithOutput(items)

	// Verify that the composite writer uses both message logging and progress reporting
	// Should have called Info and Success for overall status
	if len(mockLogger.InfoCalls) != 1 {
		t.Errorf("expected 1 Info call, got %d", len(mockLogger.InfoCalls))
	}
	if len(mockLogger.SuccessCalls) != 1 {
		t.Errorf("expected 1 Success call, got %d", len(mockLogger.SuccessCalls))
	}

	// Should have called Progress for each item
	if len(mockProgress.ProgressCalls) != 3 {
		t.Errorf("expected 3 Progress calls, got %d", len(mockProgress.ProgressCalls))
	}
}

func TestFocusedInterfaces_GeneratorWithDependencyInjection(t *testing.T) {
	// Create focused mocks
	mockOutput := &mockCompleteOutput{
		logger:    &MockMessageLogger{},
		reporter:  &MockErrorReporter{},
		formatter: &MockErrorFormatter{},
		progress:  &MockProgressReporter{},
		config:    &MockOutputConfig{QuietMode: false},
	}
	mockProgress := &MockProgressManager{}

	// Create generator with dependency injection
	config := &AppConfig{
		Theme:        "default",
		OutputFormat: "md",
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
	reporter  ErrorReporter
	formatter ErrorFormatter
	progress  ProgressReporter
	config    OutputConfig
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
func (m *mockCompleteOutput) ErrorWithSuggestions(err *errors.ContextualError) {
	m.reporter.ErrorWithSuggestions(err)
}
func (m *mockCompleteOutput) ErrorWithContext(code errors.ErrorCode, message string, context map[string]string) {
	m.reporter.ErrorWithContext(code, message, context)
}
func (m *mockCompleteOutput) ErrorWithSimpleFix(message, suggestion string) {
	m.reporter.ErrorWithSimpleFix(message, suggestion)
}
func (m *mockCompleteOutput) FormatContextualError(err *errors.ContextualError) string {
	return m.formatter.FormatContextualError(err)
}
func (m *mockCompleteOutput) Progress(format string, args ...any) {
	m.progress.Progress(format, args...)
}
func (m *mockCompleteOutput) IsQuiet() bool { return m.config.IsQuiet() }

type mockOutputWriter struct {
	logger   MessageLogger
	reporter ProgressReporter
	config   OutputConfig
}

func (m *mockOutputWriter) Info(format string, args ...any)    { m.logger.Info(format, args...) }
func (m *mockOutputWriter) Success(format string, args ...any) { m.logger.Success(format, args...) }
func (m *mockOutputWriter) Warning(format string, args ...any) { m.logger.Warning(format, args...) }
func (m *mockOutputWriter) Bold(format string, args ...any)    { m.logger.Bold(format, args...) }
func (m *mockOutputWriter) Printf(format string, args ...any)  { m.logger.Printf(format, args...) }
func (m *mockOutputWriter) Fprintf(w *os.File, format string, args ...any) {
	m.logger.Fprintf(w, format, args...)
}
func (m *mockOutputWriter) Progress(format string, args ...any) { m.reporter.Progress(format, args...) }
func (m *mockOutputWriter) IsQuiet() bool                       { return m.config.IsQuiet() }

// MockErrorFormatter implements ErrorFormatter for testing.
type MockErrorFormatter struct {
	FormatContextualErrorCalls []string
}

func (m *MockErrorFormatter) FormatContextualError(err *errors.ContextualError) string {
	if err != nil {
		formatted := err.Error()
		m.FormatContextualErrorCalls = append(m.FormatContextualErrorCalls, formatted)
		return formatted
	}
	return ""
}

// mockErrorManager implements ErrorManager for testing.
type mockErrorManager struct {
	reporter  ErrorReporter
	formatter ErrorFormatter
}

func (m *mockErrorManager) Error(format string, args ...any) { m.reporter.Error(format, args...) }
func (m *mockErrorManager) ErrorWithSuggestions(err *errors.ContextualError) {
	m.reporter.ErrorWithSuggestions(err)
}
func (m *mockErrorManager) ErrorWithContext(code errors.ErrorCode, message string, context map[string]string) {
	m.reporter.ErrorWithContext(code, message, context)
}
func (m *mockErrorManager) ErrorWithSimpleFix(message, suggestion string) {
	m.reporter.ErrorWithSimpleFix(message, suggestion)
}
func (m *mockErrorManager) FormatContextualError(err *errors.ContextualError) string {
	return m.formatter.FormatContextualError(err)
}
