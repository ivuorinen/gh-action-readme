package internal

import (
	"testing"

	"github.com/schollz/progressbar/v3"
)

func TestProgressBarManager_CreateProgressBar(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		quiet       bool
		description string
		total       int
		expectNil   bool
	}{
		{
			name:        "normal progress bar",
			quiet:       false,
			description: "Test progress",
			total:       10,
			expectNil:   false,
		},
		{
			name:        "quiet mode returns nil",
			quiet:       true,
			description: "Test progress",
			total:       10,
			expectNil:   true,
		},
		{
			name:        "single item returns nil",
			quiet:       false,
			description: "Test progress",
			total:       1,
			expectNil:   true,
		},
		{
			name:        "zero items returns nil",
			quiet:       false,
			description: "Test progress",
			total:       0,
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pm := NewProgressBarManager(tt.quiet)
			bar := pm.CreateProgressBar(tt.description, tt.total)

			if tt.expectNil {
				if bar != nil {
					t.Errorf("expected nil progress bar, got %v", bar)
				}
			} else {
				if bar == nil {
					t.Error("expected progress bar, got nil")
				}
			}
		})
	}
}

func TestProgressBarManager_CreateProgressBarForFiles(t *testing.T) {
	t.Parallel()
	pm := NewProgressBarManager(false)
	files := []string{"file1.yml", "file2.yml", "file3.yml"}

	bar := pm.CreateProgressBarForFiles("Processing files", files)

	if bar == nil {
		t.Error("expected progress bar for multiple files, got nil")
	}
}

func TestProgressBarManager_FinishProgressBar(t *testing.T) {
	t.Parallel()
	pm := NewProgressBarManager(false)

	// Test with nil bar (should not panic)
	pm.FinishProgressBar(nil)

	// Test with actual bar
	bar := pm.CreateProgressBar("Test", 5)
	if bar != nil {
		pm.FinishProgressBar(bar)
	}
}

func TestProgressBarManager_UpdateProgressBar(t *testing.T) {
	t.Parallel()
	pm := NewProgressBarManager(false)

	// Test with nil bar (should not panic)
	pm.UpdateProgressBar(nil)

	// Test with actual bar
	bar := pm.CreateProgressBar("Test", 5)
	if bar != nil {
		pm.UpdateProgressBar(bar)
	}
}

func TestProgressBarManager_ProcessWithProgressBar(t *testing.T) {
	t.Parallel()
	pm := NewProgressBarManager(false)
	items := []string{"item1", "item2", "item3"}

	processedItems := make([]string, 0)
	processFunc := func(item string, _ *progressbar.ProgressBar) {
		processedItems = append(processedItems, item)
	}

	pm.ProcessWithProgressBar("Processing items", items, processFunc)

	if len(processedItems) != len(items) {
		t.Errorf("expected %d processed items, got %d", len(items), len(processedItems))
	}

	for i, item := range items {
		if processedItems[i] != item {
			t.Errorf("expected item %s at position %d, got %s", item, i, processedItems[i])
		}
	}
}

func TestProgressBarManager_ProcessWithProgressBar_QuietMode(t *testing.T) {
	t.Parallel()
	pm := NewProgressBarManager(true) // quiet mode
	items := []string{"item1", "item2"}

	processedItems := make([]string, 0)
	processFunc := func(item string, bar *progressbar.ProgressBar) {
		processedItems = append(processedItems, item)
		// In quiet mode, bar should be nil
		if bar != nil {
			t.Error("expected nil progress bar in quiet mode")
		}
	}

	pm.ProcessWithProgressBar("Processing items", items, processFunc)

	if len(processedItems) != len(items) {
		t.Errorf("expected %d processed items, got %d", len(items), len(processedItems))
	}
}
