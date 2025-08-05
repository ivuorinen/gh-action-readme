// Package internal provides progress bar utilities for the gh-action-readme tool.
package internal

import (
	"fmt"

	"github.com/schollz/progressbar/v3"
)

// ProgressBarManager handles progress bar creation and management.
// It implements the ProgressManager interface.
type ProgressBarManager struct {
	quiet bool
}

// Compile-time interface check.
var _ ProgressManager = (*ProgressBarManager)(nil)

// NewProgressBarManager creates a new progress bar manager.
func NewProgressBarManager(quiet bool) *ProgressBarManager {
	return &ProgressBarManager{
		quiet: quiet,
	}
}

// CreateProgressBar creates a progress bar with standardized options.
func (pm *ProgressBarManager) CreateProgressBar(description string, total int) *progressbar.ProgressBar {
	if total <= 1 || pm.quiet {
		return nil
	}

	return progressbar.NewOptions(total,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
}

// CreateProgressBarForFiles creates a progress bar for processing multiple files.
func (pm *ProgressBarManager) CreateProgressBarForFiles(description string, files []string) *progressbar.ProgressBar {
	return pm.CreateProgressBar(description, len(files))
}

// FinishProgressBar completes the progress bar display.
func (pm *ProgressBarManager) FinishProgressBar(bar *progressbar.ProgressBar) {
	if bar != nil {
		_ = bar.Finish()
	}
}

// FinishProgressBarWithNewline completes the progress bar display and adds a newline.
func (pm *ProgressBarManager) FinishProgressBarWithNewline(bar *progressbar.ProgressBar) {
	pm.FinishProgressBar(bar)
	if bar != nil {
		fmt.Println()
	}
}

// ProcessWithProgressBar executes a function for each item with progress tracking.
// The processFunc receives the item and the progress bar for updating.
func (pm *ProgressBarManager) ProcessWithProgressBar(
	description string,
	items []string,
	processFunc func(item string, bar *progressbar.ProgressBar),
) {
	bar := pm.CreateProgressBarForFiles(description, items)
	defer pm.FinishProgressBarWithNewline(bar)

	for _, item := range items {
		processFunc(item, bar)
		if bar != nil {
			_ = bar.Add(1)
		}
	}
}

// UpdateProgressBar safely updates the progress bar if it exists.
func (pm *ProgressBarManager) UpdateProgressBar(bar *progressbar.ProgressBar) {
	if bar != nil {
		_ = bar.Add(1)
	}
}
