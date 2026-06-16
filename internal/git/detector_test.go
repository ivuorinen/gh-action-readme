package git

import (
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

const (
	testGitOrgOwner   = "owner"
	testGitRepoRepo   = "repo"
	testGitOrgActions = "actions"
	testGitRepoChkout = "checkout"
	testGitBranchMain = "main"
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
				testutil.SetupGitDirectory(t, tmpDir)

				// Create subdirectory to test from
				subDir := filepath.Join(tmpDir, "subdir", "nested")
				testutil.CreateTestDir(t, subDir)

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
			name: testutil.TestCaseNameNoGitRepository,
			setupFunc: func(t *testing.T, tmpDir string) string {
				t.Helper()
				// Create subdirectory without .git
				subDir := filepath.Join(tmpDir, "subdir")
				testutil.CreateTestDir(t, subDir)

				return subDir
			},
			expectError: true,
		},
		{
			name: testutil.TestCaseNameNonexistentDir,
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
				testutil.AssertFileExists(t, gitPath)
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
		createGitRepoTestCase(gitTestCase{
			name: "GitHub repository",
			configContent: `[core]
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
`,
			expectedOrg:  testGitOrgOwner,
			expectedRepo: testGitRepoRepo,
			expectedURL:  "https://github.com/" + testGitOrgOwner + "/" + testGitRepoRepo + ".git",
		}),
		createGitRepoTestCase(gitTestCase{
			name: "SSH remote URL",
			configContent: `[remote "origin"]
	url = git@github.com:owner/repo.git
	fetch = +refs/heads/*:refs/remotes/origin/*
`,
			expectedOrg:  testGitOrgOwner,
			expectedRepo: testGitRepoRepo,
			expectedURL:  "git@github.com:" + testGitOrgOwner + "/" + testGitRepoRepo + ".git",
		}),
		{
			name: testutil.TestCaseNameNoGitRepository,
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
		createGitRepoTestCase(gitTestCase{
			name: "git repository without origin remote",
			configContent: `[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
`,
			expectedOrg:  "",
			expectedRepo: "",
		}),
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
			remoteURL:    "https://github.com/" + testGitOrgOwner + "/" + testGitRepoRepo + ".git",
			expectedOrg:  testGitOrgOwner,
			expectedRepo: testGitRepoRepo,
		},
		{
			name:         testutil.TestCaseNameSSHGitHub,
			remoteURL:    "git@github.com:" + testGitOrgOwner + "/" + testGitRepoRepo + ".git",
			expectedOrg:  testGitOrgOwner,
			expectedRepo: testGitRepoRepo,
		},
		{
			name:         "GitHub URL without .git suffix",
			remoteURL:    "https://github.com/" + testGitOrgOwner + "/" + testGitRepoRepo,
			expectedOrg:  testGitOrgOwner,
			expectedRepo: testGitRepoRepo,
		},
		{
			name:         "HTTPS URL without .git suffix and dotted repo name",
			remoteURL:    "https://github.com/" + testGitOrgOwner + "/my.repo",
			expectedOrg:  testGitOrgOwner,
			expectedRepo: "my.repo",
		},
		{
			name:         "SSH URL with dotted repo name and .git suffix",
			remoteURL:    "git@github.com:" + testGitOrgOwner + "/my.repo.git",
			expectedOrg:  testGitOrgOwner,
			expectedRepo: "my.repo",
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

func TestRepoInfoGetRepositoryName(t *testing.T) {
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
				Organization: testGitOrgOwner,
			},
			expected: "",
		},
		{
			name: "only repository set",
			repoInfo: RepoInfo{
				Repository: testGitRepoRepo,
			},
			expected: "",
		},
		{
			name: "both organization and repository set",
			repoInfo: RepoInfo{
				Organization: testGitOrgOwner,
				Repository:   testGitRepoRepo,
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

// TestRepoInfoGenerateUsesStatement tests the GenerateUsesStatement method.
func TestRepoInfoGenerateUsesStatement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		repoInfo   *RepoInfo
		actionName string
		version    string
		expected   string
	}{
		{
			name: "repository-level action",
			repoInfo: &RepoInfo{
				Organization: testGitOrgActions,
				Repository:   testGitRepoChkout,
			},
			actionName: "",
			version:    "v3",
			expected:   testutil.TestActionCheckoutV3,
		},
		{
			name: "repository-level action with same name",
			repoInfo: &RepoInfo{
				Organization: testGitOrgActions,
				Repository:   testGitRepoChkout,
			},
			actionName: testGitRepoChkout,
			version:    "v3",
			expected:   testutil.TestActionCheckoutV3,
		},
		{
			name: testutil.TestCaseNameSubdirAction,
			repoInfo: &RepoInfo{
				Organization: testGitOrgActions,
				Repository:   "toolkit",
			},
			actionName: "cache",
			version:    "v2",
			expected:   "actions/toolkit/cache@v2",
		},
		{
			name: "without organization",
			repoInfo: &RepoInfo{
				Organization: "",
				Repository:   "",
			},
			actionName: "my-action",
			version:    "v1",
			expected:   "your-org/my-action@v1",
		},
		{
			name: "without organization and action name",
			repoInfo: &RepoInfo{
				Organization: "",
				Repository:   "",
			},
			actionName: "",
			version:    "v1",
			expected:   "your-org/your-action@v1",
		},
		{
			name: "with SHA version",
			repoInfo: &RepoInfo{
				Organization: testGitOrgActions,
				Repository:   testGitRepoChkout,
			},
			actionName: "",
			version:    "abc123def456",
			expected:   "actions/checkout@abc123def456",
		},
		{
			name: "with main branch",
			repoInfo: &RepoInfo{
				Organization: testGitOrgActions,
				Repository:   "setup-node",
			},
			actionName: "",
			version:    testGitBranchMain,
			expected:   "actions/setup-node@main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.repoInfo.GenerateUsesStatement(tt.actionName, tt.version)
			testutil.AssertEqual(t, tt.expected, result)
		})
	}
}

// TestGetDefaultBranch_Fallbacks tests branch detection fallback logic.
func TestGetDefaultBranchFallbacks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFunc      func(t *testing.T, tmpDir string) string
		expectedBranch string
	}{
		createDefaultBranchTestCase(defaultBranchTestCase{
			name:           "git config with main branch",
			branch:         testGitBranchMain,
			expectedBranch: testGitBranchMain,
		}),
		createDefaultBranchTestCase(defaultBranchTestCase{
			name:           "git config with master branch - returns main fallback",
			branch:         "master",
			expectedBranch: testGitBranchMain,
		}),
		createDefaultBranchTestCase(defaultBranchTestCase{
			name:           "git config with develop branch - returns main fallback",
			branch:         "develop",
			expectedBranch: testGitBranchMain,
		}),
		{
			name: "no git config - returns main fallback",
			setupFunc: func(t *testing.T, tmpDir string) string {
				t.Helper()
				_ = testutil.SetupGitDirectory(t, tmpDir)

				return tmpDir
			},
			expectedBranch: testGitBranchMain, // Falls back to "main" when git command fails
		},
		{
			name: "malformed git config - returns main fallback",
			setupFunc: func(t *testing.T, tmpDir string) string {
				t.Helper()
				gitDir := testutil.SetupGitDirectory(t, tmpDir)

				configContent := `[branch this is malformed`
				configPath := filepath.Join(gitDir, "config")
				testutil.WriteTestFile(t, configPath, configContent)

				return tmpDir
			},
			expectedBranch: testGitBranchMain, // Falls back to "main" when git command fails
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			repoDir := tt.setupFunc(t, tmpDir)
			branch := getDefaultBranch(repoDir)

			testutil.AssertEqual(t, tt.expectedBranch, branch)
		})
	}
}

// TestGetRemoteURL_AllSources tests all remote URL detection methods.
func TestGetRemoteURLAllSources(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFunc   func(t *testing.T, tmpDir string) string
		expectError bool
		expectedURL string
	}{
		createGitURLTestCase(gitURLTestCase{
			name: "remote from git config - https",
			configContent: `[remote "origin"]
	url = https://github.com/test/repo.git
`,
			expectError: false,
			expectedURL: "https://github.com/test/repo.git",
		}),
		createGitURLTestCase(gitURLTestCase{
			name: "remote from git config - ssh",
			configContent: `[remote "origin"]
	url = git@github.com:user/repo.git
`,
			expectError: false,
			expectedURL: "git@github.com:user/repo.git",
		}),
		createGitURLTestCase(gitURLTestCase{
			name: "multiple remotes - origin takes precedence",
			configContent: `[remote "upstream"]
	url = https://github.com/upstream/repo
[remote "origin"]
	url = https://github.com/origin/repo
`,
			expectError: false,
			expectedURL: "https://github.com/origin/repo",
		}),
		{
			name: "no remote configured",
			setupFunc: func(t *testing.T, tmpDir string) string {
				t.Helper()
				_ = testutil.SetupGitDirectory(t, tmpDir)

				return tmpDir
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			repoDir := tt.setupFunc(t, tmpDir)
			url, err := getRemoteURL(repoDir)

			if tt.expectError {
				testutil.AssertError(t, err)
			} else {
				testutil.AssertNoError(t, err)
				testutil.AssertEqual(t, tt.expectedURL, url)
			}
		})
	}
}

// TestGetRemoteURLFromConfig_EdgeCases tests git config parsing with edge cases.
func TestGetRemoteURLFromConfigEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		configContent string
		expectError   bool
		expectedURL   string
		description   string
	}{
		{
			name: "standard git config",
			configContent: `[remote "origin"]
	url = ` + testutil.TestURLGitHubUserRepo + `
`,
			expectError: false,
			expectedURL: testutil.TestURLGitHubUserRepo,
			description: "Standard git config",
		},
		{
			name: "config with comments",
			configContent: `# This is a comment
[remote "origin"]
	# Another comment
	url = ` + testutil.TestURLGitHubUserRepo + `
	fetch = +refs/heads/*:refs/remotes/origin/*
`,
			expectError: false,
			expectedURL: testutil.TestURLGitHubUserRepo,
			description: "Config with comments should be parsed",
		},
		{
			name:          "empty config",
			configContent: ``,
			expectError:   true,
			description:   "Empty config",
		},
		{
			name: "incomplete section",
			configContent: `[remote "origin"
	url = ` + testutil.TestURLGitHubUserRepo + `
`,
			expectError: true,
			description: "Malformed section",
		},
		{
			name: "url with spaces",
			configContent: `[remote "origin"]
	url = https://github.com/user name/repo name
`,
			expectError: false,
			expectedURL: "https://github.com/user name/repo name",
			description: "URL with spaces should be preserved",
		},
		{
			name: "multiple origin sections - first wins",
			configContent: `[remote "origin"]
	url = https://github.com/first/repo
[remote "origin"]
	url = https://github.com/second/repo
`,
			expectError: false,
			expectedURL: "https://github.com/first/repo",
			description: "First origin section takes precedence",
		},
		{
			name: "ssh url format",
			configContent: `[remote "origin"]
	url = git@gitlab.com:user/repo.git
`,
			expectError: false,
			expectedURL: "git@gitlab.com:user/repo.git",
			description: "SSH URL format",
		},
		{
			name: "url with trailing whitespace",
			configContent: `[remote "origin"]
	url = ` + testutil.TestURLGitHubUserRepo + `
`,
			expectError: false,
			expectedURL: testutil.TestURLGitHubUserRepo,
			description: "Trailing whitespace should be trimmed",
		},
		{
			name: "config without url field",
			configContent: `[remote "origin"]
	fetch = +refs/heads/*:refs/remotes/origin/*
`,
			expectError: true,
			description: "Remote without URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			gitDir := testutil.SetupGitDirectory(t, tmpDir)

			if tt.configContent != "" {
				configPath := filepath.Join(gitDir, "config")
				testutil.WriteTestFile(t, configPath, tt.configContent)
			}

			url, err := getRemoteURLFromConfig(tmpDir)

			if tt.expectError {
				testutil.AssertError(t, err)
			} else {
				testutil.AssertNoError(t, err)
				testutil.AssertEqual(t, tt.expectedURL, url)
			}
		})
	}
}

// TestFindRepositoryRoot_EdgeCases tests additional edge cases for repository root detection.
func TestFindRepositoryRootEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFunc   func(t *testing.T, tmpDir string) string
		expectError bool
		checkFunc   func(t *testing.T, tmpDir, repoRoot string)
	}{
		{
			name: "deeply nested subdirectory",
			setupFunc: func(t *testing.T, tmpDir string) string {
				t.Helper()
				testutil.SetupGitDirectory(t, tmpDir)

				deepPath := filepath.Join(tmpDir, "a", "b", "c", "d", "e")
				testutil.CreateTestDir(t, deepPath)

				return deepPath
			},
			expectError: false,
			checkFunc: func(t *testing.T, tmpDir, repoRoot string) {
				t.Helper()
				testutil.AssertEqual(t, tmpDir, repoRoot)
			},
		},
		{
			name: "git worktree with .git file",
			setupFunc: func(t *testing.T, tmpDir string) string {
				t.Helper()
				gitFile := filepath.Join(tmpDir, ".git")
				testutil.WriteTestFile(t, gitFile, "gitdir: /path/to/worktree")

				return tmpDir
			},
			expectError: false,
			checkFunc: func(t *testing.T, tmpDir, repoRoot string) {
				t.Helper()
				testutil.AssertEqual(t, tmpDir, repoRoot)
			},
		},
		{
			name: "current directory is repo root",
			setupFunc: func(t *testing.T, tmpDir string) string {
				t.Helper()
				testutil.SetupGitDirectory(t, tmpDir)

				return tmpDir
			},
			expectError: false,
			checkFunc: func(t *testing.T, tmpDir, repoRoot string) {
				t.Helper()
				testutil.AssertEqual(t, tmpDir, repoRoot)
			},
		},
		{
			name: "path with spaces",
			setupFunc: func(t *testing.T, tmpDir string) string {
				t.Helper()
				testutil.SetupGitDirectory(t, tmpDir)

				spacePath := filepath.Join(tmpDir, "path with spaces")
				testutil.CreateTestDir(t, spacePath)

				return spacePath
			},
			expectError: false,
			checkFunc: func(t *testing.T, tmpDir, repoRoot string) {
				t.Helper()
				testutil.AssertEqual(t, tmpDir, repoRoot)
			},
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
			} else {
				testutil.AssertNoError(t, err)
				tt.checkFunc(t, tmpDir, repoRoot)
			}
		})
	}
}

// TestParseGitHubURL_EdgeCases tests additional URL parsing edge cases.
func TestParseGitHubURLEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		remoteURL    string
		expectedOrg  string
		expectedRepo string
		description  string
	}{
		{
			name:         "gitlab https url",
			remoteURL:    "https://gitlab.com/owner/repo.git",
			expectedOrg:  "",
			expectedRepo: "",
			description:  "Non-GitHub URLs return empty",
		},
		{
			name:         "github url with subgroups",
			remoteURL:    "https://github.com/org/subgroup/repo.git",
			expectedOrg:  "org",
			expectedRepo: "subgroup", // Regex only captures first two path segments
			description:  "GitHub URLs with subpaths only capture org/subgroup",
		},
		{
			name:         "ssh url without git suffix",
			remoteURL:    "git@github.com:owner/repo",
			expectedOrg:  testGitOrgOwner,
			expectedRepo: testGitRepoRepo,
			description:  "SSH URL without .git suffix",
		},
		{
			name:         "url with trailing slash",
			remoteURL:    "https://github.com/owner/repo/",
			expectedOrg:  testGitOrgOwner,
			expectedRepo: testGitRepoRepo,
			description:  "Handles trailing slash",
		},
		{
			name:         "url with query parameters",
			remoteURL:    "https://github.com/owner/repo?param=value",
			expectedOrg:  testGitOrgOwner,
			expectedRepo: "repo?param=value", // Regex doesn't strip query params
			description:  "Query parameters are not stripped by regex",
		},
		{
			name:         "malformed ssh url",
			remoteURL:    "git@github.com/owner/repo.git",
			expectedOrg:  testGitOrgOwner,
			expectedRepo: testGitRepoRepo, // Actually matches the pattern
			description:  "Malformed SSH URL still matches pattern",
		},
		{
			name:         "url with username",
			remoteURL:    "https://user@github.com/owner/repo.git",
			expectedOrg:  testGitOrgOwner,
			expectedRepo: testGitRepoRepo,
			description:  "Handles URL with username",
		},
		{
			name:         "github enterprise url",
			remoteURL:    "https://github.company.com/owner/repo.git",
			expectedOrg:  "",
			expectedRepo: "",
			description:  "GitHub Enterprise URLs return empty (not github.com)",
		},
		{
			name:         "short ssh format",
			remoteURL:    "github.com:owner/repo.git",
			expectedOrg:  testGitOrgOwner,
			expectedRepo: testGitRepoRepo, // Actually matches the pattern with ':'
			description:  "Short SSH format matches the regex pattern",
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
