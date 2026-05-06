package internal

import (
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

const (
	testAnalyzerThemeDefault = appconstants.ThemeDefault
	testAnalyzerFormatMD     = appconstants.OutputFormatMarkdown
	testAnalyzerFakeToken    = "fake_token"
)

func TestCreateAnalyzer(t *testing.T) {
	t.Parallel()

	config := &AppConfig{
		Theme:        testAnalyzerThemeDefault,
		OutputFormat: testAnalyzerFormatMD,
		OutputDir:    ".",
		GitHubToken:  testAnalyzerFakeToken,
	}
	generator := NewGenerator(config)
	output := &ColoredOutput{NoColor: true}

	analyzer := CreateAnalyzer(generator, output)
	if analyzer == nil {
		t.Error("expected analyzer to be created, got nil")
	}
}

func TestCreateAnalyzer_WithoutToken(t *testing.T) {
	// Not parallel: uses t.Setenv to pin the environment to no-token state.
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_README_GITHUB_TOKEN", "")

	config := &AppConfig{
		Theme:        testAnalyzerThemeDefault,
		OutputFormat: testAnalyzerFormatMD,
		OutputDir:    ".",
	}
	generator := NewGenerator(config)
	output := &ColoredOutput{NoColor: true}

	// CreateDependencyAnalyzer succeeds without a token (skips GitHub client).
	analyzer := CreateAnalyzer(generator, output)
	if analyzer == nil {
		t.Error("expected analyzer to be created without a GitHub token, got nil")
	}
}

func TestCreateAnalyzerIntegration(t *testing.T) {
	t.Parallel()

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	config := &AppConfig{
		Theme:        testAnalyzerThemeDefault,
		OutputFormat: testAnalyzerFormatMD,
		OutputDir:    tmpDir,
		Quiet:        true,
		GitHubToken:  testAnalyzerFakeToken,
	}

	generator := NewGenerator(config)
	output := NewColoredOutput(true)

	analyzer := CreateAnalyzer(generator, output)
	if analyzer == nil {
		t.Error("expected analyzer to be created with a fake token, got nil")
	}
}
