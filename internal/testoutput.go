package internal

import (
	"os"

	"github.com/schollz/progressbar/v3"

	"github.com/ivuorinen/gh-action-readme/internal/errors"
)

// NullOutput is a no-op implementation of CompleteOutput for testing.
// All methods are no-ops to prevent cluttering test output.
type NullOutput struct{}

// Compile-time interface checks.
var (
	_ MessageLogger    = (*NullOutput)(nil)
	_ ErrorReporter    = (*NullOutput)(nil)
	_ ErrorFormatter   = (*NullOutput)(nil)
	_ ProgressReporter = (*NullOutput)(nil)
	_ OutputConfig     = (*NullOutput)(nil)
	_ CompleteOutput   = (*NullOutput)(nil)
)

// NewNullOutput creates a new null output instance for testing.
func NewNullOutput() *NullOutput {
	return &NullOutput{}
}

// IsQuiet returns true as null output is always quiet.
func (no *NullOutput) IsQuiet() bool {
	return true
}

// Success is a no-op.
func (no *NullOutput) Success(_ string, _ ...any) {}

// Error is a no-op.
func (no *NullOutput) Error(_ string, _ ...any) {}

// Warning is a no-op.
func (no *NullOutput) Warning(_ string, _ ...any) {}

// Info is a no-op.
func (no *NullOutput) Info(_ string, _ ...any) {}

// Progress is a no-op.
func (no *NullOutput) Progress(_ string, _ ...any) {}

// Bold is a no-op.
func (no *NullOutput) Bold(_ string, _ ...any) {}

// Printf is a no-op.
func (no *NullOutput) Printf(_ string, _ ...any) {}

// Fprintf is a no-op.
func (no *NullOutput) Fprintf(_ *os.File, _ string, _ ...any) {}

// ErrorWithSuggestions is a no-op.
func (no *NullOutput) ErrorWithSuggestions(_ *errors.ContextualError) {}

// ErrorWithContext is a no-op.
func (no *NullOutput) ErrorWithContext(
	_ errors.ErrorCode,
	_ string,
	_ map[string]string,
) {
}

// ErrorWithSimpleFix is a no-op.
func (no *NullOutput) ErrorWithSimpleFix(_, _ string) {}

// FormatContextualError returns empty string.
func (no *NullOutput) FormatContextualError(_ *errors.ContextualError) string {
	return ""
}

// NullProgressManager is a no-op implementation of ProgressManager for testing.
type NullProgressManager struct{}

// Compile-time interface check.
var _ ProgressManager = (*NullProgressManager)(nil)

// NewNullProgressManager creates a new null progress manager for testing.
func NewNullProgressManager() *NullProgressManager {
	return &NullProgressManager{}
}

// CreateProgressBar returns nil to suppress progress bars.
func (npm *NullProgressManager) CreateProgressBar(_ string, _ int) *progressbar.ProgressBar {
	return nil
}

// CreateProgressBarForFiles returns nil to suppress progress bars.
func (npm *NullProgressManager) CreateProgressBarForFiles(
	_ string,
	_ []string,
) *progressbar.ProgressBar {
	return nil
}

// FinishProgressBar is a no-op.
func (npm *NullProgressManager) FinishProgressBar(_ *progressbar.ProgressBar) {}

// FinishProgressBarWithNewline is a no-op.
func (npm *NullProgressManager) FinishProgressBarWithNewline(_ *progressbar.ProgressBar) {}

// UpdateProgressBar is a no-op.
func (npm *NullProgressManager) UpdateProgressBar(_ *progressbar.ProgressBar) {}

// ProcessWithProgressBar executes the function for each item without progress display.
func (npm *NullProgressManager) ProcessWithProgressBar(
	_ string,
	items []string,
	processFunc func(item string, bar *progressbar.ProgressBar),
) {
	for _, item := range items {
		processFunc(item, nil)
	}
}
