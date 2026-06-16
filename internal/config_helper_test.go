package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-github/v74/github"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

// boolFields represents the default-false boolean configuration fields used in
// merge tests. UseDefaultBranch is intentionally excluded: it defaults to true and
// merges on explicit presence (see TestMergeBooleanFieldsUseDefaultBranchPresence),
// which the "merge if true" table here cannot model.
type boolFields struct {
	AnalyzeDependencies bool
	ShowSecurityInfo    bool
	Verbose             bool
	Quiet               bool
}

// createBoolFieldMergeTest creates a test table entry for testing boolean field merging.
// This helper reduces duplication by standardizing the creation of AppConfig test structures
// with boolean fields.
func createBoolFieldMergeTest(name string, dst, src, want boolFields) struct {
	name string
	dst  *AppConfig
	src  *AppConfig
	want *AppConfig
} {
	return struct {
		name string
		dst  *AppConfig
		src  *AppConfig
		want *AppConfig
	}{
		name: name,
		dst: &AppConfig{
			AnalyzeDependencies: dst.AnalyzeDependencies,
			ShowSecurityInfo:    dst.ShowSecurityInfo,
			Verbose:             dst.Verbose,
			Quiet:               dst.Quiet,
		},
		src: &AppConfig{
			AnalyzeDependencies: src.AnalyzeDependencies,
			ShowSecurityInfo:    src.ShowSecurityInfo,
			Verbose:             src.Verbose,
			Quiet:               src.Quiet,
		},
		want: &AppConfig{
			AnalyzeDependencies: want.AnalyzeDependencies,
			ShowSecurityInfo:    want.ShowSecurityInfo,
			Verbose:             want.Verbose,
			Quiet:               want.Quiet,
		},
	}
}

// createGitRemoteTestCase creates a test table entry for git remote detection tests.
// This helper reduces duplication for tests that set up a git repo with a remote config.
func createGitRemoteTestCase(
	name, configContent, expectedResult, description string,
) struct {
	name           string
	setupFunc      func(t *testing.T) string
	expectedResult string
	description    string
} {
	return struct {
		name           string
		setupFunc      func(t *testing.T) string
		expectedResult string
		description    string
	}{
		name: name,
		setupFunc: func(t *testing.T) string {
			t.Helper()
			tmpDir, _ := testutil.TempDir(t)
			testutil.InitGitRepo(t, tmpDir)

			if configContent != "" {
				configPath := filepath.Join(tmpDir, testutil.ConfigFieldGit, "config")
				testutil.WriteTestFile(t, configPath, configContent)
			}

			return tmpDir
		},
		expectedResult: expectedResult,
		description:    description,
	}
}

// createTokenMergeTest creates a test table entry for testing token merging behavior.
// This helper reduces duplication for the 4 token merge test cases.
func createTokenMergeTest(
	name, dstToken, srcToken, wantToken string,
	allowTokens bool,
) struct {
	name        string
	dst         *AppConfig
	src         *AppConfig
	allowTokens bool
	want        *AppConfig
} {
	return struct {
		name        string
		dst         *AppConfig
		src         *AppConfig
		allowTokens bool
		want        *AppConfig
	}{
		name:        name,
		dst:         &AppConfig{GitHubToken: dstToken},
		src:         &AppConfig{GitHubToken: srcToken},
		allowTokens: allowTokens,
		want:        &AppConfig{GitHubToken: wantToken},
	}
}

// createMapMergeTest creates a test table entry for testing map field merging (permissions/variables).
// This helper reduces duplication for tests that merge map[string]string fields.
func createMapMergeTest(
	name string,
	dstMap, srcMap, expectedMap map[string]string,
	isPermissions bool,
) struct {
	name     string
	dst      *AppConfig
	src      *AppConfig
	expected *AppConfig
} {
	dst := &AppConfig{}
	src := &AppConfig{}
	expected := &AppConfig{}

	if isPermissions {
		dst.Permissions = dstMap
		src.Permissions = srcMap
		expected.Permissions = expectedMap
	} else {
		dst.Variables = dstMap
		src.Variables = srcMap
		expected.Variables = expectedMap
	}

	return struct {
		name     string
		dst      *AppConfig
		src      *AppConfig
		expected *AppConfig
	}{
		name:     name,
		dst:      dst,
		src:      src,
		expected: expected,
	}
}

// ConfigHierarchySetup contains fixture paths for creating a multi-level config hierarchy.
type ConfigHierarchySetup struct {
	GlobalFixture string // Fixture path for global config
	RepoFixture   string // Fixture path for repo config
	ActionFixture string // Fixture path for action config
}

// SetupConfigHierarchy creates a multi-level config hierarchy (global/repo/action).
// Returns global config path, repo root, and action directory.
//
// Example:
//
//	globalPath, repoRoot, actionDir := SetupConfigHierarchy(t, tmpDir, ConfigHierarchySetup{
//		GlobalFixture: testutil.TestConfigGlobalDefault,
//		RepoFixture:   testutil.TestConfigRepoSimple,
//		ActionFixture: testutil.TestConfigActionSimple,
//	})
func SetupConfigHierarchy(
	t *testing.T,
	baseDir string,
	setup ConfigHierarchySetup,
) (globalConfigPath, repoRoot, actionDir string) {
	t.Helper()
	// setupAndCreateConfigFixtures sets up config fixtures in a test directory.
	// It creates the repo directory structure unconditionally and populates config files
	// based on the provided setup.GlobalFixture, setup.RepoFixture, and
	// setup.ActionFixture. Returns globalConfigPath, repoRoot, and actionDir.

	// Create global config
	if setup.GlobalFixture != "" {
		globalConfigDir := filepath.Join(baseDir, testutil.TestDirDotConfig, testutil.TestBinaryName)
		globalConfigPath = testutil.WriteFileInDir(
			t, globalConfigDir, testutil.TestFileConfigYAML,
			testutil.MustReadFixture(setup.GlobalFixture),
		)
	}

	// Create repo config
	repoRoot = filepath.Join(baseDir, testutil.ConfigFieldRepo)
	if err := os.MkdirAll(repoRoot, 0o700); err != nil {
		t.Fatalf("failed to create repo directory: %v", err)
	}
	if setup.RepoFixture != "" {
		testutil.WriteFileInDir(
			t, repoRoot, testutil.TestFileGHReadmeYAML,
			testutil.MustReadFixture(setup.RepoFixture),
		)
	}

	// Create action config
	if setup.ActionFixture != "" {
		actionDir = filepath.Join(repoRoot, testutil.ConfigFieldAction)
		testutil.WriteFileInDir(
			t, actionDir, testutil.TestFileConfigYAML,
			testutil.MustReadFixture(setup.ActionFixture),
		)
	} else {
		actionDir = repoRoot
	}

	return globalConfigPath, repoRoot, actionDir
}

// WriteConfigFixture writes a config fixture to a directory with standard config filename.
// Returns the full path to the written config file.
//
// Example:
//
//	configPath := WriteConfigFixture(t, tmpDir, testutil.TestConfigGlobalDefault)
func WriteConfigFixture(t *testing.T, dir, fixturePath string) string {
	t.Helper()

	return testutil.WriteFileInDir(
		t, dir, testutil.TestFileConfigYAML,
		testutil.MustReadFixture(fixturePath),
	)
}

// ExpectedConfig holds expected values for config field assertions.
// Only non-zero values will be checked.
type ExpectedConfig struct {
	Theme        string
	OutputFormat string
	OutputDir    string
	Template     string
	Schema       string
	Verbose      bool
	Quiet        bool
	GitHubToken  string
}

// AssertConfigFields asserts that config matches expected values for all non-empty fields.
// Only checks fields that are set in expected (non-zero values).
//
// Example:
//
//	AssertConfigFields(t, config, ExpectedConfig{
//		Theme:        testutil.TestThemeDefault,
//		OutputFormat: "md",
//		Verbose:      true,
//	})
func AssertConfigFields(t *testing.T, config *AppConfig, expected ExpectedConfig) {
	t.Helper()
	if expected.Theme != "" {
		testutil.AssertEqual(t, expected.Theme, config.Theme)
	}
	if expected.OutputFormat != "" {
		testutil.AssertEqual(t, expected.OutputFormat, config.OutputFormat)
	}
	if expected.OutputDir != "" {
		testutil.AssertEqual(t, expected.OutputDir, config.OutputDir)
	}
	if expected.Template != "" {
		testutil.AssertEqual(t, expected.Template, config.Template)
	}
	if expected.Schema != "" {
		testutil.AssertEqual(t, expected.Schema, config.Schema)
	}
	// Always check booleans (they have meaningful zero values)
	testutil.AssertEqual(t, expected.Verbose, config.Verbose)
	testutil.AssertEqual(t, expected.Quiet, config.Quiet)
	if expected.GitHubToken != "" {
		testutil.AssertEqual(t, expected.GitHubToken, config.GitHubToken)
	}
}

// assertGitHubClient validates GitHub client creation results.
// This helper reduces test code duplication by centralizing
// the client validation logic for github.Client instances.
func assertGitHubClient(t *testing.T, client *github.Client, err error, expectError bool) {
	t.Helper()

	if expectError {
		if err == nil {
			t.Error(testutil.TestErrNoErrorGotNone)
		}
		if client != nil {
			t.Error("expected nil client on error")
		}

		return
	}

	// Success case
	if err != nil {
		t.Errorf(testutil.TestErrUnexpected, err)
	}
	if client == nil {
		t.Error("expected non-nil client")
	}
}
