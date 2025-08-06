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
				t.Helper()
				fixture, err := testutil.LoadActionFixture("actions/javascript/simple.yml")
				testutil.AssertNoError(t, err)
				testutil.WriteTestFile(t, filepath.Join(tmpDir, "action.yml"), fixture.Content)
			},
			recursive:   false,
			expectedLen: 1,
		},
		{
			name: "action.yaml variant",
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				fixture, err := testutil.LoadActionFixture("actions/javascript/simple.yml")
				testutil.AssertNoError(t, err)
				testutil.WriteTestFile(t, filepath.Join(tmpDir, "action.yaml"), fixture.Content)
			},
			recursive:   false,
			expectedLen: 1,
		},
		{
			name: "both yml and yaml files",
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				simpleFixture, err := testutil.LoadActionFixture("actions/javascript/simple.yml")
				testutil.AssertNoError(t, err)
				minimalFixture, err := testutil.LoadActionFixture("minimal-action.yml")
				testutil.AssertNoError(t, err)
				testutil.WriteTestFile(t, filepath.Join(tmpDir, "action.yml"), simpleFixture.Content)
				testutil.WriteTestFile(t, filepath.Join(tmpDir, "action.yaml"), minimalFixture.Content)
			},
			recursive:   false,
			expectedLen: 2,
		},
		{
			name: "recursive discovery",
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				simpleFixture, err := testutil.LoadActionFixture("actions/javascript/simple.yml")
				testutil.AssertNoError(t, err)
				compositeFixture, err := testutil.LoadActionFixture("actions/composite/basic.yml")
				testutil.AssertNoError(t, err)
				testutil.WriteTestFile(t, filepath.Join(tmpDir, "action.yml"), simpleFixture.Content)
				subDir := filepath.Join(tmpDir, "subdir")
				_ = os.MkdirAll(subDir, 0750) // #nosec G301 -- test directory permissions
				testutil.WriteTestFile(t, filepath.Join(subDir, "action.yml"), compositeFixture.Content)
			},
			recursive:   true,
			expectedLen: 2,
		},
		{
			name: "non-recursive skips subdirectories",
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				simpleFixture, err := testutil.LoadActionFixture("actions/javascript/simple.yml")
				testutil.AssertNoError(t, err)
				compositeFixture, err := testutil.LoadActionFixture("actions/composite/basic.yml")
				testutil.AssertNoError(t, err)
				testutil.WriteTestFile(t, filepath.Join(tmpDir, "action.yml"), simpleFixture.Content)
				subDir := filepath.Join(tmpDir, "subdir")
				_ = os.MkdirAll(subDir, 0750) // #nosec G301 -- test directory permissions
				testutil.WriteTestFile(t, filepath.Join(subDir, "action.yml"), compositeFixture.Content)
			},
			recursive:   false,
			expectedLen: 1,
		},
		{
			name: "no action files",
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
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
			actionYML:    testutil.MustReadFixture("actions/javascript/simple.yml"),
			outputFormat: "md",
			expectError:  false,
			contains:     []string{"# Simple JavaScript Action", "A simple JavaScript action for testing"},
		},
		{
			name:         "composite action to markdown",
			actionYML:    testutil.MustReadFixture("actions/composite/basic.yml"),
			outputFormat: "md",
			expectError:  false,
			contains:     []string{"# Basic Composite Action", "A simple composite action with basic steps"},
		},
		{
			name:         "action to HTML",
			actionYML:    testutil.MustReadFixture("actions/javascript/simple.yml"),
			outputFormat: "html",
			expectError:  false,
			contains: []string{
				"Simple JavaScript Action",
				"A simple JavaScript action for testing",
			}, // HTML uses same template content
		},
		{
			name:         "action to JSON",
			actionYML:    testutil.MustReadFixture("actions/javascript/simple.yml"),
			outputFormat: "json",
			expectError:  false,
			contains: []string{
				`"name": "Simple JavaScript Action"`,
				`"description": "A simple JavaScript action for testing"`,
			},
		},
		{
			name:         "invalid action file",
			actionYML:    testutil.MustReadFixture("actions/invalid/invalid-using.yml"),
			outputFormat: "md",
			expectError:  true, // Invalid runtime configuration should cause failure
			contains:     []string{},
		},
		{
			name:         "unknown output format",
			actionYML:    testutil.MustReadFixture("actions/javascript/simple.yml"),
			outputFormat: "unknown",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			// Set up test templates
			testutil.SetupTestTemplates(t, tmpDir)

			// Write action file
			actionPath := filepath.Join(tmpDir, "action.yml")
			testutil.WriteTestFile(t, actionPath, tt.actionYML)

			// Create generator with explicit template path
			config := &AppConfig{
				OutputFormat: tt.outputFormat,
				OutputDir:    tmpDir,
				Quiet:        true,
				Template:     filepath.Join(tmpDir, "templates", "readme.tmpl"),
			}
			generator := NewGenerator(config)

			// Generate output
			err := generator.GenerateFromFile(actionPath)

			if tt.expectError {
				testutil.AssertError(t, err)

				return
			}

			testutil.AssertNoError(t, err)

			// Find the generated output file based on format
			var pattern string
			switch tt.outputFormat {
			case "html":
				pattern = filepath.Join(tmpDir, "*.html")
			case "json":
				pattern = filepath.Join(tmpDir, "*.json")
			default:
				pattern = filepath.Join(tmpDir, "README*.md")
			}
			readmeFiles, _ := filepath.Glob(pattern)
			if len(readmeFiles) == 0 {
				t.Errorf("no output file was created for format %s", tt.outputFormat)

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

// countREADMEFiles counts README.md files in a directory tree.
func countREADMEFiles(t *testing.T, dir string) int {
	t.Helper()
	count := 0
	err := filepath.Walk(dir, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, "README.md") {
			count++
		}

		return nil
	})
	if err != nil {
		t.Errorf("error walking directory: %v", err)
	}

	return count
}

// logREADMELocations logs the locations of README files for debugging.
func logREADMELocations(t *testing.T, dir string) {
	t.Helper()
	_ = filepath.Walk(dir, func(path string, _ os.FileInfo, err error) error {
		if err == nil && strings.HasSuffix(path, "README.md") {
			t.Logf("Found README at: %s", path)
		}

		return nil
	})
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
				t.Helper()
				// Create separate directories for each action
				dir1 := filepath.Join(tmpDir, "action1")
				dir2 := filepath.Join(tmpDir, "action2")
				if err := os.MkdirAll(dir1, 0750); err != nil { // #nosec G301 -- test directory permissions
					t.Fatalf("failed to create dir1: %v", err)
				}
				if err := os.MkdirAll(dir2, 0750); err != nil { // #nosec G301 -- test directory permissions
					t.Fatalf("failed to create dir2: %v", err)
				}

				files := []string{
					filepath.Join(dir1, "action.yml"),
					filepath.Join(dir2, "action.yml"),
				}
				testutil.WriteTestFile(t, files[0], testutil.MustReadFixture("actions/javascript/simple.yml"))
				testutil.WriteTestFile(t, files[1], testutil.MustReadFixture("actions/composite/basic.yml"))

				return files
			},
			expectError: false,
			expectFiles: 2,
		},
		{
			name: "handle mixed valid and invalid files",
			setupFunc: func(t *testing.T, tmpDir string) []string {
				t.Helper()
				// Create separate directories for mixed test too
				dir1 := filepath.Join(tmpDir, "valid-action")
				dir2 := filepath.Join(tmpDir, "invalid-action")
				if err := os.MkdirAll(dir1, 0750); err != nil { // #nosec G301 -- test directory permissions
					t.Fatalf("failed to create dir1: %v", err)
				}
				if err := os.MkdirAll(dir2, 0750); err != nil { // #nosec G301 -- test directory permissions
					t.Fatalf("failed to create dir2: %v", err)
				}

				files := []string{
					filepath.Join(dir1, "action.yml"),
					filepath.Join(dir2, "action.yml"),
				}
				testutil.WriteTestFile(t, files[0], testutil.MustReadFixture("actions/javascript/simple.yml"))
				testutil.WriteTestFile(t, files[1], testutil.MustReadFixture("actions/invalid/invalid-using.yml"))

				return files
			},
			expectError: true, // Invalid runtime configuration should cause batch to fail
			expectFiles: 0,    // No files should be expected when batch fails
		},
		{
			name: "empty file list",
			setupFunc: func(_ *testing.T, _ string) []string {
				return []string{}
			},
			expectError: true, // ProcessBatch returns error for empty list
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

			// Set up test templates
			testutil.SetupTestTemplates(t, tmpDir)

			config := &AppConfig{
				OutputFormat: "md",
				// Don't set OutputDir so each action generates README in its own directory
				Verbose:  true, // Enable verbose to see what's happening
				Template: filepath.Join(tmpDir, "templates", "readme.tmpl"),
			}
			generator := NewGenerator(config)

			files := tt.setupFunc(t, tmpDir)
			err := generator.ProcessBatch(files)

			if tt.expectError {
				testutil.AssertError(t, err)

				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)

				return
			}

			// Count generated README files
			if tt.expectFiles > 0 {
				readmeCount := countREADMEFiles(t, tmpDir)
				if readmeCount != tt.expectFiles {
					t.Errorf("expected %d README files, got %d", tt.expectFiles, readmeCount)
					t.Logf("Expected %d files, found %d", tt.expectFiles, readmeCount)
					logREADMELocations(t, tmpDir)
				}
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
				t.Helper()
				files := []string{
					filepath.Join(tmpDir, "action1.yml"),
					filepath.Join(tmpDir, "action2.yml"),
				}
				testutil.WriteTestFile(t, files[0], testutil.MustReadFixture("actions/javascript/simple.yml"))
				testutil.WriteTestFile(t, files[1], testutil.MustReadFixture("minimal-action.yml"))

				return files
			},
			expectError: false,
		},
		{
			name: "files with validation issues",
			setupFunc: func(t *testing.T, tmpDir string) []string {
				t.Helper()
				files := []string{
					filepath.Join(tmpDir, "valid.yml"),
					filepath.Join(tmpDir, "invalid.yml"),
				}
				testutil.WriteTestFile(t, files[0], testutil.MustReadFixture("actions/javascript/simple.yml"))
				testutil.WriteTestFile(t, files[1], testutil.MustReadFixture("actions/invalid/missing-description.yml"))

				return files
			},
			expectError: true, // Validation should fail for invalid runtime configuration
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

	// Set up test templates
	testutil.SetupTestTemplates(t, tmpDir)

	actionPath := filepath.Join(tmpDir, "action.yml")
	testutil.WriteTestFile(t, actionPath, testutil.MustReadFixture("actions/javascript/simple.yml"))

	for _, theme := range themes {
		t.Run("theme_"+theme, func(t *testing.T) {
			// Change to tmpDir so templates can be found
			origDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("failed to get working directory: %v", err)
			}
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("failed to change directory: %v", err)
			}
			defer func() {
				if err := os.Chdir(origDir); err != nil {
					t.Errorf("failed to restore directory: %v", err)
				}
			}()

			config := &AppConfig{
				Theme:        theme,
				OutputFormat: "md",
				OutputDir:    tmpDir,
				Quiet:        true,
			}
			generator := NewGenerator(config)

			if err := generator.GenerateFromFile(actionPath); err != nil {
				t.Errorf("unexpected error: %v", err)

				return
			}

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
				t.Helper()
				config := &AppConfig{
					Template:     "/nonexistent/template.tmpl",
					OutputFormat: "md",
					OutputDir:    tmpDir,
					Quiet:        true,
				}
				generator := NewGenerator(config)
				actionPath := filepath.Join(tmpDir, "action.yml")
				testutil.WriteTestFile(t, actionPath, testutil.MustReadFixture("actions/javascript/simple.yml"))

				return generator, actionPath
			},
			wantError: "template",
		},
		{
			name: "permission denied on output directory",
			setupFunc: func(t *testing.T, tmpDir string) (*Generator, string) {
				t.Helper()
				// Set up test templates
				testutil.SetupTestTemplates(t, tmpDir)

				// Create a directory with no write permissions
				restrictedDir := filepath.Join(tmpDir, "restricted")
				_ = os.MkdirAll(restrictedDir, 0444) // #nosec G301 -- intentionally read-only for test

				config := &AppConfig{
					OutputFormat: "md",
					OutputDir:    restrictedDir,
					Quiet:        true,
					Template:     filepath.Join(tmpDir, "templates", "readme.tmpl"),
				}
				generator := NewGenerator(config)
				actionPath := filepath.Join(tmpDir, "action.yml")
				testutil.WriteTestFile(t, actionPath, testutil.MustReadFixture("actions/javascript/simple.yml"))

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
