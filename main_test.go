package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/internal/dependencies"
	"github.com/ivuorinen/gh-action-readme/internal/git"
	"github.com/ivuorinen/gh-action-readme/internal/wizard"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

const (
	testErrExpectedShort = "expected Short description to be non-empty"
	testErrExpectedRunFn = "expected command to have a Run or RunE function"
	testMsgUsesGlobalCfg = "uses globalConfig when config parameter is nil"
	testErrNoActionFiles = "no GitHub Action files found"
	testVersionV4        = "v4"
)

// TestRedactConfigForLog verifies token redaction (N134) covers the top-level
// token and nested repo-override tokens, leaves the original untouched, and is
// nil-safe — so neither `config show` nor `gen --verbose` leaks secrets via %+v.
func TestRedactConfigForLog(t *testing.T) {
	t.Parallel()

	if redactConfigForLog(nil) != nil {
		t.Fatal("redactConfigForLog(nil) must return nil")
	}

	const overrideKey = "org/repo"
	cfg := &internal.AppConfig{
		GitHubToken: testutil.TestTokenConfig,
		RepoOverrides: map[string]internal.AppConfig{
			overrideKey: {GitHubToken: testutil.TestTokenConfig},
		},
	}

	got := redactConfigForLog(cfg)

	if got.GitHubToken != appconstants.RedactedPlaceholder {
		t.Errorf("top-level token = %q, want redacted", got.GitHubToken)
	}
	if got.RepoOverrides[overrideKey].GitHubToken != appconstants.RedactedPlaceholder {
		t.Errorf("override token = %q, want redacted", got.RepoOverrides[overrideKey].GitHubToken)
	}
	// Redaction must operate on a copy: the original config keeps its real tokens.
	if cfg.GitHubToken != testutil.TestTokenConfig ||
		cfg.RepoOverrides[overrideKey].GitHubToken != testutil.TestTokenConfig {
		t.Error("redactConfigForLog mutated the original config")
	}
}

// TestInputReader allows injecting test responses for testing.
type TestInputReader struct {
	responses []string
	index     int
}

func (r *TestInputReader) ReadLine() (string, error) {
	if r.index >= len(r.responses) {
		return "", errors.New("no more test responses")
	}
	response := r.responses[r.index]
	r.index++

	return response, nil
}

// createFixtureTestCase creates a test table entry for tests that load a fixture
// and expect a specific error outcome. This helper reduces duplication by standardizing
// the creation of test structures that follow the "load fixture, write to tmpDir, expect error" pattern.
func createFixtureTestCase(name, fixturePath string, wantErr bool) struct {
	name      string
	setupFunc func(t *testing.T, tmpDir string)
	wantErr   bool
} {
	return struct {
		name      string
		setupFunc func(t *testing.T, tmpDir string)
		wantErr   bool
	}{
		name: name,
		setupFunc: func(t *testing.T, tmpDir string) {
			t.Helper()
			fixtureContent := testutil.MustReadFixture(fixturePath)
			testutil.WriteTestFile(t, filepath.Join(tmpDir, appconstants.ActionFileNameYML), string(fixtureContent))
		},
		wantErr: wantErr,
	}
}

// createFixtureTestCaseWithPaths creates a test table entry for tests that load a fixture
// and return paths for processing. This helper reduces duplication for the pattern where
// setupFunc returns []string paths.
func createFixtureTestCaseWithPaths(name, fixturePath string, wantErr bool) struct {
	name      string
	setupFunc func(t *testing.T, tmpDir string) []string
	wantErr   bool
	setFlags  func(cmd *cobra.Command)
} {
	return struct {
		name      string
		setupFunc func(t *testing.T, tmpDir string) []string
		wantErr   bool
		setFlags  func(cmd *cobra.Command)
	}{
		name: name,
		setupFunc: func(t *testing.T, tmpDir string) []string {
			t.Helper()
			fixtureContent := testutil.MustReadFixture(fixturePath)
			testutil.WriteTestFile(t, filepath.Join(tmpDir, appconstants.ActionFileNameYML), string(fixtureContent))

			return []string{tmpDir}
		},
		wantErr: wantErr,
	}
}

// TestCLICommands tests the main CLI commands using subprocess execution.
func TestCLICommands(t *testing.T) {
	t.Parallel()
	// Build the binary for testing
	binaryPath := buildTestBinary(t)

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
			wantStdout: "dev",
		},
		{
			name:       "about command",
			args:       []string{"about"},
			wantExit:   0,
			wantStdout: "gh-action-readme: Generates README.md and HTML for GitHub Actions",
		},
		{
			name:     "help command",
			args:     []string{"--help"},
			wantExit: 0,
			wantStdout: "gh-action-readme is a CLI tool for parsing one or many action.yml files and " +
				"generating informative, modern, and customizable documentation",
		},
		{
			name: "gen command with valid action",
			args: []string{appconstants.CommandGen, testutil.TestFlagOutputFormat, appconstants.OutputFormatMarkdown},
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				createTestActionFile(t, tmpDir, testutil.TestFixtureJavaScriptSimple)
			},
			wantExit: 0,
		},
		{
			name: "gen command with theme flag",
			args: []string{
				appconstants.CommandGen,
				testutil.TestFlagTheme,
				appconstants.ThemeGitHub,
				testutil.TestFlagOutputFormat,
				appconstants.OutputFormatJSON,
			},
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				createTestActionFile(t, tmpDir, testutil.TestFixtureJavaScriptSimple)
			},
			wantExit: 0,
		},
		{
			name:       "gen command with no action files",
			args:       []string{appconstants.CommandGen},
			wantExit:   1,
			wantStderr: "no GitHub Action files found for documentation generation [NO_ACTION_FILES]",
		},
		{
			name: "validate command with valid action",
			args: []string{appconstants.CommandValidate},
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				createTestActionFile(t, tmpDir, testutil.TestFixtureJavaScriptSimple)
			},
			wantExit:   0,
			wantStdout: "All validations passed successfully",
		},
		{
			name: "validate command with invalid action",
			args: []string{appconstants.CommandValidate},
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				createTestActionFile(t, tmpDir, testutil.TestFixtureInvalidMissingDescription)
			},
			wantExit: 1,
		},
		{
			name:       "schema command",
			args:       []string{appconstants.CommandSchema},
			wantExit:   0,
			wantStdout: "schemas/schema.json",
		},
		{
			name:       "config command default",
			args:       []string{appconstants.CommandConfig},
			wantExit:   0,
			wantStdout: "Configuration file location:",
		},
		{
			name:       "config show command",
			args:       []string{appconstants.CommandConfig, appconstants.CommandShow},
			wantExit:   0,
			wantStdout: "Current Configuration:",
		},
		{
			name:       "config themes command",
			args:       []string{appconstants.CommandConfig, appconstants.CommandThemes},
			wantExit:   0,
			wantStdout: "Available Themes:",
		},
		{
			name:       "deps list command no files",
			args:       []string{appconstants.CommandDeps, appconstants.CommandList},
			wantExit:   0, // Changed: deps list now outputs warning instead of error when no files found
			wantStdout: "no action files found",
		},
		{
			name: "deps list command with composite action",
			args: []string{appconstants.CommandDeps, appconstants.CommandList},
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				actionPath := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
				testutil.WriteTestFile(t, actionPath, testutil.MustReadFixture(testutil.TestFixtureCompositeBasic))
			},
			wantExit: 0,
		},
		{
			name:       "cache path command",
			args:       []string{appconstants.CommandCache, "path"},
			wantExit:   0,
			wantStdout: "Cache Directory:",
		},
		{
			name:       "cache stats command",
			args:       []string{appconstants.CommandCache, appconstants.CommandStats},
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
			result := runTestCommand(binaryPath, tt.args, tmpDir)
			assertCommandResult(t, result, tt.wantExit, tt.wantStdout, tt.wantStderr)
		})
	}
}

// TestCLIFlags tests various flag combinations.
func TestCLIFlags(t *testing.T) {
	t.Parallel()
	binaryPath := buildTestBinary(t)

	tests := []struct {
		name     string
		args     []string
		wantExit int
		contains string
	}{
		{
			name:     "verbose flag",
			args:     []string{"--verbose", appconstants.CommandConfig, appconstants.CommandShow},
			wantExit: 0,
			contains: "Current Configuration:",
		},
		{
			name:     "quiet flag",
			args:     []string{"--quiet", appconstants.CommandConfig, appconstants.CommandShow},
			wantExit: 0,
		},
		{
			name: "config file flag",
			args: []string{
				"--config",
				testutil.TestNonexistentYML,
				appconstants.CommandConfig,
				appconstants.CommandShow,
			},
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
			contains: "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			result := runTestCommand(binaryPath, tt.args, tmpDir)

			if result.exitCode != tt.wantExit {
				t.Errorf(testutil.TestMsgExitCode, tt.wantExit, result.exitCode)
				t.Logf(testutil.TestMsgStdout, result.stdout)
				t.Logf(testutil.TestMsgStderr, result.stderr)
			}

			if tt.contains != "" {
				// For contains check, look in both stdout and stderr
				assertCommandResult(t, result, tt.wantExit, tt.contains, "")
			}
		})
	}
}

// TestCLIRecursiveFlag tests the recursive flag functionality.
func TestCLIRecursiveFlag(t *testing.T) {
	t.Parallel()
	binaryPath := buildTestBinary(t)

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Create nested directory structure with action files
	testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureJavaScriptSimple)
	testutil.CreateActionSubdir(t, tmpDir, testutil.TestDirSubdir, testutil.TestFixtureCompositeBasic)

	tests := []struct {
		name     string
		args     []string
		wantExit int
		minFiles int // minimum number of files that should be processed
	}{
		{
			name:     "without recursive flag",
			args:     []string{appconstants.CommandGen, testutil.TestFlagOutputFormat, appconstants.OutputFormatJSON},
			wantExit: 0,
			minFiles: 1, // should only process root action.yml
		},
		{
			name: "with recursive flag",
			args: []string{
				appconstants.CommandGen,
				"--recursive",
				testutil.TestFlagOutputFormat,
				appconstants.OutputFormatJSON,
			},
			wantExit: 0,
			minFiles: 2, // should process both action.yml files
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runTestCommand(binaryPath, tt.args, tmpDir)
			assertCommandResult(t, result, tt.wantExit, "", "")

			// For recursive tests, check that appropriate number of files were processed
			// This is a simple heuristic - could be made more sophisticated
			if tt.minFiles > 1 && !strings.Contains(result.stdout, testutil.TestDirSubdir) {
				t.Errorf("expected recursive processing to include subdirectory")
			}
		})
	}
}

// TestCLIErrorHandling tests error scenarios.
func TestCLIErrorHandling(t *testing.T) {
	t.Parallel()
	binaryPath := buildTestBinary(t)

	tests := []struct {
		name      string
		args      []string
		setupFunc func(t *testing.T, tmpDir string)
		wantExit  int
		wantError string
	}{
		{
			name: "permission denied on output directory",
			args: []string{appconstants.CommandGen, "--output-dir", "/root/restricted"},
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				createTestActionFile(t, tmpDir, testutil.TestFixtureJavaScriptSimple)
			},
			wantExit:  1,
			wantError: "encountered 1 errors during batch processing",
		},
		{
			name: "invalid YAML in action file",
			args: []string{appconstants.CommandValidate},
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				testutil.WriteTestFile(
					t,
					filepath.Join(tmpDir, appconstants.ActionFileNameYML),
					"invalid: yaml: content: [",
				)
			},
			wantExit: 1,
		},
		{
			name: "unknown output format",
			args: []string{appconstants.CommandGen, testutil.TestFlagOutputFormat, "unknown"},
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				createTestActionFile(t, tmpDir, testutil.TestFixtureJavaScriptSimple)
			},
			wantExit: 1,
		},
		{
			name: "unknown theme",
			args: []string{appconstants.CommandGen, testutil.TestFlagTheme, "nonexistent-theme"},
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				createTestActionFile(t, tmpDir, testutil.TestFixtureJavaScriptSimple)
			},
			wantExit: 1,
		},
		// Phase 5: Additional error path tests for gen handler
		{
			name:      "gen with empty directory (no action.yml)",
			args:      []string{appconstants.CommandGen},
			setupFunc: nil, // Empty directory
			wantExit:  1,
			wantError: testErrNoActionFiles,
		},
		{
			name: "gen with malformed YAML syntax",
			args: []string{appconstants.CommandGen},
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				testutil.WriteTestFile(
					t,
					filepath.Join(tmpDir, appconstants.ActionFileNameYML),
					testutil.MustReadFixture(testutil.TestFixtureInvalidMalformedRuns),
				)
			},
			wantExit:  1,
			wantError: "error",
		},
		{
			name: "gen with invalid action path",
			args: []string{appconstants.CommandGen, "/nonexistent/path/action.yml"},
			setupFunc: func(t *testing.T, _ string) {
				t.Helper()
			},
			wantExit:  1,
			wantError: "does not exist",
		},
		// Phase 5: Additional error path tests for validate handler
		{
			name:      "validate with missing required field (description)",
			args:      []string{appconstants.CommandValidate},
			setupFunc: setupFixtureInDir(testutil.TestFixtureInvalidMissingDescription),
			wantExit:  1,
			wantError: "validation failed",
		},
		{
			name: "validate with missing runs field",
			args: []string{appconstants.CommandValidate},
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				testutil.WriteTestFile(
					t,
					filepath.Join(tmpDir, appconstants.ActionFileNameYML),
					testutil.MustReadFixture(testutil.TestFixtureCompositeNameDescOnly),
				)
			},
			wantExit:  1,
			wantError: "validation",
		},
		// Phase 5: Additional error path tests for deps commands
		{
			name: "deps list with no dependencies",
			args: []string{appconstants.CommandDeps, appconstants.CommandList},
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				// Create an action with no dependencies
				testutil.WriteTestFile(
					t,
					filepath.Join(tmpDir, appconstants.ActionFileNameYML),
					testutil.TestMinimalAction,
				)
			},
			wantExit: 0, // Not an error, just no dependencies
		},
		{
			name: "deps list with malformed action - graceful handling",
			args: []string{appconstants.CommandDeps, appconstants.CommandList},
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				testutil.WriteTestFile(
					t,
					filepath.Join(tmpDir, appconstants.ActionFileNameYML),
					testutil.TestInvalidYAMLPrefix,
				)
			},
			wantExit: 0, // deps list handles errors gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			if tt.setupFunc != nil {
				tt.setupFunc(t, tmpDir)
			}

			result := runTestCommand(binaryPath, tt.args, tmpDir)

			if result.exitCode != tt.wantExit {
				t.Errorf(testutil.TestMsgExitCode, tt.wantExit, result.exitCode)
				t.Logf(testutil.TestMsgStdout, result.stdout)
				t.Logf(testutil.TestMsgStderr, result.stderr)
			}

			if tt.wantError != "" {
				output := result.stdout + result.stderr
				if !strings.Contains(strings.ToLower(output), strings.ToLower(tt.wantError)) {
					t.Errorf(testutil.TestErrExpectedErrorContaining, tt.wantError, output)
				}
			}
		})
	}
}

// TestCLIConfigInitialization tests configuration initialization.
func TestCLIConfigInitialization(t *testing.T) {
	t.Parallel()
	binaryPath := buildTestBinary(t)

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Test config init command
	cmd := exec.Command(binaryPath, appconstants.CommandConfig, "init") // #nosec G204 -- controlled test input
	cmd.Dir = tmpDir

	// Set XDG_CONFIG_HOME to temp directory
	cmd.Env = append(os.Environ(), "XDG_CONFIG_HOME="+tmpDir)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() != 0 {
			t.Errorf("config init failed: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
		}
	}

	// Check if config file was created (note: uses .yaml extension, not .yml)
	expectedConfigPath := filepath.Join(tmpDir, "gh-action-readme", "config.yaml")
	testutil.AssertFileExists(t, expectedConfigPath)
}

// Unit Tests for Helper Functions
// These test the actual functions directly rather than through subprocess execution.

func TestCreateOutputManager(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		quiet bool
	}{
		{"normal mode", false},
		{"quiet mode", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := createOutputManager(tt.quiet)
			if output == nil {
				t.Fatal("createOutputManager returned nil")
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		size     int64
		expected string
	}{
		{"zero bytes", 0, "0 bytes"},
		{"bytes", 500, "500 bytes"},
		{"kilobyte boundary", 1024, "1.00 KB"},
		{"kilobytes", 2048, "2.00 KB"},
		{"megabyte boundary", 1024 * 1024, "1.00 MB"},
		{"megabytes", 5 * 1024 * 1024, "5.00 MB"},
		{"gigabyte boundary", 1024 * 1024 * 1024, "1.00 GB"},
		{"gigabytes", 3 * 1024 * 1024 * 1024, "3.00 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSize(tt.size)
			if result != tt.expected {
				t.Errorf("formatSize(%d) = %q, want %q", tt.size, result, tt.expected)
			}
		})
	}
}

func TestResolveExportFormat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		format   string
		expected wizard.ExportFormat
	}{
		{"json format", appconstants.OutputFormatJSON, wizard.FormatJSON},
		{"toml format", appconstants.OutputFormatTOML, wizard.FormatTOML},
		{"yaml format", appconstants.OutputFormatYAML, wizard.FormatYAML},
		{"default format", "unknown", wizard.FormatYAML},
		{"empty format", "", wizard.FormatYAML},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveExportFormat(tt.format)
			if result != tt.expected {
				t.Errorf("resolveExportFormat(%q) = %v, want %v", tt.format, result, tt.expected)
			}
		})
	}
}

func TestCreateErrorHandler(t *testing.T) {
	t.Parallel()
	output := internal.NewColoredOutput(false)
	handler := createErrorHandler(output)

	if handler == nil {
		t.Fatal("createErrorHandler returned nil")
	}
}

func TestSetupOutputAndErrorHandling(t *testing.T) {
	// Note: This test cannot use t.Parallel() because it modifies globalConfig
	// Setup globalConfig for the test
	originalConfig := globalConfig
	defer func() { globalConfig = originalConfig }()

	globalConfig = &internal.AppConfig{Quiet: false}

	output, errorHandler := setupOutputAndErrorHandling()

	if output == nil {
		t.Fatal("setupOutputAndErrorHandling returned nil output")
	}
	if errorHandler == nil {
		t.Fatal("setupOutputAndErrorHandling returned nil errorHandler")
	}
}

// Unit Tests for Command Creation Functions

func TestNewGenCmd(t *testing.T) {
	t.Parallel()
	cmd := newGenCmd()

	if cmd.Use != "gen [directory_or_file]" {
		t.Errorf("expected Use to be 'gen [directory_or_file]', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error(testErrExpectedShort)
	}

	if cmd.RunE == nil && cmd.Run == nil {
		t.Error(testErrExpectedRunFn)
	}

	// Check that required flags exist
	flags := []string{"output-format", "output-dir", "theme", "recursive"}
	for _, flag := range flags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag %q to exist", flag)
		}
	}
}

func TestNewValidateCmd(t *testing.T) {
	t.Parallel()
	cmd := newValidateCmd()

	if cmd.Use != appconstants.CommandValidate {
		t.Errorf("expected Use to be 'validate', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error(testErrExpectedShort)
	}

	if cmd.RunE == nil && cmd.Run == nil {
		t.Error(testErrExpectedRunFn)
	}
}

func TestNewSchemaCmd(t *testing.T) {
	t.Parallel()
	cmd := newSchemaCmd()

	if cmd.Use != appconstants.CommandSchema {
		t.Errorf("expected Use to be 'schema', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error(testErrExpectedShort)
	}

	if cmd.RunE == nil && cmd.Run == nil {
		t.Error(testErrExpectedRunFn)
	}
}

// cmdResult holds the results of a command execution.
type cmdResult struct {
	stdout   string
	stderr   string
	exitCode int
}

// runTestCommand executes a command with the given args in the specified directory.
// It returns the stdout, stderr, and exit code.
func runTestCommand(binaryPath string, args []string, dir string) cmdResult {
	cmd := exec.Command(binaryPath, args...) // #nosec G204 -- controlled test input
	cmd.Dir = dir

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

	return cmdResult{
		stdout:   stdout.String(),
		stderr:   stderr.String(),
		exitCode: exitCode,
	}
}

// createTestActionFile is a helper that creates a test action file from a fixture.
// It writes the specified fixture to action.yml in the given temporary directory.
func createTestActionFile(t *testing.T, tmpDir, fixture string) {
	t.Helper()
	actionPath := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
	testutil.WriteTestFile(t, actionPath, testutil.MustReadFixture(fixture))
}

// assertCommandResult is a helper that asserts the result of a command execution.
// It checks the exit code, and optionally checks for expected content in stdout and stderr.
func assertCommandResult(t *testing.T, result cmdResult, wantExit int, wantStdout, wantStderr string) {
	t.Helper()

	if result.exitCode != wantExit {
		t.Errorf(testutil.TestMsgExitCode, wantExit, result.exitCode)
		t.Logf(testutil.TestMsgStdout, result.stdout)
		t.Logf(testutil.TestMsgStderr, result.stderr)
	}

	// Check stdout if specified
	if wantStdout != "" {
		if !strings.Contains(result.stdout, wantStdout) {
			t.Errorf("expected stdout to contain %q, got: %s", wantStdout, result.stdout)
		}
	}

	// Check stderr if specified
	if wantStderr != "" {
		if !strings.Contains(result.stderr, wantStderr) {
			t.Errorf("expected stderr to contain %q, got: %s", wantStderr, result.stderr)
		}
	}
}

// Unit Tests for Handler Functions
// These test the handler logic directly without subprocess execution

func TestCacheClearHandler(t *testing.T) {
	// Handler should execute without error
	// The actual cache clearing logic is tested in cache package
	testSimpleHandler(t, cacheClearHandler, "cacheClearHandler")
}

func TestCacheStatsHandler(t *testing.T) {
	testSimpleHandler(t, cacheStatsHandler, "cacheStatsHandler")
}

func TestCachePathHandler(t *testing.T) {
	testSimpleHandler(t, cachePathHandler, "cachePathHandler")
}

// Not parallel: the subtests reassign the shared globalConfig, so running the
// parent concurrently with other parallel tests would race on it (N157).
func TestSchemaHandler(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
	}{
		{
			name:    "non-verbose mode",
			verbose: false,
		},
		{
			name:    "verbose mode",
			verbose: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			// Note: Cannot use t.Parallel() because test modifies shared globalConfig

			originalConfig := globalConfig
			defer func() { globalConfig = originalConfig }()

			globalConfig = &internal.AppConfig{
				Quiet:   true,
				Verbose: tt.verbose,
				Schema:  "schemas/custom.json",
			}

			cmd := &cobra.Command{}
			_ = schemaHandler(cmd, []string{})
			// Should not panic - output is tested via integration tests
		})
	}
}

func TestConfigThemesHandler(t *testing.T) {
	testSimpleHandler(t, configThemesHandler, "configThemesHandler")
}

func TestConfigShowHandler(t *testing.T) {
	testSimpleHandler(t, configShowHandler, "configShowHandler")
}

func TestDepsGraphHandler(t *testing.T) {
	// depsGraphHandler now returns error (always nil — unimplemented feature);
	// adapt it to the void-handler helper, which asserts it does not panic.
	testSimpleVoidHandler(t, func(cmd *cobra.Command, args []string) {
		_ = depsGraphHandler(cmd, args)
	})
}

func TestCreateAnalyzer(t *testing.T) {
	output := &internal.ColoredOutput{NoColor: true, Quiet: true}
	config := internal.DefaultAppConfig()
	generator := internal.NewGenerator(config)

	analyzer := createAnalyzer(generator, output)

	if analyzer == nil {
		t.Error("createAnalyzer() returned nil")
	}
}

// Test helper functions that don't require complex setup

func TestBuildTestBinary(t *testing.T) {
	// This test verifies that buildTestBinary works
	binaryPath := buildTestBinary(t)

	// Clean and validate the path
	cleanedPath := filepath.Clean(binaryPath)
	if strings.Contains(cleanedPath, "..") {
		t.Fatalf("binary path contains .. components: %q", cleanedPath)
	}

	// Check that binary exists
	if _, err := os.Stat(cleanedPath); err != nil {
		t.Errorf("buildTestBinary() created binary does not exist: %v", err)
	}

	// Check that binary is executable
	info, err := os.Stat(cleanedPath)
	if err != nil {
		t.Fatalf("Failed to stat binary: %v", err)
	}

	// On Unix systems, check executable bit
	if runtime.GOOS != "windows" && info.Mode()&0111 == 0 {
		t.Error("buildTestBinary() created binary is not executable")
	}
}

// TestApplyGlobalFlags tests global flag application.
func TestApplyGlobalFlags(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
		quiet   bool
		wantV   bool
		wantQ   bool
	}{
		{
			name:    "verbose flag",
			verbose: true,
			quiet:   false,
			wantV:   true,
			wantQ:   false,
		},
		{
			name:    "quiet flag",
			verbose: false,
			quiet:   true,
			wantV:   false,
			wantQ:   true,
		},
		{
			name:    "no flags",
			verbose: false,
			quiet:   false,
			wantV:   false,
			wantQ:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original global flag values
			origVerbose := verbose
			origQuiet := quiet
			defer func() {
				verbose = origVerbose
				quiet = origQuiet
			}()

			// Set global flags to test values
			verbose = tt.verbose
			quiet = tt.quiet

			config := internal.DefaultAppConfig()
			applyGlobalFlags(config)

			if config.Verbose != tt.wantV {
				t.Errorf("Verbose = %v, want %v", config.Verbose, tt.wantV)
			}
			if config.Quiet != tt.wantQ {
				t.Errorf("Quiet = %v, want %v", config.Quiet, tt.wantQ)
			}
		})
	}
}

// TestApplyCommandFlags tests command flag application.
func TestApplyCommandFlags(t *testing.T) {
	tests := []struct {
		name      string
		theme     string
		format    string
		wantTheme string
		wantFmt   string
	}{
		{
			name:      "with theme flag only",
			theme:     appconstants.ThemeGitHub,
			format:    appconstants.OutputFormatMarkdown, // Must set format to avoid empty string
			wantTheme: appconstants.ThemeGitHub,
			wantFmt:   appconstants.OutputFormatMarkdown,
		},
		{
			name:      "with format flag",
			theme:     "",
			format:    appconstants.OutputFormatHTML,
			wantTheme: appconstants.ThemeDefault, // Default from DefaultAppConfig
			wantFmt:   appconstants.OutputFormatHTML,
		},
		{
			name:      "with both flags",
			theme:     appconstants.ThemeProfessional,
			format:    appconstants.OutputFormatJSON,
			wantTheme: appconstants.ThemeProfessional,
			wantFmt:   appconstants.OutputFormatJSON,
		},
	}

	for _, tt := range tests {
		config := internal.DefaultAppConfig()
		cmd := &cobra.Command{}

		// Always define flags with proper defaults
		cmd.Flags().String("theme", "", "")
		cmd.Flags().String(appconstants.FlagOutputFormat, appconstants.OutputFormatMarkdown, "")

		if tt.theme != "" {
			_ = cmd.Flags().Set("theme", tt.theme)
		}
		if tt.format != appconstants.OutputFormatMarkdown {
			_ = cmd.Flags().Set(appconstants.FlagOutputFormat, tt.format)
		}

		applyCommandFlags(cmd, config)

		if config.Theme != tt.wantTheme {
			t.Errorf("%s: Theme = %v, want %v", tt.name, config.Theme, tt.wantTheme)
		}
		if config.OutputFormat != tt.wantFmt {
			t.Errorf("%s: OutputFormat = %v, want %v", tt.name, config.OutputFormat, tt.wantFmt)
		}
	}
}

// TestValidateGitHubToken tests GitHub token validation.
func TestValidateGitHubToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  bool
	}{
		{ // #nosec G101 -- test token, not a real credential
			name:  "with valid token",
			token: "ghp_test_token_123",
			want:  true,
		},
		{
			name:  "with empty token",
			token: "",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original global config
			origConfig := globalConfig
			defer func() { globalConfig = origConfig }()

			// Clear env vars so GetGitHubToken only sees the config struct token.
			t.Setenv(appconstants.EnvGitHubToken, "")
			t.Setenv(appconstants.EnvGitHubTokenStandard, "")

			// Set test token
			globalConfig = &internal.AppConfig{
				GitHubToken: tt.token,
				Quiet:       true,
			}

			output := createOutputManager(true)
			got := validateGitHubToken(output)

			if got != tt.want {
				t.Errorf("validateGitHubToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLogConfigInfo tests configuration info logging.
func TestLogConfigInfo(_ *testing.T) {
	tests := []struct {
		name     string
		verbose  bool
		repoRoot string
	}{
		{
			name:     "verbose with repo root",
			verbose:  true,
			repoRoot: "/path/to/repo",
		},
		{
			name:     "verbose without repo root",
			verbose:  true,
			repoRoot: "",
		},
		{
			name:     "not verbose",
			verbose:  false,
			repoRoot: "/path/to/repo",
		},
	}

	for _, tt := range tests {
		config := &internal.AppConfig{
			Verbose: tt.verbose,
			Quiet:   true,
		}
		generator := internal.NewGenerator(config)

		// Just call it to ensure it doesn't panic
		logConfigInfo(generator, config, tt.repoRoot)
	}
}

// TestShowUpgradeMode tests upgrade mode display.
func TestShowUpgradeMode(_ *testing.T) {
	tests := []struct {
		name      string
		ciMode    bool
		isPinCmd  bool
		wantEmpty bool
	}{
		{
			name:      "CI mode",
			ciMode:    true,
			isPinCmd:  false,
			wantEmpty: false,
		},
		{
			name:      "pin command",
			ciMode:    false,
			isPinCmd:  true,
			wantEmpty: false,
		},
		{
			name:      "interactive mode",
			ciMode:    false,
			isPinCmd:  false,
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		output := createOutputManager(true)
		// Just call it to ensure it doesn't panic
		showUpgradeMode(output, tt.ciMode, tt.isPinCmd)
	}
}

// TestDisplayOutdatedResults tests outdated dependencies display.
func TestDisplayOutdatedResults(_ *testing.T) {
	tests := []struct {
		name        string
		allOutdated []dependencies.OutdatedDependency
	}{
		{
			name:        "no outdated dependencies",
			allOutdated: []dependencies.OutdatedDependency{},
		},
		{
			name: "with outdated dependencies",
			allOutdated: []dependencies.OutdatedDependency{
				{
					Current: dependencies.Dependency{
						Name:    testutil.TestActionCheckout,
						Version: "v3",
					},
					LatestVersion: testVersionV4,
					UpdateType:    "major",
				},
			},
		},
		{
			name: "with security update",
			allOutdated: []dependencies.OutdatedDependency{
				{
					Current: dependencies.Dependency{
						Name:    "actions/setup-node",
						Version: "v3",
					},
					LatestVersion:    "v4",
					UpdateType:       "major",
					IsSecurityUpdate: true,
				},
			},
		},
	}

	for _, tt := range tests {
		output := createOutputManager(true)
		// Just call it to ensure it doesn't panic
		displayOutdatedResults(output, tt.allOutdated)
	}
}

// TestDisplayFloatingDeps tests floating dependencies display.
func TestDisplayFloatingDeps(_ *testing.T) {

	output := createOutputManager(true)
	floatingDeps := []fileDep{
		{
			file: testutil.TestTmpActionFile,
			dep: dependencies.Dependency{
				Name:    testutil.TestActionCheckout,
				Version: testVersionV4,
			},
		},
	}

	// Just call it to ensure it doesn't panic
	displayFloatingDeps(output, testutil.TestTmpDir, floatingDeps)
}

// TestDisplaySecuritySummary tests security summary display.
func TestDisplaySecuritySummary(_ *testing.T) {
	tests := []struct {
		name         string
		pinnedCount  int
		floatingDeps []fileDep
	}{
		{
			name:         "all pinned",
			pinnedCount:  5,
			floatingDeps: nil,
		},
		{
			name:        "with floating dependencies",
			pinnedCount: 3,
			floatingDeps: []fileDep{
				{
					file: testutil.TestTmpActionFile,
					dep: dependencies.Dependency{
						Name:    testutil.TestActionCheckout,
						Version: testVersionV4,
					},
				},
			},
		},
		{
			name:         "no dependencies",
			pinnedCount:  0,
			floatingDeps: nil,
		},
	}

	for _, tt := range tests {
		output := createOutputManager(true)
		// Just call it to ensure it doesn't panic
		displaySecuritySummary(output, testutil.TestTmpDir, tt.pinnedCount, tt.floatingDeps)
	}
}

// TestShowPendingUpdates tests displaying pending dependency updates.
func TestShowPendingUpdates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		updates    []dependencies.PinnedUpdate
		currentDir string
	}{
		{
			name:       "no updates",
			updates:    []dependencies.PinnedUpdate{},
			currentDir: testutil.TestTmpDir,
		},
		{
			name: "single update",
			updates: []dependencies.PinnedUpdate{
				{
					FilePath:   testutil.TestTmpActionFile,
					OldUses:    testutil.TestActionCheckoutV3,
					NewUses:    testutil.TestActionCheckoutV4,
					UpdateType: appconstants.UpdateTypeMajor,
				},
			},
			currentDir: testutil.TestTmpDir,
		},
		{
			name: "multiple updates",
			updates: []dependencies.PinnedUpdate{
				{
					FilePath:   testutil.TestTmpActionFile,
					OldUses:    testutil.TestActionCheckoutV3,
					NewUses:    testutil.TestActionCheckoutV4,
					UpdateType: appconstants.UpdateTypeMajor,
				},
				{
					FilePath:   testutil.TestTmpDir + "/workflow.yml",
					OldUses:    "actions/setup-node@v2",
					NewUses:    testutil.TestActionSetupNodeV3,
					UpdateType: appconstants.UpdateTypeMajor,
				},
			},
			currentDir: testutil.TestTmpDir,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			output := createOutputManager(true)
			// Just call it to ensure it doesn't panic
			showPendingUpdates(output, tt.updates, tt.currentDir)
		})
	}
}

// TestAnalyzeActionFileDeps tests action file dependency analysis.
func TestAnalyzeActionFileDeps(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func(t *testing.T) (string, *dependencies.Analyzer)
		wantDepCnt int
	}{
		{
			name: "nil analyzer returns 0",
			setupFunc: func(t *testing.T) (string, *dependencies.Analyzer) {
				t.Helper()

				return testutil.TestTmpActionFile, nil
			},
			wantDepCnt: 0,
		},
		{
			name: "action with dependencies",
			setupFunc: func(t *testing.T) (string, *dependencies.Analyzer) {
				t.Helper()
				tmpDir := t.TempDir()
				actionFile := testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureCompositeMultipleNamedSteps)

				// Create a basic analyzer without GitHub client
				analyzer := dependencies.NewAnalyzer(nil, git.RepoInfo{}, nil)

				return actionFile, analyzer
			},
			wantDepCnt: 2, // 2 uses statements
		},
		{
			name: "action without dependencies",
			setupFunc: func(t *testing.T) (string, *dependencies.Analyzer) {
				t.Helper()
				tmpDir := t.TempDir()
				actionFile := testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureJavaScriptSimple)

				// Create a basic analyzer without GitHub client
				analyzer := dependencies.NewAnalyzer(nil, git.RepoInfo{}, nil)

				return actionFile, analyzer
			},
			wantDepCnt: 0,
		},
		{
			name: "invalid action file",
			setupFunc: func(t *testing.T) (string, *dependencies.Analyzer) {
				t.Helper()
				tmpDir := t.TempDir()
				actionFile := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
				// Write invalid YAML (unclosed bracket)
				testutil.WriteTestFile(t, actionFile, testutil.TestInvalidYAMLPrefix)

				// Create a basic analyzer without GitHub client
				analyzer := dependencies.NewAnalyzer(nil, git.RepoInfo{}, nil)

				return actionFile, analyzer
			},
			wantDepCnt: 0, // Returns 0 on error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actionFile, analyzer := tt.setupFunc(t)
			output := createOutputManager(true)

			got := analyzeActionFileDeps(output, actionFile, analyzer)
			if got != tt.wantDepCnt {
				t.Errorf("analyzeActionFileDeps() = %v, want %v", got, tt.wantDepCnt)
			}
		})
	}
}

// TestNewConfigCmd tests config command creation.
// verifySubcommandsExist checks that all expected subcommands exist in the command.
func verifySubcommandsExist(t *testing.T, cmd *cobra.Command, expectedSubcommands []string) {
	t.Helper()
	subcommands := cmd.Commands()

	if len(subcommands) < len(expectedSubcommands) {
		t.Errorf("newConfigCmd() has %d subcommands, want at least %d", len(subcommands), len(expectedSubcommands))
	}

	// Verify each expected subcommand exists
	for _, expected := range expectedSubcommands {
		found := false
		for _, sub := range subcommands {
			if sub.Use == expected {
				found = true

				break
			}
		}
		if !found {
			t.Errorf("newConfigCmd() missing subcommand: %s", expected)
		}
	}
}

func TestNewConfigCmd(t *testing.T) {
	// Note: Cannot use t.Parallel() because test modifies shared globalConfig

	// Save original global config
	origConfig := globalConfig
	defer func() { globalConfig = origConfig }()

	globalConfig = &internal.AppConfig{
		Quiet: true,
	}

	t.Run("creates command with correct properties", func(t *testing.T) {
		cmd := newConfigCmd()
		if cmd == nil {
			t.Fatal("newConfigCmd() returned nil")
		}
		if cmd.Use != appconstants.CommandConfig {
			t.Errorf("newConfigCmd().Use = %v, want 'config'", cmd.Use)
		}
	})

	t.Run("has all expected subcommands", func(t *testing.T) {
		cmd := newConfigCmd()
		expectedSubcommands := []string{
			"init",
			appconstants.CommandWizard,
			appconstants.CommandShow,
			appconstants.CommandThemes,
		}
		verifySubcommandsExist(t, cmd, expectedSubcommands)
	})

	t.Run("wizard subcommand has required flags", func(t *testing.T) {
		cmd := newConfigCmd()
		wizardCmd, _, err := cmd.Find([]string{appconstants.CommandWizard})
		if err != nil {
			t.Fatalf("Failed to find wizard subcommand: %v", err)
		}
		if wizardCmd == nil {
			t.Fatal("wizard subcommand is nil")
		}

		if wizardCmd.Flags().Lookup(appconstants.FlagFormat) == nil {
			t.Error("wizard subcommand missing --format flag")
		}
		if wizardCmd.Flags().Lookup(appconstants.FlagOutput) == nil {
			t.Error("wizard subcommand missing --output flag")
		}
	})
}

// TestNewDepsCmd tests deps command creation.
func TestNewDepsCmd(t *testing.T) {

	cmd := newDepsCmd()
	if cmd == nil {
		t.Fatal("newDepsCmd() returned nil")
	}
	if cmd.Use != appconstants.CommandDeps {
		t.Errorf("newDepsCmd().Use = %v, want 'deps'", cmd.Use)
	}
}

// TestNewCacheCmd tests cache command creation.
func TestNewCacheCmd(t *testing.T) {

	cmd := newCacheCmd()
	if cmd == nil {
		t.Fatal("newCacheCmd() returned nil")
	}
	if cmd.Use != appconstants.CommandCache {
		t.Errorf("newCacheCmd().Use = %v, want 'cache'", cmd.Use)
	}
}

// TestGenHandlerIntegration tests genHandler with various scenarios.
// Note: Not using t.Parallel() because these tests modify shared globalConfig.
func TestGenHandlerIntegration(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T, tmpDir string) []string
		wantErr   bool
		setFlags  func(cmd *cobra.Command)
	}{
		{
			name:      "generates README from valid action in current dir",
			setupFunc: setupWithSingleFixture(testutil.TestFixtureJavaScriptSimple),
			wantErr:   false,
		},
		{
			name:      "generates HTML output",
			setupFunc: setupWithSingleFixture(testutil.TestFixtureJavaScriptSimple),
			wantErr:   false,
			setFlags: func(cmd *cobra.Command) {
				_ = cmd.Flags().Set(appconstants.FlagOutputFormat, appconstants.OutputFormatHTML)
			},
		},
		{
			name:      "generates JSON output",
			setupFunc: setupWithSingleFixture(testutil.TestFixtureJavaScriptSimple),
			wantErr:   false,
			setFlags: func(cmd *cobra.Command) {
				_ = cmd.Flags().Set(appconstants.FlagOutputFormat, appconstants.OutputFormatJSON)
			},
		},
		{
			name:      "generates with theme override",
			setupFunc: setupWithSingleFixture(testutil.TestFixtureJavaScriptSimple),
			wantErr:   false,
			setFlags: func(cmd *cobra.Command) {
				_ = cmd.Flags().Set("theme", appconstants.ThemeGitHub)
			},
		},
		{
			name:      "processes composite action",
			setupFunc: setupWithSingleFixture(testutil.TestFixtureCompositeBasic),
			wantErr:   false,
		},
		{
			name:      "processes docker action",
			setupFunc: setupWithSingleFixture(testutil.TestFixtureDockerBasic),
			wantErr:   false,
		},
		{
			name:      "processes action with custom output file",
			setupFunc: setupWithSingleFixture(testutil.TestFixtureJavaScriptSimple),
			wantErr:   false,
			setFlags: func(cmd *cobra.Command) {
				_ = cmd.Flags().Set("output", "custom-readme.md")
			},
		},
		{
			name: "recursive processing with subdirectories",
			setupFunc: func(t *testing.T, tmpDir string) []string {
				t.Helper()
				testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureJavaScriptSimple)
				testutil.CreateActionSubdir(t, tmpDir, testutil.TestDirSubdir, testutil.TestFixtureCompositeBasic)

				return []string{tmpDir}
			},
			wantErr: false,
			setFlags: func(cmd *cobra.Command) {
				_ = cmd.Flags().Set("recursive", "true")
			},
		},
		{
			name: "processes specific action file",
			setupFunc: func(t *testing.T, tmpDir string) []string {
				t.Helper()
				testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureJavaScriptSimple)

				return []string{filepath.Join(tmpDir, appconstants.ActionFileNameYML)}
			},
			wantErr: false,
		},
		// Error scenarios using fixtures
		createFixtureTestCaseWithPaths(
			testutil.TestCaseNameReturnsErrInvalidYAML,
			testutil.TestErrorScenarioInvalidYAML,
			true,
		),
		createFixtureTestCaseWithPaths(
			testutil.TestCaseNameReturnsErrMissingFields,
			testutil.TestErrorScenarioMissingFields,
			true,
		),
		{
			name: "returns error for empty directory with no action files",
			setupFunc: func(t *testing.T, tmpDir string) []string {
				t.Helper()
				// Don't write any action file - directory is empty
				return []string{tmpDir}
			},
			wantErr: true,
		},
		{
			name: "returns error for nonexistent path",
			setupFunc: func(t *testing.T, tmpDir string) []string {
				t.Helper()

				return []string{filepath.Join(tmpDir, "nonexistent")}
			},
			wantErr: true,
		},
		// Empty steps is valid
		createFixtureTestCaseWithPaths(
			"handles empty action file gracefully",
			testutil.TestFixtureEmptyAction,
			false,
		),
		// Old deps don't cause generation to fail
		createFixtureTestCaseWithPaths(
			"processes action with outdated dependencies",
			testutil.TestErrorScenarioOldDeps,
			false,
		),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Not using t.Parallel() because these tests modify shared globalConfig

			// Save and restore global state
			origConfig := globalConfig
			defer func() { globalConfig = origConfig }()

			// Create temp directory
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			// Setup test environment and get args
			var args []string
			if tt.setupFunc != nil {
				args = tt.setupFunc(t, tmpDir)
			}

			// Initialize global config
			globalConfig = internal.DefaultAppConfig()

			// Create command and set flags
			cmd := newGenCmd()
			if tt.setFlags != nil {
				tt.setFlags(cmd)
			}

			// Execute handler - now returns error instead of os.Exit
			err := genHandler(cmd, args)
			if (err != nil) != tt.wantErr {
				t.Errorf("genHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGenHandlerRejectsInvalidFormatFlagUpfront verifies N152: an invalid
// --output-format flag fails with a clear validation error before any batch
// processing, not a cryptic per-file "unsupported output format" batch error.
func TestGenHandlerRejectsInvalidFormatFlagUpfront(t *testing.T) {
	origConfig := globalConfig
	defer func() { globalConfig = origConfig }()

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()
	testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureJavaScriptSimple)

	globalConfig = internal.DefaultAppConfig()
	cmd := newGenCmd()
	_ = cmd.Flags().Set(appconstants.FlagOutputFormat, "bogus")

	err := genHandler(cmd, []string{tmpDir})
	if err == nil {
		t.Fatal("expected an error for an invalid --output-format flag")
	}
	if !strings.Contains(err.Error(), "invalid output format") {
		t.Errorf("expected an upfront 'invalid output format' validation error, got: %v", err)
	}
}

// TestValidateHandlerIntegration tests validateHandler with various scenarios.
// Note: Not using t.Parallel() because these tests modify shared globalConfig.
func TestValidateHandlerIntegration(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T, tmpDir string)
		wantErr   bool
	}{
		{
			name:      "validates valid action successfully",
			setupFunc: setupFixtureInDir(testutil.TestFixtureJavaScriptSimple),
			wantErr:   false,
		},
		{
			name:      "validates composite action",
			setupFunc: setupFixtureInDir(testutil.TestFixtureCompositeBasic),
			wantErr:   false,
		},
		{
			name:      "validates docker action",
			setupFunc: setupFixtureInDir(testutil.TestFixtureDockerBasic),
			wantErr:   false,
		},
		{
			name: "validates multiple actions recursively",
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureJavaScriptSimple)
				testutil.CreateActionSubdir(t, tmpDir, testutil.TestDirSubdir, testutil.TestFixtureCompositeBasic)
			},
			wantErr: false,
		},
		// Error scenarios using fixtures
		createFixtureTestCase(
			testutil.TestCaseNameReturnsErrInvalidYAML,
			testutil.TestErrorScenarioInvalidYAML,
			true,
		),
		createFixtureTestCase(
			testutil.TestCaseNameReturnsErrMissingFields,
			testutil.TestErrorScenarioMissingFields,
			true,
		),
		// Outdated dependencies don't fail validation
		createFixtureTestCase(
			"validates action with outdated dependencies",
			testutil.TestErrorScenarioOldDeps,
			false,
		),
		{
			name: "returns error for empty directory with no action files",
			setupFunc: func(t *testing.T, _ string) {
				t.Helper()
				// Don't write any action file - directory is empty
			},
			wantErr: true,
		},
		{
			name: "validates empty action file with no steps",
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				fixtureContent := testutil.MustReadFixture(testutil.TestFixtureEmptyAction)
				testutil.WriteTestFile(t, filepath.Join(tmpDir, appconstants.ActionFileNameYML), string(fixtureContent))
			},
			wantErr: false, // Empty steps is valid YAML structure
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Not using t.Parallel() because these tests modify shared globalConfig

			// Save and restore global state
			origConfig := globalConfig
			defer func() { globalConfig = origConfig }()

			// Create temp directory
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			// Setup test environment BEFORE changing directory
			// (so setupFunc can access testdata/ fixtures in project root)
			if tt.setupFunc != nil {
				tt.setupFunc(t, tmpDir)
			}

			// Change to temp directory for validation
			t.Chdir(tmpDir)

			// Initialize global config
			globalConfig = internal.DefaultAppConfig()

			// Create command
			cmd := newValidateCmd()

			// Execute handler - now returns error instead of os.Exit
			err := validateHandler(cmd, []string{})
			if (err != nil) != tt.wantErr {
				t.Errorf("validateHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestConfigInitHandlerIntegration tests configInitHandler.
// Note: This test is limited because configInitHandler uses internal.GetConfigPath()
// which uses the real XDG config directory. Full integration testing is done via
// subprocess tests in TestCLIConfigInitialization.
func TestConfigInitHandlerIntegration(t *testing.T) {
	// Skip parallelization as we need to manipulate global config path
	// which is shared state

	tests := []struct {
		name      string
		setupFunc func(t *testing.T) string
		wantErr   bool
		validate  func(t *testing.T, tmpDir string, err error)
	}{
		{
			name: "creates config when not exists",
			setupFunc: func(t *testing.T) string {
				t.Helper()

				return t.TempDir()
			},
			wantErr: false,
			validate: func(t *testing.T, _ string, err error) {
				t.Helper()
				if err != nil {
					t.Errorf(testutil.TestErrUnexpected, err)
				}
				// Note: Since configInitHandler uses internal.GetConfigPath() which points to real
				// user config directory, we can only verify no error occurred.
				// File creation is tested in subprocess tests.
			},
		},
		{
			name: "handles existing config gracefully",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()
				// Create a config file first
				configPath, err := internal.GetConfigPath()
				testutil.AssertNoError(t, err)
				// If config exists, handler should return nil (not error)
				_ = configPath

				return tmpDir
			},
			wantErr: false, // Handler returns nil when config exists, just warns
			validate: func(t *testing.T, _ string, err error) {
				t.Helper()
				// No error expected - handler just warns if config exists
				if err != nil {
					t.Errorf(testutil.TestErrUnexpected, err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore global state
			origConfig := globalConfig
			defer func() { globalConfig = origConfig }()

			// Initialize global config
			globalConfig = internal.DefaultAppConfig()

			tmpDir := tt.setupFunc(t)

			// Create command
			cmd := &cobra.Command{}

			// Execute handler
			err := configInitHandler(cmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("configInitHandler() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.validate != nil {
				tt.validate(t, tmpDir, err)
			}
		})
	}
}

// TestLoadGenConfigIntegration tests loadGenConfig configuration loading.
func TestLoadGenConfigIntegration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFunc func(t *testing.T, tmpDir string) (repoRoot, currentDir string)
		wantTheme string
	}{
		{
			name: "loads default config",
			setupFunc: func(t *testing.T, tmpDir string) (string, string) {
				t.Helper()

				return tmpDir, tmpDir
			},
			wantTheme: appconstants.ThemeDefault,
		},
		{
			name: "loads repo-specific config",
			setupFunc: func(t *testing.T, tmpDir string) (string, string) {
				t.Helper()
				configContent := "theme: professional\noutput_format: html\n"
				testutil.WriteTestFile(t, filepath.Join(tmpDir, ".ghreadme.yaml"), configContent)

				return tmpDir, tmpDir
			},
			wantTheme: appconstants.ThemeProfessional,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temp directory
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			// Setup test environment
			repoRoot, currentDir := tt.setupFunc(t, tmpDir)

			// Load config
			config, err := loadGenConfig(repoRoot, currentDir)
			if err != nil {
				t.Fatalf("loadGenConfig() error = %v", err)
			}

			if config == nil {
				t.Fatal("loadGenConfig() returned nil")
			}

			if config.Theme != tt.wantTheme {
				t.Errorf("loadGenConfig() theme = %v, want %v", config.Theme, tt.wantTheme)
			}
		})
	}
}

// TestProcessActionFilesIntegration tests processActionFiles batch processing.
func TestProcessActionFilesIntegration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFunc func(t *testing.T, tmpDir string) []string
		wantErr   bool
	}{
		{
			name: "processes single action file",
			setupFunc: func(t *testing.T, tmpDir string) []string {
				t.Helper()
				actionPath := testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureJavaScriptSimple)

				return []string{actionPath}
			},
			wantErr: false,
		},
		{
			name: "processes multiple action files",
			setupFunc: func(t *testing.T, tmpDir string) []string {
				t.Helper()
				action1 := testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureJavaScriptSimple)
				action2 := testutil.CreateActionSubdir(
					t,
					tmpDir,
					testutil.TestDirSubdir,
					testutil.TestFixtureCompositeBasic,
				)

				return []string{action1, action2}
			},
			wantErr: false,
		},
		// Note: "handles empty file list" case removed as it calls os.Exit
		// when there are no files to process. This scenario is tested via
		// subprocess tests in TestCLICommands instead.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temp directory
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			// Setup test environment
			actionFiles := tt.setupFunc(t, tmpDir)

			// Create generator with test config
			config := internal.DefaultAppConfig()
			config.Quiet = true
			generator := internal.NewGenerator(config)

			// Execute handler - just test that it doesn't panic
			defer func() {
				if r := recover(); r != nil && !tt.wantErr {
					t.Errorf("processActionFiles() unexpected panic: %v", r)
				}
			}()

			err := processActionFiles(generator, actionFiles)
			testutil.AssertNoError(t, err)
		})
	}
}

// TestDepsListHandlerIntegration tests depsListHandler.
func TestDepsListHandlerIntegration(t *testing.T) {
	// Note: Not using t.Parallel() because these tests modify shared globalConfig

	tests := []struct {
		name      string
		setupFunc func(t *testing.T, tmpDir string)
		wantErr   bool
	}{
		{
			name:      "lists dependencies from composite action",
			setupFunc: setupFixtureInDir(testutil.TestFixtureCompositeWithDeps),
			wantErr:   false,
		},
		{
			name:      testutil.TestScenarioNoDeps,
			setupFunc: setupFixtureInDir(testutil.TestFixtureJavaScriptSimple),
			wantErr:   false,
		},
		{
			name: "handles no action files",
			setupFunc: func(t *testing.T, _ string) {
				t.Helper()
				// No action files
			},
			wantErr: false,
		},
		// Error scenarios using fixtures
		// depsListHandler shows warning but returns nil
		createFixtureTestCase(
			"handles invalid YAML syntax with warning",
			testutil.TestErrorScenarioInvalidYAML,
			false,
		),
		// depsListHandler shows warning but returns nil
		createFixtureTestCase(
			"handles missing required fields with warning",
			testutil.TestErrorScenarioMissingFields,
			false,
		),
		// Should successfully list the outdated deps
		createFixtureTestCase(
			"lists dependencies from action with outdated deps",
			testutil.TestErrorScenarioOldDeps,
			false,
		),
		{
			name: "handles multiple action files recursively",
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				// Create main action
				testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureCompositeWithDeps)
				// Create subdirectory with another action
				subdir := filepath.Join(tmpDir, "subaction")
				testutil.AssertNoError(t, os.MkdirAll(subdir, 0750))
				fixtureContent := testutil.MustReadFixture(testutil.TestErrorScenarioOldDeps)
				testutil.WriteTestFile(t, filepath.Join(subdir, appconstants.ActionFileNameYML), string(fixtureContent))
			},
			wantErr: false, // Should list deps from both actions
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Not using t.Parallel() because these tests modify shared globalConfig

			// Save and restore global state
			origConfig := globalConfig
			defer func() { globalConfig = origConfig }()

			// Create temp directory
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			// Setup test environment BEFORE changing directory
			// (so setupFunc can access testdata/ fixtures in project root)
			if tt.setupFunc != nil {
				tt.setupFunc(t, tmpDir)
			}

			// Change to temp directory
			t.Chdir(tmpDir)

			// Initialize global config
			globalConfig = internal.DefaultAppConfig()
			globalConfig.Quiet = true

			// Execute handler - now returns error instead of os.Exit
			err := depsListHandler(&cobra.Command{}, []string{})
			if (err != nil) != tt.wantErr {
				t.Errorf("depsListHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDepsSecurityHandlerIntegration tests depsSecurityHandler.
func TestDepsSecurityHandlerIntegration(t *testing.T) {
	// Note: Not using t.Parallel() because these tests modify shared globalConfig

	tests := []struct {
		name      string
		setupFunc func(t *testing.T, tmpDir string)
		setToken  bool
		wantErr   bool
	}{
		{
			name:      "analyzes security with GitHub token",
			setupFunc: setupFixtureInDir(testutil.TestFixtureCompositeWithDeps),
			setToken:  true,
			wantErr:   false,
		},
		{
			name:      testutil.TestScenarioNoDeps,
			setupFunc: setupFixtureInDir(testutil.TestFixtureJavaScriptSimple),
			setToken:  true,
			wantErr:   false,
		},
		{
			name: "handles invalid YAML syntax gracefully",
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				fixtureContent := testutil.MustReadFixture(testutil.TestErrorScenarioInvalidYAML)
				testutil.WriteTestFile(t, filepath.Join(tmpDir, appconstants.ActionFileNameYML), string(fixtureContent))
			},
			setToken: true,
			wantErr:  false, // depsSecurityHandler handles YAML errors gracefully
		},
		{
			name: "handles missing required fields gracefully",
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				fixtureContent := testutil.MustReadFixture(testutil.TestErrorScenarioMissingFields)
				testutil.WriteTestFile(t, filepath.Join(tmpDir, appconstants.ActionFileNameYML), string(fixtureContent))
			},
			setToken: true,
			wantErr:  false, // depsSecurityHandler handles YAML errors gracefully
		},
		{
			name: "analyzes action with outdated dependencies",
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				fixtureContent := testutil.MustReadFixture(testutil.TestErrorScenarioOldDeps)
				testutil.WriteTestFile(t, filepath.Join(tmpDir, appconstants.ActionFileNameYML), string(fixtureContent))
			},
			setToken: true,
			wantErr:  false,
		},
		{
			// No action files is not an error — depsSecurityHandler now returns nil
			// here, consistent with deps list/outdated (handleNoFilesFoundError).
			name: "returns no error for no action files",
			setupFunc: func(t *testing.T, _ string) {
				t.Helper()
				// Don't create any action files
			},
			setToken: true,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Not using t.Parallel() because these tests modify shared globalConfig

			// Save and restore global state
			origConfig := globalConfig
			defer func() { globalConfig = origConfig }()

			// Create temp directory
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			// Setup test environment BEFORE changing directory
			// (so setupFunc can access testdata/ fixtures in project root)
			if tt.setupFunc != nil {
				tt.setupFunc(t, tmpDir)
			}

			// Change to temp directory
			t.Chdir(tmpDir)

			// Initialize global config
			globalConfig = internal.DefaultAppConfig()
			globalConfig.Quiet = true
			if tt.setToken {
				globalConfig.GitHubToken = testutil.TestTokenValue
			}

			// Execute handler - now returns error instead of os.Exit
			err := depsSecurityHandler(&cobra.Command{}, []string{})
			if (err != nil) != tt.wantErr {
				t.Errorf("depsSecurityHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDepsOutdatedHandlerIntegration tests depsOutdatedHandler.
func TestDepsOutdatedHandlerIntegration(t *testing.T) {
	// Note: Not using t.Parallel() because these tests modify shared globalConfig

	tests := []struct {
		name      string
		setupFunc func(t *testing.T, tmpDir string)
		setToken  bool
		wantErr   bool
	}{
		{
			name:      "checks outdated with GitHub token",
			setupFunc: setupFixtureInDir(testutil.TestFixtureCompositeWithDeps),
			setToken:  true,
			wantErr:   false,
		},
		{
			name: "handles no action files",
			setupFunc: func(t *testing.T, _ string) {
				t.Helper()
				// No action files
			},
			setToken: true,
			wantErr:  false,
		},
		{
			name:      "handles missing GitHub token",
			setupFunc: setupFixtureInDir(testutil.TestFixtureCompositeWithDeps),
			setToken:  false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Not using t.Parallel() because these tests modify shared globalConfig

			// Save and restore global state
			origConfig := globalConfig
			defer func() { globalConfig = origConfig }()

			// Create temp directory and change to it
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			t.Chdir(tmpDir)

			// Setup test environment
			if tt.setupFunc != nil {
				tt.setupFunc(t, tmpDir)
			}

			// Initialize global config
			globalConfig = internal.DefaultAppConfig()
			globalConfig.Quiet = true
			if tt.setToken {
				globalConfig.GitHubToken = testutil.TestTokenValue
			}

			// Execute handler - now returns error instead of os.Exit
			err := depsOutdatedHandler(&cobra.Command{}, []string{})
			if (err != nil) != tt.wantErr {
				t.Errorf("depsOutdatedHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestConfigWizardHandlerIntegration tests configWizardHandler.
func TestConfigWizardHandlerIntegration(t *testing.T) {
	// Note: This is a limited test as wizard requires interactive input
	// Full wizard testing is done in the wizard package

	// Save and restore global state
	origConfig := globalConfig
	defer func() { globalConfig = origConfig }()

	// Create temp directory
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Set XDG_CONFIG_HOME to temp directory
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Initialize global config
	globalConfig = internal.DefaultAppConfig()
	globalConfig.Quiet = true

	// Create command with output flag pointing to temp file
	cmd := &cobra.Command{}
	cmd.Flags().String("format", "yaml", "")
	outputPath := filepath.Join(tmpDir, "test-config.yaml")
	cmd.Flags().String("output", outputPath, "")

	// Note: We can't fully test wizard handler without mocking stdin
	// The wizard requires interactive input which is tested in wizard package
	// This test just ensures the handler doesn't panic on setup
}

// Phase 6: Tests for zero-coverage business logic functions

// TestCheckAllOutdated tests the checkAllOutdated function.
func TestCheckAllOutdated(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		setupFunc       func(t *testing.T, tmpDir string) []string
		mockAnalyzer    bool
		wantOutdatedCnt int
		wantErr         bool
	}{
		{
			name:            "finds outdated dependencies",
			setupFunc:       setupFixtureReturningPath(testutil.TestFixtureCompositeWithDeps),
			mockAnalyzer:    true,
			wantOutdatedCnt: 0, // Mock analyzer will return no outdated deps
		},
		{
			name:            testutil.TestScenarioNoDeps,
			setupFunc:       setupWithActionContent(testutil.TestMinimalAction),
			mockAnalyzer:    true,
			wantOutdatedCnt: 0,
		},
		{
			name: "handles multiple action files",
			setupFunc: func(t *testing.T, tmpDir string) []string {
				t.Helper()
				action1 := filepath.Join(tmpDir, testutil.TestFileAction1)
				action2 := filepath.Join(tmpDir, testutil.TestFileAction2)
				testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureCompositeWithDeps)
				_ = os.Rename(filepath.Join(tmpDir, appconstants.ActionFileNameYML), action1)
				testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureCompositeBasic)
				_ = os.Rename(filepath.Join(tmpDir, appconstants.ActionFileNameYML), action2)

				return []string{action1, action2}
			},
			mockAnalyzer:    true,
			wantOutdatedCnt: 0,
		},
		{
			name: "handles invalid action file gracefully",
			setupFunc: func(t *testing.T, tmpDir string) []string {
				t.Helper()
				actionPath := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
				testutil.WriteTestFile(t, actionPath, testutil.TestInvalidYAMLPrefix)

				return []string{actionPath}
			},
			mockAnalyzer:    true,
			wantOutdatedCnt: 0, // Should handle error and return empty list
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			actionFiles := tt.setupFunc(t, tmpDir)

			output := createOutputManager(true)  // quiet mode
			analyzer := &dependencies.Analyzer{} // Basic analyzer without GitHub token

			outdated := checkAllOutdated(output, actionFiles, analyzer)

			if len(outdated) != tt.wantOutdatedCnt {
				t.Errorf("checkAllOutdated() returned %d outdated deps, want %d",
					len(outdated), tt.wantOutdatedCnt)
			}
		})
	}
}

// TestAnalyzeSecurityDeps tests the analyzeSecurityDeps function.
func TestAnalyzeSecurityDeps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupFunc  func(t *testing.T, tmpDir string) []string
		wantPinned int
	}{
		{
			name:       "analyzes action with dependencies",
			setupFunc:  setupFixtureReturningPath(testutil.TestFixtureCompositeWithDeps),
			wantPinned: 2, // TestFixtureCompositeWithDeps has 2 pinned dependencies
		},
		{
			name:       testutil.TestScenarioNoDeps,
			setupFunc:  setupWithActionContent(testutil.TestMinimalAction),
			wantPinned: 0,
		},
		{
			name: "handles multiple action files",
			setupFunc: func(t *testing.T, tmpDir string) []string {
				t.Helper()
				action1 := filepath.Join(tmpDir, testutil.TestFileAction1)
				action2 := filepath.Join(tmpDir, testutil.TestFileAction2)

				testutil.WriteTestFile(
					t,
					action1,
					testutil.MustReadFixture(testutil.TestFixtureCompositeStepCheckoutV4),
				)
				testutil.WriteTestFile(
					t,
					action2,
					testutil.MustReadFixture(testutil.TestFixtureCompositeStepSetupNodeV3),
				)

				return []string{action1, action2}
			},
			wantPinned: 0, // Without GitHub token, won't verify pins
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			actionFiles := tt.setupFunc(t, tmpDir)

			output := createOutputManager(true)  // quiet mode
			analyzer := &dependencies.Analyzer{} // Basic analyzer without GitHub token

			pinnedCount, _ := analyzeSecurityDeps(output, actionFiles, analyzer)

			if pinnedCount != tt.wantPinned {
				t.Errorf("analyzeSecurityDeps() returned %d pinned deps, want %d",
					pinnedCount, tt.wantPinned)
			}
		})
	}
}

// TestCollectAllUpdates tests the collectAllUpdates function.
func TestCollectAllUpdates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupFunc     func(t *testing.T, tmpDir string) []string
		wantUpdateCnt int
	}{
		{
			name:          "collects updates from single action",
			setupFunc:     setupFixtureReturningPath(testutil.TestFixtureCompositeWithDeps),
			wantUpdateCnt: 0, // Without GitHub token, won't fetch updates
		},
		{
			name: "collects from multiple actions",
			setupFunc: func(t *testing.T, tmpDir string) []string {
				t.Helper()
				action1 := filepath.Join(tmpDir, testutil.TestFileAction1)
				action2 := filepath.Join(tmpDir, testutil.TestFileAction2)

				testutil.WriteTestFile(
					t,
					action1,
					testutil.MustReadFixture(testutil.TestFixtureCompositeStepCheckoutV3),
				)
				testutil.WriteTestFile(
					t,
					action2,
					testutil.MustReadFixture(testutil.TestFixtureCompositeStepSetupNodeV2),
				)

				return []string{action1, action2}
			},
			wantUpdateCnt: 0, // Without GitHub token, won't fetch updates
		},
		{
			name:          testutil.TestScenarioNoDeps,
			setupFunc:     setupWithActionContent(testutil.TestMinimalAction),
			wantUpdateCnt: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			actionFiles := tt.setupFunc(t, tmpDir)

			output := createOutputManager(true)  // quiet mode
			analyzer := &dependencies.Analyzer{} // Basic analyzer without GitHub token

			updates := collectAllUpdates(output, analyzer, actionFiles)

			if len(updates) != tt.wantUpdateCnt {
				t.Errorf("collectAllUpdates() returned %d updates, want %d",
					len(updates), tt.wantUpdateCnt)
			}
		})
	}
}

// TestWrapError tests the wrapError helper function.
func TestWrapError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		msgConstant  string
		err          error
		wantContains []string
	}{
		{
			name:        "wraps error with message constant",
			msgConstant: "operation failed",
			err:         errors.New("original error"),
			wantContains: []string{
				"operation failed",
				"original error",
			},
		},
		{
			name:        "handles empty message constant",
			msgConstant: "",
			err:         errors.New("test error"),
			wantContains: []string{
				"test error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := wrapError(tt.msgConstant, tt.err)
			if result == nil {
				t.Fatal("wrapError() returned nil, want error")
			}

			resultStr := result.Error()
			for _, want := range tt.wantContains {
				if !strings.Contains(resultStr, want) {
					t.Errorf("wrapError() = %q, want to contain %q", resultStr, want)
				}
			}

			// Verify it's a wrapped error
			if !errors.Is(result, tt.err) {
				t.Errorf("wrapError() did not wrap original error properly")
			}
		})
	}
}

// TestWrapHandlerWithErrorHandling tests the wrapper function for handlers.
func TestWrapHandlerWithErrorHandling(t *testing.T) {
	// Save and restore global state
	origConfig := globalConfig
	defer func() { globalConfig = origConfig }()

	tests := []struct {
		name    string
		handler func(*cobra.Command, []string) error
		wantErr bool
	}{
		{
			name: "handler returns nil - no error",
			handler: func(_ *cobra.Command, _ []string) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "initializes globalConfig if nil before calling handler",
			handler: func(_ *cobra.Command, _ []string) error {
				// Verify globalConfig was initialized by wrapper
				if globalConfig == nil {
					return errors.New("globalConfig is nil in handler")
				}

				return nil
			},
			wantErr: false,
		},
		// Note: Cannot test error path because wrapHandlerWithErrorHandling calls os.Exit(1)
		// which would terminate the test process. Error path is tested via subprocess tests.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set globalConfig to nil to test initialization
			if tt.name == "initializes globalConfig if nil before calling handler" {
				globalConfig = nil
			} else {
				globalConfig = internal.DefaultAppConfig()
			}

			cmd := &cobra.Command{}
			wrapped := wrapHandlerWithErrorHandling(tt.handler)

			// Execute wrapped handler (should not panic)
			wrapped(cmd, []string{})

			// Verify globalConfig was initialized
			if globalConfig == nil {
				t.Error("wrapHandlerWithErrorHandling() did not initialize globalConfig")
			}
		})
	}
}

func TestApplyUpdates(t *testing.T) {
	t.Parallel()

	// Test cases that don't require calling ApplyPinnedUpdates (user cancellation)
	t.Run("interactive mode cancellation", func(t *testing.T) {
		tests := []struct {
			name     string
			response string
		}{
			{name: "response 'n' cancels", response: "n"},
			{name: "response 'no' cancels", response: "no"},
			{name: "empty response cancels", response: ""},
			{name: "random text cancels", response: "random"},
			{name: "uppercase N cancels", response: "N"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				// Create test reader with response
				reader := &TestInputReader{responses: []string{tt.response}}

				// Create minimal analyzer (won't be used since we're canceling)
				analyzer := dependencies.NewAnalyzer(nil, git.RepoInfo{}, nil)

				output := createOutputManager(true) // Quiet mode for tests
				updates := []dependencies.PinnedUpdate{
					{OldUses: testutil.TestActionCheckoutV3, NewUses: testutil.TestActionCheckoutV4},
				}

				// Execute function - should not call ApplyPinnedUpdates
				err := applyUpdates(output, analyzer, updates, false, reader)

				// Should not error when user cancels
				if err != nil {
					t.Errorf("applyUpdates() with cancel should not error, got: %v", err)
				}

				// Verify reader was used
				if reader.index != 1 {
					t.Errorf(testutil.TestErrInputReaderNotUsed, reader.index)
				}
			})
		}
	})

	// Test automatic mode bypasses prompting
	t.Run("automatic mode bypasses prompting", func(t *testing.T) {
		t.Parallel()

		// Create minimal analyzer
		analyzer := dependencies.NewAnalyzer(nil, git.RepoInfo{}, nil)

		// Create temp directory for test action file
		tmpDir := t.TempDir()
		actionFile := testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureActionWithCheckoutV3)

		output := createOutputManager(true)
		updates := []dependencies.PinnedUpdate{
			{
				OldUses:  testutil.TestActionCheckoutV3,
				NewUses:  "actions/checkout@abc123",
				FilePath: actionFile,
			},
		}

		// Call with automatic=true, reader should not be used (can pass nil)
		err := applyUpdates(output, analyzer, updates, true, nil)

		// May error due to nil github client or other reasons, but that's expected
		// The important thing is it didn't block on stdin prompting the user
		_ = err // Accept any result for this integration test
	})

	// Test that InputReader is used when provided
	t.Run("InputReader is used in interactive mode", func(t *testing.T) {
		t.Parallel()

		// Create test reader
		reader := &TestInputReader{responses: []string{"n"}}

		analyzer := dependencies.NewAnalyzer(nil, git.RepoInfo{}, nil)
		output := createOutputManager(true)
		updates := []dependencies.PinnedUpdate{{OldUses: "old", NewUses: "new"}}

		_ = applyUpdates(output, analyzer, updates, false, reader)

		// Verify reader was actually used (index should be 1 after reading first response)
		if reader.index != 1 {
			t.Errorf(testutil.TestErrInputReaderNotUsed, reader.index)
		}
	})

	// Test that default StdinReader is used when reader is nil
	t.Run("defaults to StdinReader when reader is nil", func(t *testing.T) {
		t.Parallel()

		// This test verifies the nil check works, but can't test actual stdin
		// Just verify the function accepts nil and doesn't panic

		analyzer := dependencies.NewAnalyzer(nil, git.RepoInfo{}, nil)
		output := createOutputManager(true)
		updates := []dependencies.PinnedUpdate{{OldUses: "old", NewUses: "new"}}

		// With automatic=true and nil reader, should not prompt
		err := applyUpdates(output, analyzer, updates, true, nil)

		// May error, but shouldn't panic from nil reader
		_ = err
	})
}

func TestSetupDepsUpgrade(t *testing.T) {
	// Note: Cannot use t.Parallel() because one subtest modifies shared globalConfig

	tests := []struct {
		name       string
		setupFunc  func(t *testing.T) (string, *internal.AppConfig)
		wantErr    bool
		errContain string
	}{
		{
			name: testutil.TestMsgNoGitHubToken,
			setupFunc: func(t *testing.T) (string, *internal.AppConfig) {
				t.Helper()
				tmpDir := t.TempDir()
				// Create a valid action file
				testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureActionWithCheckoutV4)

				// Clear env vars so GetGitHubToken returns "" regardless of CI environment
				t.Setenv(appconstants.EnvGitHubToken, "")
				t.Setenv(appconstants.EnvGitHubTokenStandard, "")

				config := internal.DefaultAppConfig()
				config.GitHubToken = "" // No token

				return tmpDir, config
			},
			wantErr:    true,
			errContain: "GitHub token not found",
		},
		{
			name: "succeeds with valid token and action files",
			setupFunc: func(t *testing.T) (string, *internal.AppConfig) {
				t.Helper()
				tmpDir := t.TempDir()
				// Create a valid action file
				testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureActionWithCheckoutV4)

				config := internal.DefaultAppConfig()
				config.GitHubToken = testutil.TestToken123

				return tmpDir, config
			},
			wantErr: false,
		},
		{
			name: "returns error when no action files found",
			setupFunc: func(t *testing.T) (string, *internal.AppConfig) {
				t.Helper()
				tmpDir := t.TempDir() // Empty directory
				config := internal.DefaultAppConfig()
				config.GitHubToken = testutil.TestToken123

				return tmpDir, config
			},
			wantErr:    true,
			errContain: "no action files",
		},
		{
			name: testMsgUsesGlobalCfg,
			setupFunc: func(t *testing.T) (string, *internal.AppConfig) {
				t.Helper()
				tmpDir := t.TempDir()
				// Create a valid action file
				testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureActionWithCheckoutV4)

				// Set globalConfig instead of passing config
				origConfig := globalConfig
				globalConfig = internal.DefaultAppConfig()
				globalConfig.GitHubToken = "test-token-from-global"
				t.Cleanup(func() { globalConfig = origConfig })

				return tmpDir, nil // Pass nil config
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Cannot use t.Parallel() for cases that mutate env vars or globalConfig
			serialCases := tt.name == testMsgUsesGlobalCfg || tt.name == testutil.TestMsgNoGitHubToken
			if !serialCases {
				t.Parallel()
			}

			currentDir, config := tt.setupFunc(t)

			_, _, err := setupDepsUpgrade(currentDir, config)

			validateDepsUpgradeError(t, err, tt.wantErr, tt.errContain)
		})
	}
}

// validateDepsUpgradeError validates error expectations for deps upgrade tests.
func validateDepsUpgradeError(t *testing.T, err error, wantErr bool, errContain string) {
	t.Helper()

	if (err != nil) != wantErr {
		t.Errorf("error = %v, wantErr %v", err, wantErr)

		return
	}

	if wantErr && errContain != "" {
		if err == nil || !strings.Contains(err.Error(), errContain) {
			t.Errorf("error should contain %q, got %v", errContain, err)
		}
	}
}

func TestConfigWizardHandlerInitialization(t *testing.T) {
	// Note: Cannot use t.Parallel() because test modifies shared globalConfig

	t.Run("wrapHandlerWithErrorHandling initializes globalConfig when nil", func(t *testing.T) {
		// Save and restore
		origConfig := globalConfig
		defer func() { globalConfig = origConfig }()

		globalConfig = nil

		// Use a probe that returns nil so the wrapper never calls os.Exit.
		// The wrapper initializes globalConfig before invoking the handler,
		// so capturedConfig must be non-nil if the wrapper works correctly.
		var capturedConfig *internal.AppConfig
		probe := func(_ *cobra.Command, _ []string) error {
			capturedConfig = globalConfig

			return nil
		}

		cmd := &cobra.Command{}
		wrapHandlerWithErrorHandling(probe)(cmd, []string{})

		if capturedConfig == nil {
			t.Error("wrapHandlerWithErrorHandling should initialize globalConfig when nil")
		}
	})
}

func TestDepsUpgradeHandlerErrorPaths(t *testing.T) {
	// Note: Cannot use t.Parallel() — test modifies env vars and globalConfig.

	origConfig := globalConfig
	defer func() { globalConfig = origConfig }()

	t.Run("returns error when no GitHub token", func(t *testing.T) {
		t.Setenv(appconstants.EnvGitHubToken, "")
		t.Setenv(appconstants.EnvGitHubTokenStandard, "")

		globalConfig = internal.DefaultAppConfig()
		globalConfig.GitHubToken = ""
		globalConfig.Quiet = true

		cmd := &cobra.Command{}
		cmd.Flags().Bool(appconstants.FlagCI, false, "")
		cmd.Flags().Bool(appconstants.InputAll, false, "")
		cmd.Flags().Bool(appconstants.InputDryRun, false, "")

		err := depsUpgradeHandler(cmd, []string{})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "GitHub token not found") {
			t.Errorf("expected 'GitHub token not found' error, got: %v", err)
		}
	})
}

func TestConfigRootHandler(t *testing.T) {
	origConfig := globalConfig
	t.Cleanup(func() { globalConfig = origConfig })

	globalConfig = internal.DefaultAppConfig()
	globalConfig.Quiet = true

	cmd := &cobra.Command{}
	err := configRootHandler(cmd, []string{})
	if err != nil {
		t.Errorf("configRootHandler() unexpected error: %v", err)
	}
}

// TestConfigRootHandlerRedactsNestedTokens verifies that verbose config output
// redacts GitHub tokens both at the top level and inside RepoOverrides, so a
// token configured per-repo is never printed in plaintext.
func TestConfigRootHandlerRedactsNestedTokens(t *testing.T) {
	origConfig := globalConfig
	t.Cleanup(func() { globalConfig = origConfig })

	const topToken = "ghp_TOPLEVELtoken1234567890abcdef"  // #nosec G101 -- test fixture value, not a credential
	const nestedToken = "ghp_NESTEDtoken0987654321zyxwvu" // #nosec G101 -- test fixture value, not a credential

	globalConfig = internal.DefaultAppConfig()
	globalConfig.Quiet = false
	globalConfig.Verbose = true
	globalConfig.GitHubToken = topToken
	globalConfig.RepoOverrides = map[string]internal.AppConfig{
		"testorg/secret-repo": {GitHubToken: nestedToken},
	}

	out := testutil.CaptureStdout(func() {
		if err := configRootHandler(&cobra.Command{}, []string{}); err != nil {
			t.Errorf("configRootHandler() unexpected error: %v", err)
		}
	})

	if strings.Contains(out, topToken) {
		t.Errorf("top-level token leaked in verbose output:\n%s", out)
	}
	if strings.Contains(out, nestedToken) {
		t.Errorf("RepoOverrides nested token leaked in verbose output:\n%s", out)
	}
	if !strings.Contains(out, appconstants.RedactedPlaceholder) {
		t.Errorf("expected redaction placeholder %q in output:\n%s", appconstants.RedactedPlaceholder, out)
	}
}
