package testutil

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestMockHTTPClient tests the MockHTTPClient implementation.
func TestMockHTTPClient(t *testing.T) {
	t.Run("returns configured response", func(t *testing.T) {
		testMockHTTPClientConfiguredResponse(t)
	})

	t.Run("returns 404 for unconfigured endpoints", func(t *testing.T) {
		testMockHTTPClientUnconfiguredEndpoints(t)
	})

	t.Run("tracks requests", func(t *testing.T) {
		testMockHTTPClientRequestTracking(t)
	})
}

// testMockHTTPClientConfiguredResponse tests that configured responses are returned correctly.
func testMockHTTPClientConfiguredResponse(t *testing.T) {
	client := createMockHTTPClientWithResponse("GET https://api.github.com/test", 200, `{"test": "response"}`)

	req := createTestRequest(t, "GET", "https://api.github.com/test")
	resp := executeRequest(t, client, req)
	defer func() { _ = resp.Body.Close() }()

	validateResponseStatus(t, resp, 200)
	validateResponseBody(t, resp, `{"test": "response"}`)
}

// testMockHTTPClientUnconfiguredEndpoints tests that unconfigured endpoints return 404.
func testMockHTTPClientUnconfiguredEndpoints(t *testing.T) {
	client := &MockHTTPClient{
		Responses: make(map[string]*http.Response),
	}

	req := createTestRequest(t, "GET", "https://api.github.com/nonexistent")
	resp := executeRequest(t, client, req)
	defer func() { _ = resp.Body.Close() }()

	validateResponseStatus(t, resp, 404)
}

// testMockHTTPClientRequestTracking tests that requests are tracked correctly.
func testMockHTTPClientRequestTracking(t *testing.T) {
	client := &MockHTTPClient{
		Responses: make(map[string]*http.Response),
	}

	req1 := createTestRequest(t, "GET", "https://api.github.com/test1")
	req2 := createTestRequest(t, "POST", "https://api.github.com/test2")

	executeAndCloseResponse(client, req1)
	executeAndCloseResponse(client, req2)

	validateRequestTracking(t, client, 2, "https://api.github.com/test1", "POST")
}

// createMockHTTPClientWithResponse creates a mock HTTP client with a single configured response.
func createMockHTTPClientWithResponse(key string, statusCode int, body string) *MockHTTPClient {
	return &MockHTTPClient{
		Responses: map[string]*http.Response{
			key: {
				StatusCode: statusCode,
				Body:       io.NopCloser(strings.NewReader(body)),
			},
		},
	}
}

// createTestRequest creates an HTTP request for testing purposes.
func createTestRequest(t *testing.T, method, url string) *http.Request {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	return req
}

// executeRequest executes an HTTP request and returns the response.
func executeRequest(t *testing.T, client *MockHTTPClient, req *http.Request) *http.Response {
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return resp
}

// executeAndCloseResponse executes a request and closes the response body.
func executeAndCloseResponse(client *MockHTTPClient, req *http.Request) {
	if resp, _ := client.Do(req); resp != nil {
		_ = resp.Body.Close()
	}
}

// validateResponseStatus validates that the response has the expected status code.
func validateResponseStatus(t *testing.T, resp *http.Response, expectedStatus int) {
	if resp.StatusCode != expectedStatus {
		t.Errorf("expected status %d, got %d", expectedStatus, resp.StatusCode)
	}
}

// validateResponseBody validates that the response body matches the expected content.
func validateResponseBody(t *testing.T, resp *http.Response, expected string) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if string(body) != expected {
		t.Errorf("expected body %s, got %s", expected, string(body))
	}
}

// validateRequestTracking validates that requests are tracked correctly.
func validateRequestTracking(
	t *testing.T,
	client *MockHTTPClient,
	expectedCount int,
	expectedURL, expectedMethod string,
) {
	if len(client.Requests) != expectedCount {
		t.Errorf("expected %d tracked requests, got %d", expectedCount, len(client.Requests))
		return
	}

	if client.Requests[0].URL.String() != expectedURL {
		t.Errorf("unexpected first request URL: %s", client.Requests[0].URL.String())
	}

	if len(client.Requests) > 1 && client.Requests[1].Method != expectedMethod {
		t.Errorf("unexpected second request method: %s", client.Requests[1].Method)
	}
}

func TestMockGitHubClient(t *testing.T) {
	t.Run("creates client with mocked responses", func(t *testing.T) {
		responses := map[string]string{
			"GET https://api.github.com/repos/test/repo": `{"name": "repo", "full_name": "test/repo"}`,
		}

		client := MockGitHubClient(responses)
		if client == nil {
			t.Fatal("expected client to be created")
		}

		// Test that we can make a request (this would normally hit the API)
		// The mock transport should handle this
		ctx := context.Background()
		_, resp, err := client.Repositories.Get(ctx, "test", "repo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.StatusCode != 200 {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("uses MockGitHubResponses", func(t *testing.T) {
		responses := MockGitHubResponses()
		client := MockGitHubClient(responses)

		// Test a specific endpoint that we know is mocked
		ctx := context.Background()
		_, resp, err := client.Repositories.Get(ctx, "actions", "checkout")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.StatusCode != 200 {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})
}

func TestMockTransport(t *testing.T) {
	client := &MockHTTPClient{
		Responses: map[string]*http.Response{
			"GET https://api.github.com/test": {
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"success": true}`)),
			},
		},
	}

	transport := &mockTransport{client: client}

	req, err := http.NewRequest("GET", "https://api.github.com/test", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestTempDir(t *testing.T) {
	t.Run("creates temporary directory", func(t *testing.T) {
		dir, cleanup := TempDir(t)
		defer cleanup()

		// Verify directory exists
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Error("temporary directory was not created")
		}

		// Verify it's in temp location
		if !strings.Contains(dir, os.TempDir()) && !strings.Contains(dir, "/tmp") {
			t.Errorf("directory not in temp location: %s", dir)
		}

		// Verify directory name pattern
		if !strings.Contains(filepath.Base(dir), "gh-action-readme-test-") {
			t.Errorf("unexpected directory name pattern: %s", dir)
		}
	})

	t.Run("cleanup removes directory", func(t *testing.T) {
		dir, cleanup := TempDir(t)

		// Verify directory exists
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Error("temporary directory was not created")
		}

		// Clean up
		cleanup()

		// Verify directory is removed
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Error("temporary directory was not cleaned up")
		}
	})
}

func TestWriteTestFile(t *testing.T) {
	tmpDir, cleanup := TempDir(t)
	defer cleanup()

	t.Run("writes file with content", func(t *testing.T) {
		testPath := filepath.Join(tmpDir, "test.txt")
		testContent := "Hello, World!"

		WriteTestFile(t, testPath, testContent)

		// Verify file exists
		if _, err := os.Stat(testPath); os.IsNotExist(err) {
			t.Error("file was not created")
		}

		// Verify content
		content, err := os.ReadFile(testPath) // #nosec G304 -- test file path
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}

		if string(content) != testContent {
			t.Errorf("expected content %s, got %s", testContent, string(content))
		}
	})

	t.Run("creates nested directories", func(t *testing.T) {
		nestedPath := filepath.Join(tmpDir, "nested", "deep", "file.txt")
		testContent := "nested content"

		WriteTestFile(t, nestedPath, testContent)

		// Verify file exists
		if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
			t.Error("nested file was not created")
		}

		// Verify parent directories exist
		parentDir := filepath.Dir(nestedPath)
		if _, err := os.Stat(parentDir); os.IsNotExist(err) {
			t.Error("parent directories were not created")
		}
	})

	t.Run("sets correct permissions", func(t *testing.T) {
		testPath := filepath.Join(tmpDir, "perm-test.txt")
		WriteTestFile(t, testPath, "test")

		info, err := os.Stat(testPath)
		if err != nil {
			t.Fatalf("failed to stat file: %v", err)
		}

		// File should have 0600 permissions
		expectedPerm := os.FileMode(0600)
		if info.Mode().Perm() != expectedPerm {
			t.Errorf("expected permissions %v, got %v", expectedPerm, info.Mode().Perm())
		}
	})
}

func TestSetupTestTemplates(t *testing.T) {
	tmpDir, cleanup := TempDir(t)
	defer cleanup()

	SetupTestTemplates(t, tmpDir)

	// Verify template directories exist
	templatesDir := filepath.Join(tmpDir, "templates")
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		t.Error("templates directory was not created")
	}

	// Verify theme directories exist
	themes := []string{"github", "gitlab", "minimal", "professional"}
	for _, theme := range themes {
		themeDir := filepath.Join(templatesDir, "themes", theme)
		if _, err := os.Stat(themeDir); os.IsNotExist(err) {
			t.Errorf("theme directory %s was not created", theme)
		}

		// Verify theme template file exists
		templateFile := filepath.Join(themeDir, "readme.tmpl")
		if _, err := os.Stat(templateFile); os.IsNotExist(err) {
			t.Errorf("template file for theme %s was not created", theme)
		}

		// Verify template content
		content, err := os.ReadFile(templateFile) // #nosec G304 -- test file path
		if err != nil {
			t.Errorf("failed to read template file for theme %s: %v", theme, err)
		}

		if string(content) != SimpleTemplate {
			t.Errorf("template content for theme %s doesn't match SimpleTemplate", theme)
		}
	}

	// Verify default template exists
	defaultTemplate := filepath.Join(templatesDir, "readme.tmpl")
	if _, err := os.Stat(defaultTemplate); os.IsNotExist(err) {
		t.Error("default template was not created")
	}
}

func TestMockColoredOutput(t *testing.T) {
	t.Run("creates mock output", func(t *testing.T) {
		testMockColoredOutputCreation(t)
	})
	t.Run("creates quiet mock output", func(t *testing.T) {
		testMockColoredOutputQuietCreation(t)
	})
	t.Run("captures info messages", func(t *testing.T) {
		testMockColoredOutputInfoMessages(t)
	})
	t.Run("captures success messages", func(t *testing.T) {
		testMockColoredOutputSuccessMessages(t)
	})
	t.Run("captures warning messages", func(t *testing.T) {
		testMockColoredOutputWarningMessages(t)
	})
	t.Run("captures error messages", func(t *testing.T) {
		testMockColoredOutputErrorMessages(t)
	})
	t.Run("captures bold messages", func(t *testing.T) {
		testMockColoredOutputBoldMessages(t)
	})
	t.Run("captures printf messages", func(t *testing.T) {
		testMockColoredOutputPrintfMessages(t)
	})
	t.Run("quiet mode suppresses non-error messages", func(t *testing.T) {
		testMockColoredOutputQuietMode(t)
	})
	t.Run("HasMessage works correctly", func(t *testing.T) {
		testMockColoredOutputHasMessage(t)
	})
	t.Run("HasError works correctly", func(t *testing.T) {
		testMockColoredOutputHasError(t)
	})
	t.Run("Reset clears messages and errors", func(t *testing.T) {
		testMockColoredOutputReset(t)
	})
}

// testMockColoredOutputCreation tests basic mock output creation.
func testMockColoredOutputCreation(t *testing.T) {
	output := NewMockColoredOutput(false)
	validateMockOutputCreated(t, output)
	validateQuietMode(t, output, false)
	validateEmptyMessagesAndErrors(t, output)
}

// testMockColoredOutputQuietCreation tests quiet mock output creation.
func testMockColoredOutputQuietCreation(t *testing.T) {
	output := NewMockColoredOutput(true)
	validateQuietMode(t, output, true)
}

// testMockColoredOutputInfoMessages tests info message capture.
func testMockColoredOutputInfoMessages(t *testing.T) {
	output := NewMockColoredOutput(false)
	output.Info("test info: %s", "value")
	validateSingleMessage(t, output, "INFO: test info: value")
}

// testMockColoredOutputSuccessMessages tests success message capture.
func testMockColoredOutputSuccessMessages(t *testing.T) {
	output := NewMockColoredOutput(false)
	output.Success("operation completed")
	validateSingleMessage(t, output, "SUCCESS: operation completed")
}

// testMockColoredOutputWarningMessages tests warning message capture.
func testMockColoredOutputWarningMessages(t *testing.T) {
	output := NewMockColoredOutput(false)
	output.Warning("this is a warning")
	validateSingleMessage(t, output, "WARNING: this is a warning")
}

// testMockColoredOutputErrorMessages tests error message capture.
func testMockColoredOutputErrorMessages(t *testing.T) {
	output := NewMockColoredOutput(false)
	output.Error("error occurred: %d", 404)
	validateSingleError(t, output, "ERROR: error occurred: 404")

	// Test errors in quiet mode
	output.Quiet = true
	output.Error("quiet error")
	validateErrorCount(t, output, 2)
}

// testMockColoredOutputBoldMessages tests bold message capture.
func testMockColoredOutputBoldMessages(t *testing.T) {
	output := NewMockColoredOutput(false)
	output.Bold("bold text")
	validateSingleMessage(t, output, "BOLD: bold text")
}

// testMockColoredOutputPrintfMessages tests printf message capture.
func testMockColoredOutputPrintfMessages(t *testing.T) {
	output := NewMockColoredOutput(false)
	output.Printf("formatted: %s = %d", "key", 42)
	validateSingleMessage(t, output, "formatted: key = 42")
}

// testMockColoredOutputQuietMode tests quiet mode behavior.
func testMockColoredOutputQuietMode(t *testing.T) {
	output := NewMockColoredOutput(true)

	// Send various message types
	output.Info("info message")
	output.Success("success message")
	output.Warning("warning message")
	output.Bold("bold message")
	output.Printf("printf message")

	validateMessageCount(t, output, 0)

	// Errors should still be captured
	output.Error("error message")
	validateErrorCount(t, output, 1)
}

// testMockColoredOutputHasMessage tests HasMessage functionality.
func testMockColoredOutputHasMessage(t *testing.T) {
	output := NewMockColoredOutput(false)
	output.Info("test message with keyword")
	output.Success("another message")

	validateMessageContains(t, output, "keyword", true)
	validateMessageContains(t, output, "another", true)
	validateMessageContains(t, output, "nonexistent", false)
}

// testMockColoredOutputHasError tests HasError functionality.
func testMockColoredOutputHasError(t *testing.T) {
	output := NewMockColoredOutput(false)
	output.Error("connection failed")
	output.Error("timeout occurred")

	validateErrorContains(t, output, "connection", true)
	validateErrorContains(t, output, "timeout", true)
	validateErrorContains(t, output, "success", false)
}

// testMockColoredOutputReset tests Reset functionality.
func testMockColoredOutputReset(t *testing.T) {
	output := NewMockColoredOutput(false)
	output.Info("test message")
	output.Error("test error")

	validateNonEmptyMessagesAndErrors(t, output)

	output.Reset()

	validateEmptyMessagesAndErrors(t, output)
}

// Helper functions for validation

// validateMockOutputCreated validates that mock output was created successfully.
func validateMockOutputCreated(t *testing.T, output *MockColoredOutput) {
	if output == nil {
		t.Fatal("expected output to be created")
	}
}

// validateQuietMode validates the quiet mode setting.
func validateQuietMode(t *testing.T, output *MockColoredOutput, expected bool) {
	if output.Quiet != expected {
		t.Errorf("expected Quiet to be %v, got %v", expected, output.Quiet)
	}
}

// validateEmptyMessagesAndErrors validates that messages and errors are empty.
func validateEmptyMessagesAndErrors(t *testing.T, output *MockColoredOutput) {
	validateMessageCount(t, output, 0)
	validateErrorCount(t, output, 0)
}

// validateNonEmptyMessagesAndErrors validates that messages and errors are present.
func validateNonEmptyMessagesAndErrors(t *testing.T, output *MockColoredOutput) {
	if len(output.Messages) == 0 || len(output.Errors) == 0 {
		t.Fatal("expected messages and errors to be present before reset")
	}
}

// validateSingleMessage validates a single message was captured.
func validateSingleMessage(t *testing.T, output *MockColoredOutput, expected string) {
	validateMessageCount(t, output, 1)
	if output.Messages[0] != expected {
		t.Errorf("expected message %s, got %s", expected, output.Messages[0])
	}
}

// validateSingleError validates a single error was captured.
func validateSingleError(t *testing.T, output *MockColoredOutput, expected string) {
	validateErrorCount(t, output, 1)
	if output.Errors[0] != expected {
		t.Errorf("expected error %s, got %s", expected, output.Errors[0])
	}
}

// validateMessageCount validates the message count.
func validateMessageCount(t *testing.T, output *MockColoredOutput, expected int) {
	if len(output.Messages) != expected {
		t.Errorf("expected %d messages, got %d", expected, len(output.Messages))
	}
}

// validateErrorCount validates the error count.
func validateErrorCount(t *testing.T, output *MockColoredOutput, expected int) {
	if len(output.Errors) != expected {
		t.Errorf("expected %d errors, got %d", expected, len(output.Errors))
	}
}

// validateMessageContains validates that HasMessage works correctly.
func validateMessageContains(t *testing.T, output *MockColoredOutput, keyword string, expected bool) {
	if output.HasMessage(keyword) != expected {
		t.Errorf("expected HasMessage('%s') to return %v", keyword, expected)
	}
}

// validateErrorContains validates that HasError works correctly.
func validateErrorContains(t *testing.T, output *MockColoredOutput, keyword string, expected bool) {
	if output.HasError(keyword) != expected {
		t.Errorf("expected HasError('%s') to return %v", keyword, expected)
	}
}

func TestCreateTestAction(t *testing.T) {
	t.Run("creates basic action", func(t *testing.T) {
		name := "Test Action"
		description := "A test action for testing"
		inputs := map[string]string{
			"input1": "First input",
			"input2": "Second input",
		}

		action := CreateTestAction(name, description, inputs)

		if action == "" {
			t.Fatal("expected non-empty action content")
		}

		// Verify the action contains our values
		if !strings.Contains(action, name) {
			t.Errorf("action should contain name: %s", name)
		}

		if !strings.Contains(action, description) {
			t.Errorf("action should contain description: %s", description)
		}

		for inputName, inputDesc := range inputs {
			if !strings.Contains(action, inputName) {
				t.Errorf("action should contain input name: %s", inputName)
			}
			if !strings.Contains(action, inputDesc) {
				t.Errorf("action should contain input description: %s", inputDesc)
			}
		}
	})

	t.Run("creates action with no inputs", func(t *testing.T) {
		action := CreateTestAction("Simple Action", "No inputs", nil)

		if action == "" {
			t.Fatal("expected non-empty action content")
		}

		if !strings.Contains(action, "Simple Action") {
			t.Error("action should contain the name")
		}
	})
}

func TestCreateCompositeAction(t *testing.T) {
	t.Run("creates composite action with steps", func(t *testing.T) {
		name := "Composite Test"
		description := "A composite action"
		steps := []string{
			"actions/checkout@v4",
			"actions/setup-node@v4",
		}

		action := CreateCompositeAction(name, description, steps)

		if action == "" {
			t.Fatal("expected non-empty action content")
		}

		// Verify the action contains our values
		if !strings.Contains(action, name) {
			t.Errorf("action should contain name: %s", name)
		}

		if !strings.Contains(action, description) {
			t.Errorf("action should contain description: %s", description)
		}

		for _, step := range steps {
			if !strings.Contains(action, step) {
				t.Errorf("action should contain step: %s", step)
			}
		}
	})

	t.Run("creates composite action with no steps", func(t *testing.T) {
		action := CreateCompositeAction("Empty Composite", "No steps", nil)

		if action == "" {
			t.Fatal("expected non-empty action content")
		}

		if !strings.Contains(action, "Empty Composite") {
			t.Error("action should contain the name")
		}
	})
}

func TestMockAppConfig(t *testing.T) {
	t.Run("creates default config", func(t *testing.T) {
		testMockAppConfigDefaults(t)
	})

	t.Run("applies overrides", func(t *testing.T) {
		testMockAppConfigOverrides(t)
	})

	t.Run("partial overrides keep defaults", func(t *testing.T) {
		testMockAppConfigPartialOverrides(t)
	})
}

// testMockAppConfigDefaults tests default config creation.
func testMockAppConfigDefaults(t *testing.T) {
	config := MockAppConfig(nil)

	validateConfigCreated(t, config)
	validateConfigDefaults(t, config)
}

// testMockAppConfigOverrides tests full override application.
func testMockAppConfigOverrides(t *testing.T) {
	overrides := createFullOverrides()
	config := MockAppConfig(overrides)

	validateOverriddenValues(t, config)
}

// testMockAppConfigPartialOverrides tests partial override application.
func testMockAppConfigPartialOverrides(t *testing.T) {
	overrides := createPartialOverrides()
	config := MockAppConfig(overrides)

	validatePartialOverrides(t, config)
	validateRemainingDefaults(t, config)
}

// createFullOverrides creates a complete set of test overrides.
func createFullOverrides() *TestAppConfig {
	return &TestAppConfig{
		Theme:        "github",
		OutputFormat: "html",
		OutputDir:    "docs",
		Template:     "custom.tmpl",
		Schema:       "custom.schema.json",
		Verbose:      true,
		Quiet:        true,
		GitHubToken:  "test-token",
	}
}

// createPartialOverrides creates a partial set of test overrides.
func createPartialOverrides() *TestAppConfig {
	return &TestAppConfig{
		Theme:   "professional",
		Verbose: true,
	}
}

// validateConfigCreated validates that config was created successfully.
func validateConfigCreated(t *testing.T, config *TestAppConfig) {
	if config == nil {
		t.Fatal("expected config to be created")
	}
}

// validateConfigDefaults validates all default configuration values.
func validateConfigDefaults(t *testing.T, config *TestAppConfig) {
	validateStringField(t, config.Theme, "default", "theme")
	validateStringField(t, config.OutputFormat, "md", "output format")
	validateStringField(t, config.OutputDir, ".", "output dir")
	validateStringField(t, config.Schema, "schemas/action.schema.json", "schema")
	validateBoolField(t, config.Verbose, false, "verbose")
	validateBoolField(t, config.Quiet, false, "quiet")
	validateStringField(t, config.GitHubToken, "", "GitHub token")
}

// validateOverriddenValues validates all overridden configuration values.
func validateOverriddenValues(t *testing.T, config *TestAppConfig) {
	validateStringField(t, config.Theme, "github", "theme")
	validateStringField(t, config.OutputFormat, "html", "output format")
	validateStringField(t, config.OutputDir, "docs", "output dir")
	validateStringField(t, config.Template, "custom.tmpl", "template")
	validateStringField(t, config.Schema, "custom.schema.json", "schema")
	validateBoolField(t, config.Verbose, true, "verbose")
	validateBoolField(t, config.Quiet, true, "quiet")
	validateStringField(t, config.GitHubToken, "test-token", "GitHub token")
}

// validatePartialOverrides validates partially overridden values.
func validatePartialOverrides(t *testing.T, config *TestAppConfig) {
	validateStringField(t, config.Theme, "professional", "theme")
	validateBoolField(t, config.Verbose, true, "verbose")
}

// validateRemainingDefaults validates that non-overridden values remain default.
func validateRemainingDefaults(t *testing.T, config *TestAppConfig) {
	validateStringField(t, config.OutputFormat, "md", "output format")
	validateBoolField(t, config.Quiet, false, "quiet")
}

// validateStringField validates a string configuration field.
func validateStringField(t *testing.T, actual, expected, fieldName string) {
	if actual != expected {
		t.Errorf("expected %s %s, got %s", fieldName, expected, actual)
	}
}

// validateBoolField validates a boolean configuration field.
func validateBoolField(t *testing.T, actual, expected bool, fieldName string) {
	if actual != expected {
		t.Errorf("expected %s to be %v, got %v", fieldName, expected, actual)
	}
}

func TestSetEnv(t *testing.T) {
	testKey := "TEST_TESTUTIL_VAR"
	originalValue := "original"
	newValue := "new"

	// Ensure the test key is not set initially
	_ = os.Unsetenv(testKey)

	t.Run("sets new environment variable", func(t *testing.T) {
		cleanup := SetEnv(t, testKey, newValue)
		defer cleanup()

		if os.Getenv(testKey) != newValue {
			t.Errorf("expected env var to be %s, got %s", newValue, os.Getenv(testKey))
		}
	})

	t.Run("cleanup unsets new variable", func(t *testing.T) {
		cleanup := SetEnv(t, testKey, newValue)
		cleanup()

		if os.Getenv(testKey) != "" {
			t.Errorf("expected env var to be unset, got %s", os.Getenv(testKey))
		}
	})

	t.Run("overrides existing variable", func(t *testing.T) {
		// Set original value
		_ = os.Setenv(testKey, originalValue)

		cleanup := SetEnv(t, testKey, newValue)
		defer cleanup()

		if os.Getenv(testKey) != newValue {
			t.Errorf("expected env var to be %s, got %s", newValue, os.Getenv(testKey))
		}
	})

	t.Run("cleanup restores original variable", func(t *testing.T) {
		// Set original value
		_ = os.Setenv(testKey, originalValue)

		cleanup := SetEnv(t, testKey, newValue)
		cleanup()

		if os.Getenv(testKey) != originalValue {
			t.Errorf("expected env var to be restored to %s, got %s", originalValue, os.Getenv(testKey))
		}
	})

	// Clean up after test
	_ = os.Unsetenv(testKey)
}

func TestWithContext(t *testing.T) {
	t.Run("creates context with timeout", func(t *testing.T) {
		timeout := 100 * time.Millisecond
		ctx := WithContext(timeout)

		if ctx == nil {
			t.Fatal("expected context to be created")
		}

		// Check that the context has a deadline
		deadline, ok := ctx.Deadline()
		if !ok {
			t.Error("expected context to have a deadline")
		}

		// The deadline should be approximately now + timeout
		expectedDeadline := time.Now().Add(timeout)
		timeDiff := deadline.Sub(expectedDeadline)
		if timeDiff < -time.Second || timeDiff > time.Second {
			t.Errorf("deadline too far from expected: diff = %v", timeDiff)
		}
	})

	t.Run("context eventually times out", func(t *testing.T) {
		ctx := WithContext(1 * time.Millisecond)

		// Wait a bit longer than the timeout
		time.Sleep(10 * time.Millisecond)

		select {
		case <-ctx.Done():
			// Context should be done
			if ctx.Err() != context.DeadlineExceeded {
				t.Errorf("expected DeadlineExceeded error, got %v", ctx.Err())
			}
		default:
			t.Error("expected context to be done after timeout")
		}
	})
}

func TestAssertNoError(t *testing.T) {
	t.Run("passes with nil error", func(t *testing.T) {
		// This should not fail
		AssertNoError(t, nil)
	})

	// Testing the failure case is complex because AssertNoError calls t.Fatalf
	// which causes the test to exit. We can't easily test this without
	// complex mocking infrastructure, so we'll just test the success case
	// The failure case is implicitly tested throughout the codebase where
	// AssertNoError is used with actual errors.
}

func TestAssertError(t *testing.T) {
	t.Run("passes with non-nil error", func(t *testing.T) {
		// This should not fail
		AssertError(t, io.EOF)
	})

	// Similar to AssertNoError, testing the failure case is complex
	// The failure behavior is implicitly tested throughout the codebase
}

func TestAssertStringContains(t *testing.T) {
	t.Run("passes when string contains substring", func(t *testing.T) {
		AssertStringContains(t, "hello world", "world")
		AssertStringContains(t, "test string", "test")
		AssertStringContains(t, "exact match", "exact match")
	})

	// Failure case testing is complex due to t.Fatalf behavior
}

func TestAssertEqual(t *testing.T) {
	t.Run("passes with equal basic types", func(t *testing.T) {
		AssertEqual(t, 42, 42)
		AssertEqual(t, "test", "test")
		AssertEqual(t, true, true)
		AssertEqual(t, 3.14, 3.14)
	})

	t.Run("passes with equal string maps", func(t *testing.T) {
		map1 := map[string]string{"key1": "value1", "key2": "value2"}
		map2 := map[string]string{"key1": "value1", "key2": "value2"}
		AssertEqual(t, map1, map2)
	})

	t.Run("passes with empty string maps", func(t *testing.T) {
		map1 := map[string]string{}
		map2 := map[string]string{}
		AssertEqual(t, map1, map2)
	})

	// Testing failure cases is complex due to t.Fatalf behavior
	// The map comparison logic is tested implicitly throughout the codebase
}

func TestNewStringReader(t *testing.T) {
	t.Run("creates reader from string", func(t *testing.T) {
		testString := "Hello, World!"
		reader := NewStringReader(testString)

		if reader == nil {
			t.Fatal("expected reader to be created")
		}

		// Read the content
		content, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("failed to read from reader: %v", err)
		}

		if string(content) != testString {
			t.Errorf("expected content %s, got %s", testString, string(content))
		}
	})

	t.Run("creates reader from empty string", func(t *testing.T) {
		reader := NewStringReader("")
		content, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("failed to read from empty reader: %v", err)
		}

		if len(content) != 0 {
			t.Errorf("expected empty content, got %d bytes", len(content))
		}
	})

	t.Run("reader can be closed", func(t *testing.T) {
		reader := NewStringReader("test")
		err := reader.Close()
		if err != nil {
			t.Errorf("failed to close reader: %v", err)
		}
	})

	t.Run("handles large strings", func(t *testing.T) {
		largeString := strings.Repeat("test ", 10000)
		reader := NewStringReader(largeString)

		content, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("failed to read large string: %v", err)
		}

		if string(content) != largeString {
			t.Error("large string content mismatch")
		}
	})
}
