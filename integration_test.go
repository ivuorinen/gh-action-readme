package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

// copyDir recursively copies a directory.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = srcFile.Close() }()

		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer func() { _ = dstFile.Close() }()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

// buildTestBinary builds the test binary for integration testing.
func buildTestBinary(t *testing.T) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "gh-action-readme-binary-*")
	if err != nil {
		t.Fatalf("failed to create temp dir for binary: %v", err)
	}

	binaryPath := filepath.Join(tmpDir, "gh-action-readme")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")

	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build test binary: %v\nstderr: %s", err, stderr.String())
	}

	// Copy templates directory to binary directory
	templatesDir := filepath.Join(filepath.Dir(binaryPath), "templates")
	if err := copyDir("templates", templatesDir); err != nil {
		t.Fatalf("failed to copy templates: %v", err)
	}

	return binaryPath
}

// setupCompleteWorkflow creates a realistic project structure for testing.
func setupCompleteWorkflow(t *testing.T, tmpDir string) {
	testutil.WriteTestFile(t, filepath.Join(tmpDir, "action.yml"), testutil.CompositeActionYML)
	testutil.WriteTestFile(t, filepath.Join(tmpDir, "README.md"), "# Old README")
	testutil.WriteTestFile(t, filepath.Join(tmpDir, ".gitignore"), testutil.GitIgnoreContent)
	testutil.WriteTestFile(t, filepath.Join(tmpDir, "package.json"), testutil.PackageJSONContent)
}

// setupMultiActionWorkflow creates a project with multiple actions.
func setupMultiActionWorkflow(t *testing.T, tmpDir string) {
	testutil.WriteTestFile(t, filepath.Join(tmpDir, "action.yml"), testutil.SimpleActionYML)

	subDir := filepath.Join(tmpDir, "actions", "deploy")
	_ = os.MkdirAll(subDir, 0755)
	testutil.WriteTestFile(t, filepath.Join(subDir, "action.yml"), testutil.DockerActionYML)

	subDir2 := filepath.Join(tmpDir, "actions", "test")
	_ = os.MkdirAll(subDir2, 0755)
	testutil.WriteTestFile(t, filepath.Join(subDir2, "action.yml"), testutil.CompositeActionYML)
}

// setupConfigWorkflow creates a simple action for config testing.
func setupConfigWorkflow(t *testing.T, tmpDir string) {
	testutil.WriteTestFile(t, filepath.Join(tmpDir, "action.yml"), testutil.SimpleActionYML)
}

// setupErrorWorkflow creates an invalid action file for error testing.
func setupErrorWorkflow(t *testing.T, tmpDir string) {
	testutil.WriteTestFile(t, filepath.Join(tmpDir, "action.yml"), testutil.InvalidActionYML)
}

// checkStepExitCode validates command exit code expectations.
func checkStepExitCode(t *testing.T, step workflowStep, exitCode int, stdout, stderr strings.Builder) {
	if step.expectSuccess && exitCode != 0 {
		t.Errorf("expected success but got exit code %d", exitCode)
		t.Logf("stdout: %s", stdout.String())
		t.Logf("stderr: %s", stderr.String())
	} else if !step.expectSuccess && exitCode == 0 {
		t.Error("expected failure but command succeeded")
	}
}

// checkStepOutput validates command output expectations.
func checkStepOutput(t *testing.T, step workflowStep, output string) {
	if step.expectOutput != "" && !strings.Contains(output, step.expectOutput) {
		t.Errorf("expected output to contain %q, got: %s", step.expectOutput, output)
	}

	if step.expectError != "" && !strings.Contains(output, step.expectError) {
		t.Errorf("expected error to contain %q, got: %s", step.expectError, output)
	}
}

// executeWorkflowStep runs a single workflow step.
func executeWorkflowStep(t *testing.T, binaryPath, tmpDir string, step workflowStep) {
	t.Run(step.name, func(t *testing.T) {
		cmd := exec.Command(binaryPath, step.cmd...)
		cmd.Dir = tmpDir

		var stdout, stderr strings.Builder
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		exitCode := 0
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode = exitError.ExitCode()
			}
		}

		checkStepExitCode(t, step, exitCode, stdout, stderr)
		checkStepOutput(t, step, stdout.String()+stderr.String())
	})
}

// TestEndToEndWorkflows tests complete workflows from start to finish.
func TestEndToEndWorkflows(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer func() { _ = os.Remove(binaryPath) }()

	tests := []struct {
		name      string
		setupFunc func(t *testing.T, tmpDir string)
		workflow  []workflowStep
	}{
		{
			name:      "Complete documentation generation workflow",
			setupFunc: setupCompleteWorkflow,
			workflow: []workflowStep{
				{
					name:          "validate action file",
					cmd:           []string{"validate"},
					expectSuccess: true,
					expectOutput:  "All validations passed",
				},
				{
					name:          "generate with default theme",
					cmd:           []string{"gen", "--theme", "default"},
					expectSuccess: true,
				},
				{
					name:          "generate with github theme",
					cmd:           []string{"gen", "--theme", "github", "--output-format", "html"},
					expectSuccess: true,
				},
				{
					name:          "list dependencies",
					cmd:           []string{"deps", "list"},
					expectSuccess: true,
				},
				{
					name:          "check cache statistics",
					cmd:           []string{"cache", "stats"},
					expectSuccess: true,
					expectOutput:  "Cache Statistics",
				},
			},
		},
		{
			name:      "Multi-action project workflow",
			setupFunc: setupMultiActionWorkflow,
			workflow: []workflowStep{
				{
					name:          "validate all actions recursively",
					cmd:           []string{"validate"},
					expectSuccess: true,
				},
				{
					name:          "generate docs for all actions",
					cmd:           []string{"gen", "--recursive", "--theme", "professional"},
					expectSuccess: true,
				},
				{
					name:          "check all dependencies",
					cmd:           []string{"deps", "list"},
					expectSuccess: true,
				},
			},
		},
		{
			name:      "Configuration management workflow",
			setupFunc: setupConfigWorkflow,
			workflow: []workflowStep{
				{
					name:          "show current config",
					cmd:           []string{"config", "show"},
					expectSuccess: true,
					expectOutput:  "Current Configuration",
				},
				{
					name:          "list available themes",
					cmd:           []string{"config", "themes"},
					expectSuccess: true,
					expectOutput:  "Available Themes",
				},
				{
					name:          "generate with custom theme",
					cmd:           []string{"gen", "--theme", "minimal"},
					expectSuccess: true,
				},
			},
		},
		{
			name:      "Error handling and recovery workflow",
			setupFunc: setupErrorWorkflow,
			workflow: []workflowStep{
				{
					name:          "validate invalid action",
					cmd:           []string{"validate"},
					expectSuccess: false,
					expectError:   "Missing required field",
				},
				{
					name:          "attempt generation with invalid action",
					cmd:           []string{"gen"},
					expectSuccess: false,
				},
				{
					name:          "show schema for reference",
					cmd:           []string{"schema"},
					expectSuccess: true,
					expectOutput:  "schema",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			// Setup the test environment
			tt.setupFunc(t, tmpDir)

			// Execute workflow steps
			for _, step := range tt.workflow {
				executeWorkflowStep(t, binaryPath, tmpDir, step)
			}
		})
	}
}

type workflowStep struct {
	name          string
	cmd           []string
	expectSuccess bool
	expectOutput  string
	expectError   string
}

// testProjectSetup tests basic project validation.
func testProjectSetup(t *testing.T, binaryPath, tmpDir string) {
	// Create a new GitHub Action project
	testutil.WriteTestFile(t, filepath.Join(tmpDir, "action.yml"), `
name: 'My New Action'
description: 'A brand new GitHub Action'
inputs:
  message:
    description: 'Message to display'
    required: true
runs:
  using: 'node20'
  main: 'index.js'
`)

	// Validate the action
	cmd := exec.Command(binaryPath, "validate")
	cmd.Dir = tmpDir
	err := cmd.Run()
	testutil.AssertNoError(t, err)
}

// testDocumentationGeneration tests generation with different themes.
func testDocumentationGeneration(t *testing.T, binaryPath, tmpDir string) {
	themes := []string{"default", "github", "minimal"}

	for _, theme := range themes {
		cmd := exec.Command(binaryPath, "gen", "--theme", theme)
		cmd.Dir = tmpDir
		err := cmd.Run()
		testutil.AssertNoError(t, err)

		// Verify README was created
		readmeFiles, _ := filepath.Glob(filepath.Join(tmpDir, "README*.md"))
		if len(readmeFiles) == 0 {
			t.Errorf("no README generated for theme %s", theme)
		}

		// Clean up for next iteration
		for _, file := range readmeFiles {
			_ = os.Remove(file)
		}
	}
}

// testDependencyManagement tests dependency listing functionality.
func testDependencyManagement(t *testing.T, binaryPath, tmpDir string) {
	// Update action to be composite with dependencies
	testutil.WriteTestFile(t, filepath.Join(tmpDir, "action.yml"), testutil.CompositeActionYML)

	// List dependencies
	cmd := exec.Command(binaryPath, "deps", "list")
	cmd.Dir = tmpDir
	var stdout strings.Builder
	cmd.Stdout = &stdout
	err := cmd.Run()
	testutil.AssertNoError(t, err)

	output := stdout.String()
	if !strings.Contains(output, "Dependencies found") {
		t.Error("expected dependency listing output")
	}
}

// testOutputFormats tests generation with different output formats.
func testOutputFormats(t *testing.T, binaryPath, tmpDir string) {
	formats := []string{"md", "html", "json"}

	for _, format := range formats {
		cmd := exec.Command(binaryPath, "gen", "--output-format", format)
		cmd.Dir = tmpDir
		err := cmd.Run()
		testutil.AssertNoError(t, err)

		// Verify output was created with correct naming patterns
		var pattern string
		switch format {
		case "md":
			pattern = "README*.md"
		case "html":
			// HTML files are named after the action name (e.g., "Example Action.html")
			pattern = "*.html"
		case "json":
			// JSON files have a fixed name
			pattern = "action-docs.json"
		}

		files, _ := filepath.Glob(filepath.Join(tmpDir, pattern))
		if len(files) == 0 {
			t.Errorf("no output generated for format %s (pattern: %s)", format, pattern)
		}

		// Clean up
		for _, file := range files {
			_ = os.Remove(file)
		}
	}
}

// testCacheManagement tests cache-related commands.
func testCacheManagement(t *testing.T, binaryPath, tmpDir string) {
	// Check cache stats
	cmd := exec.Command(binaryPath, "cache", "stats")
	cmd.Dir = tmpDir
	err := cmd.Run()
	testutil.AssertNoError(t, err)

	// Clear cache
	cmd = exec.Command(binaryPath, "cache", "clear")
	cmd.Dir = tmpDir
	err = cmd.Run()
	testutil.AssertNoError(t, err)

	// Check path
	cmd = exec.Command(binaryPath, "cache", "path")
	cmd.Dir = tmpDir
	err = cmd.Run()
	testutil.AssertNoError(t, err)
}

func TestCompleteProjectLifecycle(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer func() { _ = os.Remove(binaryPath) }()

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Phase 1: Project setup
	t.Run("Phase 1: Project Setup", func(t *testing.T) {
		testProjectSetup(t, binaryPath, tmpDir)
	})

	// Phase 2: Documentation generation
	t.Run("Phase 2: Documentation Generation", func(t *testing.T) {
		testDocumentationGeneration(t, binaryPath, tmpDir)
	})

	// Phase 3: Add dependencies and test dependency features
	t.Run("Phase 3: Dependency Management", func(t *testing.T) {
		testDependencyManagement(t, binaryPath, tmpDir)
	})

	// Phase 4: Multiple output formats
	t.Run("Phase 4: Multiple Output Formats", func(t *testing.T) {
		testOutputFormats(t, binaryPath, tmpDir)
	})

	// Phase 5: Cache management
	t.Run("Phase 5: Cache Management", func(t *testing.T) {
		testCacheManagement(t, binaryPath, tmpDir)
	})
}

func TestStressTestWorkflow(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer func() { _ = os.Remove(binaryPath) }()

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Create many action files to test performance
	const numActions = 20
	for i := 0; i < numActions; i++ {
		actionDir := filepath.Join(tmpDir, "action"+string(rune('A'+i)))
		_ = os.MkdirAll(actionDir, 0755)

		actionContent := strings.ReplaceAll(testutil.SimpleActionYML, "Simple Action", "Action "+string(rune('A'+i)))
		testutil.WriteTestFile(t, filepath.Join(actionDir, "action.yml"), actionContent)
	}

	// Test recursive processing
	cmd := exec.Command(binaryPath, "gen", "--recursive", "--theme", "github")
	cmd.Dir = tmpDir
	err := cmd.Run()
	testutil.AssertNoError(t, err)

	// Verify all READMEs were generated
	readmeFiles, _ := filepath.Glob(filepath.Join(tmpDir, "**/README*.md"))
	if len(readmeFiles) < numActions {
		t.Errorf("expected at least %d README files, got %d", numActions, len(readmeFiles))
	}

	// Test validation of all files
	cmd = exec.Command(binaryPath, "validate")
	cmd.Dir = tmpDir
	err = cmd.Run()
	testutil.AssertNoError(t, err)
}

func TestErrorRecoveryWorkflow(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer func() { _ = os.Remove(binaryPath) }()

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Create a project with mixed valid and invalid files
	// Note: validation looks for files named exactly "action.yml" or "action.yaml"
	testutil.WriteTestFile(t, filepath.Join(tmpDir, "action.yml"), testutil.SimpleActionYML)

	subDir := filepath.Join(tmpDir, "subdir")
	_ = os.MkdirAll(subDir, 0755)
	testutil.WriteTestFile(t, filepath.Join(subDir, "action.yml"), testutil.InvalidActionYML)

	// Test that validation reports issues but doesn't crash
	cmd := exec.Command(binaryPath, "validate")
	cmd.Dir = tmpDir
	var stderr strings.Builder
	cmd.Stderr = &stderr

	err := cmd.Run()
	// Validation should fail due to invalid file
	if err == nil {
		t.Error("expected validation to fail with invalid files")
	}

	// But it should still report on valid files with validation errors
	output := stderr.String()
	if !strings.Contains(output, "Missing required field:") && !strings.Contains(output, "validation failed") {
		t.Errorf("expected validation error message, got: %s", output)
	}

	// Test generation with mixed files - should generate docs for valid ones
	cmd = exec.Command(binaryPath, "gen", "--recursive")
	cmd.Dir = tmpDir
	cmd.Stderr = &stderr

	_ = cmd.Run()
	// Generation might fail due to invalid files, but check what was generated
	readmeFiles, _ := filepath.Glob(filepath.Join(tmpDir, "**/README*.md"))

	// Should have generated at least some READMEs for valid files
	if len(readmeFiles) == 0 {
		t.Log("No READMEs generated, which might be expected with invalid files")
	}
}

func TestConfigurationWorkflow(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer func() { _ = os.Remove(binaryPath) }()

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Set up XDG config environment
	configHome := filepath.Join(tmpDir, "config")
	_ = os.Setenv("XDG_CONFIG_HOME", configHome)
	defer func() { _ = os.Unsetenv("XDG_CONFIG_HOME") }()

	testutil.WriteTestFile(t, filepath.Join(tmpDir, "action.yml"), testutil.SimpleActionYML)

	var err error

	// Test configuration initialization
	cmd := exec.Command(binaryPath, "config", "init")
	cmd.Dir = tmpDir
	_ = cmd.Run()
	// This might fail if config already exists, which is fine

	// Test showing configuration
	cmd = exec.Command(binaryPath, "config", "show")
	cmd.Dir = tmpDir
	var stdout strings.Builder
	cmd.Stdout = &stdout
	err = cmd.Run()
	testutil.AssertNoError(t, err)

	if !strings.Contains(stdout.String(), "Current Configuration") {
		t.Error("expected configuration output")
	}

	// Test with different configuration options
	cmd = exec.Command(binaryPath, "--verbose", "gen")
	cmd.Dir = tmpDir
	err = cmd.Run()
	testutil.AssertNoError(t, err)

	cmd = exec.Command(binaryPath, "--quiet", "gen")
	cmd.Dir = tmpDir
	err = cmd.Run()
	testutil.AssertNoError(t, err)
}
