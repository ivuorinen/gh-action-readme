package internal

import (
	"os"
	"testing"

	"github.com/schollz/progressbar/v3"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/apperrors"
)

const testFormatString = "test %s %d"

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

// TestNullOutputEdgeCases tests NullOutput methods with edge case inputs.
func TestNullOutputEdgeCases(t *testing.T) {
	t.Parallel()

	no := NewNullOutput()

	// Test with empty strings
	no.Success("")
	no.Error("")
	no.Warning("")
	no.Info("")
	no.Progress("")
	no.Bold("")
	no.Printf("")

	// Test with special characters
	specialChars := "\n\t\r\x00\a\b\v\f"
	no.Success(specialChars)
	no.Error(specialChars)
	no.Warning(specialChars)
	no.Info(specialChars)
	no.Progress(specialChars)
	no.Bold(specialChars)
	no.Printf(specialChars)

	// Test with unicode
	unicode := "🎉 emoji test 你好 мир"
	no.Success(unicode)
	no.Error(unicode)
	no.Warning(unicode)
	no.Info(unicode)
	no.Progress(unicode)
	no.Bold(unicode)
	no.Printf(unicode)

	// Test with format strings and nil args
	no.Printf(testFormatString, nil, nil)
	no.Success(testFormatString, nil, nil)
	no.Error(testFormatString, nil, nil)

	// Test with multiple args
	no.Success("test", "arg1", "arg2", "arg3")
	no.Error("test", 1, 2, 3, 4, 5)
	no.Printf("test %s %d %v", "str", 42, true)
}

// TestNullOutputErrorMethodsWithNil tests error methods with nil inputs.
func TestNullOutputErrorMethodsWithNil(t *testing.T) {
	t.Parallel()

	no := NewNullOutput()

	// Test with nil error
	no.ErrorWithSuggestions(nil)
	no.FormatContextualError(nil)

	// Test with nil context
	no.ErrorWithContext(appconstants.ErrCodeUnknown, "test", nil)

	// Test with empty context
	no.ErrorWithContext(appconstants.ErrCodeUnknown, "", map[string]string{})

	// Test with empty simple fix
	no.ErrorWithSimpleFix("", "")
}

// TestNullProgressManagerEdgeCases tests NullProgressManager with edge cases.
func TestNullProgressManagerEdgeCases(t *testing.T) {
	t.Parallel()

	npm := NewNullProgressManager()

	// Test with empty strings
	bar := npm.CreateProgressBar("", 0)
	if bar != nil {
		t.Error("CreateProgressBar with empty string should return nil")
	}

	// Test with negative count
	bar = npm.CreateProgressBar("test", -1)
	if bar != nil {
		t.Error("CreateProgressBar with negative count should return nil")
	}

	// Test with empty file list
	bar = npm.CreateProgressBarForFiles("test", []string{})
	if bar != nil {
		t.Error("CreateProgressBarForFiles with empty list should return nil")
	}

	// Test with nil file list
	bar = npm.CreateProgressBarForFiles("test", nil)
	if bar != nil {
		t.Error("CreateProgressBarForFiles with nil list should return nil")
	}

	// Test ProcessWithProgressBar with empty items
	callCount := 0
	npm.ProcessWithProgressBar("test", []string{}, func(_ string, _ *progressbar.ProgressBar) {
		callCount++
	})
	if callCount != 0 {
		t.Errorf("ProcessWithProgressBar with empty items called func %d times, want 0", callCount)
	}

	// Test ProcessWithProgressBar with nil items
	callCount = 0
	npm.ProcessWithProgressBar("test", nil, func(_ string, _ *progressbar.ProgressBar) {
		callCount++
	})
	if callCount != 0 {
		t.Errorf("ProcessWithProgressBar with nil items called func %d times, want 0", callCount)
	}
}

// TestNullOutputInterfaceCompliance verifies NullOutput implements CompleteOutput.
func TestNullOutputInterfaceCompliance(t *testing.T) {
	t.Parallel()

	var _ CompleteOutput = (*NullOutput)(nil)
	var _ MessageLogger = (*NullOutput)(nil)
	var _ ErrorReporter = (*NullOutput)(nil)
	var _ ErrorFormatter = (*NullOutput)(nil)
	var _ ProgressReporter = (*NullOutput)(nil)
	var _ OutputConfig = (*NullOutput)(nil)
}

// TestNullProgressManagerInterfaceCompliance verifies NullProgressManager implements ProgressManager.
func TestNullProgressManagerInterfaceCompliance(t *testing.T) {
	t.Parallel()

	var _ ProgressManager = (*NullProgressManager)(nil)
}
