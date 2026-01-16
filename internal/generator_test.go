package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/apperrors"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// defaultTestConfig returns an AppConfig with sensible test defaults.
// Sets Quiet: true to suppress output during tests.
func defaultTestConfig() *AppConfig {
	return &AppConfig{
		Theme:        appconstants.ThemeDefault,
		OutputFormat: appconstants.OutputFormatMarkdown,
		OutputDir:    ".",
		Quiet:        true,
	}
}

// assertActionFiles verifies that all files are valid action files.
func assertActionFiles(t *testing.T, files []string) {
	t.Helper()
	for _, file := range files {
		testutil.AssertFileExists(t, file)
		if !strings.HasSuffix(file, appconstants.ActionFileNameYML) &&
			!strings.HasSuffix(file, appconstants.ActionFileNameYAML) {
			t.Errorf("discovered file is not an action file: %s", file)
		}
	}
}

// createMultipleFixtureFiles writes multiple fixtures to files and returns their paths.
// This helper reduces duplication for tests that set up multiple action files.
func createMultipleFixtureFiles(
	t *testing.T,
	tmpDir string,
	filesAndFixtures map[string]string,
) []string {
	t.Helper()

	files := make([]string, 0, len(filesAndFixtures))
	for filename, fixturePath := range filesAndFixtures {
		filePath := filepath.Join(tmpDir, filename)
		testutil.WriteTestFile(t, filePath, testutil.MustReadFixture(fixturePath))
		files = append(files, filePath)
	}

	return files
}

func TestGeneratorNewGenerator(t *testing.T) {
	t.Parallel()
	config := defaultTestConfig()
	config.Quiet = false // Override for this test

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

func TestGeneratorDiscoverActionFiles(t *testing.T) {
	t.Parallel()
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
				testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureJavaScriptSimple)
			},
			recursive:   false,
			expectedLen: 1,
		},
		{
			name: "action.yaml variant",
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				testutil.WriteActionFixtureAs(
					t,
					tmpDir,
					appconstants.ActionFileNameYAML,
					testutil.TestFixtureJavaScriptSimple,
				)
			},
			recursive:   false,
			expectedLen: 1,
		},
		{
			name: "both yml and yaml files",
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureJavaScriptSimple)
				testutil.WriteActionFixtureAs(
					t,
					tmpDir,
					appconstants.ActionFileNameYAML,
					testutil.TestFixtureMinimalAction,
				)
			},
			recursive:   false,
			expectedLen: 2,
		},
		{
			name: "recursive discovery",
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureJavaScriptSimple)
				testutil.CreateActionSubdir(
					t,
					tmpDir,
					testutil.TestDirSubdir,
					testutil.TestFixtureCompositeBasic,
				)
			},
			recursive:   true,
			expectedLen: 2,
		},
		{
			name: "non-recursive skips subdirectories",
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureJavaScriptSimple)
				testutil.CreateActionSubdir(
					t,
					tmpDir,
					testutil.TestDirSubdir,
					testutil.TestFixtureCompositeBasic,
				)
			},
			recursive:   false,
			expectedLen: 1,
		},
		{
			name: "no action files",
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				testutil.WriteTestFile(t, filepath.Join(tmpDir, appconstants.ReadmeMarkdown), "# Test")
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
			t.Parallel()
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			config := defaultTestConfig()
			generator := NewGenerator(config)

			testDir := tmpDir
			if tt.setupFunc != nil {
				tt.setupFunc(t, tmpDir)
			} else if tt.expectError {
				testDir = filepath.Join(tmpDir, "nonexistent")
			}

			files, err := generator.DiscoverActionFiles(testDir, tt.recursive, []string{})

			if tt.expectError {
				testutil.AssertError(t, err)

				return
			}

			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, tt.expectedLen, len(files))

			assertActionFiles(t, files)
		})
	}
}

func TestGeneratorDiscoverActionFilesVerbose(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		recursive bool
	}{
		{
			name:      "verbose non-recursive",
			recursive: false,
		},
		{
			name:      "verbose recursive",
			recursive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			// Create test action file
			testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureJavaScriptSimple)
			if tt.recursive {
				testutil.CreateActionSubdir(t, tmpDir, "subdir", testutil.TestFixtureCompositeBasic)
			}

			// Create generator with verbose mode enabled
			config := defaultTestConfig()
			config.Verbose = true
			generator := NewGenerator(config)

			files, err := generator.DiscoverActionFiles(tmpDir, tt.recursive, []string{})

			testutil.AssertNoError(t, err)
			if tt.recursive {
				testutil.AssertEqual(t, 2, len(files))
			} else {
				testutil.AssertEqual(t, 1, len(files))
			}
		})
	}
}

// getOutputPattern returns the expected output filename pattern for the given format.
func getOutputPattern(format string) string {
	switch format {
	case appconstants.OutputFormatHTML:
		return "*.html"
	case appconstants.OutputFormatJSON:
		return "*.json"
	case appconstants.OutputFormatASCIIDoc:
		return "*.adoc"
	default:
		return "README*.md"
	}
}

// validateGeneratedContent validates that the generated content contains expected strings.
func validateGeneratedContent(t *testing.T, content []byte, expectedStrings []string) {
	t.Helper()

	for _, expected := range expectedStrings {
		if !strings.Contains(string(content), expected) {
			t.Errorf("Output missing expected string: %q", expected)
		}
	}
}

func TestGeneratorGenerateFromFile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		actionYML    string
		outputFormat string
		expectError  bool
		contains     []string
	}{
		{
			name:         "simple action to markdown",
			actionYML:    testutil.MustReadFixture(testutil.TestFixtureJavaScriptSimple),
			outputFormat: appconstants.OutputFormatMarkdown,
			expectError:  false,
			contains:     []string{"# Simple JavaScript Action", "A simple JavaScript action for testing"},
		},
		{
			name:         "composite action to markdown",
			actionYML:    testutil.MustReadFixture(testutil.TestFixtureCompositeBasic),
			outputFormat: appconstants.OutputFormatMarkdown,
			expectError:  false,
			contains:     []string{"# Basic Composite Action", "A simple composite action with basic steps"},
		},
		{
			name:         "action to HTML",
			actionYML:    testutil.MustReadFixture(testutil.TestFixtureJavaScriptSimple),
			outputFormat: appconstants.OutputFormatHTML,
			expectError:  false,
			contains: []string{
				"Simple JavaScript Action",
				"A simple JavaScript action for testing",
			}, // HTML uses same template content
		},
		{
			name:         "action to JSON",
			actionYML:    testutil.MustReadFixture(testutil.TestFixtureJavaScriptSimple),
			outputFormat: appconstants.OutputFormatJSON,
			expectError:  false,
			contains: []string{
				`"name": "Simple JavaScript Action"`,
				`"description": "A simple JavaScript action for testing"`,
			},
		},
		{
			name:         "invalid action file",
			actionYML:    testutil.MustReadFixture(testutil.TestFixtureInvalidInvalidUsing),
			outputFormat: appconstants.OutputFormatMarkdown,
			expectError:  true, // Invalid runtime configuration should cause failure
			contains:     []string{},
		},
		{
			name:         "unknown output format",
			actionYML:    testutil.MustReadFixture(testutil.TestFixtureJavaScriptSimple),
			outputFormat: "unknown",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			// Set up test templates
			testutil.SetupTestTemplates(t, tmpDir)

			// Write action file
			actionPath := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
			testutil.WriteTestFile(t, actionPath, tt.actionYML)

			// Create generator with explicit template path
			config := &AppConfig{
				OutputFormat: tt.outputFormat,
				OutputDir:    tmpDir,
				Quiet:        true,
				Template:     filepath.Join(tmpDir, "templates", appconstants.TemplateReadme),
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
			filename := getOutputPattern(tt.outputFormat)
			pattern := filepath.Join(tmpDir, filename)
			readmeFiles, _ := filepath.Glob(pattern)
			if len(readmeFiles) == 0 {
				t.Errorf("no output file was created for format %s", tt.outputFormat)

				return
			}

			// Read and verify output content
			content, err := os.ReadFile(readmeFiles[0])
			testutil.AssertNoError(t, err)
			validateGeneratedContent(t, content, tt.contains)
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
		if strings.HasSuffix(path, appconstants.ReadmeMarkdown) {
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
		if err == nil && strings.HasSuffix(path, appconstants.ReadmeMarkdown) {
			t.Logf("Found README at: %s", path)
		}

		return nil
	})
}

func TestGeneratorProcessBatch(t *testing.T) {
	t.Parallel()
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
				dirs := createTestDirs(t, tmpDir, "action1", "action2")

				files := []string{
					filepath.Join(dirs[0], appconstants.ActionFileNameYML),
					filepath.Join(dirs[1], appconstants.ActionFileNameYML),
				}
				testutil.WriteTestFile(t, files[0], testutil.MustReadFixture(testutil.TestFixtureJavaScriptSimple))
				testutil.WriteTestFile(t, files[1], testutil.MustReadFixture(testutil.TestFixtureCompositeBasic))

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
				dirs := createTestDirs(t, tmpDir, "valid-action", "invalid-action")

				files := []string{
					filepath.Join(dirs[0], appconstants.ActionFileNameYML),
					filepath.Join(dirs[1], appconstants.ActionFileNameYML),
				}
				testutil.WriteTestFile(t, files[0], testutil.MustReadFixture(testutil.TestFixtureJavaScriptSimple))
				testutil.WriteTestFile(
					t,
					files[1],
					testutil.MustReadFixture(testutil.TestFixtureInvalidInvalidUsing),
				)

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
			t.Parallel()
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			// Set up test templates
			testutil.SetupTestTemplates(t, tmpDir)

			config := defaultTestConfig()
			config.OutputFormat = appconstants.OutputFormatMarkdown
			// Don't set OutputDir so each action generates README in its own directory
			config.OutputDir = ""
			config.Verbose = true // Enable verbose to see what's happening
			config.Template = filepath.Join(tmpDir, "templates", appconstants.TemplateReadme)
			generator := NewGenerator(config)

			files := tt.setupFunc(t, tmpDir)
			err := generator.ProcessBatch(files)

			if tt.expectError {
				testutil.AssertError(t, err)

				return
			}

			if err != nil {
				t.Errorf(testutil.TestErrUnexpected, err)

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

func TestGeneratorValidateFiles(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T, tmpDir string) []string
		expectError bool
	}{
		{
			name: "all valid files",
			setupFunc: func(t *testing.T, tmpDir string) []string {
				t.Helper()

				return createMultipleFixtureFiles(t, tmpDir, map[string]string{
					"action1.yml": testutil.TestFixtureJavaScriptSimple,
					"action2.yml": testutil.TestFixtureMinimalAction,
				})
			},
			expectError: false,
		},
		{
			name: "files with validation issues",
			setupFunc: func(t *testing.T, tmpDir string) []string {
				t.Helper()

				return createMultipleFixtureFiles(t, tmpDir, map[string]string{
					"valid.yml":   testutil.TestFixtureJavaScriptSimple,
					"invalid.yml": testutil.TestFixtureInvalidMissingDescription,
				})
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
			t.Parallel()
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			config := defaultTestConfig()
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

func TestGeneratorCreateDependencyAnalyzer(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			config := defaultTestConfig()
			config.GitHubToken = tt.token
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

func TestGeneratorWithDifferentThemes(t *testing.T) {
	t.Parallel()
	themes := []string{
		appconstants.ThemeDefault,
		appconstants.ThemeGitHub,
		appconstants.ThemeGitLab,
		appconstants.ThemeMinimal,
		appconstants.ThemeProfessional,
	}

	for _, theme := range themes {
		t.Run("theme_"+theme, func(t *testing.T) {
			t.Parallel()
			// Create separate temp directory for each theme test
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			// Set up test templates for this theme test
			testutil.SetupTestTemplates(t, tmpDir)

			actionPath := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
			testutil.WriteTestFile(t, actionPath, testutil.MustReadFixture(testutil.TestFixtureJavaScriptSimple))

			config := defaultTestConfig()
			config.Theme = theme
			config.OutputDir = tmpDir
			generator := NewGenerator(config)

			if err := generator.GenerateFromFile(actionPath); err != nil {
				t.Errorf(testutil.TestErrUnexpected, err)

				return
			}

			// Verify output was created
			readmeFiles, _ := filepath.Glob(filepath.Join(tmpDir, "README*.md"))
			if len(readmeFiles) == 0 {
				t.Errorf("no output file was created for theme %s", theme)
			}
		})
	}
}

func TestGeneratorErrorHandling(t *testing.T) {
	t.Parallel()
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
					OutputDir:    tmpDir,
					OutputFormat: appconstants.OutputFormatMarkdown,
					Quiet:        true,
				}
				generator := NewGenerator(config)
				actionPath := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
				testutil.WriteTestFile(
					t,
					actionPath,
					testutil.MustReadFixture(testutil.TestFixtureJavaScriptSimple),
				)

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

				config := defaultTestConfig()
				config.OutputDir = restrictedDir
				config.Template = filepath.Join(tmpDir, "templates", appconstants.TemplateReadme)
				generator := NewGenerator(config)
				actionPath := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
				testutil.WriteTestFile(
					t,
					actionPath,
					testutil.MustReadFixture(testutil.TestFixtureJavaScriptSimple),
				)

				return generator, actionPath
			},
			wantError: "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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

// createTestDirs is a helper that creates multiple directories within tmpDir for testing.
// Returns the full paths of all created directories.
func createTestDirs(t *testing.T, tmpDir string, names ...string) []string {
	t.Helper()
	dirs := make([]string, len(names))
	for i, name := range names {
		dirPath := filepath.Join(tmpDir, name)
		testutil.CreateTestDir(t, dirPath)
		dirs[i] = dirPath
	}

	return dirs
}

// TestGeneratorDiscoverActionFilesWithValidation tests the validation wrapper.
// validateDiscoveryResult validates the result of action file discovery.
func validateDiscoveryResult(t *testing.T, files []string, err error, wantErr bool) {
	t.Helper()

	if (err != nil) != wantErr {
		t.Errorf("DiscoverActionFilesWithValidation() error = %v, wantErr %v", err, wantErr)

		return
	}

	if !wantErr && len(files) == 0 {
		t.Error("Expected files but got none")
	}
}

func TestGeneratorDiscoverActionFilesWithValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		dir       string
		recursive bool
		context   string
		wantErr   bool
		setupFunc func(t *testing.T) string
	}{
		{
			name:      "nonexistent directory",
			dir:       "/nonexistent/path/does/not/exist",
			recursive: false,
			context:   "test context",
			wantErr:   true,
		},
		{
			name:      "empty directory",
			recursive: false,
			context:   "empty dir test",
			wantErr:   true,
			setupFunc: func(t *testing.T) string {
				t.Helper()

				return t.TempDir()
			},
		},
		{
			name:      "valid directory with action file",
			recursive: false,
			context:   "valid test",
			wantErr:   false,
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()
				actionPath := filepath.Clean(filepath.Join(tmpDir, appconstants.ActionFileNameYML))
				if actionPath != filepath.Join(tmpDir, appconstants.ActionFileNameYML) ||
					strings.Contains(actionPath, "..") {
					t.Fatalf("invalid path: %q", actionPath)
				}
				content := "name: Test\ndescription: Test\nruns:\n  using: composite\n  steps: []"
				testutil.WriteTestFile(t, actionPath, content)

				return tmpDir
			},
		},
		{
			name:      "path with parent traversal - .. component",
			dir:       "../outside",
			recursive: false,
			context:   "path traversal test",
			wantErr:   true,
		},
		{
			name: "path with .. in middle",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()
				// Return path with .. that would escape
				return filepath.Join(tmpDir, "..", "escape")
			},
			recursive: false,
			context:   "path traversal test",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config := DefaultAppConfig()
			config.Quiet = true
			gen := NewGenerator(config)

			dir := tt.dir
			if tt.setupFunc != nil {
				dir = tt.setupFunc(t)
			}

			files, err := gen.DiscoverActionFilesWithValidation(dir, tt.recursive, []string{}, tt.context)
			validateDiscoveryResult(t, files, err, tt.wantErr)
		})
	}
}

// TestGeneratorResolveOutputPath tests output path resolution.
// validateResolveOutputPathResult validates the result of resolveOutputPath call.
func validateResolveOutputPathResult(
	t *testing.T,
	gotPath string,
	err error,
	wantPath string,
	wantErr bool,
	errContains string,
) {
	t.Helper()

	if wantErr {
		if err == nil {
			t.Errorf("resolveOutputPath() expected error but got nil")

			return
		}
		if errContains != "" && !strings.Contains(err.Error(), errContains) {
			t.Errorf("error message %q does not contain %q", err.Error(), errContains)
		}
	} else {
		if err != nil {
			t.Errorf("resolveOutputPath() unexpected error: %v", err)

			return
		}
		if gotPath != wantPath {
			t.Errorf("resolveOutputPath() = %q, want %q", gotPath, wantPath)
		}
	}
}

func TestGeneratorResolveOutputPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		outputFilename  string
		outputDir       string
		defaultFilename string
		wantPath        string // Expected path (if no error)
		wantErr         bool   // Whether error is expected
		errContains     string // Error message substring (if wantErr)
	}{
		// LEGITIMATE PATHS - Should succeed
		{
			name:            "no custom filename",
			outputFilename:  "",
			outputDir:       testutil.TestOutputPath,
			defaultFilename: appconstants.ReadmeMarkdown,
			wantPath:        "/tmp/output/README.md",
			wantErr:         false,
		},
		{
			name:            "relative custom filename",
			outputFilename:  "custom.md",
			outputDir:       testutil.TestOutputPath,
			defaultFilename: appconstants.ReadmeMarkdown,
			wantPath:        "/tmp/output/custom.md",
			wantErr:         false,
		},
		{
			name:            "absolute custom filename",
			outputFilename:  "/absolute/path/output.md",
			outputDir:       testutil.TestOutputPath,
			defaultFilename: appconstants.ReadmeMarkdown,
			wantPath:        "/absolute/path/output.md",
			wantErr:         false,
		},
		{
			name:            "custom filename with subdirectory",
			outputFilename:  "docs/output.md",
			outputDir:       testutil.TestOutputPath,
			defaultFilename: appconstants.ReadmeMarkdown,
			wantPath:        "/tmp/output/docs/output.md",
			wantErr:         false,
		},
		{
			name:            "outputDir with .. component (filename is clean)",
			outputFilename:  "file.md",
			outputDir:       "/tmp/output/../escape",
			defaultFilename: appconstants.ReadmeMarkdown,
			wantPath:        "/tmp/escape/file.md",
			wantErr:         false,
		},

		// PATH TRAVERSAL ATTEMPTS - Should error
		{
			name:            "path traversal attempt with ../",
			outputFilename:  "../escape.md",
			outputDir:       testutil.TestOutputPath,
			defaultFilename: appconstants.ReadmeMarkdown,
			wantErr:         true,
			errContains:     testutil.TestErrPathTraversal,
		},
		{
			name:            "path traversal with ../ in middle",
			outputFilename:  "sub/../escape.md",
			outputDir:       testutil.TestOutputPath,
			defaultFilename: appconstants.ReadmeMarkdown,
			wantErr:         true,
			errContains:     testutil.TestErrPathTraversal,
		},
		{
			name:            "multiple ../ escaping directory",
			outputFilename:  "../../escape.md",
			outputDir:       testutil.TestOutputPath,
			defaultFilename: appconstants.ReadmeMarkdown,
			wantErr:         true,
			errContains:     testutil.TestErrPathTraversal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config := DefaultAppConfig()
			config.OutputFilename = tt.outputFilename
			config.Quiet = true
			gen := NewGenerator(config)

			gotPath, err := gen.resolveOutputPath(tt.outputDir, tt.defaultFilename)

			validateResolveOutputPathResult(t, gotPath, err, tt.wantPath, tt.wantErr, tt.errContains)
		})
	}
}

// TestGeneratorDiscoverActionFilesErrorPaths tests error handling in file discovery.
func TestGeneratorDiscoverActionFilesErrorPaths(t *testing.T) {
	t.Parallel()

	config := DefaultAppConfig()
	config.Quiet = true
	gen := NewGenerator(config)

	// Test with non-existent directory
	_, err := gen.DiscoverActionFiles("/nonexistent/directory", false, []string{})
	if err == nil {
		t.Error("Expected error for non-existent directory, got nil")
	}

	// Test with unreadable directory (if we can create one)
	tmpDir := t.TempDir()
	unreadableDir := filepath.Join(tmpDir, "unreadable")
	err = os.Mkdir(unreadableDir, 0000)
	if err != nil {
		t.Skip("Cannot create unreadable directory for testing")
	}
	defer func() { _ = os.Chmod(unreadableDir, 0700) }() //nolint:gosec // Test cleanup needs to restore permissions

	_, _ = gen.DiscoverActionFiles(unreadableDir, true, []string{})
	// May succeed or fail depending on platform permissions
	// Just ensure it doesn't panic
}

// TestGeneratorParseAndValidateActionErrorPaths tests validation error scenarios.
func TestGeneratorParseAndValidateActionErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		content   string
		wantErr   bool
		wantValid bool
	}{
		{
			name:      "valid action",
			content:   "name: Test\ndescription: Test\nruns:\n  using: composite\n  steps: []",
			wantErr:   false,
			wantValid: true,
		},
		{
			name:      "missing name",
			content:   "description: Test\nruns:\n  using: composite\n  steps: []",
			wantErr:   true,
			wantValid: false,
		},
		{
			name:      "missing description",
			content:   "name: Test\nruns:\n  using: composite\n  steps: []",
			wantErr:   true,
			wantValid: false,
		},
		{
			name:      "missing runs",
			content:   "name: Test\ndescription: Test",
			wantErr:   true,
			wantValid: false,
		},
		{
			name:    "invalid yaml",
			content: "name: Test\ninvalid: [\n  - item",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpPath := testutil.CreateTempActionFile(t, tt.content)

			config := DefaultAppConfig()
			config.Quiet = true
			gen := NewGenerator(config)

			action, err := gen.parseAndValidateAction(tmpPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseAndValidateAction() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && action == nil {
				t.Error("Expected action to be non-nil when no error")
			}
		})
	}
}

// TestGeneratorGenerateHTMLErrorPaths tests HTML generation error handling.
func TestGeneratorGenerateHTMLErrorPaths(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a valid action
	action := &ActionYML{
		Name:        testutil.TestActionName,
		Description: testutil.TestActionDesc,
		Runs:        map[string]any{"using": "composite"},
	}

	config := DefaultAppConfig()
	config.Quiet = true
	gen := NewGenerator(config)

	// Test with valid directory
	err := gen.generateHTML(action, tmpDir, filepath.Join(tmpDir, appconstants.ActionFileNameYML))
	if err != nil {
		t.Errorf("generateHTML() unexpected error = %v", err)
	}

	// Verify HTML file was created
	// HTML filename is based on action.Name + ".html"
	htmlPath := filepath.Join(tmpDir, "Test Action.html")
	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		t.Error("Expected Test Action.html to be created")
	}
}

// TestGeneratorGenerateJSONErrorPaths tests JSON generation error handling.
func TestGeneratorGenerateJSONErrorPaths(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a valid action
	action := &ActionYML{
		Name:        testutil.TestActionName,
		Description: testutil.TestActionDesc,
		Runs:        map[string]any{"using": "composite"},
	}

	config := DefaultAppConfig()
	config.Quiet = true
	gen := NewGenerator(config)

	// Test with valid directory
	err := gen.generateJSON(action, tmpDir)
	if err != nil {
		t.Errorf("generateJSON() unexpected error = %v", err)
	}

	// Verify JSON file was created
	// JSON filename is appconstants.ActionDocsJSON = "action-docs.json"
	jsonPath := filepath.Join(tmpDir, "action-docs.json")
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		t.Error("Expected action-docs.json to be created")
	}
}

// TestGeneratorGenerateASCIIDocErrorPaths tests AsciiDoc generation error handling.
func TestGeneratorGenerateASCIIDocErrorPaths(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a valid action
	action := &ActionYML{
		Name:        testutil.TestActionName,
		Description: testutil.TestActionDesc,
		Runs:        map[string]any{"using": "composite"},
	}

	config := DefaultAppConfig()
	config.Quiet = true
	gen := NewGenerator(config)

	// Test with valid directory
	err := gen.generateASCIIDoc(action, tmpDir, filepath.Join(tmpDir, appconstants.ActionFileNameYML))
	if err != nil {
		t.Errorf("generateASCIIDoc() unexpected error = %v", err)
	}

	// Verify AsciiDoc file was created
	adocPath := filepath.Join(tmpDir, "README.adoc")
	if _, err := os.Stat(adocPath); os.IsNotExist(err) {
		t.Error("Expected README.adoc to be created")
	}
}

// TestGeneratorReportResultsEdgeCases tests result reporting edge cases.
func TestGeneratorReportResultsEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		successCount int
		errors       []string
		wantPanic    bool
	}{
		{
			name:         "all successful",
			successCount: 5,
			errors:       []string{},
			wantPanic:    false,
		},
		{
			name:         "all failed",
			successCount: 0,
			errors:       []string{"error1", "error2"},
			wantPanic:    false,
		},
		{
			name:         "mixed results",
			successCount: 3,
			errors:       []string{"error1"},
			wantPanic:    false,
		},
		{
			name:         "zero files",
			successCount: 0,
			errors:       []string{},
			wantPanic:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config := DefaultAppConfig()
			config.Quiet = true
			gen := NewGenerator(config)

			defer func() {
				if r := recover(); r != nil && !tt.wantPanic {
					t.Errorf("reportResults() panicked unexpectedly: %v", r)
				}
			}()

			gen.reportResults(tt.successCount, tt.errors)
		})
	}
}

// testCapturedOutput wraps testutil.CapturedOutput for reportResults testing.
type testCapturedOutput struct {
	*testutil.CapturedOutput
}

// ErrorWithSuggestions wraps the testutil version to match interface signature.
func (c *testCapturedOutput) ErrorWithSuggestions(err *apperrors.ContextualError) {
	if err != nil {
		c.ErrorMessages = append(c.ErrorMessages, err.Error())
	}
}

// FormatContextualError wraps the testutil version to match interface signature.
func (c *testCapturedOutput) FormatContextualError(err *apperrors.ContextualError) string {
	if err != nil {
		return err.Error()
	}

	return ""
}

// verifyReportResultsOutput checks expected vs actual output message counts.
func verifyReportResultsOutput(t *testing.T, output *testCapturedOutput, wantBold, wantError bool) {
	t.Helper()

	// Verify Bold message
	gotBold := len(output.BoldMessages) > 0
	if wantBold && !gotBold {
		t.Error("expected Bold message, got none")
	} else if !wantBold && gotBold {
		t.Errorf("expected no Bold messages, got %d", len(output.BoldMessages))
	}

	// Verify Error messages
	gotError := len(output.ErrorMessages) > 0
	if wantError && !gotError {
		t.Error("expected Error messages, got none")
	} else if !wantError && gotError {
		t.Errorf("expected no Error messages, got %d", len(output.ErrorMessages))
	}
}

// TestGeneratorReportResultsOutput tests reportResults output in non-quiet mode.
func TestGeneratorReportResultsOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		quiet        bool
		verbose      bool
		successCount int
		errors       []string
		wantBold     bool
		wantError    bool
	}{
		{
			name:         "quiet mode - no output",
			quiet:        true,
			verbose:      false,
			successCount: 5,
			errors:       []string{"error1"},
			wantBold:     false,
			wantError:    false,
		},
		{
			name:         "non-quiet, no errors",
			quiet:        false,
			verbose:      false,
			successCount: 5,
			errors:       []string{},
			wantBold:     true,
			wantError:    false,
		},
		{
			name:         "non-quiet, verbose, with errors",
			quiet:        false,
			verbose:      true,
			successCount: 3,
			errors:       []string{"error1", "error2"},
			wantBold:     true,
			wantError:    true,
		},
		{
			name:         "non-quiet, non-verbose, with errors",
			quiet:        false,
			verbose:      false,
			successCount: 2,
			errors:       []string{"error1"},
			wantBold:     true,
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := &testCapturedOutput{
				CapturedOutput: &testutil.CapturedOutput{},
			}
			config := DefaultAppConfig()
			config.Quiet = tt.quiet
			config.Verbose = tt.verbose

			gen := NewGeneratorWithDependencies(config, output, nil)
			gen.reportResults(tt.successCount, tt.errors)

			verifyReportResultsOutput(t, output, tt.wantBold, tt.wantError)
		})
	}
}

// TestGeneratorIsUnitTestEnvironment tests unit test detection.
func TestGeneratorIsUnitTestEnvironment(t *testing.T) {
	// This test runs in a test environment, so should return true
	if !isUnitTestEnvironment() {
		t.Error("Expected isUnitTestEnvironment() to return true in test context")
	}
}

// TestGeneratorNewGeneratorEdgeCases tests generator initialization edge cases.
func TestGeneratorNewGeneratorEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config *AppConfig
	}{
		{
			name:   "nil config",
			config: nil,
		},
		{
			name:   "default config",
			config: DefaultAppConfig(),
		},
		{
			name: "custom config",
			config: &AppConfig{
				Theme:        appconstants.ThemeGitHub,
				OutputFormat: appconstants.OutputFormatHTML,
				Quiet:        true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				if r := recover(); r != nil {
					t.Errorf("NewGenerator() panicked with config %v: %v", tt.config, r)
				}
			}()

			gen := NewGenerator(tt.config)

			if gen == nil {
				t.Error("NewGenerator() returned nil")
			}
		})
	}
}
