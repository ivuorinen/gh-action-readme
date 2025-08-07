// Package testutil provides table-driven testing infrastructure and test helpers for gh-action-readme.
package testutil

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-github/v74/github"
)

// File constants.
const (
	readmeFilename = "README.md"
)

// TestExecutor is a function type for executing specific types of tests.
type TestExecutor func(t *testing.T, testCase TestCase, ctx *TestContext) *TestResult

// TestCase represents a parameterized test case.
type TestCase struct {
	Name        string
	Description string
	Fixture     string
	Config      *TestConfig
	Mocks       *MockConfig
	Expected    *ExpectedResult
	SetupFunc   func(*testing.T, *TestContext) error
	CleanupFunc func(*testing.T, *TestContext) error
	SkipReason  string
	Executor    TestExecutor // Custom test execution logic
}

// TestSuite represents a collection of related test cases.
type TestSuite struct {
	Name          string
	Description   string
	Cases         []TestCase
	GlobalSetup   func(*testing.T) (*TestContext, error)
	GlobalCleanup func(*testing.T, *TestContext) error
	Parallel      bool
}

// TestConfig holds configuration options for a test case.
type TestConfig struct {
	Theme        string
	OutputFormat string
	OutputDir    string
	Verbose      bool
	Quiet        bool
	Recursive    bool
	Validate     bool
	ExtraFlags   map[string]string
}

// MockConfig specifies which mocks to set up for a test case.
type MockConfig struct {
	GitHubClient    bool
	GitHubResponses map[string]string
	ColoredOutput   bool
	FileSystem      bool
	Environment     map[string]string
	TempDir         bool
}

// ExpectedResult defines what a test case should produce.
type ExpectedResult struct {
	ShouldSucceed    bool
	ShouldFail       bool
	ExpectedError    string
	ExpectedOutput   []string
	ExpectedFiles    []string
	ExpectedExitCode int
	CustomValidation func(*testing.T, *TestResult) error
}

// TestContext holds the context for a running test.
type TestContext struct {
	TempDir        string
	FixtureManager *FixtureManager
	Mocks          *MockSuite
	Config         *TestConfig
	Cleanup        []func() error
}

// TestResult holds the results of a test execution.
type TestResult struct {
	Success  bool
	Error    error
	Output   string
	Files    []string
	ExitCode int
	Duration int64
	Context  *TestContext
}

// MockSuite holds all configured mocks for a test.
type MockSuite struct {
	GitHubClient  *github.Client
	ColoredOutput *MockColoredOutput
	HTTPClient    *MockHTTPClient
	Environment   map[string]string
	TempDirs      []string
}

// ActionTestCase represents a test case specifically for action-related testing.
type ActionTestCase struct {
	TestCase
	ActionType     ActionType
	ExpectValid    bool
	ExpectAnalysis bool
	ExpectDeps     int
}

// GeneratorTestCase represents a test case for generator testing.
type GeneratorTestCase struct {
	TestCase
	Theme         string
	OutputFormat  string
	ExpectFiles   []string
	ExpectContent map[string][]string
}

// ValidationTestCase represents a test case for validation testing.
type ValidationTestCase struct {
	TestCase
	ExpectValid     bool
	ExpectedErrors  []string
	ValidationLevel string
}

// RunTestSuite executes a complete test suite with common setup/teardown.
func RunTestSuite(t *testing.T, suite TestSuite) {
	t.Helper()

	if suite.Name == "" {
		t.Fatal("test suite must have a name")
	}

	t.Run(suite.Name, func(t *testing.T) {
		if suite.Parallel {
			t.Parallel()
		}

		globalContext := setupGlobalContext(t, suite)
		runAllTestCases(t, suite, globalContext)
	})
}

// setupGlobalContext handles global setup and cleanup for test suite.
func setupGlobalContext(t *testing.T, suite TestSuite) *TestContext {
	t.Helper()

	var globalContext *TestContext
	if suite.GlobalSetup != nil {
		var err error
		globalContext, err = suite.GlobalSetup(t)
		if err != nil {
			t.Fatalf("global setup failed: %v", err)
		}
	}

	// Set up global cleanup
	if suite.GlobalCleanup != nil && globalContext != nil {
		t.Cleanup(func() {
			if err := suite.GlobalCleanup(t, globalContext); err != nil {
				t.Errorf("global cleanup failed: %v", err)
			}
		})
	}

	return globalContext
}

// runAllTestCases executes all test cases in the suite.
func runAllTestCases(t *testing.T, suite TestSuite, globalContext *TestContext) {
	t.Helper()

	for _, testCase := range suite.Cases {
		testCase := testCase // capture loop variable

		if testCase.SkipReason != "" {
			runSkippedTest(t, testCase)

			continue
		}

		runIndividualTest(t, suite, testCase, globalContext)
	}
}

// runSkippedTest handles skipped test cases.
func runSkippedTest(t *testing.T, testCase TestCase) {
	t.Helper()

	t.Run(testCase.Name, func(t *testing.T) {
		t.Skip(testCase.SkipReason)
	})
}

// runIndividualTest executes a single test case.
func runIndividualTest(t *testing.T, suite TestSuite, testCase TestCase, globalContext *TestContext) {
	t.Helper()

	t.Run(testCase.Name, func(t *testing.T) {
		if suite.Parallel {
			t.Parallel()
		}

		runTestCase(t, testCase, globalContext)
	})
}

// runTestCase executes a single test case.
func runTestCase(t *testing.T, testCase TestCase, globalContext *TestContext) {
	t.Helper()

	// Create test context
	ctx := createTestContext(t, testCase, globalContext)

	// Setup test-specific cleanup
	defer func() {
		for i := len(ctx.Cleanup) - 1; i >= 0; i-- {
			if err := ctx.Cleanup[i](); err != nil {
				t.Errorf("cleanup failed: %v", err)
			}
		}
	}()

	// Run test case setup
	if testCase.SetupFunc != nil {
		if err := testCase.SetupFunc(t, ctx); err != nil {
			t.Fatalf("test setup failed: %v", err)
		}
	}

	// Execute the test
	result := executeTest(t, testCase, ctx)

	// Validate results
	validateTestResult(t, testCase, result)

	// Run test case cleanup
	if testCase.CleanupFunc != nil {
		if err := testCase.CleanupFunc(t, ctx); err != nil {
			t.Errorf("test cleanup failed: %v", err)
		}
	}
}

// createTestContext creates a test context for a test case.
func createTestContext(t *testing.T, testCase TestCase, globalContext *TestContext) *TestContext {
	t.Helper()

	ctx := &TestContext{
		FixtureManager: GetFixtureManager(),
		Config:         testCase.Config,
		Cleanup:        make([]func() error, 0),
	}

	// Inherit from global context if available
	if globalContext != nil {
		if ctx.Config == nil {
			ctx.Config = globalContext.Config
		}
		if globalContext.TempDir != "" {
			ctx.TempDir = globalContext.TempDir
		}
	}

	// Set up temporary directory if needed
	if testCase.Mocks != nil && testCase.Mocks.TempDir {
		tempDir, cleanup := TempDir(t)
		ctx.TempDir = tempDir
		ctx.Cleanup = append(ctx.Cleanup, func() error {
			cleanup()

			return nil
		})
	}

	// Set up mocks
	if testCase.Mocks != nil {
		ctx.Mocks = createMockSuite(t, testCase.Mocks)
	}

	return ctx
}

// createMockSuite creates a mock suite based on configuration.
func createMockSuite(t *testing.T, config *MockConfig) *MockSuite {
	t.Helper()

	suite := &MockSuite{
		Environment: make(map[string]string),
		TempDirs:    make([]string, 0),
	}

	// Set up GitHub client mock
	if config.GitHubClient {
		responses := config.GitHubResponses
		if responses == nil {
			responses = MockGitHubResponses()
		}
		suite.GitHubClient = MockGitHubClient(responses)
	}

	// Set up colored output mock
	if config.ColoredOutput {
		suite.ColoredOutput = &MockColoredOutput{
			Messages: make([]string, 0),
		}
	}

	// Set up HTTP client mock
	if config.GitHubClient {
		suite.HTTPClient = &MockHTTPClient{
			Responses: make(map[string]*http.Response),
			Requests:  make([]*http.Request, 0),
		}
	}

	// Set up environment variables
	for key, value := range config.Environment {
		suite.Environment[key] = value
	}

	return suite
}

// executeTest runs the actual test logic.
func executeTest(t *testing.T, testCase TestCase, ctx *TestContext) *TestResult {
	t.Helper()

	// Use custom executor if provided
	if testCase.Executor != nil {
		return testCase.Executor(t, testCase, ctx)
	}

	// Default execution: just create fixture and return success
	result := &TestResult{
		Context: ctx,
	}

	// If we have a fixture, load it and create action file
	if testCase.Fixture != "" {
		fixture, err := ctx.FixtureManager.LoadActionFixture(testCase.Fixture)
		if err != nil {
			result.Error = fmt.Errorf("failed to load fixture %s: %w", testCase.Fixture, err)

			return result
		}

		// Create temporary action file
		actionPath := filepath.Join(ctx.TempDir, "action.yml")
		WriteTestFile(t, actionPath, fixture.Content)
	}

	// Default success for non-generator tests
	result.Success = true

	return result
}

// validateTestResult validates the test results against expectations.
func validateTestResult(t *testing.T, testCase TestCase, result *TestResult) {
	t.Helper()

	if testCase.Expected == nil {
		return
	}

	expected := testCase.Expected
	validateSuccessFailure(t, expected, result)
	validateError(t, expected, result)
	validateOutput(t, expected, result)
	validateFiles(t, expected, result)
	validateExitCode(t, expected, result)
	validateCustom(t, expected, result)
}

// validateSuccessFailure checks success/failure expectations.
func validateSuccessFailure(t *testing.T, expected *ExpectedResult, result *TestResult) {
	t.Helper()

	if expected.ShouldSucceed && !result.Success {
		t.Errorf("expected test to succeed, but it failed: %v", result.Error)
	}

	if expected.ShouldFail && result.Success {
		t.Error("expected test to fail, but it succeeded")
	}
}

// validateError checks expected error conditions.
func validateError(t *testing.T, expected *ExpectedResult, result *TestResult) {
	t.Helper()

	if expected.ExpectedError == "" {
		return
	}

	if result.Error == nil {
		t.Errorf("expected error %q, but got no error", expected.ExpectedError)

		return
	}

	if result.Error.Error() != expected.ExpectedError {
		t.Errorf("expected error %q, but got %q", expected.ExpectedError, result.Error.Error())
	}
}

// validateOutput checks expected output conditions.
func validateOutput(t *testing.T, expected *ExpectedResult, result *TestResult) {
	t.Helper()

	for _, expectedOutput := range expected.ExpectedOutput {
		if !containsString(result.Output, expectedOutput) {
			t.Errorf("expected output to contain %q, but it didn't. Output: %s", expectedOutput, result.Output)
		}
	}
}

// validateFiles checks expected file conditions.
func validateFiles(t *testing.T, expected *ExpectedResult, result *TestResult) {
	t.Helper()

	for _, expectedFile := range expected.ExpectedFiles {
		if strings.Contains(expectedFile, "*") {
			// Handle wildcard patterns like "*.html"
			found := false
			pattern := strings.TrimPrefix(expectedFile, "*")
			for _, actualFile := range result.Files {
				if strings.HasSuffix(actualFile, pattern) {
					found = true

					break
				}
			}
			if !found {
				t.Errorf(
					"expected file matching pattern %q to be created, but none found. Files: %v",
					expectedFile,
					result.Files,
				)
			}
		} else if !containsString(result.Files, expectedFile) {
			// Handle exact filename matches
			t.Errorf("expected file %q to be created, but it wasn't. Files: %v", expectedFile, result.Files)
		}
	}
}

// validateExitCode checks expected exit code.
func validateExitCode(t *testing.T, expected *ExpectedResult, result *TestResult) {
	t.Helper()

	if expected.ExpectedExitCode != 0 && result.ExitCode != expected.ExpectedExitCode {
		t.Errorf("expected exit code %d, but got %d", expected.ExpectedExitCode, result.ExitCode)
	}
}

// validateCustom runs custom validation if provided.
func validateCustom(t *testing.T, expected *ExpectedResult, result *TestResult) {
	t.Helper()

	if expected.CustomValidation == nil {
		return
	}

	if err := expected.CustomValidation(t, result); err != nil {
		t.Errorf("custom validation failed: %v", err)
	}
}

// Helper functions for specific test types

// RunActionTests executes action-related test cases.
func RunActionTests(t *testing.T, cases []ActionTestCase) {
	t.Helper()

	testCases := make([]TestCase, len(cases))
	for i, actionCase := range cases {
		testCases[i] = actionCase.TestCase
	}

	suite := TestSuite{
		Name:     "Action Tests",
		Cases:    testCases,
		Parallel: true,
	}

	RunTestSuite(t, suite)
}

// RunGeneratorTests executes generator test cases.
func RunGeneratorTests(t *testing.T, cases []GeneratorTestCase) {
	t.Helper()

	testCases := make([]TestCase, len(cases))
	for i, genCase := range cases {
		testCases[i] = genCase.TestCase
	}

	suite := TestSuite{
		Name:     "Generator Tests",
		Cases:    testCases,
		Parallel: true,
	}

	RunTestSuite(t, suite)
}

// RunValidationTests executes validation test cases.
func RunValidationTests(t *testing.T, cases []ValidationTestCase) {
	t.Helper()

	testCases := make([]TestCase, len(cases))
	for i, valCase := range cases {
		testCases[i] = valCase.TestCase
	}

	suite := TestSuite{
		Name:     "Validation Tests",
		Cases:    testCases,
		Parallel: true,
	}

	RunTestSuite(t, suite)
}

// Utility functions

// containsString checks if a slice contains a string.
func containsString(slice any, item string) bool {
	switch s := slice.(type) {
	case []string:
		for _, v := range s {
			if v == item {
				return true
			}
		}
	case string:
		return len(s) > 0 && s == item
	}

	return false
}

// DetectGeneratedFiles finds files that were generated in the output directory.
// This is exported so tests in other packages can use it.
func DetectGeneratedFiles(outputDir string, outputFormat string) []string {
	var files []string

	// Different output formats create different files:
	// - md: README.md
	// - html: <action-name>.html (name varies)
	// - json: action-docs.json
	// - asciidoc: README.adoc

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return files
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			// Skip the action.yml we created for testing
			if name == "action.yml" {
				continue
			}

			// Check if this file matches the expected output format
			isGenerated := false
			switch outputFormat {
			case "md":
				isGenerated = name == readmeFilename
			case "html":
				isGenerated = strings.HasSuffix(name, ".html")
			case "json":
				isGenerated = name == "action-docs.json"
			case "asciidoc":
				isGenerated = name == "README.adoc"
			default:
				isGenerated = name == readmeFilename
			}

			if isGenerated {
				files = append(files, name)
			}
		}
	}

	return files
}

// DefaultTestConfig returns a default test configuration.
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		Theme:        "default",
		OutputFormat: "md",
		OutputDir:    ".",
		Verbose:      false,
		Quiet:        false,
		Recursive:    false,
		Validate:     true,
		ExtraFlags:   make(map[string]string),
	}
}

// DefaultMockConfig returns a default mock configuration.
func DefaultMockConfig() *MockConfig {
	return &MockConfig{
		GitHubClient:    true,
		GitHubResponses: MockGitHubResponses(),
		ColoredOutput:   true,
		FileSystem:      false,
		Environment:     make(map[string]string),
		TempDir:         true,
	}
}

// Advanced test environment and helper functions (consolidated from helpers.go)

// EnvironmentConfig configures a test environment.
type EnvironmentConfig struct {
	ActionFixtures []string
	ConfigFixture  string
	WithMocks      bool
	ExtraFiles     map[string]string
}

// TestEnvironment represents a complete test environment.
type TestEnvironment struct {
	TempDir        string
	ActionPaths    []string
	ConfigPath     string
	FixtureManager *FixtureManager
	Mocks          *MockSuite
	Cleanup        []func() error
}

// CreateTemporaryAction creates a temporary action file for testing.
func CreateTemporaryAction(t *testing.T, fixture string) string {
	t.Helper()

	// Load the fixture
	actionFixture, err := LoadActionFixture(fixture)
	if err != nil {
		t.Fatalf("failed to load action fixture %s: %v", fixture, err)
	}

	// Create temporary directory
	tempDir, cleanup := TempDir(t)
	t.Cleanup(cleanup)

	// Write action file
	actionPath := filepath.Join(tempDir, "action.yml")
	WriteTestFile(t, actionPath, actionFixture.Content)

	return actionPath
}

// CreateTemporaryActionDir creates a temporary directory with an action file.
func CreateTemporaryActionDir(t *testing.T, fixture string) string {
	t.Helper()

	// Load the fixture
	actionFixture, err := LoadActionFixture(fixture)
	if err != nil {
		t.Fatalf("failed to load action fixture %s: %v", fixture, err)
	}

	// Create temporary directory
	tempDir, cleanup := TempDir(t)
	t.Cleanup(cleanup)

	// Write action file
	actionPath := filepath.Join(tempDir, "action.yml")
	WriteTestFile(t, actionPath, actionFixture.Content)

	return tempDir
}

// CreateTestEnvironment sets up a complete test environment.
func CreateTestEnvironment(t *testing.T, config *EnvironmentConfig) *TestEnvironment {
	t.Helper()

	if config == nil {
		config = &EnvironmentConfig{}
	}

	env := &TestEnvironment{
		TempDir:        "",
		ActionPaths:    make([]string, 0),
		ConfigPath:     "",
		FixtureManager: GetFixtureManager(),
		Cleanup:        make([]func() error, 0),
	}

	// Create temporary directory
	tempDir, cleanup := TempDir(t)
	env.TempDir = tempDir
	env.Cleanup = append(env.Cleanup, func() error {
		cleanup()

		return nil
	})

	// Create action files from fixtures
	for _, fixture := range config.ActionFixtures {
		actionPath := CreateTemporaryAction(t, fixture)
		env.ActionPaths = append(env.ActionPaths, actionPath)
	}

	// Create config file if specified
	if config.ConfigFixture != "" {
		configFixture, err := LoadConfigFixture(config.ConfigFixture)
		if err != nil {
			t.Fatalf("failed to load config fixture %s: %v", config.ConfigFixture, err)
		}

		configPath := filepath.Join(tempDir, "config.yml")
		WriteTestFile(t, configPath, configFixture.Content)
		env.ConfigPath = configPath
	}

	// Set up mocks if requested
	if config.WithMocks {
		env.Mocks = CreateMockSuite(DefaultMockConfig())
	}

	return env
}

// CreateMockSuite creates a complete mock environment (enhanced from helpers.go).
func CreateMockSuite(config *MockConfig) *MockSuite {
	if config == nil {
		config = DefaultMockConfig()
	}

	suite := &MockSuite{
		Environment: make(map[string]string),
		TempDirs:    make([]string, 0),
	}

	// Set up GitHub client mock
	if config.GitHubClient {
		responses := config.GitHubResponses
		if responses == nil {
			responses = MockGitHubResponses()
		}
		suite.GitHubClient = MockGitHubClient(responses)
	}

	// Set up colored output mock
	if config.ColoredOutput {
		suite.ColoredOutput = &MockColoredOutput{
			Messages: make([]string, 0),
		}
	}

	// Set up HTTP client mock
	if config.GitHubClient {
		suite.HTTPClient = &MockHTTPClient{
			Responses: make(map[string]*http.Response),
			Requests:  make([]*http.Request, 0),
		}
	}

	// Set up environment variables
	for key, value := range config.Environment {
		suite.Environment[key] = value
	}

	return suite
}

// SetupGitHubMocks configures GitHub API mocks for test scenarios.
func SetupGitHubMocks(scenarios []string) map[string]string {
	responses := MockGitHubResponses()

	// Add scenario-specific responses
	for _, scenario := range scenarios {
		switch scenario {
		case "rate-limit":
			responses["GET https://api.github.com/rate_limit"] = GitHubRateLimitResponse
		case "not-found":
			responses["GET https://api.github.com/repos/nonexistent/repo"] = GitHubErrorResponse
		case "latest-release":
			responses["GET https://api.github.com/repos/actions/checkout/releases/latest"] = GitHubReleaseResponse
		}
	}

	return responses
}

// ValidateActionFixture validates that a fixture contains expected content.
func ValidateActionFixture(t *testing.T, fixture *ActionFixture) {
	t.Helper()

	if fixture == nil {
		t.Fatal("fixture is nil")
	}

	if fixture.Content == "" {
		t.Error("fixture content is empty")
	}

	if fixture.Name == "" {
		t.Error("fixture name is empty")
	}

	// Basic YAML validation is done in the fixture loader
	if !fixture.IsValid && fixture.Scenario != nil && fixture.Scenario.ExpectValid {
		t.Errorf("fixture %s should be valid according to scenario but validation failed", fixture.Name)
	}
}

// TestAllThemes runs a test with all available themes.
func TestAllThemes(t *testing.T, testFunc func(*testing.T, string)) {
	t.Helper()

	themes := []string{"default", "github", "minimal", "professional"}

	for _, theme := range themes {
		theme := theme // capture loop variable
		t.Run("theme_"+theme, func(t *testing.T) {
			t.Parallel()
			testFunc(t, theme)
		})
	}
}

// TestAllFormats runs a test with all available output formats.
func TestAllFormats(t *testing.T, testFunc func(*testing.T, string)) {
	t.Helper()

	formats := []string{"md", "html", "json", "asciidoc"}

	for _, format := range formats {
		format := format // capture loop variable
		t.Run("format_"+format, func(t *testing.T) {
			t.Parallel()
			testFunc(t, format)
		})
	}
}

// TestValidationScenarios runs validation tests for all invalid fixtures.
func TestValidationScenarios(t *testing.T, validatorFunc func(*testing.T, string) error) {
	t.Helper()

	invalidFixtures := GetInvalidFixtures()

	for _, fixture := range invalidFixtures {
		fixture := fixture // capture loop variable
		t.Run("invalid_"+strings.ReplaceAll(fixture, "/", "_"), func(t *testing.T) {
			t.Parallel()

			err := validatorFunc(t, fixture)
			if err == nil {
				t.Errorf("expected validation error for fixture %s, but got none", fixture)
			}
		})
	}
}

// CreateGitHubMockSuite creates a GitHub API mock suite.
func CreateGitHubMockSuite(scenarios []string) *MockSuite {
	config := &MockConfig{
		GitHubClient:    true,
		GitHubResponses: SetupGitHubMocks(scenarios),
		ColoredOutput:   true,
		Environment:     make(map[string]string),
		TempDir:         true,
	}

	return CreateMockSuite(config)
}

// AssertFixtureValid asserts that a fixture should be valid.
func AssertFixtureValid(t *testing.T, fixtureName string) {
	t.Helper()

	fixture, err := LoadActionFixture(fixtureName)
	AssertNoError(t, err)

	if !fixture.IsValid {
		t.Errorf("fixture %s should be valid but failed validation", fixtureName)
	}

	ValidateActionFixture(t, fixture)
}

// AssertFixtureInvalid asserts that a fixture should be invalid.
func AssertFixtureInvalid(t *testing.T, fixtureName string) {
	t.Helper()

	fixture, err := LoadActionFixture(fixtureName)

	// The fixture might load but be marked as invalid
	if err == nil && fixture.IsValid {
		t.Errorf("fixture %s should be invalid but passed validation", fixtureName)
	}
}

// CreateActionTestCases creates test cases for action-related testing.
func CreateActionTestCases() []ActionTestCase {
	validFixtures := GetValidFixtures()
	invalidFixtures := GetInvalidFixtures()

	cases := make([]ActionTestCase, 0, len(validFixtures)+len(invalidFixtures))

	// Add valid fixture test cases
	for _, fixture := range validFixtures {
		actionFixture, err := LoadActionFixture(fixture)
		if err != nil {
			continue // Skip fixtures that can't be loaded
		}

		cases = append(cases, ActionTestCase{
			TestCase: TestCase{
				Name:        "valid_" + strings.ReplaceAll(fixture, "/", "_"),
				Description: "Test valid action fixture: " + fixture,
				Fixture:     fixture,
				Config:      DefaultTestConfig(),
				Mocks:       DefaultMockConfig(),
				Expected: &ExpectedResult{
					ShouldSucceed: true,
					ShouldFail:    false,
				},
			},
			ActionType:     actionFixture.ActionType,
			ExpectValid:    true,
			ExpectAnalysis: true,
		})
	}

	// Add invalid fixture test cases
	for _, fixture := range invalidFixtures {
		actionFixture, _ := LoadActionFixture(fixture) // Might fail, that's ok
		actionType := ActionTypeInvalid
		if actionFixture != nil {
			actionType = actionFixture.ActionType
		}

		cases = append(cases, ActionTestCase{
			TestCase: TestCase{
				Name:        "invalid_" + strings.ReplaceAll(fixture, "/", "_"),
				Description: "Test invalid action fixture: " + fixture,
				Fixture:     fixture,
				Config:      DefaultTestConfig(),
				Mocks:       DefaultMockConfig(),
				Expected: &ExpectedResult{
					ShouldSucceed: false,
					ShouldFail:    true,
				},
			},
			ActionType:     actionType,
			ExpectValid:    false,
			ExpectAnalysis: false,
		})
	}

	return cases
}

// getExpectedFilename returns the expected filename for a given output format.
func getExpectedFilename(outputFormat string) string {
	switch outputFormat {
	case "md":
		return "README.md"
	case "html":
		// HTML files have variable names based on action name, so we'll use a pattern
		// The DetectGeneratedFiles function will find any .html file
		return "*.html"
	case "json":
		return "action-docs.json"
	case "asciidoc":
		return "README.adoc"
	default:
		return "README.md"
	}
}

// CreateGeneratorTestCases creates test cases for generator testing.
func CreateGeneratorTestCases() []GeneratorTestCase {
	validFixtures := GetValidFixtures()
	themes := []string{"default", "github", "minimal", "professional"}
	formats := []string{"md", "html", "json", "asciidoc"}

	cases := make([]GeneratorTestCase, 0)

	// Create test cases for each valid fixture with each theme/format combination
	for _, fixture := range validFixtures {
		for _, theme := range themes {
			for _, format := range formats {
				expectedFilename := getExpectedFilename(format)

				cases = append(cases, GeneratorTestCase{
					TestCase: TestCase{
						Name: fmt.Sprintf("%s_%s_%s", strings.ReplaceAll(fixture, "/", "_"), theme, format),
						Description: fmt.Sprintf(
							"Generate %s documentation for %s using %s theme",
							format,
							fixture,
							theme,
						),
						Fixture: fixture,
						Config: &TestConfig{
							Theme:        theme,
							OutputFormat: format,
							OutputDir:    ".",
							Verbose:      false,
							Quiet:        false,
						},
						Mocks: DefaultMockConfig(),
						Expected: &ExpectedResult{
							ShouldSucceed: true,
							ExpectedFiles: []string{expectedFilename},
						},
					},
					Theme:        theme,
					OutputFormat: format,
					ExpectFiles:  []string{expectedFilename},
				})
			}
		}
	}

	return cases
}

// CreateValidationTestCases creates test cases for validation testing.
func CreateValidationTestCases() []ValidationTestCase {
	fm := GetFixtureManager()
	cases := make([]ValidationTestCase, 0)

	// Add test cases for all scenarios
	for _, scenario := range fm.scenarios {
		cases = append(cases, ValidationTestCase{
			TestCase: TestCase{
				Name:        "validate_" + scenario.ID,
				Description: scenario.Description,
				Fixture:     scenario.Fixture,
				Config:      DefaultTestConfig(),
				Expected: &ExpectedResult{
					ShouldSucceed: scenario.ExpectValid,
					ShouldFail:    !scenario.ExpectValid,
				},
			},
			ExpectValid:     scenario.ExpectValid,
			ValidationLevel: "standard",
		})
	}

	return cases
}
