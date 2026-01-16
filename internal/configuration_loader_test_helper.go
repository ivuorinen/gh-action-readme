package internal

import (
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

// repoOverrideTestCase defines the structure for repository override test cases.
type repoOverrideTestCase struct {
	name           string
	setupFunc      func(t *testing.T) (config *AppConfig, repoRoot string)
	expectedTheme  string
	expectedFormat string
	description    string
}

// runRepoOverrideTest executes a test case for repository override functionality.
// This helper reduces duplication in TestConfigurationLoaderApplyRepoOverrides tests.
func runRepoOverrideTest(t *testing.T, tc repoOverrideTestCase) {
	t.Helper()

	config, repoRoot := tc.setupFunc(t)

	loader := NewConfigurationLoader()
	loader.applyRepoOverrides(config, repoRoot)

	// Verify expected values
	testutil.AssertEqual(t, tc.expectedTheme, config.Theme)
	testutil.AssertEqual(t, tc.expectedFormat, config.OutputFormat)
}

// repoOverrideTestParams holds parameters for creating repo override test cases.
type repoOverrideTestParams struct {
	name, remoteURL, overrideKey  string
	overrideTheme, overrideFormat string
	expectedTheme, expectedFormat string
	description                   string
}

// createRepoOverrideTestCase creates a repo override test case with git repo setup.
// This helper reduces duplication when creating test cases that need git repositories.
func createRepoOverrideTestCase(params repoOverrideTestParams) repoOverrideTestCase {
	return repoOverrideTestCase{
		name: params.name,
		setupFunc: func(t *testing.T) (*AppConfig, string) {
			t.Helper()
			tmpDir, _ := testutil.TempDir(t)

			if params.remoteURL != "" {
				testutil.CreateGitRepoWithRemote(t, tmpDir, params.remoteURL)
			}

			config := &AppConfig{
				Theme:        testutil.TestThemeDefault,
				OutputFormat: "md",
				RepoOverrides: map[string]AppConfig{
					params.overrideKey: {
						Theme:        params.overrideTheme,
						OutputFormat: params.overrideFormat,
					},
				},
			}

			return config, tmpDir
		},
		expectedTheme:  params.expectedTheme,
		expectedFormat: params.expectedFormat,
		description:    params.description,
	}
}

// configLoaderTestCase defines the structure for configuration loader test cases.
type configLoaderTestCase struct {
	name        string
	setupFunc   func(t *testing.T) string
	expectError bool
	checkFunc   func(t *testing.T, config *AppConfig)
	description string
}

// runConfigLoaderTest executes a test case for configuration loading functionality.
// This helper reduces duplication between LoadGlobalConfig and loadActionConfig tests.
func runConfigLoaderTest(
	t *testing.T,
	tc configLoaderTestCase,
	loadFunc func(loader *ConfigurationLoader, path string) (*AppConfig, error),
) {
	t.Helper()
	t.Parallel()

	path := tc.setupFunc(t)

	loader := NewConfigurationLoader()
	config, err := loadFunc(loader, path)

	if tc.expectError {
		testutil.AssertError(t, err)
	} else {
		testutil.AssertNoError(t, err)

		if tc.checkFunc != nil {
			tc.checkFunc(t, config)
		}
	}
}

// checkThemeAndFormat is a helper that creates a checkFunc for verifying theme and output format.
// This reduces duplication in test cases that only need to verify these two fields.
func checkThemeAndFormat(expectedTheme, expectedFormat string) func(t *testing.T, config *AppConfig) {
	return func(t *testing.T, config *AppConfig) {
		t.Helper()
		testutil.AssertEqual(t, expectedTheme, config.Theme)
		testutil.AssertEqual(t, expectedFormat, config.OutputFormat)
	}
}
