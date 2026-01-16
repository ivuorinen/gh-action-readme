package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// testHTMLGeneration tests HTML generation creates the expected output file.
func testHTMLGeneration(t *testing.T) {
	t.Helper()
	t.Parallel()

	tmpDir := t.TempDir()
	action := createTestAction()
	gen := createQuietGenerator()

	actionPath := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
	err := gen.generateHTML(action, tmpDir, actionPath)
	if err != nil {
		t.Errorf("generateHTML() unexpected error = %v", err)
	}

	// HTML filename is based on action.Name + ".html"
	verifyFileExists(t, filepath.Join(tmpDir, "Test Action.html"), "Test Action.html")
}

// testJSONGeneration tests JSON generation creates the expected output file.
func testJSONGeneration(t *testing.T) {
	t.Helper()
	t.Parallel()

	tmpDir := t.TempDir()
	action := createTestAction()
	gen := createQuietGenerator()

	err := gen.generateJSON(action, tmpDir)
	if err != nil {
		t.Errorf("generateJSON() unexpected error = %v", err)
	}

	verifyFileExists(t, filepath.Join(tmpDir, "action-docs.json"), "action-docs.json")
}

// testASCIIDocGeneration tests AsciiDoc generation creates the expected output file.
func testASCIIDocGeneration(t *testing.T) {
	t.Helper()
	t.Parallel()

	tmpDir := t.TempDir()
	action := createTestAction()
	gen := createQuietGenerator()

	actionPath := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
	err := gen.generateASCIIDoc(action, tmpDir, actionPath)
	if err != nil {
		t.Errorf("generateASCIIDoc() unexpected error = %v", err)
	}

	verifyFileExists(t, filepath.Join(tmpDir, "README.adoc"), "README.adoc")
}

// createTestAction creates a basic test action for generator tests.
func createTestAction() *ActionYML {
	return &ActionYML{
		Name:        testutil.TestActionName,
		Description: testutil.TestActionDesc,
		Runs:        map[string]any{"using": "composite"},
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
