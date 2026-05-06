package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// testFormatGeneration is a generic helper for testing format generation methods.
// It consolidates the common pattern across HTML, JSON, and AsciiDoc generation tests.
func testFormatGeneration(
	t *testing.T,
	generateFunc func(*Generator, *ActionYML, string, string) error,
	expectedFile, formatName string,
	needsActionPath bool,
) {
	t.Helper()
	t.Parallel()

	tmpDir := t.TempDir()
	action := createTestAction()
	gen := createQuietGenerator()

	var err error
	if needsActionPath {
		actionPath := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
		err = generateFunc(gen, action, tmpDir, actionPath)
	} else {
		// For JSON which doesn't need actionPath
		err = generateFunc(gen, action, tmpDir, "")
	}

	if err != nil {
		t.Errorf("%s generation unexpected error = %v", formatName, err)
	}

	verifyFileExists(t, filepath.Join(tmpDir, expectedFile), expectedFile)
}

// testHTMLGeneration tests HTML generation creates the expected output file.
func testHTMLGeneration(t *testing.T) {
	t.Helper()

	testFormatGeneration(
		t,
		func(g *Generator, a *ActionYML, out, path string) error {
			return g.generateHTML(a, out, path)
		},
		"Test Action.html",
		"HTML",
		true, // needs actionPath
	)
}

// testJSONGeneration tests JSON generation creates the expected output file.
func testJSONGeneration(t *testing.T) {
	t.Helper()

	testFormatGeneration(
		t,
		func(g *Generator, a *ActionYML, out, _ string) error {
			return g.generateJSON(a, out)
		},
		"action-docs.json",
		"JSON",
		false, // doesn't need actionPath
	)
}

// testASCIIDocGeneration tests AsciiDoc generation creates the expected output file.
func testASCIIDocGeneration(t *testing.T) {
	t.Helper()

	testFormatGeneration(
		t,
		func(g *Generator, a *ActionYML, out, path string) error {
			return g.generateASCIIDoc(a, out, path)
		},
		"README.adoc",
		"AsciiDoc",
		true, // needs actionPath
	)
}

// createTestAction creates a basic test action for generator tests.
func createTestAction() *ActionYML {
	return &ActionYML{
		Name:        testutil.TestActionName,
		Description: testutil.TestActionDesc,
		Runs:        map[string]any{testGenRunsUsing: appconstants.ActionTypeComposite},
	}
}

// createQuietGenerator creates a generator with quiet output for testing.
func createQuietGenerator() *Generator {
	config := DefaultAppConfig()
	config.Quiet = true

	return NewGenerator(config)
}

// verifyFileExists checks that a file was created at the expected path.
func verifyFileExists(t *testing.T, fullPath, expectedFileName string) {
	t.Helper()

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Errorf("Expected %s to be created", expectedFileName)
	}
}

// createTestDirs creates multiple test directories with given names.
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

// createMultiActionSetup creates a setupFunc for batch processing tests with multiple actions.
// It generates separate directories for each action and writes the specified fixtures.
func createMultiActionSetup(dirNames, fixtures []string) func(t *testing.T, tmpDir string) []string {
	return func(t *testing.T, tmpDir string) []string {
		t.Helper()

		// Create separate directories for each action
		dirs := createTestDirs(t, tmpDir, dirNames...)

		// Build file paths and write fixtures
		files := make([]string, len(dirs))
		for i, dir := range dirs {
			files[i] = filepath.Join(dir, appconstants.ActionFileNameYML)
			testutil.WriteTestFile(t, files[i], testutil.MustReadFixture(fixtures[i]))
		}

		return files
	}
}

// setupNonexistentFiles returns a setupFunc that creates paths to nonexistent files.
// This is used in multiple tests to verify error handling for missing files.
func setupNonexistentFiles(filename string) func(*testing.T, string) []string {
	return func(_ *testing.T, tmpDir string) []string {
		return []string{filepath.Join(tmpDir, filename)}
	}
}

// validationSummaryTestCase defines a test case for validation summary tests.
// This helper reduces duplication in test case definitions by providing
// a factory function with sensible defaults.
type validationSummaryTestCase struct {
	name        string
	totalFiles  int
	validFiles  int
	totalIssues int
	resultCount int
	errorCount  int
	wantBold    int
	wantSuccess int
	wantWarning int
	wantError   int
	wantInfo    int
}

// validationSummaryParams holds parameters for creating validation summary test cases.
type validationSummaryParams struct {
	name                                                         string
	totalFiles, validFiles, totalIssues, resultCount, errorCount int
	wantWarning, wantError, wantInfo                             int
}

// createValidationSummaryTest creates a validation summary test case with defaults.
// Default values: wantBold=1, wantSuccess=1, wantWarning=0, wantError=0, wantInfo=0
// Only provide the fields that differ from defaults.
func createValidationSummaryTest(params validationSummaryParams) validationSummaryTestCase {
	return validationSummaryTestCase{
		name:        params.name,
		totalFiles:  params.totalFiles,
		validFiles:  params.validFiles,
		totalIssues: params.totalIssues,
		resultCount: params.resultCount,
		errorCount:  params.errorCount,
		wantBold:    1, // Always 1
		wantSuccess: 1, // Always 1
		wantWarning: params.wantWarning,
		wantError:   params.wantError,
		wantInfo:    params.wantInfo,
	}
}
