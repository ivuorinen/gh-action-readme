package internal

import (
	"os"
	"testing"

	"github.com/schollz/progressbar/v3"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/apperrors"
)

func TestNullOutput(t *testing.T) {
	t.Parallel()

	no := NewNullOutput()
	if no == nil {
		t.Fatal("NewNullOutput() returned nil")
	}

	// Test IsQuiet
	if !no.IsQuiet() {
		t.Error("NullOutput.IsQuiet() should return true")
	}

	// Test all no-op methods don't panic
	no.Success("test")
	no.Error("test")
	no.Warning("test")
	no.Info("test")
	no.Progress("test")
	no.Bold("test")
	no.Printf("test")
	no.Fprintf(os.Stdout, "test")

	// Test error methods
	err := apperrors.New(appconstants.ErrCodeUnknown, "test error")
	no.ErrorWithSuggestions(err)
	no.ErrorWithContext(appconstants.ErrCodeUnknown, "test", map[string]string{})
	no.ErrorWithSimpleFix("test", "fix")

	// Test FormatContextualError
	formatted := no.FormatContextualError(err)
	if formatted != "" {
		t.Errorf("NullOutput.FormatContextualError() = %q, want empty string", formatted)
	}
}

func TestNullProgressManager(t *testing.T) {
	t.Parallel()

	npm := NewNullProgressManager()
	if npm == nil {
		t.Fatal("NewNullProgressManager() returned nil")
	}

	// Test CreateProgressBar returns nil
	bar := npm.CreateProgressBar("test", 10)
	if bar != nil {
		t.Error("NullProgressManager.CreateProgressBar() should return nil")
	}

	// Test CreateProgressBarForFiles returns nil
	bar = npm.CreateProgressBarForFiles("test", []string{"file1", "file2"})
	if bar != nil {
		t.Error("NullProgressManager.CreateProgressBarForFiles() should return nil")
	}

	// Test no-op methods don't panic
	npm.FinishProgressBar(nil)
	npm.FinishProgressBarWithNewline(nil)
	npm.UpdateProgressBar(nil)

	// Test ProcessWithProgressBar executes function for each item
	var count int
	items := []string{"item1", "item2", "item3"}
	npm.ProcessWithProgressBar("test", items, func(_ string, _ *progressbar.ProgressBar) {
		count++
	})

	if count != len(items) {
		t.Errorf("ProcessWithProgressBar processed %d items, want %d", count, len(items))
	}
}
