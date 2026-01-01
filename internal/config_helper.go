package internal

import (
	"os"
	"path/filepath"
)

// findFirstExistingConfig searches for the first existing config file
// from a list of config names within a base directory.
// Returns the full path to the first existing config file, or empty string if none exist.
func findFirstExistingConfig(basePath string, configNames []string) (string, bool) {
	for _, name := range configNames {
		path := filepath.Join(basePath, name)
		if _, err := os.Stat(path); err == nil {
			return path, true
		}
	}

	return "", false
}
