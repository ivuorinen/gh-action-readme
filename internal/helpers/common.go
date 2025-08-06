// Package helpers provides helper functions used across the application.
package helpers

import (
	"fmt"
	"os"

	"github.com/ivuorinen/gh-action-readme/internal"
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

// SetupGeneratorContext creates a generator with proper setup and current directory.
func SetupGeneratorContext(config *internal.AppConfig) (*internal.Generator, string, error) {
	generator := internal.NewGenerator(config)
	output := generator.Output

	if config.Verbose {
		output.Info("Using config: %+v", config)
	}

	currentDir, err := GetCurrentDir()
	if err != nil {
		return nil, "", err
	}

	return generator, currentDir, nil
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
