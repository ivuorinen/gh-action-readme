package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ValidateTestPath validates a path for use in tests.
// It ensures the path doesn't contain traversal attempts and is within expected boundaries.
func ValidateTestPath(t *testing.T, path string, expectedRoot string) string {
	t.Helper()

	// Clean the path
	cleanPath := filepath.Clean(path)

	// Check if Clean modified the path (indicates traversal attempt)
	if cleanPath != path {
		t.Fatalf("path contains traversal attempt: original=%q clean=%q", path, cleanPath)
	}

	// Check for explicit ".." components
	if strings.Contains(cleanPath, "..") {
		t.Fatalf("path contains .. traversal: %q", cleanPath)
	}

	// If expectedRoot provided, verify path is within boundary
	if expectedRoot != "" {
		cleanRoot := filepath.Clean(expectedRoot)
		relPath, err := filepath.Rel(cleanRoot, cleanPath)
		if err != nil || strings.HasPrefix(relPath, "..") {
			t.Fatalf("path escapes expected root: path=%q root=%q (rel=%q, err=%v)", cleanPath, cleanRoot, relPath, err)
		}
	}

	return cleanPath
}

// SafeReadFile reads a file after validating the path.
// For temp directory paths, pass the temp dir as expectedRoot.
// For fixture paths, use MustReadFixture instead.
func SafeReadFile(t *testing.T, path string, expectedRoot string) []byte {
	t.Helper()

	validPath := ValidateTestPath(t, path, expectedRoot)

	content, err := os.ReadFile(validPath) // #nosec G304 -- path validated
	if err != nil {
		t.Fatalf("failed to read file %q: %v", validPath, err)
	}

	return content
}
