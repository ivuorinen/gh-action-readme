// Package testutil provides testing utilities and mocks for gh-action-readme.
package testutil

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-github/v74/github"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

// MockHTTPClient is a mock HTTP client for testing.
//
// The Requests slice is mutex-guarded, and each Do returns a shallow copy of the
// configured response with a FRESH body reader (the configured body is buffered
// once, on first use, into bodies). That makes a single mock client safe to share
// across parallel callers requesting the same method+URL — e.g. the bounded
// worker pool in Analyzer.CheckOutdated — without racing on a single-use body.
type MockHTTPClient struct {
	mu        sync.Mutex
	Responses map[string]*http.Response
	Requests  []*http.Request
	bodies    map[string][]byte // buffered response bodies, keyed by method+URL
}

// Do implements the http.Client interface.
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Requests = append(m.Requests, req)

	key := req.Method + " " + req.URL.String()
	if resp, ok := m.Responses[key]; ok {
		body := m.bufferBodyLocked(key, resp)
		clone := *resp
		clone.Body = io.NopCloser(bytes.NewReader(body))

		return &clone, nil
	}

	// Default 404 response
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader(`{"error": "not found"}`)),
	}, nil
}

// bufferBodyLocked reads a configured response body once into a byte buffer so
// every subsequent Do can serve an independent reader. Must be called with m.mu
// held.
func (m *MockHTTPClient) bufferBodyLocked(key string, resp *http.Response) []byte {
	if m.bodies == nil {
		m.bodies = make(map[string][]byte)
	}
	if buf, ok := m.bodies[key]; ok {
		return buf
	}

	var buf []byte
	if resp.Body != nil {
		buf, _ = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
	}
	m.bodies[key] = buf

	return buf
}

// MockGitHubClient creates a GitHub client with mocked responses.
func MockGitHubClient(responses map[string]string) *github.Client {
	client, _ := MockGitHubClientWithSpy(responses)

	return client
}

// MockGitHubClientWithSpy is like MockGitHubClient but also returns the underlying
// MockHTTPClient so tests can inspect how many upstream requests were made (via its
// Requests slice) — e.g. to prove a cache prevented a second fetch.
func MockGitHubClientWithSpy(responses map[string]string) (*github.Client, *MockHTTPClient) {
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

	return client, mockClient
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

// mustLoadActionFixture loads an action fixture and fails the test on error.
// This helper consolidates the load + assertion pattern.
func mustLoadActionFixture(t *testing.T, path string) *ActionFixture {
	t.Helper()
	fixture, err := LoadActionFixture(path)
	AssertNoError(t, err)

	return fixture
}

// WriteTestFile writes a test file to the given path.
func WriteTestFile(t *testing.T, path, content string) {
	t.Helper()

	dir := filepath.Dir(path)
	// #nosec G301 -- test directory permissions
	if err := os.MkdirAll(dir, appconstants.FilePermDir); err != nil {
		t.Fatalf("failed to create dir %s: %v", dir, err)
	}

	// #nosec G306 G703 -- test file permissions, path is controlled by test infrastructure
	if err := os.WriteFile(path, []byte(content), appconstants.FilePermDefault); err != nil {
		t.Fatalf("failed to write test file %s: %v", path, err)
	}
}

// WriteFileInDir writes a file with the given filename in the specified directory.
// This is a convenience wrapper that combines filepath.Join + WriteTestFile.
// Eliminates the pattern: path := filepath.Join(dir, filename); WriteTestFile(t, path, content).
func WriteFileInDir(t *testing.T, dir, filename, content string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	WriteTestFile(t, path, content)

	return path
}

// WriteActionFixture writes an action fixture to a standard action.yml file.
func WriteActionFixture(t *testing.T, dir, fixturePath string) string {
	t.Helper()
	actionPath := filepath.Join(dir, appconstants.ActionFileNameYML)
	fixture := mustLoadActionFixture(t, fixturePath)
	WriteTestFile(t, actionPath, fixture.Content)

	return actionPath
}

// WriteActionFixtureAs writes an action fixture with a custom filename.
func WriteActionFixtureAs(t *testing.T, dir, filename, fixturePath string) string {
	t.Helper()
	actionPath := filepath.Join(dir, filename)
	fixture := mustLoadActionFixture(t, fixturePath)
	WriteTestFile(t, actionPath, fixture.Content)

	return actionPath
}

// CreateActionInTempDir creates a temporary directory with an action.yml file.
// This is a convenience wrapper for the common pattern of t.TempDir() + WriteTestFile.
// Returns the temp directory path and the full path to the action.yml file.
//
// Example:
//
//	tmpDir, actionPath := testutil.CreateActionInTempDir(t, "name: Test")
func CreateActionInTempDir(t *testing.T, yamlContent string) (tmpDir, actionPath string) {
	t.Helper()

	tmpDir = t.TempDir()
	actionPath = filepath.Join(tmpDir, appconstants.ActionFileNameYML)
	WriteTestFile(t, actionPath, yamlContent)

	return tmpDir, actionPath
}

// CreateNestedAction creates a nested action directory structure with an action.yml file.
// This is useful for testing monorepo scenarios with multiple actions in subdirectories.
// Returns the subdirectory path and the full path to the action.yml file.
//
// Example:
//
//	dirPath, actionPath := testutil.CreateNestedAction(t, tmpDir, "actions/build", "name: Build")
func CreateNestedAction(t *testing.T, baseDir, subdir, yamlContent string) (dirPath, actionPath string) {
	t.Helper()

	dirPath = filepath.Join(baseDir, subdir)
	// #nosec G301 -- test directory permissions
	if err := os.MkdirAll(dirPath, appconstants.FilePermDir); err != nil {
		t.Fatalf("failed to create nested directory %s: %v", subdir, err)
	}

	actionPath = filepath.Join(dirPath, appconstants.ActionFileNameYML)
	WriteTestFile(t, actionPath, yamlContent)

	return dirPath, actionPath
}

// CreateTestSubdir creates a subdirectory within the base directory.
// This is useful for test setup that needs directory structures without action files.
// Returns the full path to the created subdirectory.
//
// Example:
//
//	subdir := testutil.CreateTestSubdir(t, tmpDir, ".config", "gh-action-readme")
//	// Creates tmpDir/.config/gh-action-readme
func CreateTestSubdir(t *testing.T, baseDir string, subdirs ...string) string {
	t.Helper()

	pathParts := append([]string{baseDir}, subdirs...)
	fullPath := filepath.Join(pathParts...)

	// #nosec G301 -- test directory permissions
	if err := os.MkdirAll(fullPath, appconstants.FilePermDir); err != nil {
		t.Fatalf("failed to create test subdirectory %s: %v", fullPath, err)
	}

	return fullPath
}

// CreateTestDir creates a directory with test-appropriate permissions (0750).
// Automatically fails the test if directory creation fails.
// This is a convenience wrapper to reduce the 30+ instances of:
//
//	if err := os.MkdirAll(dir, 0750); err != nil { t.Fatalf(...) }
//
// Example:
//
//	testutil.CreateTestDir(t, filepath.Join(tmpDir, ".git"))
func CreateTestDir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0750); err != nil { // #nosec G301 -- test directory permissions
		t.Fatalf("failed to create directory %s: %v", path, err)
	}
}

// RunBinaryCommand executes the built binary with arguments in the given directory.
// Returns the combined output (stdout + stderr) and error for verification in tests.
// This helper consolidates the common pattern of running subprocess commands in integration tests.
//
// Example:
//
//	output, err := testutil.RunBinaryCommand(t, binaryPath, tmpDir, "gen", "--theme", "github")
//	testutil.AssertNoError(t, err)
//	if !strings.Contains(output, "Generated") {
//	    t.Error("expected success message in output")
//	}
//
// RunBinaryCommand executes the compiled binary under test with the given args.
// TEST-ONLY: binaryPath is always the compiled test binary, never user-supplied input.
func RunBinaryCommand(t *testing.T, binaryPath, dir string, args ...string) (output string, err error) {
	t.Helper()

	cmd := exec.Command(binaryPath, args...) // #nosec G204 -- test-only: binaryPath is always the compiled test binary
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()

	return string(out), err
}

// WriteConfigFile writes a config file to the standard location.
func WriteConfigFile(t *testing.T, baseDir, content string) string {
	t.Helper()
	configDir := filepath.Join(baseDir, TestDirConfigGhActionReadme)
	// #nosec G301 -- test directory permissions
	if err := os.MkdirAll(configDir, appconstants.FilePermDir); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	configPath := filepath.Join(configDir, appconstants.ConfigFileNameFull)
	WriteTestFile(t, configPath, content)

	return configPath
}

// SetupConfigEnvironment sets up HOME and XDG_CONFIG_HOME environment variables for testing.
// This is commonly needed for config hierarchy tests.
//
// Example:
//
//	testutil.SetupConfigEnvironment(t, tmpDir)
func SetupConfigEnvironment(t *testing.T, tmpDir string) {
	t.Helper()
	t.Setenv(EnvVarHOME, tmpDir)
	t.Setenv(EnvVarXDGConfigHome, filepath.Join(tmpDir, TestDirDotConfig))
}

// CreateGitRepoWithRemote initializes a git repository and sets up a remote.
// Returns the path to the git config file for further customization if needed.
//
// Example:
//
//	testutil.CreateGitRepoWithRemote(t, tmpDir, "https://github.com/user/repo.git")
func CreateGitRepoWithRemote(t *testing.T, tmpDir, remoteURL string) string {
	t.Helper()

	InitGitRepo(t, tmpDir)

	gitDir := filepath.Join(tmpDir, ConfigFieldGit)
	configPath := filepath.Join(gitDir, "config")

	configContent := fmt.Sprintf(`[remote "origin"]
	url = %s
	fetch = +refs/heads/*:refs/remotes/origin/*
[branch "main"]
	remote = origin
	merge = refs/heads/main
`, remoteURL)

	WriteTestFile(t, configPath, configContent)

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

// SetupTestTemplates creates template files for testing.
func SetupTestTemplates(t *testing.T, dir string) {
	t.Helper()

	// Create templates directory structure
	templatesDir := filepath.Join(dir, "templates")
	themesDir := filepath.Join(templatesDir, "themes")

	// Create directories
	for _, theme := range []string{TestThemeGitHub, TestThemeGitLab, TestThemeMinimal, TestThemeProfessional} {
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
		fmt.Fprintf(&stepsYAML, "  - name: Step %d\n    uses: %s\n", i+1, step)
	}

	result := fmt.Sprintf(appconstants.YAMLFieldName, name)
	result += fmt.Sprintf(appconstants.YAMLFieldDescription, description)
	result += appconstants.YAMLFieldRuns
	result += "  using: 'composite'\n"
	result += "  steps:\n"
	result += stepsYAML.String()

	return result
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

	if ok, msg := equalCheck(expected, actual); !ok {
		t.Fatal(msg)
	}
}

// equalCheck reports whether actual equals expected, special-casing
// map[string]string (which is not directly comparable), and returns a failure
// message when they differ. It is a pure function so the comparison — including
// the failure path — can be unit-tested directly; AssertEqual itself takes a
// *testing.T, which cannot be mocked to assert that it actually fails on
// unequal values.
func equalCheck(expected, actual any) (ok bool, msg string) {
	if expectedMap, isMap := expected.(map[string]string); isMap {
		actualMap, isMap := actual.(map[string]string)
		if !isMap {
			return false, fmt.Sprintf("expected map[string]string, got %T", actual)
		}
		if len(expectedMap) != len(actualMap) {
			return false, fmt.Sprintf("expected map with %d entries, got %d", len(expectedMap), len(actualMap))
		}
		for k, v := range expectedMap {
			// Use the two-value form: a missing key reads as "" and would falsely
			// match an expected "" value when the lengths happen to be equal.
			actualVal, exists := actualMap[k]
			if !exists || actualVal != v {
				return false, fmt.Sprintf("expected map[%s] = %s, got %s", k, v, actualVal)
			}
		}

		return true, ""
	}

	if expected != actual {
		return false, fmt.Sprintf("expected %v, got %v", expected, actual)
	}

	return true, ""
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
				t.Setenv(appconstants.EnvGitHubToken, "priority-token")
				t.Setenv(appconstants.EnvGitHubTokenStandard, appconstants.TokenFallback)

				return func() {}
			},
			ExpectedToken: "priority-token",
		},
		{
			Name: "GITHUB_TOKEN as fallback",
			SetupFunc: func(t *testing.T) func() {
				t.Helper()
				_ = os.Unsetenv(appconstants.EnvGitHubToken)
				t.Setenv(appconstants.EnvGitHubTokenStandard, appconstants.TokenFallback)

				return func() {}
			},
			ExpectedToken: appconstants.TokenFallback,
		},
		{
			Name: "no environment variables",
			SetupFunc: func(t *testing.T) func() {
				t.Helper()
				_ = os.Unsetenv(appconstants.EnvGitHubToken)
				_ = os.Unsetenv(appconstants.EnvGitHubTokenStandard)

				return func() {
					// No cleanup required: environment variables explicitly unset for this scenario.
				}
			},
			ExpectedToken: "",
		},
	}
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
	cmd := exec.Command(appconstants.GitCommand, "init") // #nosec G204 -- test helper with controlled input
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	// Configure git user for commits. Also disable commit signing so the helper
	// is hermetic: a developer's global commit.gpgsign (e.g. a 1Password/SSH
	// signer) must not be invoked here, or commits fail non-interactively.
	configCmds := [][]string{
		{appconstants.GitCommand, ConfigFieldName, "user.name", "Test User"},
		{appconstants.GitCommand, ConfigFieldName, "user.email", "test@example.com"},
		{appconstants.GitCommand, ConfigFieldName, "commit.gpgsign", "false"},
		{appconstants.GitCommand, ConfigFieldName, "tag.gpgsign", "false"},
	}

	for _, args := range configCmds {
		cmd := exec.Command(args[0], args[1:]...) // #nosec G204 -- test helper
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to configure git: %v", err)
		}
	}

	// Create an initial commit
	readmePath := filepath.Join(dir, appconstants.ReadmeMarkdown)
	if err := os.WriteFile(readmePath, []byte("# Test Repository\n"), appconstants.FilePermDefault); err != nil {
		t.Fatalf("Failed to create README: %v", err)
	}

	addCmd := exec.Command(appconstants.GitCommand, "add", appconstants.ReadmeMarkdown) // #nosec G204 -- test helper
	addCmd.Dir = dir
	if err := addCmd.Run(); err != nil {
		t.Fatalf("Failed to add file to git: %v", err)
	}

	commitCmd := exec.Command(appconstants.GitCommand, "commit", "-m", "Initial commit") // #nosec G204 -- test helper
	commitCmd.Dir = dir
	if err := commitCmd.Run(); err != nil {
		t.Fatalf("Failed to create initial commit: %v", err)
	}
}

// CaptureStdout captures stdout output during function execution.
// Useful for testing functions that write to os.Stdout.
func CaptureStdout(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	_ = w.Close() // Ignore error in test helper
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Ignore error in test helper

	return buf.String()
}

// CaptureStderr captures stderr output during function execution.
// Useful for testing functions that write to os.Stderr.
func CaptureStderr(f func()) string {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	f()

	_ = w.Close() // Ignore error in test helper
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Ignore error in test helper

	return buf.String()
}

// CreateTempActionFile creates a temporary action.yml file with content.
// Returns the file path. File is automatically cleaned up by t.TempDir().
// Used to eliminate duplication in parser tests (4 occurrences).
func CreateTempActionFile(t *testing.T, content string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp(t.TempDir(), TestActionFilePattern)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		_ = tmpFile.Close()
		t.Fatalf("failed to write temp file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	return tmpFile.Name()
}

// SetupTestEnvironment creates a temp directory and sets up config environment variables.
// Returns temp directory path and cleanup function.
// Consolidates the common pattern: TempDir + XDG_CONFIG_HOME + HOME setup.
//
// Example:
//
//	tmpDir, cleanup := testutil.SetupTestEnvironment(t)
//	defer cleanup()
func SetupTestEnvironment(t *testing.T) (tmpDir string, cleanup func()) {
	t.Helper()
	tmpDir, cleanup = TempDir(t)
	t.Setenv(EnvVarXDGConfigHome, tmpDir)
	t.Setenv(EnvVarHOME, tmpDir)

	return tmpDir, cleanup
}

// SetupTokenEnv sets up GitHub token environment variables for testing.
// Pass empty string to clear a token.
//
// Example:
//
//	testutil.SetupTokenEnv(t, "tool-token", "standard-token")
func SetupTokenEnv(t *testing.T, toolToken, standardToken string) {
	t.Helper()
	t.Setenv(appconstants.EnvGitHubToken, toolToken)
	t.Setenv(appconstants.EnvGitHubTokenStandard, standardToken)
}

// SetupXDGEnv sets XDG_CONFIG_HOME and HOME environment variables.
// Pass an empty string to explicitly clear (unset) that variable.
//
// Example:
//
//	testutil.SetupXDGEnv(t, tmpDir, "")  // Set XDG, clear HOME
func SetupXDGEnv(t *testing.T, xdgConfigHome, home string) {
	t.Helper()
	t.Setenv(EnvVarXDGConfigHome, xdgConfigHome)
	t.Setenv(EnvVarHOME, home)
}
