package testutil

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMockHTTPClient tests the MockHTTPClient implementation.
func TestMockHTTPClient(t *testing.T) {
	t.Parallel()
	t.Run("returns configured response", func(t *testing.T) {
		t.Parallel()
		testMockHTTPClientConfiguredResponse(t)
	})

	t.Run("returns 404 for unconfigured endpoints", func(t *testing.T) {
		t.Parallel()
		testMockHTTPClientUnconfiguredEndpoints(t)
	})

	t.Run("tracks requests", func(t *testing.T) {
		t.Parallel()
		testMockHTTPClientRequestTracking(t)
	})
}

// testMockHTTPClientConfiguredResponse tests that configured responses are returned correctly.
func testMockHTTPClientConfiguredResponse(t *testing.T) {
	t.Helper()
	client := createMockHTTPClientWithResponse("GET https://api.github.com/test", 200, `{"test": "response"}`)

	req := createTestRequest(t, "GET", ""+TestURLGitHubAPI+"test")
	resp := executeRequest(t, client, req)
	defer func() { _ = resp.Body.Close() }()

	validateResponseStatus(t, resp, 200)
	validateResponseBody(t, resp, `{"test": "response"}`)
}

// testMockHTTPClientUnconfiguredEndpoints tests that unconfigured endpoints return 404.
func testMockHTTPClientUnconfiguredEndpoints(t *testing.T) {
	t.Helper()
	client := &MockHTTPClient{
		Responses: make(map[string]*http.Response),
	}

	req := createTestRequest(t, "GET", ""+TestURLGitHubAPI+"nonexistent")
	resp := executeRequest(t, client, req)
	defer func() { _ = resp.Body.Close() }()

	validateResponseStatus(t, resp, 404)
}

// testMockHTTPClientRequestTracking tests that requests are tracked correctly.
func testMockHTTPClientRequestTracking(t *testing.T) {
	t.Helper()
	client := &MockHTTPClient{
		Responses: make(map[string]*http.Response),
	}

	req1 := createTestRequest(t, "GET", ""+TestURLGitHubAPI+"test1")
	req2 := createTestRequest(t, "POST", ""+TestURLGitHubAPI+"test2")

	executeAndCloseResponse(client, req1)
	executeAndCloseResponse(client, req2)

	validateRequestTracking(t, client, 2, ""+TestURLGitHubAPI+"test1", "POST")
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
	t.Helper()
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	return req
}

// executeRequest executes an HTTP request and returns the response.
func executeRequest(t *testing.T, client *MockHTTPClient, req *http.Request) *http.Response {
	t.Helper()
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf(TestErrUnexpected, err)
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
	t.Helper()
	if resp.StatusCode != expectedStatus {
		t.Errorf("expected status %d, got %d", expectedStatus, resp.StatusCode)
	}
}

// validateResponseBody validates that the response body matches the expected content.
func validateResponseBody(t *testing.T, resp *http.Response, expected string) {
	t.Helper()
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
	t.Helper()
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
	t.Parallel()
	t.Run("creates client with mocked responses", func(t *testing.T) {
		t.Parallel()
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
			t.Fatalf(TestErrUnexpected, err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf(TestErrStatusCode, resp.StatusCode)
		}
	})

	t.Run("uses MockGitHubResponses", func(t *testing.T) {
		t.Parallel()
		responses := MockGitHubResponses()
		client := MockGitHubClient(responses)

		// Test a specific endpoint that we know is mocked
		ctx := context.Background()
		_, resp, err := client.Repositories.Get(ctx, "actions", "checkout")
		if err != nil {
			t.Fatalf(TestErrUnexpected, err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf(TestErrStatusCode, resp.StatusCode)
		}
	})
}

func TestMockTransport(t *testing.T) {
	t.Parallel()
	client := &MockHTTPClient{
		Responses: map[string]*http.Response{
			"GET https://api.github.com/test": {
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"success": true}`)),
			},
		},
	}

	transport := &MockTransport{Client: client}

	req, err := http.NewRequest(http.MethodGet, ""+TestURLGitHubAPI+"test", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf(TestErrUnexpected, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf(TestErrStatusCode, resp.StatusCode)
	}
}

func TestTempDir(t *testing.T) {
	t.Parallel()
	t.Run("creates temporary directory", func(t *testing.T) {
		t.Parallel()
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

		// Verify directory name pattern (t.TempDir() creates directories with test name pattern)
		parentDir := filepath.Base(filepath.Dir(dir))
		if !strings.Contains(parentDir, "TestTempDir") {
			t.Errorf("parent directory name should contain TestTempDir: %s", parentDir)
		}
	})

	t.Run("cleanup removes directory", func(t *testing.T) {
		t.Parallel()
		dir, cleanup := TempDir(t)

		// Verify directory exists
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Error("temporary directory was not created")
		}

		// Clean up - this is now a no-op since t.TempDir() handles cleanup automatically
		cleanup()

		// Note: We can't verify directory removal here because t.TempDir() only
		// cleans up at the end of the test, not when cleanup() is called.
		// The directory will be automatically cleaned up when the test ends.
	})
}

func TestWriteTestFile(t *testing.T) {
	t.Parallel()
	tmpDir, cleanup := TempDir(t)
	defer cleanup()

	t.Run("writes file with content", func(t *testing.T) {
		t.Parallel()
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
		t.Parallel()
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
		t.Parallel()
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
	t.Parallel()
	tmpDir, cleanup := TempDir(t)
	defer cleanup()

	SetupTestTemplates(t, tmpDir)

	// Verify template directories exist
	templatesDir := filepath.Join(tmpDir, "templates")
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		t.Error("templates directory was not created")
	}

	// Verify theme directories exist
	themes := []string{TestThemeGitHub, TestThemeGitLab, TestThemeMinimal, TestThemeProfessional}
	for _, theme := range themes {
		themeDir := filepath.Join(templatesDir, "themes", theme)
		if _, err := os.Stat(themeDir); os.IsNotExist(err) {
			t.Errorf("theme directory %s was not created", theme)
		}

		// Verify theme template file exists
		templateFile := filepath.Join(themeDir, TestTemplateReadme)
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
	defaultTemplate := filepath.Join(templatesDir, TestTemplateReadme)
	if _, err := os.Stat(defaultTemplate); os.IsNotExist(err) {
		t.Error("default template was not created")
	}
}

func TestCreateCompositeAction(t *testing.T) {
	t.Parallel()
	t.Run("creates composite action with steps", func(t *testing.T) {
		t.Parallel()
		name := "Composite Test"
		description := "A composite action"
		steps := []string{
			TestActionCheckoutV4,
			"actions/setup-node@v4",
		}

		action := CreateCompositeAction(name, description, steps)

		if action == "" {
			t.Fatal(TestErrNonEmptyAction)
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
		t.Parallel()
		action := CreateCompositeAction("Empty Composite", "No steps", nil)

		if action == "" {
			t.Fatal(TestErrNonEmptyAction)
		}

		if !strings.Contains(action, "Empty Composite") {
			t.Error("action should contain the name")
		}
	})
}

func TestAssertNoError(t *testing.T) {
	t.Parallel()
	t.Run("passes with nil error", func(t *testing.T) {
		t.Parallel()
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
	t.Parallel()
	t.Run("passes with non-nil error", func(t *testing.T) {
		t.Parallel()
		// This should not fail
		AssertError(t, io.EOF)
	})

	// Similar to AssertNoError, testing the failure case is complex
	// The failure behavior is implicitly tested throughout the codebase
}

func TestAssertStringContains(t *testing.T) {
	t.Parallel()
	t.Run("passes when string contains substring", func(t *testing.T) {
		t.Parallel()
		AssertStringContains(t, "hello world", "world")
		AssertStringContains(t, "test string", "test")
		AssertStringContains(t, "exact match", "exact match")
	})

	// Failure case testing is complex due to t.Fatalf behavior
}

func TestAssertEqual(t *testing.T) {
	t.Parallel()
	t.Run("passes with equal basic types", func(t *testing.T) {
		t.Parallel()
		AssertEqual(t, 42, 42)
		AssertEqual(t, "test", "test")
		AssertEqual(t, true, true)
		AssertEqual(t, 3.14, 3.14)
	})

	t.Run("passes with equal string maps", func(t *testing.T) {
		t.Parallel()
		map1 := map[string]string{CacheTestKey1: CacheTestValue1, CacheTestKey2: CacheTestValue2}
		map2 := map[string]string{CacheTestKey1: CacheTestValue1, CacheTestKey2: CacheTestValue2}
		AssertEqual(t, map1, map2)
	})

	t.Run("passes with empty string maps", func(t *testing.T) {
		t.Parallel()
		map1 := map[string]string{}
		map2 := map[string]string{}
		AssertEqual(t, map1, map2)
	})

	// The failure path is asserted via equalCheck (the pure function AssertEqual
	// delegates to), since a *testing.T cannot be mocked to observe t.Fatal.
}

// TestEqualCheck asserts both the pass and FAIL paths of the comparison that
// backs AssertEqual. Without this, a regression that made AssertEqual a no-op
// would silently pass every assertion in the codebase.
func TestEqualCheck(t *testing.T) {
	t.Parallel()

	t.Run("equal values report ok", func(t *testing.T) {
		t.Parallel()
		for _, c := range []struct{ a, b any }{
			{42, 42}, {"x", "x"}, {true, true},
		} {
			if ok, msg := equalCheck(c.a, c.b); !ok {
				t.Errorf("equalCheck(%v, %v) = false (%q), want true", c.a, c.b, msg)
			}
		}
		if ok, _ := equalCheck(
			map[string]string{"k": "v"}, map[string]string{"k": "v"},
		); !ok {
			t.Error("equalCheck of equal maps = false, want true")
		}
	})

	t.Run("unequal values report not-ok with message", func(t *testing.T) {
		t.Parallel()
		for _, c := range []struct{ a, b any }{
			{1, 2}, {"x", "y"}, {true, false},
			{map[string]string{"k": "v"}, map[string]string{"k": "w"}}, // value differs
			{map[string]string{"k": "v"}, map[string]string{}},         // length differs
			{map[string]string{"k": "v"}, "not a map"},                 // type differs
		} {
			ok, msg := equalCheck(c.a, c.b)
			if ok {
				t.Errorf("equalCheck(%v, %v) = true, want false", c.a, c.b)
			}
			if msg == "" {
				t.Errorf("equalCheck(%v, %v) returned empty failure message", c.a, c.b)
			}
		}
	})
}

func TestNewStringReader(t *testing.T) {
	t.Parallel()
	t.Run("creates reader from string", func(t *testing.T) {
		t.Parallel()
		testNewStringReaderBasic(t)
	})

	t.Run("creates reader from empty string", func(t *testing.T) {
		t.Parallel()
		testNewStringReaderEmpty(t)
	})

	t.Run("reader can be closed", func(t *testing.T) {
		t.Parallel()
		testNewStringReaderClose(t)
	})

	t.Run("handles large strings", func(t *testing.T) {
		t.Parallel()
		testNewStringReaderLarge(t)
	})
}

// testNewStringReaderBasic tests basic string reader creation and reading.
func testNewStringReaderBasic(t *testing.T) {
	t.Helper()
	testString := "Hello, World!"
	reader := NewStringReader(testString)

	if reader == nil {
		t.Fatal("expected reader to be created")
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read from reader: %v", err)
	}

	if string(content) != testString {
		t.Errorf("expected content %s, got %s", testString, string(content))
	}
}

// testNewStringReaderEmpty tests string reader with empty string.
func testNewStringReaderEmpty(t *testing.T) {
	t.Helper()
	reader := NewStringReader("")
	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read from empty reader: %v", err)
	}

	if len(content) != 0 {
		t.Errorf("expected empty content, got %d bytes", len(content))
	}
}

// testNewStringReaderClose tests that the reader can be closed.
func testNewStringReaderClose(t *testing.T) {
	t.Helper()
	reader := NewStringReader("test")
	err := reader.Close()
	if err != nil {
		t.Errorf("failed to close reader: %v", err)
	}
}

// testNewStringReaderLarge tests reading large strings.
func testNewStringReaderLarge(t *testing.T) {
	t.Helper()
	largeString := strings.Repeat("test ", 10000)
	reader := NewStringReader(largeString)

	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read large string: %v", err)
	}

	if string(content) != largeString {
		t.Error("large string content mismatch")
	}
}

func TestCaptureStdout(t *testing.T) {
	// Note: Cannot run in parallel as it manipulates global os.Stdout

	output := CaptureStdout(func() {
		fmt.Print("test output")
	})

	if output != "test output" {
		t.Errorf("expected 'test output', got %q", output)
	}
}

func TestCaptureStderr(t *testing.T) {
	// Note: Cannot run in parallel as it manipulates global os.Stderr

	output := CaptureStderr(func() {
		fmt.Fprint(os.Stderr, "test error")
	})

	if output != "test error" {
		t.Errorf("expected 'test error', got %q", output)
	}
}
