package internal

import (
	"os"

	"github.com/schollz/progressbar/v3"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/apperrors"
)

// NullOutput is a no-op implementation of CompleteOutput for testing.
// All methods are no-ops to prevent cluttering test output.
type NullOutput struct{}

// Compile-time interface checks.
var (
	_ MessageLogger    = (*NullOutput)(nil)
	_ ProgressReporter = (*NullOutput)(nil)
	_ QuietChecker     = (*NullOutput)(nil)
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
func (no *NullOutput) Success(_ string, _ ...any) {
	// Intentionally empty: NullOutput suppresses all output for testing.
}

// Error is a no-op.
func (no *NullOutput) Error(_ string, _ ...any) {
	// Intentionally empty: NullOutput suppresses all output for testing.
}

// Warning is a no-op.
func (no *NullOutput) Warning(_ string, _ ...any) {
	// Intentionally empty: NullOutput suppresses all output for testing.
}

// Info is a no-op.
func (no *NullOutput) Info(_ string, _ ...any) {
	// Intentionally empty: NullOutput suppresses all output for testing.
}

// Progress is a no-op.
func (no *NullOutput) Progress(_ string, _ ...any) {
	// Intentionally empty: NullOutput suppresses all output for testing.
}

// Bold is a no-op.
func (no *NullOutput) Bold(_ string, _ ...any) {
	// Intentionally empty: NullOutput suppresses all output for testing.
}

// Printf is a no-op.
func (no *NullOutput) Printf(_ string, _ ...any) {
	// Intentionally empty: NullOutput suppresses all output for testing.
}

// Fprintf is a no-op.
func (no *NullOutput) Fprintf(_ *os.File, _ string, _ ...any) {
	// Intentionally empty: NullOutput suppresses all output for testing.
}

// ErrorWithSuggestions is a no-op.
func (no *NullOutput) ErrorWithSuggestions(_ *apperrors.ContextualError) {
	// Intentionally empty - no-op implementation for testing
}

// ErrorWithContext is a no-op.
func (no *NullOutput) ErrorWithContext(
	_ appconstants.ErrorCode,
	_ string,
	_ map[string]string,
) {
	// Intentionally empty: NullOutput suppresses all output for testing.
}

// ErrorWithSimpleFix is a no-op.
func (no *NullOutput) ErrorWithSimpleFix(_, _ string) {
	// Intentionally empty: NullOutput suppresses all output for testing.
}

// FormatContextualError returns empty string.
func (no *NullOutput) FormatContextualError(_ *apperrors.ContextualError) string {
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
func (npm *NullProgressManager) FinishProgressBar(_ *progressbar.ProgressBar) {
	// Intentionally empty: NullProgressManager suppresses progress output for testing.
}

// FinishProgressBarWithNewline is a no-op.
func (npm *NullProgressManager) FinishProgressBarWithNewline(_ *progressbar.ProgressBar) {
	// Intentionally empty: NullProgressManager suppresses progress output for testing.
}

// UpdateProgressBar is a no-op.
func (npm *NullProgressManager) UpdateProgressBar(_ *progressbar.ProgressBar) {
	// Intentionally empty: NullProgressManager suppresses progress output for testing.
}

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
