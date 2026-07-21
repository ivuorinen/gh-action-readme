// Package helpers provides helper functions used across the application.
package helpers

import (
	"fmt"
	"os"

	"github.com/ivuorinen/gh-action-readme/internal/git"
)

// GetCurrentDir gets current working directory with standardized error handling.
func GetCurrentDir() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current directory: %w", err)
	}

	return currentDir, nil
}

// FindGitRepoRoot finds git repository root with standardized error handling.
func FindGitRepoRoot(currentDir string) string {
	repoRoot, _ := git.FindRepositoryRoot(currentDir)

	return repoRoot
}
