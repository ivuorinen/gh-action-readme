package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

func TestGenerator_NewGenerator(t *testing.T) {
	config := &AppConfig{
		Theme:        "default",
		OutputFormat: "md",
		OutputDir:    ".",
		Verbose:      false,
		Quiet:        false,
	}

	generator := NewGenerator(config)

	if generator == nil {
		t.Fatal("expected generator to be created")
	}

	if generator.Config != config {
		t.Error("expected generator to have the provided config")
	}

	if generator.Output == nil {
		t.Error("expected generator to have output initialized")
	}
}

func TestGenerator_DiscoverActionFiles(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T, tmpDir string)
		recursive   bool
		expectedLen int
		expectError bool
	}{
		{
			name: "single action.yml in root",
			setupFunc: func(t *testing.T, tmpDir string) {
				testutil.WriteTestFile(t, filepath.Join(tmpDir, "action.yml"), testutil.SimpleActionYML)
			},
			recursive:   false,
			expectedLen: 1,
		},
		{
			name: "action.yaml variant",
			setupFunc: func(t *testing.T, tmpDir string) {
				testutil.WriteTestFile(t, filepath.Join(tmpDir, "action.yaml"), testutil.SimpleActionYML)
			},
			recursive:   false,
			expectedLen: 1,
		},
		{
			name: "both yml and yaml files",
			setupFunc: func(t *testing.T, tmpDir string) {
				testutil.WriteTestFile(t, filepath.Join(tmpDir, "action.yml"), testutil.SimpleActionYML)
				testutil.WriteTestFile(t, filepath.Join(tmpDir, "action.yaml"), testutil.MinimalActionYML)
			},
			recursive:   false,
			expectedLen: 2,
		},
		{
			name: "recursive discovery",
			setupFunc: func(t *testing.T, tmpDir string) {
				testutil.WriteTestFile(t, filepath.Join(tmpDir, "action.yml"), testutil.SimpleActionYML)
				subDir := filepath.Join(tmpDir, "subdir")
				_ = os.MkdirAll(subDir, 0755)
				testutil.WriteTestFile(t, filepath.Join(subDir, "action.yml"), testutil.CompositeActionYML)
			},
			recursive:   true,
			expectedLen: 2,
		},
		{
			name: "non-recursive skips subdirectories",
			setupFunc: func(t *testing.T, tmpDir string) {
				testutil.WriteTestFile(t, filepath.Join(tmpDir, "action.yml"), testutil.SimpleActionYML)
				subDir := filepath.Join(tmpDir, "subdir")
				_ = os.MkdirAll(subDir, 0755)
				testutil.WriteTestFile(t, filepath.Join(subDir, "action.yml"), testutil.CompositeActionYML)
			},
			recursive:   false,
			expectedLen: 1,
		},
		{
			name: "no action files",
			setupFunc: func(t *testing.T, tmpDir string) {
				testutil.WriteTestFile(t, filepath.Join(tmpDir, "README.md"), "# Test")
			},
			recursive:   false,
			expectedLen: 0,
		},
		{
			name:        "nonexistent directory",
			setupFunc:   nil,
			recursive:   false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			config := &AppConfig{Quiet: true}
			generator := NewGenerator(config)

			testDir := tmpDir
			if tt.setupFunc != nil {
				tt.setupFunc(t, tmpDir)
			} else if tt.expectError {
				testDir = filepath.Join(tmpDir, "nonexistent")
			}

			files, err := generator.DiscoverActionFiles(testDir, tt.recursive)

			if tt.expectError {
				testutil.AssertError(t, err)
				return
			}

			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, tt.expectedLen, len(files))

			// Verify all returned files exist and are action files
			for _, file := range files {
				if _, err := os.Stat(file); os.IsNotExist(err) {
					t.Errorf("discovered file does not exist: %s", file)
				}

				if !strings.HasSuffix(file, "action.yml") && !strings.HasSuffix(file, "action.yaml") {
					t.Errorf("discovered file is not an action file: %s", file)
				}
			}
		})
	}
}

func TestGenerator_GenerateFromFile(t *testing.T) {
	tests := []struct {
		name         string
		actionYML    string
		outputFormat string
		expectError  bool
		contains     []string
	}{
		{
			name:         "simple action to markdown",
			actionYML:    testutil.SimpleActionYML,
			outputFormat: "md",
			expectError:  false,
			contains:     []string{"# Simple Action", "A simple test action"},
		},
		{
			name:         "composite action to markdown",
			actionYML:    testutil.CompositeActionYML,
			outputFormat: "md",
			expectError:  false,
			contains:     []string{"# Composite Action", "A composite action with dependencies"},
		},
		{
			name:         "action to HTML",
			actionYML:    testutil.SimpleActionYML,
			outputFormat: "html",
			expectError:  false,
			contains:     []string{"<html>", "<h1>Simple Action</h1>"},
		},
		{
			name:         "action to JSON",
			actionYML:    testutil.SimpleActionYML,
			outputFormat: "json",
			expectError:  false,
			contains:     []string{`"name":"Simple Action"`, `"description":"A simple test action"`},
		},
		{
			name:         "invalid action file",
			actionYML:    testutil.InvalidActionYML,
			outputFormat: "md",
			expectError:  true,
		},
		{
			name:         "unknown output format",
			actionYML:    testutil.SimpleActionYML,
			outputFormat: "unknown",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			// Write action file
			actionPath := filepath.Join(tmpDir, "action.yml")
			testutil.WriteTestFile(t, actionPath, tt.actionYML)

			// Create generator
			config := &AppConfig{
				OutputFormat: tt.outputFormat,
				OutputDir:    tmpDir,
				Quiet:        true,
			}
			generator := NewGenerator(config)

			// Generate output
			err := generator.GenerateFromFile(actionPath)

			if tt.expectError {
				testutil.AssertError(t, err)
				return
			}

			testutil.AssertNoError(t, err)

			// Find the generated output file
			readmeFiles, _ := filepath.Glob(filepath.Join(tmpDir, "README*.md"))
			if len(readmeFiles) == 0 {
				t.Error("no output file was created")
				return
			}

			// Read and verify output content
			content, err := os.ReadFile(readmeFiles[0])
			testutil.AssertNoError(t, err)

			contentStr := string(content)
			for _, expectedStr := range tt.contains {
				if !strings.Contains(contentStr, expectedStr) {
					t.Errorf("output does not contain expected string %q", expectedStr)
					t.Logf("Output content: %s", contentStr)
				}
			}
		})
	}
}

func TestGenerator_ProcessBatch(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T, tmpDir string) []string
		expectError bool
		expectFiles int
	}{
		{
			name: "process multiple valid files",
			setupFunc: func(t *testing.T, tmpDir string) []string {
				files := []string{
					filepath.Join(tmpDir, "action1.yml"),
					filepath.Join(tmpDir, "action2.yml"),
				}
				testutil.WriteTestFile(t, files[0], testutil.SimpleActionYML)
				testutil.WriteTestFile(t, files[1], testutil.CompositeActionYML)
				return files
			},
			expectError: false,
			expectFiles: 2,
		},
		{
			name: "handle mixed valid and invalid files",
			setupFunc: func(t *testing.T, tmpDir string) []string {
				files := []string{
					filepath.Join(tmpDir, "valid.yml"),
					filepath.Join(tmpDir, "invalid.yml"),
				}
				testutil.WriteTestFile(t, files[0], testutil.SimpleActionYML)
				testutil.WriteTestFile(t, files[1], testutil.InvalidActionYML)
				return files
			},
			expectError: true, // Should fail due to invalid file
		},
		{
			name: "empty file list",
			setupFunc: func(_ *testing.T, _ string) []string {
				return []string{}
			},
			expectError: false,
			expectFiles: 0,
		},
		{
			name: "nonexistent files",
			setupFunc: func(_ *testing.T, tmpDir string) []string {
				return []string{filepath.Join(tmpDir, "nonexistent.yml")}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			config := &AppConfig{
				OutputFormat: "md",
				OutputDir:    tmpDir,
				Quiet:        true,
			}
			generator := NewGenerator(config)

			files := tt.setupFunc(t, tmpDir)
			err := generator.ProcessBatch(files)

			if tt.expectError {
				testutil.AssertError(t, err)
				return
			}

			testutil.AssertNoError(t, err)

			// Count generated README files
			readmeFiles, _ := filepath.Glob(filepath.Join(tmpDir, "README*.md"))
			if len(readmeFiles) != tt.expectFiles {
				t.Errorf("expected %d README files, got %d", tt.expectFiles, len(readmeFiles))
			}
		})
	}
}

func TestGenerator_ValidateFiles(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T, tmpDir string) []string
		expectError bool
	}{
		{
			name: "all valid files",
			setupFunc: func(t *testing.T, tmpDir string) []string {
				files := []string{
					filepath.Join(tmpDir, "action1.yml"),
					filepath.Join(tmpDir, "action2.yml"),
				}
				testutil.WriteTestFile(t, files[0], testutil.SimpleActionYML)
				testutil.WriteTestFile(t, files[1], testutil.MinimalActionYML)
				return files
			},
			expectError: false,
		},
		{
			name: "files with validation issues",
			setupFunc: func(t *testing.T, tmpDir string) []string {
				files := []string{
					filepath.Join(tmpDir, "valid.yml"),
					filepath.Join(tmpDir, "invalid.yml"),
				}
				testutil.WriteTestFile(t, files[0], testutil.SimpleActionYML)
				testutil.WriteTestFile(t, files[1], testutil.InvalidActionYML)
				return files
			},
			expectError: true,
		},
		{
			name: "nonexistent files",
			setupFunc: func(_ *testing.T, tmpDir string) []string {
				return []string{filepath.Join(tmpDir, "nonexistent.yml")}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			config := &AppConfig{Quiet: true}
			generator := NewGenerator(config)

			files := tt.setupFunc(t, tmpDir)
			err := generator.ValidateFiles(files)

			if tt.expectError {
				testutil.AssertError(t, err)
			} else {
				testutil.AssertNoError(t, err)
			}
		})
	}
}

func TestGenerator_CreateDependencyAnalyzer(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "with GitHub token",
			token:       "test-token",
			expectError: false,
		},
		{
			name:        "without GitHub token",
			token:       "",
			expectError: false, // Should not error, but analyzer might have limitations
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &AppConfig{
				GitHubToken: tt.token,
				Quiet:       true,
			}
			generator := NewGenerator(config)

			analyzer, err := generator.CreateDependencyAnalyzer()

			if tt.expectError {
				testutil.AssertError(t, err)
				return
			}

			testutil.AssertNoError(t, err)

			if analyzer == nil {
				t.Error("expected analyzer to be created")
			}
		})
	}
}

func TestGenerator_WithDifferentThemes(t *testing.T) {
	themes := []string{"default", "github", "gitlab", "minimal", "professional"}

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	actionPath := filepath.Join(tmpDir, "action.yml")
	testutil.WriteTestFile(t, actionPath, testutil.SimpleActionYML)

	for _, theme := range themes {
		t.Run("theme_"+theme, func(t *testing.T) {
			config := &AppConfig{
				Theme:        theme,
				OutputFormat: "md",
				OutputDir:    tmpDir,
				Quiet:        true,
			}
			generator := NewGenerator(config)

			err := generator.GenerateFromFile(actionPath)
			testutil.AssertNoError(t, err)

			// Verify output was created
			readmeFiles, _ := filepath.Glob(filepath.Join(tmpDir, "README*.md"))
			if len(readmeFiles) == 0 {
				t.Errorf("no output file was created for theme %s", theme)
			}

			// Clean up for next test
			for _, file := range readmeFiles {
				_ = os.Remove(file)
			}
		})
	}
}

func TestGenerator_ErrorHandling(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T, tmpDir string) (*Generator, string)
		wantError string
	}{
		{
			name: "invalid template path",
			setupFunc: func(t *testing.T, tmpDir string) (*Generator, string) {
				config := &AppConfig{
					Template:     "/nonexistent/template.tmpl",
					OutputFormat: "md",
					OutputDir:    tmpDir,
					Quiet:        true,
				}
				generator := NewGenerator(config)
				actionPath := filepath.Join(tmpDir, "action.yml")
				testutil.WriteTestFile(t, actionPath, testutil.SimpleActionYML)
				return generator, actionPath
			},
			wantError: "template",
		},
		{
			name: "permission denied on output directory",
			setupFunc: func(t *testing.T, tmpDir string) (*Generator, string) {
				// Create a directory with no write permissions
				restrictedDir := filepath.Join(tmpDir, "restricted")
				_ = os.MkdirAll(restrictedDir, 0444) // Read-only

				config := &AppConfig{
					OutputFormat: "md",
					OutputDir:    restrictedDir,
					Quiet:        true,
				}
				generator := NewGenerator(config)
				actionPath := filepath.Join(tmpDir, "action.yml")
				testutil.WriteTestFile(t, actionPath, testutil.SimpleActionYML)
				return generator, actionPath
			},
			wantError: "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			generator, actionPath := tt.setupFunc(t, tmpDir)
			err := generator.GenerateFromFile(actionPath)

			testutil.AssertError(t, err)
			if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.wantError)) {
				t.Errorf("expected error containing %q, got: %v", tt.wantError, err)
			}
		})
	}
}
