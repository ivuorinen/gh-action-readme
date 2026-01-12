// Package testutil provides testing utilities and mocks for gh-action-readme.
package testutil

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v74/github"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

// MockHTTPClient is a mock HTTP client for testing.
type MockHTTPClient struct {
	Responses map[string]*http.Response
	Requests  []*http.Request
}

// HTTPResponse represents a mock HTTP response.
type HTTPResponse struct {
	StatusCode int
	Body       string
	Headers    map[string]string
}

// HTTPRequest represents a captured HTTP request.
type HTTPRequest struct {
	Method  string
	URL     string
	Body    string
	Headers map[string]string
}

// Do implements the http.Client interface.
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.Requests = append(m.Requests, req)

	key := req.Method + " " + req.URL.String()
	if resp, ok := m.Responses[key]; ok {
		return resp, nil
	}

	// Default 404 response
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader(`{"error": "not found"}`)),
	}, nil
}

// MockGitHubClient creates a GitHub client with mocked responses.
func MockGitHubClient(responses map[string]string) *github.Client {
	mockClient := &MockHTTPClient{
		Responses: make(map[string]*http.Response),
	}

	for key, body := range responses {
		mockClient.Responses[key] = &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}
	}

	client := github.NewClient(&http.Client{Transport: &MockTransport{Client: mockClient}})

	return client
}

// MockTransport implements http.RoundTripper for testing HTTP clients.
type MockTransport struct {
	Client *MockHTTPClient
}

// RoundTrip implements http.RoundTripper interface.
func (t *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.Client.Do(req)
}

// TempDir creates a temporary directory for testing and returns cleanup function.
func TempDir(t *testing.T) (string, func()) {
	t.Helper()

	dir := t.TempDir()

	return dir, func() {
		// t.TempDir() automatically cleans up, so no action needed
	}
}

// CleanupCache provides a standard cache cleanup helper for deferred cleanup.
// It returns a function that closes the cache and fails the test on errors.
func CleanupCache(tb testing.TB, cache interface{ Close() error }) func() {
	tb.Helper()

	return func() {
		tb.Helper()
		if err := cache.Close(); err != nil {
			tb.Fatalf("failed to close cache: %v", err)
		}
	}
}

// ExpectPanic asserts that the provided function panics with a message containing the expected substring.
// This helper reduces panic recovery test boilerplate from 12-15 lines to 3-4 lines.
func ExpectPanic(t *testing.T, fn func(), expectedSubstring string) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic but got none")
		} else {
			var errStr string
			switch v := r.(type) {
			case string:
				errStr = v
			case error:
				errStr = v.Error()
			default:
				errStr = fmt.Sprintf("%v", v)
			}
			if !strings.Contains(errStr, expectedSubstring) {
				t.Errorf("expected panic message containing %q, got: %v", expectedSubstring, r)
			}
		}
	}()
	fn()
}

// MustLoadActionFixture loads an action fixture and fails the test on error.
// This helper consolidates the load + assertion pattern.
func MustLoadActionFixture(t *testing.T, path string) *ActionFixture {
	t.Helper()
	fixture, err := LoadActionFixture(path)
	AssertNoError(t, err)

	return fixture
}

// LoadAndWriteFixture loads an action fixture and writes it to the specified path.
// This helper reduces the common 3-line pattern to a single line.
func LoadAndWriteFixture(t *testing.T, fixturePath, targetPath string) {
	t.Helper()
	fixture := MustLoadActionFixture(t, fixturePath)
	WriteTestFile(t, targetPath, fixture.Content)
}

// WriteTestFile writes a test file to the given path.
func WriteTestFile(t *testing.T, path, content string) {
	t.Helper()

	dir := filepath.Dir(path)
	// #nosec G301 -- test directory permissions
	if err := os.MkdirAll(dir, appconstants.FilePermDir); err != nil {
		t.Fatalf("failed to create dir %s: %v", dir, err)
	}

	// #nosec G306 -- test file permissions
	if err := os.WriteFile(path, []byte(content), appconstants.FilePermDefault); err != nil {
		t.Fatalf("failed to write test file %s: %v", path, err)
	}
}

// WriteActionFixture writes an action fixture to a standard action.yml file.
func WriteActionFixture(t *testing.T, dir, fixturePath string) string {
	t.Helper()
	actionPath := filepath.Join(dir, appconstants.ActionFileNameYML)
	fixture := MustLoadActionFixture(t, fixturePath)
	WriteTestFile(t, actionPath, fixture.Content)

	return actionPath
}

// WriteActionFixtureAs writes an action fixture with a custom filename.
func WriteActionFixtureAs(t *testing.T, dir, filename, fixturePath string) string {
	t.Helper()
	actionPath := filepath.Join(dir, filename)
	fixture := MustLoadActionFixture(t, fixturePath)
	WriteTestFile(t, actionPath, fixture.Content)

	return actionPath
}

// CreateConfigDir creates a standard .config/gh-action-readme directory.
func CreateConfigDir(t *testing.T, baseDir string) string {
	t.Helper()
	configDir := filepath.Join(baseDir, TestDirConfigGhActionReadme)
	// #nosec G301 -- test directory permissions
	if err := os.MkdirAll(configDir, appconstants.FilePermDir); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	return configDir
}

// WriteConfigFile writes a config file to the standard location.
func WriteConfigFile(t *testing.T, baseDir, content string) string {
	t.Helper()
	configDir := CreateConfigDir(t, baseDir)
	configPath := filepath.Join(configDir, appconstants.ConfigFileNameFull)
	WriteTestFile(t, configPath, content)

	return configPath
}

// CreateActionSubdir creates a subdirectory and writes an action fixture to it.
func CreateActionSubdir(t *testing.T, baseDir, subdirName, fixturePath string) string {
	t.Helper()
	subDir := filepath.Join(baseDir, subdirName)
	// #nosec G301 -- test directory permissions
	if err := os.MkdirAll(subDir, appconstants.FilePermDir); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	return WriteActionFixture(t, subDir, fixturePath)
}

// AssertFileExists fails if the file does not exist.
func AssertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("expected file to exist: %s", path)
	}
}

// AssertFileNotExists fails if the file exists.
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	if err == nil {
		// File exists
		t.Fatalf("expected file not to exist: %s", path)
	}
	if err != nil && !os.IsNotExist(err) {
		// Error occurred but it's not a "does not exist" error
		t.Fatalf("error checking file existence: %v", err)
	}
	// err != nil && os.IsNotExist(err) - this is the success case
}

// CreateTestAction creates a test action.yml file content.
func CreateTestAction(name, description string, inputs map[string]string) string {
	var inputsYAML bytes.Buffer
	for key, desc := range inputs {
		inputsYAML.WriteString(fmt.Sprintf("  %s:\n    description: %s\n    required: true\n", key, desc))
	}

	result := fmt.Sprintf(appconstants.YAMLFieldName, name)
	result += fmt.Sprintf(appconstants.YAMLFieldDescription, description)
	result += "inputs:\n"
	result += inputsYAML.String()
	result += "outputs:\n"
	result += "  result:\n"
	result += "    description: 'The result'\n"
	result += appconstants.YAMLFieldRuns
	result += "  using: 'node20'\n"
	result += "  main: 'index.js'\n"
	result += "branding:\n"
	result += "  icon: 'zap'\n"
	result += "  color: 'yellow'\n"

	return result
}

// SetupTestTemplates creates template files for testing.
func SetupTestTemplates(t *testing.T, dir string) {
	t.Helper()

	// Create templates directory structure
	templatesDir := filepath.Join(dir, "templates")
	themesDir := filepath.Join(templatesDir, "themes")

	// Create directories
	for _, theme := range []string{"github", "gitlab", "minimal", "professional"} {
		themeDir := filepath.Join(themesDir, theme)
		// #nosec G301 -- test directory permissions
		if err := os.MkdirAll(themeDir, appconstants.FilePermDir); err != nil {
			t.Fatalf("failed to create theme dir %s: %v", themeDir, err)
		}
		// Write theme template
		templatePath := filepath.Join(themeDir, appconstants.TemplateReadme)
		WriteTestFile(t, templatePath, SimpleTemplate)
	}

	// Create default template
	defaultTemplatePath := filepath.Join(templatesDir, appconstants.TemplateReadme)
	WriteTestFile(t, defaultTemplatePath, SimpleTemplate)
}

// CreateCompositeAction creates a test composite action with dependencies.
func CreateCompositeAction(name, description string, steps []string) string {
	var stepsYAML bytes.Buffer
	for i, step := range steps {
		stepsYAML.WriteString(fmt.Sprintf("  - name: Step %d\n    uses: %s\n", i+1, step))
	}

	result := fmt.Sprintf(appconstants.YAMLFieldName, name)
	result += fmt.Sprintf(appconstants.YAMLFieldDescription, description)
	result += appconstants.YAMLFieldRuns
	result += "  using: 'composite'\n"
	result += "  steps:\n"
	result += stepsYAML.String()

	return result
}

// TestAppConfig represents a test configuration structure.
type TestAppConfig struct {
	Theme        string
	OutputFormat string
	OutputDir    string
	Template     string
	Schema       string
	Verbose      bool
	Quiet        bool
	GitHubToken  string
}

// MockAppConfig creates a test configuration.
func MockAppConfig(overrides *TestAppConfig) *TestAppConfig {
	config := &TestAppConfig{
		Theme:        "default",
		OutputFormat: "md",
		OutputDir:    ".",
		Template:     "",
		Schema:       "schemas/action.schema.json",
		Verbose:      false,
		Quiet:        false,
		GitHubToken:  "",
	}

	if overrides != nil {
		if overrides.Theme != "" {
			config.Theme = overrides.Theme
		}
		if overrides.OutputFormat != "" {
			config.OutputFormat = overrides.OutputFormat
		}
		if overrides.OutputDir != "" {
			config.OutputDir = overrides.OutputDir
		}
		if overrides.Template != "" {
			config.Template = overrides.Template
		}
		if overrides.Schema != "" {
			config.Schema = overrides.Schema
		}
		config.Verbose = overrides.Verbose
		config.Quiet = overrides.Quiet
		if overrides.GitHubToken != "" {
			config.GitHubToken = overrides.GitHubToken
		}
	}

	return config
}

// SetEnv sets an environment variable for testing and returns cleanup function.
func SetEnv(t *testing.T, key, value string) func() {
	t.Helper()

	t.Setenv(key, value)

	return func() {
		// t.Setenv() automatically handles cleanup, so no action needed
	}
}

// WithContext creates a context with timeout for testing.
func WithContext(timeout time.Duration) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	_ = cancel // Avoid lostcancel - we're intentionally creating a context without cleanup for testing

	return ctx
}

// AssertNoError fails the test if err is not nil.
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// AssertError fails the test if err is nil.
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}

// AssertStringContains fails the test if str doesn't contain substring.
func AssertStringContains(t *testing.T, str, substring string) {
	t.Helper()
	if !strings.Contains(str, substring) {
		t.Fatalf("expected string to contain %q, got: %s", substring, str)
	}
}

// AssertEqual fails the test if expected != actual.
func AssertEqual(t *testing.T, expected, actual any) {
	t.Helper()

	// Handle maps which can't be compared directly
	if expectedMap, ok := expected.(map[string]string); ok {
		actualMap, ok := actual.(map[string]string)
		if !ok {
			t.Fatalf("expected map[string]string, got %T", actual)
		}

		if len(expectedMap) != len(actualMap) {
			t.Fatalf("expected map with %d entries, got %d", len(expectedMap), len(actualMap))
		}

		for k, v := range expectedMap {
			if actualMap[k] != v {
				t.Fatalf("expected map[%s] = %s, got %s", k, v, actualMap[k])
			}
		}

		return
	}

	if expected != actual {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}

// AssertSliceContainsAll fails if any of expectedSubstrings is not found in any item of the slice.
// This is useful for checking that suggestions or messages contain expected content.
func AssertSliceContainsAll(t *testing.T, slice []string, expectedSubstrings []string) {
	t.Helper()

	if len(slice) == 0 {
		t.Fatal("slice is empty")
	}

	allItems := strings.Join(slice, " ")
	for _, expected := range expectedSubstrings {
		if !strings.Contains(allItems, expected) {
			t.Errorf(
				"expected to find %q in slice, got:\n%s",
				expected,
				strings.Join(slice, "\n"),
			)
		}
	}
}

// NewStringReader creates an io.ReadCloser from a string.
func NewStringReader(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(s))
}

// GitHubTokenTestCase represents a test case for GitHub token hierarchy testing.
type GitHubTokenTestCase struct {
	Name          string
	SetupFunc     func(t *testing.T) func()
	ExpectedToken string
}

// GetGitHubTokenHierarchyTests returns shared test cases for GitHub token hierarchy.
func GetGitHubTokenHierarchyTests() []GitHubTokenTestCase {
	return []GitHubTokenTestCase{
		{
			Name: "GH_README_GITHUB_TOKEN has highest priority",
			SetupFunc: func(t *testing.T) func() {
				t.Helper()
				cleanup1 := SetEnv(t, appconstants.EnvGitHubToken, "priority-token")
				cleanup2 := SetEnv(t, appconstants.EnvGitHubTokenStandard, appconstants.TokenFallback)

				return func() {
					cleanup1()
					cleanup2()
				}
			},
			ExpectedToken: "priority-token",
		},
		{
			Name: "GITHUB_TOKEN as fallback",
			SetupFunc: func(t *testing.T) func() {
				t.Helper()
				_ = os.Unsetenv(appconstants.EnvGitHubToken)
				cleanup := SetEnv(t, appconstants.EnvGitHubTokenStandard, appconstants.TokenFallback)

				return cleanup
			},
			ExpectedToken: appconstants.TokenFallback,
		},
		{
			Name: "no environment variables",
			SetupFunc: func(t *testing.T) func() {
				t.Helper()
				_ = os.Unsetenv(appconstants.EnvGitHubToken)
				_ = os.Unsetenv(appconstants.EnvGitHubTokenStandard)

				return func() {}
			},
			ExpectedToken: "",
		},
	}
}

// ErrCreateFile returns a formatted error message for file creation failures.
func ErrCreateFile(name string) string {
	return fmt.Sprintf("Failed to create %s: %s", name, "%v")
}

// ErrCreateDir returns a formatted error message for directory creation failures.
func ErrCreateDir(name string) string {
	return fmt.Sprintf("Failed to create %s dir: %s", name, "%v")
}

// ErrDiscoverActionFiles returns the error format string for DiscoverActionFiles failures.
func ErrDiscoverActionFiles() string {
	return "DiscoverActionFiles() error = %v"
}

// InitGitRepo initializes a git repository in the given directory.
// It runs git init and creates an initial commit.
func InitGitRepo(t *testing.T, dir string) {
	t.Helper()

	// Initialize git repo
	cmd := exec.Command("git", "init") // #nosec G204 -- test helper with controlled input
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	// Configure git user for commits
	configCmds := [][]string{
		{"git", "config", "user.name", "Test User"},
		{"git", "config", "user.email", "test@example.com"},
	}

	for _, args := range configCmds {
		cmd := exec.Command(args[0], args[1:]...) // #nosec G204 -- test helper
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to configure git: %v", err)
		}
	}

	// Create an initial commit
	readmePath := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test Repository\n"), 0600); err != nil {
		t.Fatalf("Failed to create README: %v", err)
	}

	addCmd := exec.Command("git", "add", "README.md") // #nosec G204 -- test helper
	addCmd.Dir = dir
	if err := addCmd.Run(); err != nil {
		t.Fatalf("Failed to add file to git: %v", err)
	}

	commitCmd := exec.Command("git", "commit", "-m", "Initial commit") // #nosec G204 -- test helper
	commitCmd.Dir = dir
	if err := commitCmd.Run(); err != nil {
		t.Fatalf("Failed to create initial commit: %v", err)
	}
}
