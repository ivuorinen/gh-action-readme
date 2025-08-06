// Package internal demonstrates how to use focused interfaces for better separation of concerns.
package internal

import (
	"fmt"

	"github.com/ivuorinen/gh-action-readme/internal/errors"
)

// SimpleLogger demonstrates a component that only needs basic message logging.
// It depends only on MessageLogger, not the entire output system.
type SimpleLogger struct {
	logger MessageLogger
}

// NewSimpleLogger creates a logger that only needs message logging capabilities.
func NewSimpleLogger(logger MessageLogger) *SimpleLogger {
	return &SimpleLogger{logger: logger}
}

// LogOperation logs a simple operation with different message types.
func (sl *SimpleLogger) LogOperation(operation string, success bool) {
	sl.logger.Info("Starting operation: %s", operation)

	if success {
		sl.logger.Success("Operation completed: %s", operation)
	} else {
		sl.logger.Warning("Operation had issues: %s", operation)
	}
}

// FocusedErrorManager demonstrates a component that only handles error reporting.
// It depends only on ErrorReporter and ErrorFormatter, not the entire output system.
type FocusedErrorManager struct {
	manager ErrorManager
}

// NewFocusedErrorManager creates an error manager with focused dependencies.
func NewFocusedErrorManager(manager ErrorManager) *FocusedErrorManager {
	return &FocusedErrorManager{
		manager: manager,
	}
}

// HandleValidationError handles validation errors with context and suggestions.
func (fem *FocusedErrorManager) HandleValidationError(file string, missingFields []string) {
	context := map[string]string{
		"file":           file,
		"missing_fields": fmt.Sprintf("%v", missingFields),
	}

	fem.manager.ErrorWithContext(
		errors.ErrCodeValidation,
		fmt.Sprintf("Validation failed for %s", file),
		context,
	)
}

// TaskProgress demonstrates a component that only needs progress reporting.
// It depends only on ProgressReporter, not the entire output system.
type TaskProgress struct {
	reporter ProgressReporter
}

// NewTaskProgress creates a progress reporter with focused dependencies.
func NewTaskProgress(reporter ProgressReporter) *TaskProgress {
	return &TaskProgress{reporter: reporter}
}

// ReportProgress reports progress for a long-running task.
func (tp *TaskProgress) ReportProgress(task string, step int, total int) {
	tp.reporter.Progress("Task %s: step %d of %d", task, step, total)
}

// ConfigAwareComponent demonstrates a component that only needs to check configuration.
// It depends only on OutputConfig, not the entire output system.
type ConfigAwareComponent struct {
	config OutputConfig
}

// NewConfigAwareComponent creates a component that checks output configuration.
func NewConfigAwareComponent(config OutputConfig) *ConfigAwareComponent {
	return &ConfigAwareComponent{config: config}
}

// ShouldOutput determines whether output should be generated based on quiet mode.
func (cac *ConfigAwareComponent) ShouldOutput() bool {
	return !cac.config.IsQuiet()
}

// CompositeOutputWriter demonstrates how to compose interfaces for specific needs.
// It combines MessageLogger and ProgressReporter without error handling.
type CompositeOutputWriter struct {
	writer OutputWriter
}

// NewCompositeOutputWriter creates a writer that combines message and progress reporting.
func NewCompositeOutputWriter(writer OutputWriter) *CompositeOutputWriter {
	return &CompositeOutputWriter{writer: writer}
}

// ProcessWithOutput demonstrates processing with both messages and progress.
func (cow *CompositeOutputWriter) ProcessWithOutput(items []string) {
	if cow.writer.IsQuiet() {
		return
	}

	cow.writer.Info("Processing %d items", len(items))

	for i, item := range items {
		cow.writer.Progress("Processing item %d: %s", i+1, item)
		// Process the item...
	}

	cow.writer.Success("All items processed successfully")
}

// ValidationComponent demonstrates combining error handling interfaces.
type ValidationComponent struct {
	errorManager ErrorManager
	logger       MessageLogger
}

// NewValidationComponent creates a validator with focused error handling and logging.
func NewValidationComponent(errorManager ErrorManager, logger MessageLogger) *ValidationComponent {
	return &ValidationComponent{
		errorManager: errorManager,
		logger:       logger,
	}
}

// ValidateAndReport validates an item and reports results using focused interfaces.
func (vc *ValidationComponent) ValidateAndReport(item string, isValid bool, err error) {
	if isValid {
		vc.logger.Success("Validation passed for: %s", item)

		return
	}

	if err != nil {
		if contextualErr, ok := err.(*errors.ContextualError); ok {
			vc.errorManager.ErrorWithSuggestions(contextualErr)
		} else {
			vc.errorManager.Error("Validation failed for %s: %v", item, err)
		}
	} else {
		vc.errorManager.ErrorWithSimpleFix(
			fmt.Sprintf("Validation failed for %s", item),
			"Please check the item configuration and try again",
		)
	}
}
