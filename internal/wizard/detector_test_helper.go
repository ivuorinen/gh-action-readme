package wizard

import (
	"testing"

	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// NewTestDetector creates a ProjectDetector configured for testing.
// Reduces the 3-line detector initialization pattern to a single line.
//
// Example:
//
//	detector := NewTestDetector(t, tempDir)
func NewTestDetector(t *testing.T, currentDir string) *ProjectDetector {
	t.Helper()
	output := internal.NewColoredOutput(true)

	return &ProjectDetector{
		output:     output,
		currentDir: currentDir,
	}
}

// SetupDetectorWithFiles creates a detector and writes test files to its directory.
// Returns the detector and temp directory path.
//
// Example:
//
//	detector, tmpDir := SetupDetectorWithFiles(t, map[string]string{
//		"action.yml": "name: Test",
//		"package.json": `{"version": "1.0.0"}`,
//	})
func SetupDetectorWithFiles(
	t *testing.T,
	files map[string]string,
) (*ProjectDetector, string) {
	t.Helper()
	tmpDir := t.TempDir()

	for filename, content := range files {
		testutil.WriteFileInDir(t, tmpDir, filename, content)
	}

	return NewTestDetector(t, tmpDir), tmpDir
}
