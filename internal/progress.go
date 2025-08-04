// Package internal provides progress bar utilities for the gh-action-readme tool.
package internal

import (
	"github.com/schollz/progressbar/v3"
)

// ProgressBarManager handles progress bar creation and management.
type ProgressBarManager struct {
	quiet bool
}

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
