// Package testutil provides testing utilities and mocks for gh-action-readme.
package testutil

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v57/github"
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

	client := github.NewClient(&http.Client{Transport: &mockTransport{client: mockClient}})

	return client
}

type mockTransport struct {
	client *MockHTTPClient
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.client.Do(req)
}

// TempDir creates a temporary directory for testing and returns cleanup function.
func TempDir(t *testing.T) (string, func()) {
	t.Helper()

	dir, err := os.MkdirTemp("", "gh-action-readme-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	return dir, func() {
		_ = os.RemoveAll(dir)
	}
}

// WriteTestFile writes a test file to the given path.
func WriteTestFile(t *testing.T, path, content string) {
	t.Helper()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil { // #nosec G301 -- test directory permissions
		t.Fatalf("failed to create dir %s: %v", dir, err)
	}

	if err := os.WriteFile(path, []byte(content), 0600); err != nil { // #nosec G306 -- test file permissions
		t.Fatalf("failed to write test file %s: %v", path, err)
	}
}

// MockColoredOutput captures output for testing.
type MockColoredOutput struct {
	Messages []string
	Errors   []string
	Quiet    bool
}

// NewMockColoredOutput creates a new mock colored output.
func NewMockColoredOutput(quiet bool) *MockColoredOutput {
	return &MockColoredOutput{Quiet: quiet}
}

// Info captures info messages.
func (m *MockColoredOutput) Info(format string, args ...any) {
	if !m.Quiet {
		m.Messages = append(m.Messages, fmt.Sprintf("INFO: "+format, args...))
	}
}

// Success captures success messages.
func (m *MockColoredOutput) Success(format string, args ...any) {
	if !m.Quiet {
		m.Messages = append(m.Messages, fmt.Sprintf("SUCCESS: "+format, args...))
	}
}

// Warning captures warning messages.
func (m *MockColoredOutput) Warning(format string, args ...any) {
	if !m.Quiet {
		m.Messages = append(m.Messages, fmt.Sprintf("WARNING: "+format, args...))
	}
}

// Error captures error messages.
func (m *MockColoredOutput) Error(format string, args ...any) {
	m.Errors = append(m.Errors, fmt.Sprintf("ERROR: "+format, args...))
}

// Bold captures bold messages.
func (m *MockColoredOutput) Bold(format string, args ...any) {
	if !m.Quiet {
		m.Messages = append(m.Messages, fmt.Sprintf("BOLD: "+format, args...))
	}
}

// Printf captures printf messages.
func (m *MockColoredOutput) Printf(format string, args ...any) {
	if !m.Quiet {
		m.Messages = append(m.Messages, fmt.Sprintf(format, args...))
	}
}

// Reset clears all captured messages.
func (m *MockColoredOutput) Reset() {
	m.Messages = nil
	m.Errors = nil
}

// HasMessage checks if a message contains the given substring.
func (m *MockColoredOutput) HasMessage(substring string) bool {
	for _, msg := range m.Messages {
		if strings.Contains(msg, substring) {
			return true
		}
	}

	return false
}

// HasError checks if an error contains the given substring.
func (m *MockColoredOutput) HasError(substring string) bool {
	for _, err := range m.Errors {
		if strings.Contains(err, substring) {
			return true
		}
	}

	return false
}

// CreateTestAction creates a test action.yml file content.
func CreateTestAction(name, description string, inputs map[string]string) string {
	var inputsYAML bytes.Buffer
	for key, desc := range inputs {
		inputsYAML.WriteString(fmt.Sprintf("  %s:\n    description: %s\n    required: true\n", key, desc))
	}

	result := fmt.Sprintf("name: %s\n", name)
	result += fmt.Sprintf("description: %s\n", description)
	result += "inputs:\n"
	result += inputsYAML.String()
	result += "outputs:\n"
	result += "  result:\n"
	result += "    description: 'The result'\n"
	result += "runs:\n"
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
		if err := os.MkdirAll(themeDir, 0750); err != nil { // #nosec G301 -- test directory permissions
			t.Fatalf("failed to create theme dir %s: %v", themeDir, err)
		}
		// Write theme template
		templatePath := filepath.Join(themeDir, "readme.tmpl")
		WriteTestFile(t, templatePath, SimpleTemplate)
	}

	// Create default template
	defaultTemplatePath := filepath.Join(templatesDir, "readme.tmpl")
	WriteTestFile(t, defaultTemplatePath, SimpleTemplate)
}

// CreateCompositeAction creates a test composite action with dependencies.
func CreateCompositeAction(name, description string, steps []string) string {
	var stepsYAML bytes.Buffer
	for i, step := range steps {
		stepsYAML.WriteString(fmt.Sprintf("  - name: Step %d\n    uses: %s\n", i+1, step))
	}

	result := fmt.Sprintf("name: %s\n", name)
	result += fmt.Sprintf("description: %s\n", description)
	result += "runs:\n"
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

	original := os.Getenv(key)
	_ = os.Setenv(key, value)

	return func() {
		if original == "" {
			_ = os.Unsetenv(key)
		} else {
			_ = os.Setenv(key, original)
		}
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

// NewStringReader creates an io.ReadCloser from a string.
func NewStringReader(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(s))
}
