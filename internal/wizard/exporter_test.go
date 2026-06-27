package wizard

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrg/xdg"
	"github.com/goccy/go-yaml"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

const (
	testWizardOrg      = testutil.WizardOrgTest
	testWizardRepo     = testutil.WizardRepoTest
	testWizardVersion  = "1.0.0"
	testWizardThemeGH  = appconstants.ThemeGitHub
	testWizardFormatMD = appconstants.OutputFormatMarkdown
	testWizardThemeDef = appconstants.ThemeDefault
)

func TestConfigExporterExportConfig(t *testing.T) {
	t.Parallel()
	output := internal.NewColoredOutput(true) // quiet mode for testing
	exporter := NewConfigExporter(output)

	// Create test config
	config := createTestConfig()

	// Test YAML export
	t.Run("export YAML", func(t *testing.T) {
		t.Parallel()
		testYAMLExport(exporter, config)(t)
	})

	// Test JSON export
	t.Run("export JSON", func(t *testing.T) {
		t.Parallel()
		testJSONExport(exporter, config)(t)
	})

	// Test TOML export
	t.Run("export TOML", func(t *testing.T) {
		t.Parallel()
		testTOMLExport(exporter, config)(t)
	})
}

// createTestConfig creates a test configuration for testing.
func createTestConfig() *internal.AppConfig {
	return &internal.AppConfig{
		Organization:        testWizardOrg,
		Repository:          testWizardRepo,
		Version:             testWizardVersion,
		Theme:               testWizardThemeGH,
		OutputFormat:        testWizardFormatMD,
		OutputDir:           ".",
		AnalyzeDependencies: true,
		ShowSecurityInfo:    false,
		Variables:           map[string]string{"TEST_VAR": "test_value"},
		Permissions:         map[string]string{appconstants.PermScopeContents: appconstants.PermissionRead},
		RunsOn:              []string{"ubuntu-latest"},
	}
}

// testYAMLExport tests YAML export functionality.
func testYAMLExport(exporter *ConfigExporter, config *internal.AppConfig) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()
		tempDir := t.TempDir()
		outputPath := filepath.Join(tempDir, testutil.TestFileConfigYAML)

		err := exporter.ExportConfig(config, FormatYAML, outputPath)
		if err != nil {
			t.Fatalf(testutil.TestMsgExportConfigError, err)
		}

		testutil.AssertFileExists(t, outputPath)
		verifyYAMLContent(t, outputPath, config)
	}
}

// testJSONExport tests JSON export functionality.
func testJSONExport(exporter *ConfigExporter, config *internal.AppConfig) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()
		tempDir := t.TempDir()
		outputPath := filepath.Join(tempDir, "config.json")

		err := exporter.ExportConfig(config, FormatJSON, outputPath)
		if err != nil {
			t.Fatalf(testutil.TestMsgExportConfigError, err)
		}

		testutil.AssertFileExists(t, outputPath)
		verifyJSONContent(t, outputPath, config)
	}
}

// testTOMLExport tests TOML export functionality.
func testTOMLExport(exporter *ConfigExporter, config *internal.AppConfig) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()
		tempDir := t.TempDir()
		outputPath := filepath.Join(tempDir, "config.toml")

		err := exporter.ExportConfig(config, FormatTOML, outputPath)
		if err != nil {
			t.Fatalf(testutil.TestMsgExportConfigError, err)
		}

		testutil.AssertFileExists(t, outputPath)
		verifyTOMLContent(t, outputPath)
	}
}

// verifyYAMLContent verifies YAML content is valid and contains expected data.
func verifyYAMLContent(t *testing.T, outputPath string, expected *internal.AppConfig) {
	t.Helper()
	data, err := os.ReadFile(outputPath) // #nosec G304 -- test output path
	if err != nil {
		t.Fatalf(testutil.TestMsgFailedReadOutput, err)
	}

	var yamlConfig internal.AppConfig
	if err := yaml.Unmarshal(data, &yamlConfig); err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	if yamlConfig.Organization != expected.Organization {
		t.Errorf("Organization = %v, want %v", yamlConfig.Organization, expected.Organization)
	}
	if yamlConfig.Theme != expected.Theme {
		t.Errorf("Theme = %v, want %v", yamlConfig.Theme, expected.Theme)
	}
}

// verifyJSONContent verifies JSON content is valid and contains expected data.
func verifyJSONContent(t *testing.T, outputPath string, expected *internal.AppConfig) {
	t.Helper()
	data, err := os.ReadFile(outputPath) // #nosec G304 -- test output path
	if err != nil {
		t.Fatalf(testutil.TestMsgFailedReadOutput, err)
	}

	var jsonConfig internal.AppConfig
	if err := json.Unmarshal(data, &jsonConfig); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if jsonConfig.Repository != expected.Repository {
		t.Errorf("Repository = %v, want %v", jsonConfig.Repository, expected.Repository)
	}
	if jsonConfig.OutputFormat != expected.OutputFormat {
		t.Errorf("OutputFormat = %v, want %v", jsonConfig.OutputFormat, expected.OutputFormat)
	}
}

// verifyTOMLContent verifies TOML content contains expected fields.
func verifyTOMLContent(t *testing.T, outputPath string) {
	t.Helper()
	data, err := os.ReadFile(outputPath) // #nosec G304 -- test output path
	if err != nil {
		t.Fatalf(testutil.TestMsgFailedReadOutput, err)
	}

	content := string(data)
	if !strings.Contains(content, `organization = "`+testWizardOrg+`"`) {
		t.Error("TOML should contain organization field")
	}
	if !strings.Contains(content, `theme = "`+testWizardThemeGH+`"`) {
		t.Error("TOML should contain theme field")
	}
}

func TestConfigExporterSanitizeConfig(t *testing.T) {
	t.Parallel()
	output := internal.NewColoredOutput(true)
	exporter := NewConfigExporter(output)

	config := &internal.AppConfig{
		Organization: testWizardOrg,
		Repository:   testWizardRepo,
		GitHubToken:  "ghp_secret_token",
		RepoOverrides: map[string]internal.AppConfig{
			"test/repo": {Theme: testWizardThemeGH},
		},
	}

	sanitized := exporter.sanitizeConfig(config)

	// Verify sensitive data is removed
	if sanitized.GitHubToken != "" {
		t.Error("Expected GitHubToken to be empty after sanitization")
	}

	if sanitized.RepoOverrides != nil {
		t.Error("Expected RepoOverrides to be nil after sanitization")
	}

	// Verify non-sensitive data is preserved
	if sanitized.Organization != config.Organization {
		t.Errorf("Organization = %v, want %v", sanitized.Organization, config.Organization)
	}
	if sanitized.Repository != config.Repository {
		t.Errorf("Repository = %v, want %v", sanitized.Repository, config.Repository)
	}
}

func TestConfigExporterGetSupportedFormats(t *testing.T) {
	t.Parallel()
	output := internal.NewColoredOutput(true)
	exporter := NewConfigExporter(output)

	formats := exporter.GetSupportedFormats()

	expectedFormats := []ExportFormat{FormatYAML, FormatJSON, FormatTOML}
	if len(formats) != len(expectedFormats) {
		t.Errorf("GetSupportedFormats() returned %d formats, want %d", len(formats), len(expectedFormats))
	}

	// Check that all expected formats are present
	formatMap := make(map[ExportFormat]bool)
	for _, format := range formats {
		formatMap[format] = true
	}

	for _, expected := range expectedFormats {
		if !formatMap[expected] {
			t.Errorf("Expected format %v not found in supported formats", expected)
		}
	}
}

func TestConfigExporterGetDefaultOutputPath(t *testing.T) {
	t.Parallel()
	output := internal.NewColoredOutput(true)
	exporter := NewConfigExporter(output)

	tests := []struct {
		format   ExportFormat
		expected string
	}{
		{FormatYAML, testutil.TestFileConfigYAML},
		{FormatJSON, "config.json"},
		{FormatTOML, "config.toml"},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			t.Parallel()
			path, err := exporter.GetDefaultOutputPath(tt.format)
			if err != nil {
				t.Fatalf("GetDefaultOutputPath() error = %v", err)
			}

			if !strings.HasSuffix(path, tt.expected) {
				t.Errorf("GetDefaultOutputPath() = %v, should end with %v", path, tt.expected)
			}
		})
	}

	// Test invalid format
	t.Run("invalid format", func(t *testing.T) {
		_, err := exporter.GetDefaultOutputPath("invalid")
		if err == nil {
			t.Error("Expected error for invalid format")
		}
	})
}

func TestExportConfigRefusesLiveConfigOverwrite(t *testing.T) {
	// Point XDG at a sandbox dir so GetConfigPath() never resolves to the user's
	// real config. xdg caches its dirs at init, so Reload() after Setenv is
	// required. The cleanup is registered before Setenv so it runs AFTER the env
	// is restored (LIFO), reloading xdg back to the real dirs to avoid leaking the
	// temp path into other tests. Mutating global xdg state precludes t.Parallel.
	t.Cleanup(xdg.Reload)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "xdg"))
	xdg.Reload()

	cfgPath, err := internal.GetConfigPath()
	testutil.AssertNoError(t, err)

	// Create the live config on disk so the overwrite guard has something to protect.
	testutil.AssertNoError(t, os.MkdirAll(filepath.Dir(cfgPath), appconstants.FilePermDir))
	const liveContent = "theme: github\n"
	testutil.AssertNoError(t, os.WriteFile(cfgPath, []byte(liveContent), appconstants.FilePermDefault))

	exporter := NewConfigExporter(internal.NewColoredOutput(true))
	err = exporter.ExportConfig(createTestConfig(), FormatYAML, cfgPath)
	if err == nil {
		t.Fatal("expected ExportConfig to refuse overwriting the live config")
	}
	if !strings.Contains(err.Error(), "refusing to overwrite the live config") {
		t.Errorf("unexpected error: %v", err)
	}

	// The guard must reject before any write — original content intact.
	got, readErr := os.ReadFile(cfgPath) // #nosec G304 -- cfgPath from GetConfigPath under a test-controlled XDG dir
	testutil.AssertNoError(t, readErr)
	if string(got) != liveContent {
		t.Errorf("live config was modified; got %q, want %q", string(got), liveContent)
	}
}

func TestExportConfigRejectsTraversal(t *testing.T) {
	// Isolate the filesystem so a regressed guard cannot write outside the test
	// sandbox or depend on the ambient working directory. t.Chdir disallows
	// t.Parallel, which is fine here.
	t.Chdir(t.TempDir())

	exporter := NewConfigExporter(internal.NewColoredOutput(true))
	config := createTestConfig()

	const traversalPath = "../../etc/evil.yaml"
	err := exporter.ExportConfig(config, FormatYAML, traversalPath)
	if err == nil {
		t.Fatal("expected ExportConfig to reject a path containing '..'")
	}
	if !strings.Contains(err.Error(), "..") {
		t.Errorf("expected a traversal-rejection error mentioning '..', got: %v", err)
	}
	// The guard must reject before any write; the target must not exist.
	if _, statErr := os.Stat(traversalPath); !os.IsNotExist(statErr) {
		t.Errorf("traversal target should not have been created; os.Stat err = %v", statErr)
	}
}
