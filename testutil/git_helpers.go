package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

// SetupGitDirectory creates a .git directory in the given path.
// Returns the path to the created .git directory.
// Used in git detector tests to reduce duplication.
func SetupGitDirectory(t *testing.T, tmpDir string) string {
	t.Helper()
	gitDir := filepath.Join(tmpDir, appconstants.DirGit)
	err := os.MkdirAll(gitDir, appconstants.FilePermDir)
	AssertNoError(t, err)

	return gitDir
}

// SetupGitConfig creates a git config file with the given remote URL.
// The config file is created in the specified gitDir.
// Used in git detector tests to reduce duplication.
func SetupGitConfig(t *testing.T, gitDir, remoteURL string) {
	t.Helper()
	configPath := filepath.Join(gitDir, TestCmdConfig)
	config := fmt.Sprintf(`[remote "origin"]
	url = %s
	fetch = +refs/heads/*:refs/remotes/origin/*
`, remoteURL)
	WriteTestFile(t, configPath, config)
}

// CreateGitConfigWithRemote creates a git config file with remote and branch configuration.
// Consolidates 6+ duplicated git config setups in detector_test.go.
// Returns the path to the created config file.
func CreateGitConfigWithRemote(t *testing.T, gitDir, remoteURL, branchName string) string {
	t.Helper()

	configContent := fmt.Sprintf(`[remote "origin"]
	url = %s
	fetch = +refs/heads/*:refs/remotes/origin/*
[branch "%s"]
	remote = origin
	merge = refs/heads/%s
`, remoteURL, branchName, branchName)

	configPath := filepath.Join(gitDir, TestCmdConfig)
	WriteTestFile(t, configPath, configContent)

	return configPath
}

// WriteGitConfigFile creates a .git directory and writes a config file.
// Returns the path to the config file for further assertions.
// This is a convenience wrapper combining SetupGitDirectory + file writing.
//
// Example:
//
//	configPath := testutil.WriteGitConfigFile(t, tmpDir, `[remote "origin"]...`)
func WriteGitConfigFile(t *testing.T, baseDir, configContent string) string {
	t.Helper()

	gitDir := filepath.Join(baseDir, appconstants.DirGit)
	CreateTestDir(t, gitDir)

	configPath := filepath.Join(gitDir, TestCmdConfig)
	WriteTestFile(t, configPath, configContent)

	return configPath
}
