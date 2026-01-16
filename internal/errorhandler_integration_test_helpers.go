package internal

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// spawnTestSubprocess creates and configures a test subprocess.
// This helper reduces cognitive complexity in integration tests by centralizing
// the subprocess creation logic.
//
//nolint:unused // Prepared for future use in errorhandler integration tests
func spawnTestSubprocess(t *testing.T, testType string) *exec.Cmd {
	t.Helper()

	//nolint:gosec // G204: Controlled test arguments, not user input
	cmd := exec.Command(os.Args[0], "-test.run=TestErrorHandlerIntegration")
	cmd.Env = append(os.Environ(), "GO_TEST_SUBPROCESS=1", "TEST_TYPE="+testType)

	return cmd
}

// assertSubprocessExit validates subprocess exit code and stderr.
// This helper reduces cognitive complexity in integration tests by centralizing
// the subprocess validation logic that was repeated across test cases.
//
//nolint:unused // Prepared for future use in errorhandler integration tests
func assertSubprocessExit(t *testing.T, cmd *exec.Cmd, expectedExitCode int, stderrPattern string) {
	t.Helper()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("failed to create stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start subprocess: %v", err)
	}

	stderrBytes, _ := io.ReadAll(stderr)
	stderrStr := string(stderrBytes)

	err = cmd.Wait()

	// Validate exit code
	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	}

	if exitCode != expectedExitCode {
		t.Errorf("exit code = %d, want %d", exitCode, expectedExitCode)
	}

	// Validate stderr contains pattern
	if stderrPattern != "" && !strings.Contains(stderrStr, stderrPattern) {
		t.Errorf("stderr does not contain %q, got: %s", stderrPattern, stderrStr)
	}
}
