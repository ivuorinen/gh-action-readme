package internal

import (
	"errors"
	"os"
	"path/filepath"
)

func stringContains(a, b string) bool {
	for i := 0; i <= len(a)-len(b); i++ {
		if a[i:i+len(b)] == b {
			return true
		}
	}

	return false
}

func containsString(b []byte, substr string) bool {
	return len(b) > 0 &&
		substr != "" &&
		(len(substr) <= len(b)) &&
		stringContains(string(b), substr)
}

// fileExists returns true if the file or directory exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

// FindProjectRoot searches upwards for a marker file (.git or go.mod)
// and returns the directory path.
func FindProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		gitPath := filepath.Join(dir, ".git")
		goModPath := filepath.Join(dir, "go.mod")
		if fileExists(gitPath) || fileExists(goModPath) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", errors.New("project root marker (.git or go.mod) not found")
}
