package validation

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

const (
	testValidationOrgOwner = "owner"
	testValidationRepoRepo = "repo"
)

func TestValidateActionYMLPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFunc   func(t *testing.T, tmpDir string) string
		expectError bool
		errorMsg    string // substring the error must contain (N156)
		wantErrIs   error  // sentinel the error must wrap (N156)
	}{
		{
			name: "valid action.yml file",
			setupFunc: func(t *testing.T, tmpDir string) string {
				t.Helper()

				return testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureJavaScriptSimple)
			},
			expectError: false,
		},
		{
			name: "valid action.yaml file",
			setupFunc: func(t *testing.T, tmpDir string) string {
				t.Helper()

				return testutil.WriteActionFixtureAs(
					t,
					tmpDir,
					appconstants.ActionFileNameYAML,
					testutil.TestFixtureMinimalAction,
				)
			},
			expectError: false,
		},
		{
			name: "nonexistent file",
			setupFunc: func(_ *testing.T, tmpDir string) string {
				return filepath.Join(tmpDir, testutil.TestNonexistentYML)
			},
			expectError: true,
			errorMsg:    "cannot access action file",
		},
		{
			name: "file with wrong extension",
			setupFunc: func(t *testing.T, tmpDir string) string {
				t.Helper()

				return testutil.WriteActionFixtureAs(t, tmpDir, "action.txt", testutil.TestFixtureJavaScriptSimple)
			},
			expectError: true,
			wantErrIs:   os.ErrInvalid,
		},
		{
			name: "empty file path",
			setupFunc: func(_ *testing.T, _ string) string {
				return ""
			},
			expectError: true,
			errorMsg:    "cannot access action file",
		},
		{
			name: "directory named action.yml is rejected",
			setupFunc: func(t *testing.T, tmpDir string) string {
				t.Helper()
				dirPath := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
				if err := os.Mkdir(dirPath, appconstants.FilePermDir); err != nil {
					t.Fatalf("failed to create directory: %v", err)
				}

				return dirPath
			},
			expectError: true,
			errorMsg:    "is a directory, not a file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			actionPath := tt.setupFunc(t, tmpDir)

			err := ValidateActionYMLPath(actionPath)

			if tt.expectError {
				testutil.AssertError(t, err)
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got: %v", tt.errorMsg, err)
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("expected error wrapping %v, got: %v", tt.wantErrIs, err)
				}
			} else {
				testutil.AssertNoError(t, err)
			}
		})
	}
}

func TestIsCommitSHA(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{
			name:     "full commit SHA",
			version:  testutil.TestSHAForTesting,
			expected: true,
		},
		{
			name:     testutil.TestCaseNameShortCommitSHA,
			version:  "8f4b7f8",
			expected: true,
		},
		{
			name:     testutil.TestCaseNameSemanticVersion,
			version:  testutil.TestVersionSemantic,
			expected: false,
		},
		{
			name:     testutil.TestCaseNameBranchName,
			version:  testutil.TestBranchMain,
			expected: false,
		},
		{
			name:     testutil.TestCaseNameEmpty,
			version:  "",
			expected: false,
		},
		{
			name:     "non-hex characters",
			version:  "not-a-sha",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := IsCommitSHA(tt.version)
			testutil.AssertEqual(t, tt.expected, result)
		})
	}
}

func TestIsSemanticVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{
			name:     "semantic version with v prefix",
			version:  testutil.TestVersionSemantic,
			expected: true,
		},
		{
			name:     "semantic version without v prefix",
			version:  testutil.TestVersionPlain,
			expected: true,
		},
		{
			name:     "semantic version with prerelease",
			version:  "v1.2.3-alpha.1",
			expected: true,
		},
		{
			name:     "semantic version with build metadata",
			version:  "v1.2.3+20230101",
			expected: true,
		},
		{
			name:     testutil.TestCaseNameMajorVersionOnly,
			version:  appconstants.VersionTagV1,
			expected: false,
		},
		{
			name:     testutil.TestCaseNameCommitSHA,
			version:  testutil.TestSHAForTesting,
			expected: false,
		},
		{
			name:     testutil.TestCaseNameBranchName,
			version:  testutil.TestBranchMain,
			expected: false,
		},
		{
			name:     testutil.TestCaseNameEmpty,
			version:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := IsSemanticVersion(tt.version)
			testutil.AssertEqual(t, tt.expected, result)
		})
	}
}

func TestIsVersionPinned(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{
			name:     "full semantic version",
			version:  testutil.TestVersionSemantic,
			expected: true,
		},
		{
			name:     "full commit SHA",
			version:  testutil.TestSHAForTesting,
			expected: true,
		},
		{
			name:     testutil.TestCaseNameMajorVersionOnly,
			version:  appconstants.VersionTagV1,
			expected: false,
		},
		{
			name:     "major.minor version",
			version:  "v1.2",
			expected: false,
		},
		{
			name:     testutil.TestCaseNameBranchName,
			version:  testutil.TestBranchMain,
			expected: false,
		},
		{
			name:     testutil.TestCaseNameShortCommitSHA,
			version:  "8f4b7f8",
			expected: false,
		},
		{
			name:     testutil.TestCaseNameEmpty,
			version:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := IsVersionPinned(tt.version)
			testutil.AssertEqual(t, tt.expected, result)
		})
	}
}

// initGitRepo creates a real git repository with one commit and returns the
// default branch name (e.g. "main" or "master" depending on git config).
func initGitRepo(t *testing.T, dir string) string {
	t.Helper()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...) // #nosec G204 -- test helper, args are hardcoded
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			t.Fatalf("git command %v failed: %v", args, err)
		}
	}

	run("git", "init")
	run("git", "config", "user.email", "test@example.com")
	run("git", "config", "user.name", "Test")
	// Disable commit/tag signing so a developer's global gpgsign (e.g. a
	// 1Password/SSH signer) is not invoked; it fails non-interactively.
	run("git", "config", "commit.gpgsign", "false")
	run("git", "config", "tag.gpgsign", "false")
	testutil.WriteTestFile(t, filepath.Join(dir, "README.md"), "# test")
	run("git", "add", ".")
	run("git", "commit", "-m", "init")

	out, err := exec.Command("git", "-C", dir, "branch", "--show-current").Output() // #nosec G204
	if err != nil {
		t.Fatalf("git branch --show-current failed: %v", err)
	}

	return strings.TrimSpace(string(out))
}

func TestValidateGitBranch(t *testing.T) {
	t.Parallel()

	t.Run("existing branch returns true", func(t *testing.T) {
		t.Parallel()

		tmpDir, cleanup := testutil.TempDir(t)
		defer cleanup()

		branch := initGitRepo(t, tmpDir)
		testutil.AssertEqual(t, true, ValidateGitBranch(tmpDir, branch))
	})

	t.Run("non-existent branch returns false", func(t *testing.T) {
		t.Parallel()

		tmpDir, cleanup := testutil.TempDir(t)
		defer cleanup()

		initGitRepo(t, tmpDir)
		testutil.AssertEqual(t, false, ValidateGitBranch(tmpDir, "definitely-nonexistent-branch"))
	})

	t.Run("non-git directory returns false", func(t *testing.T) {
		t.Parallel()

		tmpDir, cleanup := testutil.TempDir(t)
		defer cleanup()

		testutil.AssertEqual(t, false, ValidateGitBranch(tmpDir, testutil.TestBranchMain))
	})

	t.Run("empty branch name returns false", func(t *testing.T) {
		t.Parallel()

		tmpDir, cleanup := testutil.TempDir(t)
		defer cleanup()

		testutil.AssertEqual(t, false, ValidateGitBranch(tmpDir, ""))
	})
}

func TestIsGitRepository(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFunc func(t *testing.T, tmpDir string) string
		expected  bool
	}{
		{
			name: "directory with .git folder",
			setupFunc: func(_ *testing.T, tmpDir string) string {
				_ = testutil.SetupGitDirectory(t, tmpDir)

				return tmpDir
			},
			expected: true,
		},
		{
			name: "directory with .git file",
			setupFunc: func(t *testing.T, tmpDir string) string {
				t.Helper()
				gitFile := filepath.Join(tmpDir, appconstants.DirGit)
				testutil.WriteTestFile(t, gitFile, "gitdir: /path/to/git/dir")

				return tmpDir
			},
			expected: true,
		},
		{
			name: "directory without .git",
			setupFunc: func(_ *testing.T, tmpDir string) string {
				return tmpDir
			},
			expected: false,
		},
		{
			name: "nonexistent path",
			setupFunc: func(_ *testing.T, tmpDir string) string {
				return filepath.Join(tmpDir, "nonexistent")
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			testPath := tt.setupFunc(t, tmpDir)
			result := IsGitRepository(testPath)
			testutil.AssertEqual(t, tt.expected, result)
		})
	}
}

func TestCleanVersionString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "version with v prefix",
			input:    testutil.TestVersionSemantic,
			expected: testutil.TestVersionPlain,
		},
		{
			name:     "version without v prefix",
			input:    testutil.TestVersionPlain,
			expected: testutil.TestVersionPlain,
		},
		{
			name:     "version with leading/trailing spaces",
			input:    "  v1.2.3  ",
			expected: testutil.TestVersionPlain,
		},
		{
			name:     testutil.TestCaseNameEmpty,
			input:    "",
			expected: "",
		},
		{
			name:     testutil.TestCaseNameCommitSHA,
			input:    testutil.TestSHAForTesting,
			expected: testutil.TestSHAForTesting,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := CleanVersionString(tt.input)
			testutil.AssertEqual(t, tt.expected, result)
		})
	}
}

func TestParseGitHubURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		url          string
		expectedOrg  string
		expectedRepo string
	}{
		{
			name:         "HTTPS GitHub URL",
			url:          "https://github.com/" + testValidationOrgOwner + "/" + testValidationRepoRepo,
			expectedOrg:  testValidationOrgOwner,
			expectedRepo: testValidationRepoRepo,
		},
		{
			name:         "GitHub URL with .git suffix",
			url:          "https://github.com/" + testValidationOrgOwner + "/" + testValidationRepoRepo + ".git",
			expectedOrg:  testValidationOrgOwner,
			expectedRepo: testValidationRepoRepo,
		},
		{
			name:         testutil.TestCaseNameSSHGitHub,
			url:          "git@github.com:" + testValidationOrgOwner + "/" + testValidationRepoRepo + ".git",
			expectedOrg:  testValidationOrgOwner,
			expectedRepo: testValidationRepoRepo,
		},
		{
			name:         "Invalid URL",
			url:          "not-a-url",
			expectedOrg:  "",
			expectedRepo: "",
		},
		{
			name:         "Empty URL",
			url:          "",
			expectedOrg:  "",
			expectedRepo: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			org, repo := ParseGitHubURL(tt.url)
			testutil.AssertEqual(t, tt.expectedOrg, org)
			testutil.AssertEqual(t, tt.expectedRepo, repo)
		})
	}
}

func TestSanitizeActionName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercases and hyphenates spaces",
			input:    "My Action",
			expected: "my-action",
		},
		{
			name:     "trims surrounding whitespace",
			input:    "  My Action  ",
			expected: "my-action",
		},
		{
			name:     "uppercase is lowered",
			input:    "ALLCAPS",
			expected: "allcaps",
		},
		{
			// Documents actual behavior: each space maps to one hyphen; runs of
			// spaces are NOT collapsed, so two spaces yield two hyphens.
			name:     "consecutive spaces are not collapsed",
			input:    "My  Action",
			expected: "my--action",
		},
		{
			// N155: '/' must not survive — it would inject an extra path segment
			// into the generated uses: statement (your-org/my-sub-action@v1).
			name:     "slash is replaced, not preserved",
			input:    "My/Sub Action",
			expected: "my-sub-action",
		},
		{
			name:     "ampersand and other unsafe chars become hyphens",
			input:    "Foo & Bar",
			expected: "foo---bar",
		},
		{
			name:     "underscores and dots are preserved",
			input:    "my_action.v2",
			expected: "my_action.v2",
		},
		{
			name:     testutil.TestCaseNameEmpty,
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := SanitizeActionName(tt.input); got != tt.expected {
				t.Errorf("SanitizeActionName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGetBinaryDir(t *testing.T) {
	t.Parallel()

	dir, err := GetBinaryDir()
	testutil.AssertNoError(t, err)

	if dir == "" {
		t.Error("expected non-empty binary directory")
	}

	// Verify the directory exists
	testutil.AssertFileExists(t, dir)
}

func TestEnsureAbsolutePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		isAbsolute bool
	}{
		{
			name:       "absolute path",
			input:      "/path/to/file",
			isAbsolute: true,
		},
		{
			name:       testutil.TestCaseNameRelativePath,
			input:      "./file",
			isAbsolute: false,
		},
		{
			name:       "just filename",
			input:      "file.txt",
			isAbsolute: false,
		},
		{
			name:       testutil.TestCaseNameEmptyPath,
			input:      "",
			isAbsolute: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := EnsureAbsolutePath(tt.input)

			if tt.input == "" {
				// Empty input might cause an error
				if err != nil {
					return // This is acceptable
				}
			} else {
				testutil.AssertNoError(t, err)
			}

			// Result should always be absolute
			if result != "" && !filepath.IsAbs(result) {
				t.Errorf("expected absolute path, got: %s", result)
			}
		})
	}
}
