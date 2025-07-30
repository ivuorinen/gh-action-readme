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

// GetCurrentDirOrExit gets current working directory or exits with error.
func GetCurrentDirOrExit(output *internal.ColoredOutput) string {
	currentDir, err := GetCurrentDir()
	if err != nil {
		output.Error("Error getting current directory: %v", err)
		os.Exit(1)
	}
	return currentDir
}

// SetupGeneratorContext creates a generator with proper setup and current directory.
func SetupGeneratorContext(config *internal.AppConfig) (*internal.Generator, string) {
	generator := internal.NewGenerator(config)
	output := generator.Output

	if config.Verbose {
		output.Info("Using config: %+v", config)
	}

	currentDir := GetCurrentDirOrExit(output)
	return generator, currentDir
}

// DiscoverAndValidateFiles discovers action files with error handling.
func DiscoverAndValidateFiles(generator *internal.Generator, currentDir string, recursive bool) []string {
	actionFiles, err := generator.DiscoverActionFiles(currentDir, recursive)
	if err != nil {
		generator.Output.Error("Error discovering action files: %v", err)
		os.Exit(1)
	}

	if len(actionFiles) == 0 {
		generator.Output.Error("No action.yml or action.yaml files found in %s", currentDir)
		generator.Output.Info("Please run this command in a directory containing GitHub Action files.")
		os.Exit(1)
	}
	return actionFiles
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
