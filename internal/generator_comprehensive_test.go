package internal

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

const (
	testGenThemeDefault = appconstants.ThemeDefault
	testGenThemeGitHub  = appconstants.ThemeGitHub
	testGenThemeMinimal = appconstants.ThemeMinimal
	testGenFormatMD     = appconstants.OutputFormatMarkdown
	testGenFormatHTML   = appconstants.OutputFormatHTML
	testGenFormatJSON   = appconstants.OutputFormatJSON
)

// TestGeneratorComprehensiveGeneration demonstrates the new table-driven testing framework
// by testing generation across all fixtures, themes, and formats systematically.
func TestGeneratorComprehensiveGeneration(t *testing.T) {
	t.Parallel()
	// Create test cases using the new helper functions
	cases := testutil.CreateGeneratorTestCases()

	// Filter to a subset for demonstration (full test would be very large)
	filteredCases := make([]testutil.GeneratorTestCase, 0)
	for _, testCase := range cases {
		// Only test a few combinations for demonstration
		if (testCase.Theme == testGenThemeDefault && testCase.OutputFormat == testGenFormatMD) ||
			(testCase.Theme == testGenThemeGitHub && testCase.OutputFormat == testGenFormatHTML) ||
			(testCase.Theme == testGenThemeMinimal && testCase.OutputFormat == testGenFormatJSON) {
			// Add custom executor for generator tests
			testCase.Executor = createGeneratorTestExecutor()
			filteredCases = append(filteredCases, testCase)
		}
	}

	// Run the test suite
	testutil.RunGeneratorTests(t, filteredCases)
}

// TestGeneratorAllValidFixtures tests generation with all valid fixtures.
func TestGeneratorAllValidFixtures(t *testing.T) {
	t.Parallel()
	validFixtures := testutil.GetValidFixtures()

	for _, fixture := range validFixtures {
		fixture := fixture // capture loop variable
		t.Run(fixture, func(t *testing.T) {
			t.Parallel()

			// Create temporary action from fixture
			actionPath := testutil.CreateTemporaryAction(t, fixture)

			// Test with default configuration
			config := &AppConfig{
				Theme:        testGenThemeDefault,
				OutputFormat: testGenFormatMD,
				OutputDir:    ".",
				Quiet:        true,
			}

			generator := NewGenerator(config)

			// Generate documentation
			err := generator.GenerateFromFile(actionPath)
			if err != nil {
				t.Errorf("failed to generate documentation for fixture %s: %v", fixture, err)
			}
		})
	}
}

// TestGeneratorAllInvalidFixtures tests that invalid fixtures produce expected errors.
func TestGeneratorAllInvalidFixtures(t *testing.T) {
	t.Parallel()
	invalidFixtures := testutil.GetInvalidFixtures()

	for _, fixture := range invalidFixtures {
		fixture := fixture // capture loop variable
		t.Run(fixture, func(t *testing.T) {
			t.Parallel()

			// Some invalid fixtures might not be loadable
			actionFixture, err := testutil.LoadActionFixture(fixture)
			if err != nil {
				// This is expected for some invalid fixtures
				return
			}

			// Create temporary action from fixture
			tempDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			testutil.WriteTestFile(t, tempDir+"/action.yml", actionFixture.Content)

			// Test with default configuration
			config := &AppConfig{
				Theme:        testGenThemeDefault,
				OutputFormat: testGenFormatMD,
				OutputDir:    ".",
				Quiet:        true,
			}

			generator := NewGenerator(config)

			// Generate documentation - should fail
			err = generator.GenerateFromFile(tempDir + "/action.yml")
			if err == nil {
				t.Errorf("expected generation to fail for invalid fixture %s, but it succeeded", fixture)
			}
		})
	}
}

// TestGeneratorAllThemes demonstrates theme testing using helper functions.
func TestGeneratorAllThemes(t *testing.T) {
	t.Parallel()
	// Use the helper function to test all themes
	testutil.TestAllThemes(t, func(t *testing.T, theme string) {
		t.Helper()
		// Create a simple action for testing
		actionPath := testutil.CreateTemporaryAction(t, testutil.TestFixtureJavaScriptSimple)

		config := &AppConfig{
			Theme:        theme,
			OutputFormat: testGenFormatMD,
			OutputDir:    ".",
			Quiet:        true,
		}

		generator := NewGenerator(config)
		err := generator.GenerateFromFile(actionPath)

		testutil.AssertNoError(t, err)
	})
}

// TestGeneratorAllFormats demonstrates format testing using helper functions.
func TestGeneratorAllFormats(t *testing.T) {
	t.Parallel()
	// Use the helper function to test all formats
	testutil.TestAllFormats(t, func(t *testing.T, format string) {
		t.Helper()
		// Create a simple action for testing
		actionPath := testutil.CreateTemporaryAction(t, testutil.TestFixtureJavaScriptSimple)

		config := &AppConfig{
			Theme:        testGenThemeDefault,
			OutputFormat: format,
			OutputDir:    ".",
			Quiet:        true,
		}

		generator := NewGenerator(config)
		err := generator.GenerateFromFile(actionPath)

		testutil.AssertNoError(t, err)
	})
}

// TestGeneratorByActionType demonstrates testing by action type.
func TestGeneratorByActionType(t *testing.T) {
	t.Parallel()
	actionTypes := []testutil.ActionType{
		testutil.ActionTypeJavaScript,
		testutil.ActionTypeComposite,
		testutil.ActionTypeDocker,
	}

	for _, actionType := range actionTypes {
		actionType := actionType // capture loop variable
		t.Run(string(actionType), func(t *testing.T) {
			t.Parallel()

			fixtures := testutil.GetFixturesByActionType(actionType)
			if len(fixtures) == 0 {
				t.Skipf("no fixtures available for action type %s", actionType)
			}

			// Test the first fixture of this type
			fixture := fixtures[0]
			actionPath := testutil.CreateTemporaryAction(t, fixture)

			config := &AppConfig{
				Theme:        testGenThemeDefault,
				OutputFormat: testGenFormatMD,
				OutputDir:    ".",
				Quiet:        true,
			}

			generator := NewGenerator(config)
			err := generator.GenerateFromFile(actionPath)

			testutil.AssertNoError(t, err)
		})
	}
}

// TestGeneratorWithMockEnvironment demonstrates testing with a complete mock environment.
func TestGeneratorWithMockEnvironment(t *testing.T) {
	t.Parallel()
	// Create a complete test environment
	envConfig := &testutil.EnvironmentConfig{
		ActionFixtures: []string{"actions/composite/with-dependencies.yml"},
		WithMocks:      true,
	}

	env := testutil.CreateTestEnvironment(t, envConfig)

	// Clean up environment
	defer func() {
		for _, cleanup := range env.Cleanup {
			if err := cleanup(); err != nil {
				t.Errorf("cleanup failed: %v", err)
			}
		}
	}()

	if len(env.ActionPaths) == 0 {
		t.Fatal("expected at least one action path")
	}

	config := &AppConfig{
		Theme:        testGenThemeGitHub,
		OutputFormat: testGenFormatMD,
		OutputDir:    ".",
		Quiet:        true,
	}

	generator := NewGenerator(config)
	err := generator.GenerateFromFile(env.ActionPaths[0])

	testutil.AssertNoError(t, err)
}

// TestGeneratorFixtureValidation demonstrates fixture validation.
func TestGeneratorFixtureValidation(t *testing.T) {
	t.Parallel()
	// Test that all valid fixtures pass validation
	validFixtures := testutil.GetValidFixtures()

	for _, fixtureName := range validFixtures {
		t.Run(fixtureName, func(t *testing.T) {
			testutil.AssertFixtureValid(t, fixtureName)
		})
	}

	// Test that all invalid fixtures fail validation
	invalidFixtures := testutil.GetInvalidFixtures()

	for _, fixtureName := range invalidFixtures {
		t.Run(fixtureName, func(t *testing.T) {
			t.Parallel()
			testutil.AssertFixtureInvalid(t, fixtureName)
		})
	}
}

// createGeneratorTestExecutor returns a test executor function for generator tests.
func createGeneratorTestExecutor() testutil.TestExecutor {
	return func(t *testing.T, testCase testutil.TestCase, ctx *testutil.TestContext) *testutil.TestResult {
		t.Helper()

		result := &testutil.TestResult{
			Context: ctx,
		}

		var actionPath string

		// If we have a fixture, load it and create action file
		if testCase.Fixture != "" {
			fixture, err := ctx.FixtureManager.LoadActionFixture(testCase.Fixture)
			if err != nil {
				result.Error = fmt.Errorf("failed to load fixture %s: %w", testCase.Fixture, err)

				return result
			}

			// Create temporary action file
			actionPath = filepath.Join(ctx.TempDir, appconstants.ActionFileNameYML)
			testutil.WriteTestFile(t, actionPath, fixture.Content)
		}

		// If we don't have an action file to test, just return success
		if actionPath == "" {
			result.Success = true

			return result
		}

		// Create generator configuration from test config
		config := createGeneratorConfigFromTestConfig(ctx.Config, ctx.TempDir)

		// Debug: Log the template path (no working directory changes needed with embedded templates)
		t.Logf("Using template path: %s", config.Template)

		// Create and run generator
		generator := NewGenerator(config)
		err := generator.GenerateFromFile(actionPath)

		if err != nil {
			result.Error = err
			result.Success = false
		} else {
			result.Success = true
			// Detect generated files
			result.Files = testutil.DetectGeneratedFiles(ctx.TempDir, config.OutputFormat)
		}

		return result
	}
}

// createGeneratorConfigFromTestConfig converts TestConfig to AppConfig.
func createGeneratorConfigFromTestConfig(testConfig *testutil.TestConfig, outputDir string) *AppConfig {
	config := &AppConfig{
		Theme:        testGenThemeDefault,
		OutputFormat: testGenFormatMD,
		OutputDir:    outputDir,
		Template:     "templates/readme.tmpl",
		Schema:       "schemas/schema.json",
		Verbose:      false,
		Quiet:        true, // Default to quiet for tests
		GitHubToken:  "",
	}

	// Override with test-specific settings
	if testConfig != nil {
		if testConfig.Theme != "" {
			config.Theme = testConfig.Theme
		}
		if testConfig.OutputFormat != "" {
			config.OutputFormat = testConfig.OutputFormat
		}
		if testConfig.OutputDir != "" {
			config.OutputDir = testConfig.OutputDir
		}
		config.Verbose = testConfig.Verbose
		config.Quiet = testConfig.Quiet
	}

	// Set appropriate template path based on theme - embedded templates will handle resolution
	config.Template = resolveThemeTemplate(config.Theme)

	return config
}
