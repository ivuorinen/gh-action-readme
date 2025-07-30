// Package validation provides common utility functions for the gh-action-readme tool.
package validation

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetBinaryDir returns the directory containing the current executable.
func GetBinaryDir() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	return filepath.Dir(executable), nil
}

// EnsureAbsolutePath converts a relative path to an absolute path.
func EnsureAbsolutePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	return filepath.Abs(path)
}
