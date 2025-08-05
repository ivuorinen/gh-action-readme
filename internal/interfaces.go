// Package internal defines focused interfaces following Interface Segregation Principle.
package internal

import (
	"os"

	"github.com/schollz/progressbar/v3"

	"github.com/ivuorinen/gh-action-readme/internal/errors"
)

// MessageLogger handles informational output messages.
type MessageLogger interface {
	Info(format string, args ...any)
	Success(format string, args ...any)
	Warning(format string, args ...any)
	Bold(format string, args ...any)
	Printf(format string, args ...any)
	Fprintf(w *os.File, format string, args ...any)
}

// ErrorReporter handles error output and reporting.
type ErrorReporter interface {
	Error(format string, args ...any)
	ErrorWithSuggestions(err *errors.ContextualError)
	ErrorWithContext(code errors.ErrorCode, message string, context map[string]string)
	ErrorWithSimpleFix(message, suggestion string)
}

// ErrorFormatter handles formatting of contextual errors.
type ErrorFormatter interface {
	FormatContextualError(err *errors.ContextualError) string
}

// ProgressReporter handles progress indication and status updates.
type ProgressReporter interface {
	Progress(format string, args ...any)
}

// OutputConfig provides configuration queries for output behavior.
type OutputConfig interface {
	IsQuiet() bool
}

// ProgressManager handles progress bar creation and management.
type ProgressManager interface {
	CreateProgressBar(description string, total int) *progressbar.ProgressBar
	CreateProgressBarForFiles(description string, files []string) *progressbar.ProgressBar
	FinishProgressBar(bar *progressbar.ProgressBar)
	FinishProgressBarWithNewline(bar *progressbar.ProgressBar)
	UpdateProgressBar(bar *progressbar.ProgressBar)
	ProcessWithProgressBar(
		description string,
		items []string,
		processFunc func(item string, bar *progressbar.ProgressBar),
	)
}

// OutputWriter combines message logging and progress reporting for general output needs.
type OutputWriter interface {
	MessageLogger
	ProgressReporter
	OutputConfig
}

// ErrorManager combines error reporting and formatting for comprehensive error handling.
type ErrorManager interface {
	ErrorReporter
	ErrorFormatter
}

// CompleteOutput combines all output interfaces for backward compatibility.
// This should be used sparingly and only where all capabilities are truly needed.
type CompleteOutput interface {
	MessageLogger
	ErrorReporter
	ErrorFormatter
	ProgressReporter
	OutputConfig
}
