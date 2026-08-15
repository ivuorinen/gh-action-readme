package helpers

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

func TestGetCurrentDir(t *testing.T) {
	t.Parallel()

	t.Run("successfully get current directory", func(t *testing.T) {
		currentDir, err := GetCurrentDir()

		testutil.AssertNoError(t, err)

		if currentDir == "" {
			// Stop here: the path checks below would each report a second,
			// misleading failure for this one root cause.
			t.Fatal("expected non-empty current directory")
		}

		// Verify it's an absolute path
		if !filepath.IsAbs(currentDir) {
			t.Errorf("expected absolute path, got: %s", currentDir)
		}

		// Verify the directory actually exists
		testutil.AssertFileExists(t, currentDir)
	})
}

func TestFindGitRepoRoot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFunc func(t *testing.T, tmpDir string) string
		expectGit bool
	}{
		{
			name: "directory with git repository",
			setupFunc: func(t *testing.T, tmpDir string) string {
				t.Helper()
				// Create .git directory
				_ = testutil.SetupGitDirectory(t, tmpDir)

				// Create subdirectory to test from
				subDir := filepath.Join(tmpDir, "subdir")
				testutil.CreateTestDir(t, subDir)

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
				t.Helper()
				// Create .git directory at root
				_ = testutil.SetupGitDirectory(t, tmpDir)

				// Create deeply nested subdirectory
				nestedDir := filepath.Join(tmpDir, "a", "b", "c")
				testutil.CreateTestDir(t, nestedDir)

				return nestedDir
			},
			expectGit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			testDir := tt.setupFunc(t, tmpDir)
			repoRoot := FindGitRepoRoot(testDir)

			if tt.expectGit {
				if repoRoot == "" {
					// Stop here: the containment check below would report a second,
					// misleading failure for this one root cause.
					t.Fatal("expected to find git repository root, got empty string")
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
	t.Parallel()

	t.Run("valid git repository with complete info", func(t *testing.T) {
		t.Parallel()

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
		t.Parallel()

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
		t.Parallel()

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
	t.Helper()
	// Create .git directory
	gitDir := testutil.SetupGitDirectory(t, tmpDir)

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
	testutil.WriteTestFile(t, configPath, configContent)

	return tmpDir
}

func setupMinimalGitRepo(t *testing.T, tmpDir string) string {
	t.Helper()
	// Create .git directory but with minimal content
	_ = testutil.SetupGitDirectory(t, tmpDir)

	return tmpDir
}

func verifyRepoRoot(t *testing.T, repoRoot, tmpDir string) {
	t.Helper()
	if repoRoot != "" && !strings.Contains(repoRoot, tmpDir) {
		t.Errorf("expected repo root to be within %s, got %s", tmpDir, repoRoot)
	}
}

// Test error handling in GetGitRepoRootAndInfo.
func TestGetGitRepoRootAndInfoErrorHandling(t *testing.T) {
	t.Parallel()

	t.Run(testutil.TestCaseNameNonexistentDir, func(t *testing.T) {
		t.Parallel()

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
