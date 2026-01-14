package internal_test

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/internal/apperrors"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// verifyExitCode checks that the command exited with the expected exit code.
func verifyExitCode(t *testing.T, err error, expectedExit int) {
	t.Helper()
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != expectedExit {
			t.Errorf("expected exit code %d, got %d", expectedExit, exitErr.ExitCode())
		}

		return
	}
	if err != nil {
		t.Fatalf(testutil.TestErrUnexpected, err)
	}
	if expectedExit != 0 {
		t.Errorf("expected exit code %d, but process exited successfully", expectedExit)
	}
}

// execSubprocessTest spawns a subprocess and returns its stderr output and error.
func execSubprocessTest(t *testing.T, testType string) (string, error) {
	t.Helper()
	//nolint:gosec // Controlled test arguments
	cmd := exec.Command(os.Args[0], "-test.run=^TestErrorHandlerIntegration$")
	cmd.Env = append(os.Environ(),
		"GO_TEST_SUBPROCESS=1",
		"TEST_TYPE="+testType,
	)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("failed to get stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start subprocess: %v", err)
	}

	stderrOutput := make([]byte, 4096)
	n, _ := stderr.Read(stderrOutput)
	stderrStr := string(stderrOutput[:n])

	return stderrStr, cmd.Wait()
}

// runSubprocessErrorTest executes a subprocess test and verifies exit code and stderr.
// Consolidates 15 duplicated test loops.
func runSubprocessErrorTest(t *testing.T, testType string, expectedExit int, expectedStderr string) {
	t.Helper()

	stderrStr, err := execSubprocessTest(t, testType)
	verifyExitCode(t, err, expectedExit)

	if !strings.Contains(strings.ToLower(stderrStr), strings.ToLower(expectedStderr)) {
		t.Errorf("stderr missing expected text %q, got: %s", expectedStderr, stderrStr)
	}
}

// TestErrorHandlerIntegration tests error handler methods that call os.Exit()
// using subprocess pattern.
func TestErrorHandlerIntegration(t *testing.T) {
	t.Parallel()

	// Check if this is the subprocess
	if os.Getenv("GO_TEST_SUBPROCESS") == "1" {
		runSubprocessTest()

		return
	}

	tests := []struct {
		name           string
		testType       string
		expectedExit   int
		expectedStderr string
	}{
		{
			name:           "HandleError with file not found",
			testType:       "handle_error_file_not_found",
			expectedExit:   appconstants.ExitCodeError,
			expectedStderr: "file not found",
		},
		{
			name:           "HandleError with validation error",
			testType:       "handle_error_validation",
			expectedExit:   appconstants.ExitCodeError,
			expectedStderr: "validation failed",
		},
		{
			name:           "HandleError with context",
			testType:       "handle_error_with_context",
			expectedExit:   appconstants.ExitCodeError,
			expectedStderr: "config file",
		},
		{
			name:           "HandleError with suggestions",
			testType:       "handle_error_with_suggestions",
			expectedExit:   appconstants.ExitCodeError,
			expectedStderr: "file error",
		},
		{
			name:           "HandleFatalError with permission denied",
			testType:       "handle_fatal_error_permission",
			expectedExit:   appconstants.ExitCodeError,
			expectedStderr: "permission denied",
		},
		{
			name:           "HandleFatalError with config error",
			testType:       "handle_fatal_error_config",
			expectedExit:   appconstants.ExitCodeError,
			expectedStderr: "configuration error",
		},
		{
			name:           "HandleSimpleError with generic error",
			testType:       "handle_simple_error_generic",
			expectedExit:   appconstants.ExitCodeError,
			expectedStderr: "operation failed",
		},
		{
			name:           "HandleSimpleError with file not found pattern",
			testType:       "handle_simple_error_not_found",
			expectedExit:   appconstants.ExitCodeError,
			expectedStderr: "file error",
		},
		{
			name:           "HandleSimpleError with permission pattern",
			testType:       "handle_simple_error_permission",
			expectedExit:   appconstants.ExitCodeError,
			expectedStderr: "access error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			runSubprocessErrorTest(t, tt.testType, tt.expectedExit, tt.expectedStderr)
		})
	}
}

// runSubprocessTest executes the actual error handler call based on TEST_TYPE.
func runSubprocessTest() {
	testType := os.Getenv("TEST_TYPE")
	output := internal.NewColoredOutput(false) // quiet=false
	handler := internal.NewErrorHandler(output)

	switch testType {
	case "handle_error_file_not_found":
		err := apperrors.New(appconstants.ErrCodeFileNotFound, "file not found")
		handler.HandleError(err)

	case "handle_error_validation":
		err := apperrors.New(appconstants.ErrCodeValidation, "validation failed")
		handler.HandleError(err)

	case "handle_error_with_context":
		err := apperrors.New(appconstants.ErrCodeConfiguration, "config file missing")
		err = err.WithDetails(map[string]string{
			"path": "/invalid/path/config.yaml",
			"type": "application",
		})
		handler.HandleError(err)

	case "handle_error_with_suggestions":
		err := apperrors.New(appconstants.ErrCodeFileNotFound, "file error occurred")
		err = err.WithSuggestions("Check that the file exists", "Verify file permissions")
		handler.HandleError(err)

	case "handle_fatal_error_permission":
		handler.HandleFatalError(
			appconstants.ErrCodePermission,
			"permission denied accessing file",
			map[string]string{"file": "/etc/passwd"},
		)

	case "handle_fatal_error_config":
		handler.HandleFatalError(
			appconstants.ErrCodeConfiguration,
			"configuration error in settings",
			map[string]string{
				"section": "github",
				"key":     "token",
			},
		)

	case "handle_simple_error_generic":
		handler.HandleSimpleError("operation failed", errors.New("generic error occurred"))

	case "handle_simple_error_not_found":
		handler.HandleSimpleError("file error", errors.New("no such file or directory"))

	case "handle_simple_error_permission":
		handler.HandleSimpleError("access error", errors.New("permission denied"))

	default:
		os.Exit(99) // Unexpected test type
	}
}

// TestErrorHandlerAllErrorCodes tests that all error codes produce correct exit codes.
func TestErrorHandlerAllErrorCodes(t *testing.T) {
	t.Parallel()

	// Check if this is the subprocess
	if os.Getenv("GO_TEST_SUBPROCESS") == "1" {
		runErrorCodeTest()

		return
	}

	errorCodes := []struct {
		code        appconstants.ErrorCode
		description string
	}{
		{appconstants.ErrCodeFileNotFound, "file not found"},
		{appconstants.ErrCodePermission, "permission denied"},
		{appconstants.ErrCodeInvalidYAML, "invalid yaml"},
		{appconstants.ErrCodeInvalidAction, "invalid action"},
		{appconstants.ErrCodeNoActionFiles, "no action files"},
		{appconstants.ErrCodeGitHubAPI, "github api error"},
		{appconstants.ErrCodeGitHubRateLimit, "rate limit"},
		{appconstants.ErrCodeGitHubAuth, "auth error"},
		{appconstants.ErrCodeConfiguration, "configuration error"},
		{appconstants.ErrCodeValidation, "validation error"},
		{appconstants.ErrCodeTemplateRender, "template error"},
		{appconstants.ErrCodeFileWrite, "file write error"},
		{appconstants.ErrCodeDependencyAnalysis, "dependency error"},
		{appconstants.ErrCodeCacheAccess, "cache error"},
		{appconstants.ErrCodeUnknown, "unknown error"},
	}

	for _, tc := range errorCodes {
		t.Run(string(tc.code), func(t *testing.T) {
			t.Parallel()

			//nolint:gosec // Controlled test arguments
			cmd := exec.Command(os.Args[0], "-test.run=^TestErrorHandlerAllErrorCodes$/^"+string(tc.code)+"$")
			cmd.Env = append(os.Environ(),
				"GO_TEST_SUBPROCESS=1",
				"ERROR_CODE="+string(tc.code),
				"ERROR_DESC="+tc.description,
			)

			stderr, _ := cmd.StderrPipe()
			_ = cmd.Start()

			stderrOutput := make([]byte, 4096)
			n, _ := stderr.Read(stderrOutput)
			stderrStr := string(stderrOutput[:n])

			err := cmd.Wait()

			// All errors should exit with ExitCodeError (1)
			if exitErr, ok := err.(*exec.ExitError); ok {
				if exitErr.ExitCode() != appconstants.ExitCodeError {
					t.Errorf("expected exit code %d, got %d", appconstants.ExitCodeError, exitErr.ExitCode())
				}
			} else if err != nil {
				t.Fatalf(testutil.TestErrUnexpected, err)
			} else {
				t.Error("expected non-zero exit code")
			}

			// Verify error message appears in output
			if !strings.Contains(strings.ToLower(stderrStr), strings.ToLower(tc.description)) {
				t.Errorf("stderr missing expected error description %q, got: %s", tc.description, stderrStr)
			}
		})
	}
}

// runErrorCodeTest handles subprocess execution for error code tests.
func runErrorCodeTest() {
	code := appconstants.ErrorCode(os.Getenv("ERROR_CODE"))
	desc := os.Getenv("ERROR_DESC")

	output := internal.NewColoredOutput(false)
	handler := internal.NewErrorHandler(output)

	err := apperrors.New(code, desc)
	handler.HandleError(err)
}

// TestErrorHandlerWithComplexContext tests error handler with multiple context values and suggestions.
func TestErrorHandlerWithComplexContext(t *testing.T) {
	t.Parallel()

	if os.Getenv("GO_TEST_SUBPROCESS") == "1" {
		runComplexContextTest()

		return
	}

	//nolint:gosec // Controlled test arguments
	cmd := exec.Command(os.Args[0], "-test.run=^TestErrorHandlerWithComplexContext$")
	cmd.Env = append(os.Environ(), "GO_TEST_SUBPROCESS=1")

	stderr, _ := cmd.StderrPipe()
	_ = cmd.Start()

	stderrOutput := make([]byte, 8192)
	n, _ := stderr.Read(stderrOutput)
	stderrStr := string(stderrOutput[:n])

	_ = cmd.Wait()

	// Verify all context keys are displayed
	contextKeys := []string{"path", "action", "reason"}
	for _, key := range contextKeys {
		if !strings.Contains(stderrStr, key) {
			t.Errorf("stderr missing context key %q", key)
		}
	}

	// Verify suggestions are displayed
	suggestions := []string{"Check the file path", "Verify YAML syntax", "Consult documentation"}
	for _, suggestion := range suggestions {
		if !strings.Contains(stderrStr, suggestion) {
			t.Errorf("stderr missing suggestion %q", suggestion)
		}
	}
}

// runComplexContextTest handles subprocess execution for complex context test.
func runComplexContextTest() {
	output := internal.NewColoredOutput(false)
	handler := internal.NewErrorHandler(output)

	err := apperrors.New(appconstants.ErrCodeInvalidYAML, "YAML parsing failed")
	err = err.WithDetails(map[string]string{
		"path":   "/path/to/action.yml",
		"action": "parse-workflow",
		"reason": "invalid syntax at line 42",
	})
	err = err.WithSuggestions(
		"Check the file path is correct",
		"Verify YAML syntax is valid",
		"Consult documentation for proper format",
	)

	handler.HandleError(err)
}
