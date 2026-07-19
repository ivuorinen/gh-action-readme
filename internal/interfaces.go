// Package internal defines focused interfaces following Interface Segregation Principle.
package internal

import (
	"os"

	"github.com/schollz/progressbar/v3"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/apperrors"
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

// ProgressReporter handles progress indication and status updates.
type ProgressReporter interface {
	Progress(format string, args ...any)
}

// QuietChecker provides queries for quiet mode behavior.
type QuietChecker interface {
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
	QuietChecker
}

// CompleteOutput combines every output capability the generator and CLI use.
// Use sparingly — only where message logging, error reporting and formatting,
// progress, and quiet-mode are all genuinely needed.
type CompleteOutput interface {
	MessageLogger
	// Error reporting.
	Error(format string, args ...any)
	ErrorWithSuggestions(err *apperrors.ContextualError)
	ErrorWithContext(code appconstants.ErrorCode, message string, context map[string]string)
	ErrorWithSimpleFix(message, suggestion string)
	// Contextual error formatting.
	FormatContextualError(err *apperrors.ContextualError) string
	ProgressReporter
	QuietChecker
}
