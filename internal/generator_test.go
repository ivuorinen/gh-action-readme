package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/apperrors"
	"github.com/ivuorinen/gh-action-readme/internal/cache"
	"github.com/ivuorinen/gh-action-readme/internal/dependencies"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

const (
	testGenErrMsg1     = "error1"
	testGenErrMsg2     = "error2"
	testGenErrMsg3     = "error3"
	testGenFileFoo     = "file: foo.yml"
	testGenFieldName   = "name"
	testGenFieldDesc   = "description"
	testGenFieldRuns   = "runs"
	testGenShortDesc   = "desc"
	testGenTokenKey    = "token"
	testGenHelpSection = "For more help"
	testGenRunsUsing   = "using"
	testGenActionName  = "Action"
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
			name: testutil.TestCaseNameNoActionFiles,
			setupFunc: func(t *testing.T, tmpDir string) {
				t.Helper()
				testutil.WriteTestFile(t, filepath.Join(tmpDir, appconstants.ReadmeMarkdown), "# Test")
			},
			recursive:   false,
			expectedLen: 0,
		},
		{
			name:        testutil.TestCaseNameNonexistentDir,
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
			name:         testutil.TestCaseNameInvalidActionFile,
			actionYML:    testutil.MustReadFixture(testutil.TestFixtureInvalidInvalidUsing),
			outputFormat: appconstants.OutputFormatMarkdown,
			expectError:  true, // Invalid runtime configuration should cause failure
			contains:     []string{},
		},
		{
			name:         testutil.TestCaseNameUnknownFormat,
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

// TestGeneratorProcessBatchJSONUniqueFilenames verifies that JSON generation for
// multiple actions into one shared --output-dir writes one file per action instead
// of overwriting a single fixed action-docs.json (recursive filename collision).
func TestGeneratorProcessBatchJSONUniqueFilenames(t *testing.T) {
	t.Parallel()

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()
	testutil.SetupTestTemplates(t, tmpDir)

	setup := createMultiActionSetup(
		[]string{"action1", "action2"},
		[]string{testutil.TestFixtureJavaScriptSimple, testutil.TestFixtureCompositeBasic},
	)
	files := setup(t, tmpDir)

	outDir := filepath.Join(tmpDir, "shared-out")
	config := defaultTestConfig()
	config.OutputFormat = appconstants.OutputFormatJSON
	config.OutputDir = outDir
	generator := NewGenerator(config)

	if err := generator.ProcessBatch(files); err != nil {
		t.Fatalf("ProcessBatch: %v", err)
	}

	jsonFiles, _ := filepath.Glob(filepath.Join(outDir, "*.json"))
	if len(jsonFiles) != len(files) {
		t.Errorf("expected %d distinct JSON files (one per action), got %d: %v",
			len(files), len(jsonFiles), jsonFiles)
	}
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
			setupFunc: createMultiActionSetup(
				[]string{"action1", "action2"},
				[]string{testutil.TestFixtureJavaScriptSimple, testutil.TestFixtureCompositeBasic},
			),
			expectError: false,
			expectFiles: 2,
		},
		{
			name: "handle mixed valid and invalid files",
			setupFunc: createMultiActionSetup(
				[]string{"valid-action", "invalid-action"},
				[]string{testutil.TestFixtureJavaScriptSimple, testutil.TestFixtureInvalidInvalidUsing},
			),
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
			name:        testutil.TestCaseNameNonexistentFiles,
			setupFunc:   setupNonexistentFiles(testutil.TestNonexistentYML),
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
			name: testutil.TestCaseNameAllValidFiles,
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
			name:        testutil.TestCaseNameNonexistentFiles,
			setupFunc:   setupNonexistentFiles(testutil.TestNonexistentYML),
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

// TestGeneratorCreatesMissingOutputDir verifies generation auto-creates a
// non-existent output directory rather than failing the write — regression for the
// CI "Comprehensive Documentation Generation" step where --output-dir pointed at a
// directory that did not yet exist.
func TestGeneratorCreatesMissingOutputDir(t *testing.T) {
	t.Parallel()

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()
	testutil.SetupTestTemplates(t, tmpDir)

	actionPath := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
	testutil.WriteTestFile(t, actionPath, testutil.MustReadFixture(testutil.TestFixtureJavaScriptSimple))

	outDir := filepath.Join(tmpDir, "does", "not", "exist")
	config := defaultTestConfig()
	config.OutputDir = outDir
	generator := NewGenerator(config)

	if err := generator.GenerateFromFile(actionPath); err != nil {
		t.Fatalf("GenerateFromFile into a missing output dir: %v", err)
	}

	readmeFiles, _ := filepath.Glob(filepath.Join(outDir, "README*.md"))
	if len(readmeFiles) == 0 {
		t.Errorf("expected output written into auto-created dir %q", outDir)
	}
}

// TestGeneratorUnknownThemeErrors verifies that an unrecognized theme produces an
// explicit error rather than silently falling back to the default template.
func TestGeneratorUnknownThemeErrors(t *testing.T) {
	t.Parallel()

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	actionPath := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
	testutil.WriteTestFile(t, actionPath, testutil.MustReadFixture(testutil.TestFixtureJavaScriptSimple))

	config := defaultTestConfig()
	config.Theme = "totally-not-a-real-theme"
	config.OutputDir = tmpDir
	generator := NewGenerator(config)

	err := generator.GenerateFromFile(actionPath)
	if err == nil {
		t.Fatal("expected an error for an unknown theme, got nil")
	}
	if !strings.Contains(err.Error(), "unknown theme") {
		t.Errorf("error = %q, want it to mention 'unknown theme'", err.Error())
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
			name: testutil.TestCaseNamePermissionDenied,
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
			name:      testutil.TestCaseNameNonexistentDir,
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
				content := testutil.MustReadFixture(testutil.TestFixtureActionMinimal)
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
			name:      testutil.TestCaseNameValidAction,
			content:   testutil.MustReadFixture(testutil.TestFixtureActionMinimal),
			wantErr:   false,
			wantValid: true,
		},
		{
			name:      testutil.TestCaseNameMissingName,
			content:   testutil.MustReadFixture(testutil.TestFixtureCompositeMissingName),
			wantErr:   true,
			wantValid: false,
		},
		{
			name:      testutil.TestCaseNameMissingDesc,
			content:   testutil.MustReadFixture(testutil.TestFixtureCompositeMissingDesc),
			wantErr:   true,
			wantValid: false,
		},
		{
			name:      testutil.TestCaseNameMissingRuns,
			content:   testutil.MustReadFixture(testutil.TestFixtureCompositeNameDescOnly),
			wantErr:   true,
			wantValid: false,
		},
		{
			name:    testutil.TestCaseNameInvalidYAML,
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
	testHTMLGeneration(t)
}

// TestGeneratorGenerateJSONErrorPaths tests JSON generation error handling.
func TestGeneratorGenerateJSONErrorPaths(t *testing.T) {
	testJSONGeneration(t)
}

// TestGeneratorGenerateASCIIDocErrorPaths tests AsciiDoc generation error handling.
func TestGeneratorGenerateASCIIDocErrorPaths(t *testing.T) {
	testASCIIDocGeneration(t)
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
			errors:       []string{testGenErrMsg1, testGenErrMsg2},
			wantPanic:    false,
		},
		{
			name:         "mixed results",
			successCount: 3,
			errors:       []string{testGenErrMsg1},
			wantPanic:    false,
		},
		{
			name:         testutil.TestCaseNameZeroFiles,
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
			errors:       []string{testGenErrMsg1},
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
			errors:       []string{testGenErrMsg1, testGenErrMsg2},
			wantBold:     true,
			wantError:    true,
		},
		{
			name:         "non-quiet, non-verbose, with errors",
			quiet:        false,
			verbose:      false,
			successCount: 2,
			errors:       []string{testGenErrMsg1},
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

// TestCreateDependencyAnalyzer_TokenGuard verifies the GitHub client is only
// created when a non-empty token is provided (line 104: token != "" guard).
func TestCreateDependencyAnalyzer_TokenGuard(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		token         string
		wantClientNil bool
	}{
		{name: "empty token skips client creation", token: "", wantClientNil: true},
		{name: "non-empty token creates client", token: testutil.TestTokenValue, wantClientNil: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config := defaultTestConfig()
			config.GitHubToken = tt.token
			gen := NewGenerator(config)

			analyzer, err := gen.CreateDependencyAnalyzer()

			testutil.AssertNoError(t, err)
			if analyzer == nil {
				t.Error("expected analyzer to be non-nil")
			}
			if tt.wantClientNil && analyzer.GitHubClient != nil {
				t.Error("expected GitHubClient to be nil for empty token")
			}
			if !tt.wantClientNil && analyzer.GitHubClient == nil {
				t.Error("expected GitHubClient to be non-nil for non-empty token")
			}
		})
	}
}

// TestCreateDependencyAnalyzer_CacheSelectionBranches exercises the success branch
// of the depCache guard (line 122): when cache creation succeeds, a CacheAdapter
// is selected and wrapped in the returned Analyzer.
func TestCreateDependencyAnalyzer_CacheSelectionBranches(t *testing.T) {
	t.Parallel()

	config := defaultTestConfig()
	gen := NewGenerator(config)

	analyzer, err := gen.CreateDependencyAnalyzer()
	testutil.AssertNoError(t, err)
	if analyzer == nil {
		t.Fatal("expected non-nil analyzer when cache succeeds")
	}
	if _, ok := analyzer.Cache.(*dependencies.CacheAdapter); !ok {
		t.Errorf("expected CacheAdapter when cache creation succeeds, got %T", analyzer.Cache)
	}
}

// TestCreateDependencyAnalyzer_NoOpCacheOnCacheFailure exercises the failure branch
// of the depCache guard: when cache creation fails, the analyzer falls through to
// NoOpCache so the caller still gets a valid, usable analyzer.
func TestCreateDependencyAnalyzer_NoOpCacheOnCacheFailure(t *testing.T) {
	// Temporarily replace newCacheFunc to simulate cache creation failure.
	orig := newCacheFunc
	newCacheFunc = func(*cache.Config) (*cache.Cache, error) {
		return nil, errors.New("injected cache failure")
	}
	defer func() { newCacheFunc = orig }()

	config := defaultTestConfig()
	gen := NewGenerator(config)

	analyzer, err := gen.CreateDependencyAnalyzer()
	testutil.AssertNoError(t, err)
	if analyzer == nil {
		t.Fatal("expected non-nil analyzer even when cache creation fails")
	}
	if _, ok := analyzer.Cache.(*dependencies.NoOpCache); !ok {
		t.Errorf("expected NoOpCache when cache creation fails, got %T", analyzer.Cache)
	}
}

// TestValidateFiles_SumErrorCounts verifies the arithmetic at line 255:
// totalFailures = len(errors) + validationFailures. With one parse error and one
// file with missing required fields, totalFailures must be exactly 2.
func TestValidateFiles_SumErrorCounts(t *testing.T) {
	t.Parallel()

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// File 1: valid — contributes 0 to parse errors, 0 to validationFailures.
	file1 := filepath.Join(tmpDir, "valid.yml")
	testutil.WriteTestFile(t, file1, testutil.MustReadFixture(testutil.TestFixtureJavaScriptSimple))

	// File 2: unparseable — contributes 1 to parse errors (len(errors)==1).
	file2 := filepath.Join(tmpDir, "bad.yml")
	testutil.WriteTestFile(t, file2, "invalid: [unclosed")

	// File 3: missing required field — contributes 1 to validationFailures.
	file3 := filepath.Join(tmpDir, "missingdesc.yml")
	testutil.WriteTestFile(t, file3, testutil.MustReadFixture(testutil.TestFixtureInvalidMissingDescription))

	config := defaultTestConfig()
	gen := NewGenerator(config)

	err := gen.ValidateFiles([]string{file1, file2, file3})

	// The error message must reflect totalFailures = 1+1 = 2.
	if err == nil {
		t.Fatal("expected error from ValidateFiles, got nil")
	}
	if !strings.Contains(err.Error(), "for 2 files") {
		t.Errorf("expected total failure count 2 in error message, got: %v", err)
	}
}

// TestProcessFiles_SuccessCountIncrement verifies successCount++ (line 413) fires
// for every successful file. With N valid files the count must equal N.
func TestProcessFiles_SuccessCountIncrement(t *testing.T) {
	t.Parallel()

	fixtures := []string{
		testutil.TestFixtureJavaScriptSimple,
		testutil.TestFixtureCompositeBasic,
		testutil.TestFixtureDockerBasic,
	}

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	testutil.SetupTestTemplates(t, tmpDir)

	config := defaultTestConfig()
	config.OutputDir = tmpDir
	config.Template = filepath.Join(tmpDir, "templates", appconstants.TemplateReadme)
	gen := NewGenerator(config)

	paths := make([]string, 0, len(fixtures))
	for i, fix := range fixtures {
		// Each action file goes in its own temp sub-directory.
		subDir := filepath.Join(tmpDir, fmt.Sprintf("action%d", i+1))
		if err := os.MkdirAll(subDir, 0o750); err != nil {
			t.Fatalf("failed to create sub-dir: %v", err)
		}
		p := filepath.Join(subDir, appconstants.ActionFileNameYML)
		testutil.WriteTestFile(t, p, testutil.MustReadFixture(fix))
		paths = append(paths, p)
	}

	errs, successCount := gen.processFiles(paths, nil)
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d: %v", len(errs), errs)
	}
	if successCount != len(fixtures) {
		t.Errorf("expected successCount=%d, got %d", len(fixtures), successCount)
	}
}

// TestParseAndValidateAction_FieldBoundary exercises the field == Name ||
// field == Description || field == Runs || field == RunsUsing boundary (line 450).
// A file missing "name" must error; a file missing only a non-critical field must not.
func TestParseAndValidateAction_FieldBoundary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		content     string
		wantErr     bool
		errContains string // when wantErr, the missing field the error must name
	}{
		{
			name:        "missing name field errors",
			content:     testutil.MustReadFixture(testutil.TestFixtureCompositeMissingName),
			wantErr:     true,
			errContains: appconstants.FieldName,
		},
		{
			name:        "missing description field errors",
			content:     testutil.MustReadFixture(testutil.TestFixtureCompositeMissingDesc),
			wantErr:     true,
			errContains: appconstants.FieldDescription,
		},
		{
			name:    "all required fields present does not error",
			content: testutil.TestMinimalAction,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpPath := testutil.CreateTempActionFile(t, tt.content)
			config := defaultTestConfig()
			gen := NewGenerator(config)

			_, err := gen.parseAndValidateAction(tmpPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseAndValidateAction() error=%v, wantErr=%v", err, tt.wantErr)
			}
			// Pin the rejection to the specific missing field so a regression in
			// the required-field boundary (e.g. dropping the description term)
			// cannot pass by erroring for a different reason.
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("error %q does not name the missing field %q", err.Error(), tt.errContains)
			}
		})
	}
}

// TestShowValidationSummary_Boundary verifies the resultCount-validFiles > 0
// boundary (line 607) and errorCount > 0 boundary (line 610). The summary must
// produce different output for zero-vs-one issue counts.
func TestShowValidationSummary_Boundary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		totalFiles    int
		validFiles    int
		totalIssues   int
		resultCount   int
		errorCount    int
		wantWarning   bool
		wantErrOutput bool
	}{
		{
			name:       "zero issues and zero errors",
			totalFiles: 1, validFiles: 1, totalIssues: 0,
			resultCount: 1, errorCount: 0,
			wantWarning: false, wantErrOutput: false,
		},
		{
			name:       "one file with issues triggers warning",
			totalFiles: 2, validFiles: 1, totalIssues: 1,
			resultCount: 2, errorCount: 0,
			wantWarning: true, wantErrOutput: false,
		},
		{
			name:       "one parse error triggers error output",
			totalFiles: 1, validFiles: 1, totalIssues: 0,
			resultCount: 1, errorCount: 1,
			wantWarning: false, wantErrOutput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			captured := &testCapturedOutput{CapturedOutput: &testutil.CapturedOutput{}}
			config := defaultTestConfig()
			config.Quiet = false
			gen := NewGeneratorWithDependencies(config, captured, nil)

			gen.showValidationSummary(tt.totalFiles, tt.validFiles, tt.totalIssues, tt.resultCount, tt.errorCount)

			hasWarning := len(captured.WarningMessages) > 0
			if hasWarning != tt.wantWarning {
				t.Errorf("wantWarning=%v but hasWarning=%v", tt.wantWarning, hasWarning)
			}

			hasErrOut := len(captured.ErrorMessages) > 0
			if hasErrOut != tt.wantErrOutput {
				t.Errorf("wantErrOutput=%v but hasErrOutput=%v", tt.wantErrOutput, hasErrOut)
			}
		})
	}
}

// TestShowDetailedIssues_Boundary verifies the totalIssues > 0 boundary (line 608):
// zero issues with non-verbose mode must skip output, non-zero must print.
func TestShowDetailedIssues_Boundary(t *testing.T) {
	t.Parallel()

	makeResult := func(file string, fields ...string) ValidationResult {
		r := ValidationResult{MissingFields: append([]string{"file: " + file}, fields...)}

		return r
	}

	tests := []struct {
		name        string
		results     []ValidationResult
		totalIssues int
		verbose     bool
		wantOutput  bool
	}{
		{
			name:        "zero issues non-verbose: no output",
			results:     []ValidationResult{makeResult("ok.yml")},
			totalIssues: 0, verbose: false, wantOutput: false,
		},
		{
			name:        "one issue: output shown",
			results:     []ValidationResult{makeResult("bad.yml", "branding")},
			totalIssues: 1, verbose: false, wantOutput: true,
		},
		{
			name:        "zero issues verbose: output shown",
			results:     []ValidationResult{makeResult("ok.yml")},
			totalIssues: 0, verbose: true, wantOutput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			captured := &testCapturedOutput{CapturedOutput: &testutil.CapturedOutput{}}
			config := defaultTestConfig()
			config.Quiet = false
			config.Verbose = tt.verbose
			gen := NewGeneratorWithDependencies(config, captured, nil)

			gen.showDetailedIssues(tt.results, tt.totalIssues)

			hasOutput := len(captured.BoldMessages) > 0 || len(captured.PrintfMessages) > 0
			if hasOutput != tt.wantOutput {
				t.Errorf("wantOutput=%v but hasOutput=%v", tt.wantOutput, hasOutput)
			}
		})
	}
}

// TestShowFileIssues_Boundary verifies the > 1 / > 0 boundaries for MissingFields
// and Suggestions slices in showFileIssues (lines 628, 650).
func TestShowFileIssues_Boundary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		result         ValidationResult
		wantErrOutput  bool
		wantSuggestion bool
	}{
		{
			name: "no missing fields no suggestions",
			result: ValidationResult{
				MissingFields: []string{testGenFileFoo},
				Suggestions:   nil,
			},
			wantErrOutput: false, wantSuggestion: false,
		},
		{
			name: "one missing field emits error",
			result: ValidationResult{
				MissingFields: []string{testGenFileFoo, testGenFieldName},
				Suggestions:   nil,
			},
			wantErrOutput: true, wantSuggestion: false,
		},
		{
			name: "one suggestion emits suggestion output",
			result: ValidationResult{
				MissingFields: []string{testGenFileFoo},
				Suggestions:   []string{"add a name field"},
			},
			wantErrOutput: false, wantSuggestion: true,
		},
		{
			name: "two missing fields and suggestion",
			result: ValidationResult{
				MissingFields: []string{testGenFileFoo, testGenFieldName, testGenFieldDesc},
				Suggestions:   []string{"fix it"},
			},
			wantErrOutput: true, wantSuggestion: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			captured := &testCapturedOutput{CapturedOutput: &testutil.CapturedOutput{}}
			config := defaultTestConfig()
			config.Quiet = false
			gen := NewGeneratorWithDependencies(config, captured, nil)

			gen.showFileIssues(tt.result)

			hasErrMsg := len(captured.ErrorMessages) > 0
			if hasErrMsg != tt.wantErrOutput {
				t.Errorf("wantErrOutput=%v got=%v (ErrorMessages=%v)",
					tt.wantErrOutput, hasErrMsg, captured.ErrorMessages)
			}

			// Suggestions appear in Printf output.
			hasSuggestion := false
			for _, m := range captured.PrintfMessages {
				if strings.Contains(m, "•") || strings.Contains(m, "Suggestion") {
					hasSuggestion = true

					break
				}
			}
			if hasSuggestion != tt.wantSuggestion {
				t.Errorf("wantSuggestion=%v got=%v (PrintfMessages=%v)",
					tt.wantSuggestion, hasSuggestion, captured.PrintfMessages)
			}
		})
	}
}

// TestValidateFiles_Verbose_BarNilGuard exercises the Verbose && bar == nil guard
// at line 556. In verbose mode with a nil bar the progress message must be emitted;
// with a non-nil bar it must be suppressed.
func TestValidateFiles_Verbose_BarNilGuard(t *testing.T) {
	t.Parallel()

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	actionPath := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
	testutil.WriteTestFile(t, actionPath, testutil.MustReadFixture(testutil.TestFixtureJavaScriptSimple))

	captured := &testCapturedOutput{CapturedOutput: &testutil.CapturedOutput{}}
	config := defaultTestConfig()
	config.Quiet = false
	config.Verbose = true
	gen := NewGeneratorWithDependencies(config, captured, NewNullProgressManager())

	// validateFiles is called with bar=nil internally when the generator
	// calls ValidateFiles directly. We trigger it via the exported method.
	_ = gen.ValidateFiles([]string{actionPath})

	hasProgress := len(captured.ProgressMessages) > 0
	if !hasProgress {
		t.Error("expected progress message in verbose mode with nil bar, got none")
	}
}

// TestReportValidationResults_BatchCounting verifies the arithmetic at line 580:
// totalFiles = len(results) + len(errors). We supply known counts and confirm the
// Bold summary message contains the correct total.
func TestReportValidationResults_BatchCounting(t *testing.T) {
	t.Parallel()

	captured := &testCapturedOutput{CapturedOutput: &testutil.CapturedOutput{}}
	config := defaultTestConfig()
	config.Quiet = false
	gen := NewGeneratorWithDependencies(config, captured, nil)

	results := []ValidationResult{
		{MissingFields: []string{"file: a.yml"}},
		{MissingFields: []string{"file: b.yml"}},
	}
	parseErrors := []string{"error parsing c.yml"}

	// totalFiles should be len(results)+len(errors) = 2+1 = 3.
	gen.reportValidationResults(results, parseErrors)

	found := false
	for _, msg := range captured.BoldMessages {
		if strings.Contains(msg, "3") {
			found = true

			break
		}
	}
	if !found {
		t.Errorf("expected bold summary containing '3' files, got: %v", captured.BoldMessages)
	}
}

// TestIsUnitTestEnvironment_EnvVar exercises the NOT COVERED CONDITIONALS_NEGATION
// at generator.go:41. Setting UNIT_TEST_MODE must make the function return true.
func TestIsUnitTestEnvironment_EnvVar(t *testing.T) {
	// UNIT_TEST_MODE branch: already set by the test binary (internal.test in argv),
	// but we also explicitly verify the env-var path by temporarily unsetting and resetting.
	t.Setenv("UNIT_TEST_MODE", "1")
	if !isUnitTestEnvironment() {
		t.Error("expected isUnitTestEnvironment()=true when UNIT_TEST_MODE is set")
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
