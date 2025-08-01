package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

// TestCLICommands tests the main CLI commands using subprocess execution.
func TestCLICommands(t *testing.T) {
	// Build the binary for testing
	binaryPath := buildTestBinary(t)
	defer func() { _ = os.Remove(binaryPath) }()

	tests := []struct {
		name       string
		args       []string
		setupFunc  func(t *testing.T, tmpDir string)
		wantExit   int
		wantStdout string
		wantStderr string
	}{
		{
			name:       "version command",
			args:       []string{"version"},
			wantExit:   0,
			wantStdout: "0.1.0",
		},
		{
			name:       "about command",
			args:       []string{"about"},
			wantExit:   0,
			wantStdout: "gh-action-readme: Generates README.md and HTML for GitHub Actions",
		},
		{
			name:       "help command",
			args:       []string{"--help"},
			wantExit:   0,
			wantStdout: "Auto-generate beautiful README and HTML documentation for GitHub Actions",
		},
		{
			name: "gen command with valid action",
			args: []string{"gen", "--output-format", "md"},
			setupFunc: func(t *testing.T, tmpDir string) {
				actionPath := filepath.Join(tmpDir, "action.yml")
				testutil.WriteTestFile(t, actionPath, testutil.SimpleActionYML)
			},
			wantExit: 0,
		},
		{
			name: "gen command with theme flag",
			args: []string{"gen", "--theme", "github", "--output-format", "json"},
			setupFunc: func(t *testing.T, tmpDir string) {
				actionPath := filepath.Join(tmpDir, "action.yml")
				testutil.WriteTestFile(t, actionPath, testutil.SimpleActionYML)
			},
			wantExit: 0,
		},
		{
			name:       "gen command with no action files",
			args:       []string{"gen"},
			wantExit:   1,
			wantStderr: "No action.yml or action.yaml files found",
		},
		{
			name: "validate command with valid action",
			args: []string{"validate"},
			setupFunc: func(t *testing.T, tmpDir string) {
				actionPath := filepath.Join(tmpDir, "action.yml")
				testutil.WriteTestFile(t, actionPath, testutil.SimpleActionYML)
			},
			wantExit:   0,
			wantStdout: "All validations passed successfully",
		},
		{
			name: "validate command with invalid action",
			args: []string{"validate"},
			setupFunc: func(t *testing.T, tmpDir string) {
				actionPath := filepath.Join(tmpDir, "action.yml")
				testutil.WriteTestFile(t, actionPath, testutil.InvalidActionYML)
			},
			wantExit: 1,
		},
		{
			name:       "schema command",
			args:       []string{"schema"},
			wantExit:   0,
			wantStdout: "schemas/action.schema.json",
		},
		{
			name:       "config command default",
			args:       []string{"config"},
			wantExit:   0,
			wantStdout: "Configuration file location:",
		},
		{
			name:       "config show command",
			args:       []string{"config", "show"},
			wantExit:   0,
			wantStdout: "Current Configuration:",
		},
		{
			name:       "config themes command",
			args:       []string{"config", "themes"},
			wantExit:   0,
			wantStdout: "Available Themes:",
		},
		{
			name:       "deps list command no files",
			args:       []string{"deps", "list"},
			wantExit:   0,
			wantStdout: "No action files found",
		},
		{
			name: "deps list command with composite action",
			args: []string{"deps", "list"},
			setupFunc: func(t *testing.T, tmpDir string) {
				actionPath := filepath.Join(tmpDir, "action.yml")
				testutil.WriteTestFile(t, actionPath, testutil.CompositeActionYML)
			},
			wantExit: 0,
		},
		{
			name:       "cache path command",
			args:       []string{"cache", "path"},
			wantExit:   0,
			wantStdout: "Cache Directory:",
		},
		{
			name:       "cache stats command",
			args:       []string{"cache", "stats"},
			wantExit:   0,
			wantStdout: "Cache Statistics:",
		},
		{
			name:       "invalid command",
			args:       []string{"invalid-command"},
			wantExit:   1,
			wantStderr: "unknown command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			// Setup test environment if needed
			if tt.setupFunc != nil {
				tt.setupFunc(t, tmpDir)
			}

			// Run the command in the temporary directory
			cmd := exec.Command(binaryPath, tt.args...)
			cmd.Dir = tmpDir

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			// Check exit code
			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				} else {
					t.Fatalf("unexpected error running command: %v", err)
				}
			}

			if exitCode != tt.wantExit {
				t.Errorf("expected exit code %d, got %d", tt.wantExit, exitCode)
				t.Logf("stdout: %s", stdout.String())
				t.Logf("stderr: %s", stderr.String())
			}

			// Check stdout if specified
			if tt.wantStdout != "" {
				if !strings.Contains(stdout.String(), tt.wantStdout) {
					t.Errorf("expected stdout to contain %q, got: %s", tt.wantStdout, stdout.String())
				}
			}

			// Check stderr if specified
			if tt.wantStderr != "" {
				if !strings.Contains(stderr.String(), tt.wantStderr) {
					t.Errorf("expected stderr to contain %q, got: %s", tt.wantStderr, stderr.String())
				}
			}
		})
	}
}

// TestCLIFlags tests various flag combinations.
func TestCLIFlags(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer func() { _ = os.Remove(binaryPath) }()

	tests := []struct {
		name     string
		args     []string
		wantExit int
		contains string
	}{
		{
			name:     "verbose flag",
			args:     []string{"--verbose", "config", "show"},
			wantExit: 0,
			contains: "Current Configuration:",
		},
		{
			name:     "quiet flag",
			args:     []string{"--quiet", "config", "show"},
			wantExit: 0,
		},
		{
			name:     "config file flag",
			args:     []string{"--config", "nonexistent.yml", "config", "show"},
			wantExit: 1,
		},
		{
			name:     "help flag",
			args:     []string{"-h"},
			wantExit: 0,
			contains: "Usage:",
		},
		{
			name:     "version short flag",
			args:     []string{"-v", "version"}, // -v is verbose, not version
			wantExit: 0,
			contains: "0.1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			cmd := exec.Command(binaryPath, tt.args...)
			cmd.Dir = tmpDir

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				}
			}

			if exitCode != tt.wantExit {
				t.Errorf("expected exit code %d, got %d", tt.wantExit, exitCode)
				t.Logf("stdout: %s", stdout.String())
				t.Logf("stderr: %s", stderr.String())
			}

			if tt.contains != "" {
				output := stdout.String() + stderr.String()
				if !strings.Contains(output, tt.contains) {
					t.Errorf("expected output to contain %q, got: %s", tt.contains, output)
				}
			}
		})
	}
}

// TestCLIRecursiveFlag tests the recursive flag functionality.
func TestCLIRecursiveFlag(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer func() { _ = os.Remove(binaryPath) }()

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Create nested directory structure with action files
	subDir := filepath.Join(tmpDir, "subdir")
	_ = os.MkdirAll(subDir, 0755)

	// Write action files
	testutil.WriteTestFile(t, filepath.Join(tmpDir, "action.yml"), testutil.SimpleActionYML)
	testutil.WriteTestFile(t, filepath.Join(subDir, "action.yml"), testutil.CompositeActionYML)

	tests := []struct {
		name     string
		args     []string
		wantExit int
		minFiles int // minimum number of files that should be processed
	}{
		{
			name:     "without recursive flag",
			args:     []string{"gen", "--output-format", "json"},
			wantExit: 0,
			minFiles: 1, // should only process root action.yml
		},
		{
			name:     "with recursive flag",
			args:     []string{"gen", "--recursive", "--output-format", "json"},
			wantExit: 0,
			minFiles: 2, // should process both action.yml files
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			cmd.Dir = tmpDir

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				}
			}

			if exitCode != tt.wantExit {
				t.Errorf("expected exit code %d, got %d", tt.wantExit, exitCode)
				t.Logf("stdout: %s", stdout.String())
				t.Logf("stderr: %s", stderr.String())
			}

			// For recursive tests, check that appropriate number of files were processed
			// This is a simple heuristic - could be made more sophisticated
			output := stdout.String()
			if tt.minFiles > 1 && !strings.Contains(output, "subdir") {
				t.Errorf("expected recursive processing to include subdirectory")
			}
		})
	}
}

// TestCLIErrorHandling tests error scenarios.
func TestCLIErrorHandling(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer func() { _ = os.Remove(binaryPath) }()

	tests := []struct {
		name      string
		args      []string
		setupFunc func(t *testing.T, tmpDir string)
		wantExit  int
		wantError string
	}{
		{
			name: "permission denied on output directory",
			args: []string{"gen", "--output-dir", "/root/restricted"},
			setupFunc: func(t *testing.T, tmpDir string) {
				testutil.WriteTestFile(t, filepath.Join(tmpDir, "action.yml"), testutil.SimpleActionYML)
			},
			wantExit:  1,
			wantError: "permission denied",
		},
		{
			name: "invalid YAML in action file",
			args: []string{"validate"},
			setupFunc: func(t *testing.T, tmpDir string) {
				testutil.WriteTestFile(t, filepath.Join(tmpDir, "action.yml"), "invalid: yaml: content: [")
			},
			wantExit: 1,
		},
		{
			name: "unknown output format",
			args: []string{"gen", "--output-format", "unknown"},
			setupFunc: func(t *testing.T, tmpDir string) {
				testutil.WriteTestFile(t, filepath.Join(tmpDir, "action.yml"), testutil.SimpleActionYML)
			},
			wantExit: 1,
		},
		{
			name: "unknown theme",
			args: []string{"gen", "--theme", "nonexistent-theme"},
			setupFunc: func(t *testing.T, tmpDir string) {
				testutil.WriteTestFile(t, filepath.Join(tmpDir, "action.yml"), testutil.SimpleActionYML)
			},
			wantExit: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			if tt.setupFunc != nil {
				tt.setupFunc(t, tmpDir)
			}

			cmd := exec.Command(binaryPath, tt.args...)
			cmd.Dir = tmpDir

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				}
			}

			if exitCode != tt.wantExit {
				t.Errorf("expected exit code %d, got %d", tt.wantExit, exitCode)
				t.Logf("stdout: %s", stdout.String())
				t.Logf("stderr: %s", stderr.String())
			}

			if tt.wantError != "" {
				output := stdout.String() + stderr.String()
				if !strings.Contains(strings.ToLower(output), strings.ToLower(tt.wantError)) {
					t.Errorf("expected error containing %q, got: %s", tt.wantError, output)
				}
			}
		})
	}
}

// TestCLIConfigInitialization tests configuration initialization.
func TestCLIConfigInitialization(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer func() { _ = os.Remove(binaryPath) }()

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Test config init command
	cmd := exec.Command(binaryPath, "config", "init")
	cmd.Dir = tmpDir

	// Set XDG_CONFIG_HOME to temp directory
	cmd.Env = append(os.Environ(), fmt.Sprintf("XDG_CONFIG_HOME=%s", tmpDir))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() != 0 {
			t.Errorf("config init failed: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
		}
	}

	// Check if config file was created
	expectedConfigPath := filepath.Join(tmpDir, "gh-action-readme", "config.yml")
	if _, err := os.Stat(expectedConfigPath); os.IsNotExist(err) {
		t.Errorf("config file was not created at expected path: %s", expectedConfigPath)
	}
}
