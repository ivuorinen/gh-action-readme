package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

// TestGenerator_ComprehensiveGeneration demonstrates the new table-driven testing framework
// by testing generation across all fixtures, themes, and formats systematically.
func TestGenerator_ComprehensiveGeneration(t *testing.T) {
	// Create test cases using the new helper functions
	cases := testutil.CreateGeneratorTestCases()

	// Filter to a subset for demonstration (full test would be very large)
	filteredCases := make([]testutil.GeneratorTestCase, 0)
	for _, testCase := range cases {
		// Only test a few combinations for demonstration
		if (testCase.Theme == "default" && testCase.OutputFormat == "md") ||
			(testCase.Theme == "github" && testCase.OutputFormat == "html") ||
			(testCase.Theme == "minimal" && testCase.OutputFormat == "json") {
			// Add custom executor for generator tests
			testCase.Executor = createGeneratorTestExecutor()
			filteredCases = append(filteredCases, testCase)
		}
	}

	// Run the test suite
	testutil.RunGeneratorTests(t, filteredCases)
}

// TestGenerator_AllValidFixtures tests generation with all valid fixtures.
func TestGenerator_AllValidFixtures(t *testing.T) {
	validFixtures := testutil.GetValidFixtures()

	for _, fixture := range validFixtures {
		fixture := fixture // capture loop variable
		t.Run(fixture, func(t *testing.T) {
			t.Parallel()

			// Create temporary action from fixture
			actionPath := testutil.CreateTemporaryAction(t, fixture)

			// Test with default configuration
			config := &AppConfig{
				Theme:        "default",
				OutputFormat: "md",
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

// TestGenerator_AllInvalidFixtures tests that invalid fixtures produce expected errors.
func TestGenerator_AllInvalidFixtures(t *testing.T) {
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
				Theme:        "default",
				OutputFormat: "md",
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

// TestGenerator_AllThemes demonstrates theme testing using helper functions.
func TestGenerator_AllThemes(t *testing.T) {
	// Use the helper function to test all themes
	testutil.TestAllThemes(t, func(t *testing.T, theme string) {
		t.Helper()
		// Create a simple action for testing
		actionPath := testutil.CreateTemporaryAction(t, "actions/javascript/simple.yml")

		config := &AppConfig{
			Theme:        theme,
			OutputFormat: "md",
			OutputDir:    ".",
			Quiet:        true,
		}

		generator := NewGenerator(config)
		err := generator.GenerateFromFile(actionPath)

		testutil.AssertNoError(t, err)
	})
}

// TestGenerator_AllFormats demonstrates format testing using helper functions.
func TestGenerator_AllFormats(t *testing.T) {
	// Use the helper function to test all formats
	testutil.TestAllFormats(t, func(t *testing.T, format string) {
		t.Helper()
		// Create a simple action for testing
		actionPath := testutil.CreateTemporaryAction(t, "actions/javascript/simple.yml")

		config := &AppConfig{
			Theme:        "default",
			OutputFormat: format,
			OutputDir:    ".",
			Quiet:        true,
		}

		generator := NewGenerator(config)
		err := generator.GenerateFromFile(actionPath)

		testutil.AssertNoError(t, err)
	})
}

// TestGenerator_ByActionType demonstrates testing by action type.
func TestGenerator_ByActionType(t *testing.T) {
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
				Theme:        "default",
				OutputFormat: "md",
				OutputDir:    ".",
				Quiet:        true,
			}

			generator := NewGenerator(config)
			err := generator.GenerateFromFile(actionPath)

			testutil.AssertNoError(t, err)
		})
	}
}

// TestGenerator_WithMockEnvironment demonstrates testing with a complete mock environment.
func TestGenerator_WithMockEnvironment(t *testing.T) {
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
		Theme:        "github",
		OutputFormat: "md",
		OutputDir:    ".",
		Quiet:        true,
	}

	generator := NewGenerator(config)
	err := generator.GenerateFromFile(env.ActionPaths[0])

	testutil.AssertNoError(t, err)
}

// TestGenerator_FixtureValidation demonstrates fixture validation.
func TestGenerator_FixtureValidation(t *testing.T) {
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
			actionPath = filepath.Join(ctx.TempDir, "action.yml")
			testutil.WriteTestFile(t, actionPath, fixture.Content)
		}

		// If we don't have an action file to test, just return success
		if actionPath == "" {
			result.Success = true

			return result
		}

		// Create generator configuration from test config
		config := createGeneratorConfigFromTestConfig(ctx.Config, ctx.TempDir)

		// Save current working directory and change to project root for template resolution
		originalWd, err := os.Getwd()
		if err != nil {
			result.Error = fmt.Errorf("failed to get working directory: %w", err)

			return result
		}

		// Use runtime.Caller to find project root relative to this file
		_, currentFile, _, ok := runtime.Caller(0)
		if !ok {
			result.Error = errors.New("failed to get current file path")

			return result
		}

		// Get the project root (go up from internal/generator_comprehensive_test.go to project root)
		projectRoot := filepath.Dir(filepath.Dir(currentFile))
		if err := os.Chdir(projectRoot); err != nil {
			result.Error = fmt.Errorf("failed to change to project root %s: %w", projectRoot, err)

			return result
		}

		// Debug: Log the working directory and template path
		currentWd, _ := os.Getwd()
		t.Logf("Test working directory: %s, template path: %s", currentWd, config.Template)

		// Restore working directory after test
		defer func() {
			if err := os.Chdir(originalWd); err != nil {
				// Log error but don't fail the test
				t.Logf("Failed to restore working directory: %v", err)
			}
		}()

		// Create and run generator
		generator := NewGenerator(config)
		err = generator.GenerateFromFile(actionPath)

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
		Theme:        "default",
		OutputFormat: "md",
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

	// Set appropriate template path based on theme and output format
	config.Template = resolveTemplatePathForTest(config.Theme, config.OutputFormat)

	return config
}

// resolveTemplatePathForTest resolves the correct template path for testing.
func resolveTemplatePathForTest(theme, _ string) string {
	switch theme {
	case "github":
		return "templates/themes/github/readme.tmpl"
	case "gitlab":
		return "templates/themes/gitlab/readme.tmpl"
	case "minimal":
		return "templates/themes/minimal/readme.tmpl"
	case "professional":
		return "templates/themes/professional/readme.tmpl"
	default:
		return "templates/readme.tmpl"
	}
}
