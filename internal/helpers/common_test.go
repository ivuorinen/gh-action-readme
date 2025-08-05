package helpers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

func TestGetCurrentDir(t *testing.T) {
	t.Run("successfully get current directory", func(t *testing.T) {
		currentDir, err := GetCurrentDir()

		testutil.AssertNoError(t, err)

		if currentDir == "" {
			t.Error("expected non-empty current directory")
		}

		// Verify it's an absolute path
		if !filepath.IsAbs(currentDir) {
			t.Errorf("expected absolute path, got: %s", currentDir)
		}

		// Verify the directory actually exists
		if _, err := os.Stat(currentDir); os.IsNotExist(err) {
			t.Errorf("current directory does not exist: %s", currentDir)
		}
	})
}

func TestSetupGeneratorContext(t *testing.T) {
	tests := []struct {
		name   string
		config *internal.AppConfig
	}{
		{
			name: "basic config",
			config: &internal.AppConfig{
				Theme:        "default",
				OutputFormat: "md",
				OutputDir:    ".",
				Verbose:      false,
				Quiet:        false,
			},
		},
		{
			name: "verbose config",
			config: &internal.AppConfig{
				Theme:        "github",
				OutputFormat: "html",
				OutputDir:    "/tmp",
				Verbose:      true,
				Quiet:        false,
			},
		},
		{
			name: "quiet config",
			config: &internal.AppConfig{
				Theme:        "minimal",
				OutputFormat: "json",
				OutputDir:    ".",
				Verbose:      false,
				Quiet:        true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, currentDir, err := SetupGeneratorContext(tt.config)

			// Verify no error occurred
			testutil.AssertNoError(t, err)

			// Verify generator was created
			if generator == nil {
				t.Error("expected generator to be created")
				return
			}

			// Verify current directory is returned
			if currentDir == "" {
				t.Error("expected non-empty current directory")
			}

			if !filepath.IsAbs(currentDir) {
				t.Errorf("expected absolute path, got: %s", currentDir)
			}

			// Verify generator has the correct config
			if generator.Config != tt.config {
				t.Error("expected generator to have the provided config")
			}
		})
	}
}

func TestFindGitRepoRoot(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T, tmpDir string) string
		expectGit bool
	}{
		{
			name: "directory with git repository",
			setupFunc: func(t *testing.T, tmpDir string) string {
				// Create .git directory
				gitDir := filepath.Join(tmpDir, ".git")
				err := os.MkdirAll(gitDir, 0750) // #nosec G301 -- test directory permissions
				testutil.AssertNoError(t, err)

				// Create subdirectory to test from
				subDir := filepath.Join(tmpDir, "subdir")
				err = os.MkdirAll(subDir, 0750) // #nosec G301 -- test directory permissions
				testutil.AssertNoError(t, err)

				return subDir
			},
			expectGit: true,
		},
		{
			name: "directory without git repository",
			setupFunc: func(_ *testing.T, tmpDir string) string {
				// Just return the temp directory without .git
				return tmpDir
			},
			expectGit: false,
		},
		{
			name: "nested directory in git repository",
			setupFunc: func(t *testing.T, tmpDir string) string {
				// Create .git directory at root
				gitDir := filepath.Join(tmpDir, ".git")
				err := os.MkdirAll(gitDir, 0750) // #nosec G301 -- test directory permissions
				testutil.AssertNoError(t, err)

				// Create deeply nested subdirectory
				nestedDir := filepath.Join(tmpDir, "a", "b", "c")
				err = os.MkdirAll(nestedDir, 0750) // #nosec G301 -- test directory permissions
				testutil.AssertNoError(t, err)

				return nestedDir
			},
			expectGit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			testDir := tt.setupFunc(t, tmpDir)
			repoRoot := FindGitRepoRoot(testDir)

			if tt.expectGit {
				if repoRoot == "" {
					t.Error("expected to find git repository root, got empty string")
				}
				if !strings.Contains(repoRoot, tmpDir) {
					t.Errorf("expected repo root to be within %s, got %s", tmpDir, repoRoot)
				}
			} else if repoRoot != "" {
				t.Errorf("expected empty string for non-git directory, got %s", repoRoot)
			}
		})
	}
}

func TestGetGitRepoRootAndInfo(t *testing.T) {
	t.Run("valid git repository with complete info", func(t *testing.T) {
		tmpDir, cleanup := testutil.TempDir(t)
		defer cleanup()

		testDir := setupCompleteGitRepo(t, tmpDir)
		repoRoot, gitInfo, err := GetGitRepoRootAndInfo(testDir)

		testutil.AssertNoError(t, err)
		verifyRepoRoot(t, repoRoot, tmpDir)
		if gitInfo == nil {
			t.Error("expected git info to be returned, got nil")
		}
	})

	t.Run("git repository but info detection fails", func(t *testing.T) {
		tmpDir, cleanup := testutil.TempDir(t)
		defer cleanup()

		testDir := setupMinimalGitRepo(t, tmpDir)
		repoRoot, gitInfo, err := GetGitRepoRootAndInfo(testDir)

		testutil.AssertNoError(t, err)
		verifyRepoRoot(t, repoRoot, tmpDir)
		if gitInfo != nil {
			t.Logf("got unexpected git info: %+v", gitInfo)
		}
	})

	t.Run("directory without git repository", func(t *testing.T) {
		tmpDir, cleanup := testutil.TempDir(t)
		defer cleanup()

		repoRoot, gitInfo, err := GetGitRepoRootAndInfo(tmpDir)

		if err == nil {
			t.Error("expected error, got nil")
		}
		if repoRoot != "" {
			t.Errorf("expected empty repo root, got: %s", repoRoot)
		}
		if gitInfo != nil {
			t.Errorf("expected nil git info, got: %+v", gitInfo)
		}
	})
}

// Helper functions to reduce complexity.
func setupCompleteGitRepo(t *testing.T, tmpDir string) string {
	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	err := os.MkdirAll(gitDir, 0750) // #nosec G301 -- test directory permissions
	testutil.AssertNoError(t, err)

	// Create a basic git config to make it look like a real repo
	configContent := `[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
[remote "origin"]
	url = https://github.com/test/repo.git
	fetch = +refs/heads/*:refs/remotes/origin/*
[branch "main"]
	remote = origin
	merge = refs/heads/main
`
	configPath := filepath.Join(gitDir, "config")
	err = os.WriteFile(configPath, []byte(configContent), 0600) // #nosec G306 -- test file permissions
	testutil.AssertNoError(t, err)

	return tmpDir
}

func setupMinimalGitRepo(t *testing.T, tmpDir string) string {
	// Create .git directory but with minimal content
	gitDir := filepath.Join(tmpDir, ".git")
	err := os.MkdirAll(gitDir, 0750) // #nosec G301 -- test directory permissions
	testutil.AssertNoError(t, err)

	return tmpDir
}

func verifyRepoRoot(t *testing.T, repoRoot, tmpDir string) {
	if repoRoot != "" && !strings.Contains(repoRoot, tmpDir) {
		t.Errorf("expected repo root to be within %s, got %s", tmpDir, repoRoot)
	}
}

// Test error handling in GetGitRepoRootAndInfo.
func TestGetGitRepoRootAndInfo_ErrorHandling(t *testing.T) {
	t.Run("nonexistent directory", func(t *testing.T) {
		nonexistentPath := "/this/path/should/not/exist"
		repoRoot, gitInfo, err := GetGitRepoRootAndInfo(nonexistentPath)

		if err == nil {
			t.Error("expected error for nonexistent directory")
		}

		if repoRoot != "" {
			t.Errorf("expected empty repo root, got: %s", repoRoot)
		}

		if gitInfo != nil {
			t.Errorf("expected nil git info, got: %+v", gitInfo)
		}
	})
}
