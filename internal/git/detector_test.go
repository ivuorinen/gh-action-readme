package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

func TestFindRepositoryRoot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFunc   func(t *testing.T, tmpDir string) string
		expectError bool
		expectEmpty bool
	}{
		{
			name: "git repository with .git directory",
			setupFunc: func(t *testing.T, tmpDir string) string {
				t.Helper()
				// Create .git directory
				gitDir := filepath.Join(tmpDir, ".git")
				err := os.MkdirAll(gitDir, 0750) // #nosec G301 -- test directory permissions
				if err != nil {
					t.Fatalf("failed to create .git directory: %v", err)
				}

				// Create subdirectory to test from
				subDir := filepath.Join(tmpDir, "subdir", "nested")
				err = os.MkdirAll(subDir, 0750) // #nosec G301 -- test directory permissions
				if err != nil {
					t.Fatalf("failed to create subdirectory: %v", err)
				}

				return subDir
			},
			expectError: false,
			expectEmpty: false,
		},
		{
			name: "git repository with .git file",
			setupFunc: func(t *testing.T, tmpDir string) string {
				t.Helper()
				// Create .git file (for git worktrees)
				gitFile := filepath.Join(tmpDir, ".git")
				testutil.WriteTestFile(t, gitFile, "gitdir: /path/to/git/dir")

				return tmpDir
			},
			expectError: false,
			expectEmpty: false,
		},
		{
			name: "no git repository",
			setupFunc: func(t *testing.T, tmpDir string) string {
				t.Helper()
				// Create subdirectory without .git
				subDir := filepath.Join(tmpDir, "subdir")
				err := os.MkdirAll(subDir, 0750) // #nosec G301 -- test directory permissions
				if err != nil {
					t.Fatalf("failed to create subdirectory: %v", err)
				}

				return subDir
			},
			expectError: true,
		},
		{
			name: "nonexistent directory",
			setupFunc: func(_ *testing.T, tmpDir string) string {
				t.Helper()

				return filepath.Join(tmpDir, "nonexistent")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			testDir := tt.setupFunc(t, tmpDir)

			repoRoot, err := FindRepositoryRoot(testDir)

			if tt.expectError {
				testutil.AssertError(t, err)

				return
			}

			testutil.AssertNoError(t, err)

			if tt.expectEmpty {
				if repoRoot != "" {
					t.Errorf("expected empty repository root, got: %s", repoRoot)
				}
			} else {
				if repoRoot == "" {
					t.Error("expected non-empty repository root")
				}

				// Verify the returned path contains a .git directory or file
				gitPath := filepath.Join(repoRoot, ".git")
				if _, err := os.Stat(gitPath); os.IsNotExist(err) {
					t.Errorf("repository root does not contain .git: %s", repoRoot)
				}
			}
		})
	}
}

func TestDetectGitRepository(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFunc func(t *testing.T, tmpDir string) string
		checkFunc func(t *testing.T, info *RepoInfo)
	}{
		{
			name: "GitHub repository",
			setupFunc: func(t *testing.T, tmpDir string) string {
				t.Helper()
				// Create .git directory
				gitDir := filepath.Join(tmpDir, ".git")
				err := os.MkdirAll(gitDir, 0750) // #nosec G301 -- test directory permissions
				if err != nil {
					t.Fatalf("failed to create .git directory: %v", err)
				}

				// Create config file with GitHub remote
				configContent := `[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
	logallrefupdates = true
[remote "origin"]
	url = https://github.com/owner/repo.git
	fetch = +refs/heads/*:refs/remotes/origin/*
[branch "main"]
	remote = origin
	merge = refs/heads/main
`
				configPath := filepath.Join(gitDir, "config")
				testutil.WriteTestFile(t, configPath, configContent)

				return tmpDir
			},
			checkFunc: func(t *testing.T, info *RepoInfo) {
				t.Helper()
				testutil.AssertEqual(t, "owner", info.Organization)
				testutil.AssertEqual(t, "repo", info.Repository)
				testutil.AssertEqual(t, "https://github.com/owner/repo.git", info.RemoteURL)
			},
		},
		{
			name: "SSH remote URL",
			setupFunc: func(t *testing.T, tmpDir string) string {
				t.Helper()
				gitDir := filepath.Join(tmpDir, ".git")
				err := os.MkdirAll(gitDir, 0750) // #nosec G301 -- test directory permissions
				if err != nil {
					t.Fatalf("failed to create .git directory: %v", err)
				}

				configContent := `[remote "origin"]
	url = git@github.com:owner/repo.git
	fetch = +refs/heads/*:refs/remotes/origin/*
`
				configPath := filepath.Join(gitDir, "config")
				testutil.WriteTestFile(t, configPath, configContent)

				return tmpDir
			},
			checkFunc: func(t *testing.T, info *RepoInfo) {
				t.Helper()
				testutil.AssertEqual(t, "owner", info.Organization)
				testutil.AssertEqual(t, "repo", info.Repository)
				testutil.AssertEqual(t, "git@github.com:owner/repo.git", info.RemoteURL)
			},
		},
		{
			name: "no git repository",
			setupFunc: func(_ *testing.T, tmpDir string) string {
				return tmpDir
			},
			checkFunc: func(t *testing.T, info *RepoInfo) {
				t.Helper()
				testutil.AssertEqual(t, false, info.IsGitRepo)
				testutil.AssertEqual(t, "", info.Organization)
				testutil.AssertEqual(t, "", info.Repository)
			},
		},
		{
			name: "git repository without origin remote",
			setupFunc: func(t *testing.T, tmpDir string) string {
				t.Helper()
				gitDir := filepath.Join(tmpDir, ".git")
				err := os.MkdirAll(gitDir, 0750) // #nosec G301 -- test directory permissions
				if err != nil {
					t.Fatalf("failed to create .git directory: %v", err)
				}

				configContent := `[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
`
				configPath := filepath.Join(gitDir, "config")
				testutil.WriteTestFile(t, configPath, configContent)

				return tmpDir
			},
			checkFunc: func(t *testing.T, info *RepoInfo) {
				t.Helper()
				testutil.AssertEqual(t, true, info.IsGitRepo)
				testutil.AssertEqual(t, "", info.Organization)
				testutil.AssertEqual(t, "", info.Repository)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			testDir := tt.setupFunc(t, tmpDir)

			repoInfo, _ := DetectRepository(testDir)

			if repoInfo == nil {
				repoInfo = &RepoInfo{}
			}
			tt.checkFunc(t, repoInfo)
		})
	}
}

func TestParseGitHubURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		remoteURL    string
		expectedOrg  string
		expectedRepo string
	}{
		{
			name:         "HTTPS GitHub URL",
			remoteURL:    "https://github.com/owner/repo.git",
			expectedOrg:  "owner",
			expectedRepo: "repo",
		},
		{
			name:         "SSH GitHub URL",
			remoteURL:    "git@github.com:owner/repo.git",
			expectedOrg:  "owner",
			expectedRepo: "repo",
		},
		{
			name:         "GitHub URL without .git suffix",
			remoteURL:    "https://github.com/owner/repo",
			expectedOrg:  "owner",
			expectedRepo: "repo",
		},
		{
			name:         "Invalid URL",
			remoteURL:    "not-a-valid-url",
			expectedOrg:  "",
			expectedRepo: "",
		},
		{
			name:         "Empty URL",
			remoteURL:    "",
			expectedOrg:  "",
			expectedRepo: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			org, repo := parseGitHubURL(tt.remoteURL)

			testutil.AssertEqual(t, tt.expectedOrg, org)
			testutil.AssertEqual(t, tt.expectedRepo, repo)
		})
	}
}

func TestRepoInfo_GetRepositoryName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		repoInfo RepoInfo
		expected string
	}{
		{
			name:     "empty repo info",
			repoInfo: RepoInfo{},
			expected: "",
		},
		{
			name: "only organization set",
			repoInfo: RepoInfo{
				Organization: "owner",
			},
			expected: "",
		},
		{
			name: "only repository set",
			repoInfo: RepoInfo{
				Repository: "repo",
			},
			expected: "",
		},
		{
			name: "both organization and repository set",
			repoInfo: RepoInfo{
				Organization: "owner",
				Repository:   "repo",
			},
			expected: "owner/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.repoInfo.GetRepositoryName()
			testutil.AssertEqual(t, tt.expected, result)
		})
	}
}
