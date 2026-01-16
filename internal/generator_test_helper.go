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
