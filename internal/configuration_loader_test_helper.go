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

// createRepoOverrideTestCase creates a repo override test case with git repo setup.
// This helper reduces duplication when creating test cases that need git repositories.
func createRepoOverrideTestCase(
	name, remoteURL, overrideKey string,
	overrideTheme, overrideFormat string,
	expectedTheme, expectedFormat string,
	description string,
) repoOverrideTestCase {
	return repoOverrideTestCase{
		name: name,
		setupFunc: func(t *testing.T) (*AppConfig, string) {
			t.Helper()
			tmpDir, _ := testutil.TempDir(t)

			if remoteURL != "" {
				testutil.CreateGitRepoWithRemote(t, tmpDir, remoteURL)
			}

			config := &AppConfig{
				Theme:        testutil.TestThemeDefault,
				OutputFormat: "md",
				RepoOverrides: map[string]AppConfig{
					overrideKey: {
						Theme:        overrideTheme,
						OutputFormat: overrideFormat,
					},
				},
			}

			return config, tmpDir
		},
		expectedTheme:  expectedTheme,
		expectedFormat: expectedFormat,
		description:    description,
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
