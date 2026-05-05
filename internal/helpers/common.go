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

// GetGitRepoRootAndInfo gets git repository root and info with error handling.
func GetGitRepoRootAndInfo(startPath string) (string, *git.RepoInfo, error) {
	repoRoot, err := git.FindRepositoryRoot(startPath)
	if err != nil {
		return "", nil, err
	}

	gitInfo, err := git.DetectRepository(repoRoot)
	if err != nil {
		return repoRoot, nil, err
	}

	return repoRoot, gitInfo, nil
}
