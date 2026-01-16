package git

import (
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

// gitTestCase defines the configuration for a git repository test case.
type gitTestCase struct {
	name           string
	configContent  string
	expectedOrg    string
	expectedRepo   string
	expectedBranch string
	expectedURL    string
}

// createGitRepoTestCase creates a test table entry for git repository detection tests.
// setupGitTestRepo creates a test git directory with the specified config content.
// This helper is used by multiple test case creators to eliminate duplicate setup logic.
func setupGitTestRepo(t *testing.T, tmpDir, configContent string) string {
	t.Helper()
	gitDir := testutil.SetupGitDirectory(t, tmpDir)
	configPath := filepath.Join(gitDir, "config")
	testutil.WriteTestFile(t, configPath, configContent)

	return tmpDir
}

// This helper reduces duplication by standardizing the setup and assertion patterns
// for git repository test cases.
func createGitRepoTestCase(tc gitTestCase) struct {
	name      string
	setupFunc func(t *testing.T, tmpDir string) string
	checkFunc func(t *testing.T, info *RepoInfo)
} {
	return struct {
		name      string
		setupFunc func(t *testing.T, tmpDir string) string
		checkFunc func(t *testing.T, info *RepoInfo)
	}{
		name: tc.name,
		setupFunc: func(t *testing.T, tmpDir string) string {
			t.Helper()

			return setupGitTestRepo(t, tmpDir, tc.configContent)
		},
		checkFunc: func(t *testing.T, info *RepoInfo) {
			t.Helper()
			testutil.AssertEqual(t, tc.expectedOrg, info.Organization)
			testutil.AssertEqual(t, tc.expectedRepo, info.Repository)
			if tc.expectedBranch != "" {
				testutil.AssertEqual(t, tc.expectedBranch, info.DefaultBranch)
			}
			if tc.expectedURL != "" {
				testutil.AssertEqual(t, tc.expectedURL, info.RemoteURL)
			}
		},
	}
}

// gitURLTestCase defines the configuration for git remote URL test cases.
type gitURLTestCase struct {
	name          string
	configContent string
	expectError   bool
	expectedURL   string
}

// createGitURLTestCase creates a test table entry for git remote URL detection tests.
// This helper reduces duplication by standardizing the setup pattern for URL tests.
func createGitURLTestCase(tc gitURLTestCase) struct {
	name        string
	setupFunc   func(t *testing.T, tmpDir string) string
	expectError bool
	expectedURL string
} {
	return struct {
		name        string
		setupFunc   func(t *testing.T, tmpDir string) string
		expectError bool
		expectedURL string
	}{
		name: tc.name,
		setupFunc: func(t *testing.T, tmpDir string) string {
			t.Helper()

			return setupGitTestRepo(t, tmpDir, tc.configContent)
		},
		expectError: tc.expectError,
		expectedURL: tc.expectedURL,
	}
}

// defaultBranchTestCase defines the configuration for default branch detection tests.
type defaultBranchTestCase struct {
	name           string
	branch         string
	expectedBranch string
}

// createDefaultBranchTestCase creates a test table entry for default branch tests.
// This helper reduces duplication for tests that set up git repos with different branches.
func createDefaultBranchTestCase(tc defaultBranchTestCase) struct {
	name           string
	setupFunc      func(t *testing.T, tmpDir string) string
	expectedBranch string
} {
	return struct {
		name           string
		setupFunc      func(t *testing.T, tmpDir string) string
		expectedBranch string
	}{
		name: tc.name,
		setupFunc: func(t *testing.T, tmpDir string) string {
			t.Helper()
			gitDir := testutil.SetupGitDirectory(t, tmpDir)
			testutil.CreateGitConfigWithRemote(t, gitDir, testutil.TestURLGitHubUserRepo, tc.branch)

			return tmpDir
		},
		expectedBranch: tc.expectedBranch,
	}
}
