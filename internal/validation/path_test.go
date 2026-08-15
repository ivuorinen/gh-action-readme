package validation

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGetBinaryDirMatchesExecutable pins GetBinaryDir to the directory of the
// running executable. Callers resolve bundled assets relative to it, so a value
// that is merely non-empty is not enough — it must be the actual parent directory.
func TestGetBinaryDirMatchesExecutable(t *testing.T) {
	t.Parallel()

	got, err := GetBinaryDir()
	if err != nil {
		t.Fatalf("GetBinaryDir() unexpected error: %v", err)
	}

	exe, err := os.Executable()
	if err != nil {
		t.Skipf("os.Executable() unavailable on this platform: %v", err)
	}

	if want := filepath.Dir(exe); got != want {
		t.Errorf("GetBinaryDir() = %q, want %q", got, want)
	}
}

// TestGetBinaryDirIsAbsolute guards the property callers actually depend on: the
// result is joined onto asset names, so a relative path would resolve against the
// working directory and silently find the wrong files (or none).
func TestGetBinaryDirIsAbsolute(t *testing.T) {
	t.Parallel()

	got, err := GetBinaryDir()
	if err != nil {
		t.Fatalf("GetBinaryDir() unexpected error: %v", err)
	}

	if !filepath.IsAbs(got) {
		t.Errorf("GetBinaryDir() = %q, want an absolute path", got)
	}
}
