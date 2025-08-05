package wizard

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/ivuorinen/gh-action-readme/internal"
)

func TestConfigExporter_ExportConfig(t *testing.T) {
	output := internal.NewColoredOutput(true) // quiet mode for testing
	exporter := NewConfigExporter(output)

	// Create test config
	config := createTestConfig()

	// Test YAML export
	t.Run("export YAML", testYAMLExport(exporter, config))

	// Test JSON export
	t.Run("export JSON", testJSONExport(exporter, config))

	// Test TOML export
	t.Run("export TOML", testTOMLExport(exporter, config))
}

// createTestConfig creates a test configuration for testing.
func createTestConfig() *internal.AppConfig {
	return &internal.AppConfig{
		Organization:        "testorg",
		Repository:          "testrepo",
		Version:             "1.0.0",
		Theme:               "github",
		OutputFormat:        "md",
		OutputDir:           ".",
		AnalyzeDependencies: true,
		ShowSecurityInfo:    false,
		Variables:           map[string]string{"TEST_VAR": "test_value"},
		Permissions:         map[string]string{"contents": "read"},
		RunsOn:              []string{"ubuntu-latest"},
	}
}

// testYAMLExport tests YAML export functionality.
func testYAMLExport(exporter *ConfigExporter, config *internal.AppConfig) func(*testing.T) {
	return func(t *testing.T) {
		tempDir := t.TempDir()
		outputPath := filepath.Join(tempDir, "config.yaml")

		err := exporter.ExportConfig(config, FormatYAML, outputPath)
		if err != nil {
			t.Fatalf("ExportConfig() error = %v", err)
		}

		verifyFileExists(t, outputPath)
		verifyYAMLContent(t, outputPath, config)
	}
}

// testJSONExport tests JSON export functionality.
func testJSONExport(exporter *ConfigExporter, config *internal.AppConfig) func(*testing.T) {
	return func(t *testing.T) {
		tempDir := t.TempDir()
		outputPath := filepath.Join(tempDir, "config.json")

		err := exporter.ExportConfig(config, FormatJSON, outputPath)
		if err != nil {
			t.Fatalf("ExportConfig() error = %v", err)
		}

		verifyFileExists(t, outputPath)
		verifyJSONContent(t, outputPath, config)
	}
}

// testTOMLExport tests TOML export functionality.
func testTOMLExport(exporter *ConfigExporter, config *internal.AppConfig) func(*testing.T) {
	return func(t *testing.T) {
		tempDir := t.TempDir()
		outputPath := filepath.Join(tempDir, "config.toml")

		err := exporter.ExportConfig(config, FormatTOML, outputPath)
		if err != nil {
			t.Fatalf("ExportConfig() error = %v", err)
		}

		verifyFileExists(t, outputPath)
		verifyTOMLContent(t, outputPath)
	}
}

// verifyFileExists checks that a file exists at the given path.
func verifyFileExists(t *testing.T, outputPath string) {
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("Expected output file to exist")
	}
}

// verifyYAMLContent verifies YAML content is valid and contains expected data.
func verifyYAMLContent(t *testing.T, outputPath string, expected *internal.AppConfig) {
	data, err := os.ReadFile(outputPath) // #nosec G304 -- test output path
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
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
	data, err := os.ReadFile(outputPath) // #nosec G304 -- test output path
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
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
	data, err := os.ReadFile(outputPath) // #nosec G304 -- test output path
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, `organization = "testorg"`) {
		t.Error("TOML should contain organization field")
	}
	if !strings.Contains(content, `theme = "github"`) {
		t.Error("TOML should contain theme field")
	}
}

func TestConfigExporter_sanitizeConfig(t *testing.T) {
	output := internal.NewColoredOutput(true)
	exporter := NewConfigExporter(output)

	config := &internal.AppConfig{
		Organization: "testorg",
		Repository:   "testrepo",
		GitHubToken:  "ghp_secret_token",
		RepoOverrides: map[string]internal.AppConfig{
			"test/repo": {Theme: "github"},
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

func TestConfigExporter_GetSupportedFormats(t *testing.T) {
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

func TestConfigExporter_GetDefaultOutputPath(t *testing.T) {
	output := internal.NewColoredOutput(true)
	exporter := NewConfigExporter(output)

	tests := []struct {
		format   ExportFormat
		expected string
	}{
		{FormatYAML, "config.yaml"},
		{FormatJSON, "config.json"},
		{FormatTOML, "config.toml"},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
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
