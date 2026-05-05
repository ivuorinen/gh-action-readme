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

	tests := []struct {
		name           string
		setupConfig    func() *AppConfig
		expectAnalyzer bool
	}{
		{
			name: "successful analyzer creation with valid config",
			setupConfig: func() *AppConfig {
				return &AppConfig{
					Theme:        testAnalyzerThemeDefault,
					OutputFormat: testAnalyzerFormatMD,
					OutputDir:    ".",
					GitHubToken:  testAnalyzerFakeToken,
				}
			},
			expectAnalyzer: true,
		},
		{
			name: "analyzer creation without GitHub token",
			setupConfig: func() *AppConfig {
				return &AppConfig{
					Theme:        testAnalyzerThemeDefault,
					OutputFormat: testAnalyzerFormatMD,
					OutputDir:    ".",
				}
			},
			expectAnalyzer: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config := tt.setupConfig()
			generator := NewGenerator(config)
			output := &ColoredOutput{NoColor: true}

			analyzer := CreateAnalyzer(generator, output)

			if tt.expectAnalyzer && analyzer == nil {
				t.Error("expected analyzer to be created, got nil")
			}

			if !tt.expectAnalyzer && analyzer != nil {
				t.Error("expected analyzer to be nil, got non-nil")
			}
		})
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

	if analyzer != nil {
		t.Log("Analyzer created successfully")
	} else {
		t.Log("Analyzer creation failed - expected without valid GitHub token")
	}
}
