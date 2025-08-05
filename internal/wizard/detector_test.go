package wizard

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/internal"
)

func TestProjectDetector_analyzeProjectFiles(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	// Create test files (go.mod should be processed last to be the final language)
	testFiles := map[string]string{
		"Dockerfile":     "FROM alpine",
		"action.yml":     "name: Test Action",
		"next.config.js": "module.exports = {}",
		"package.json":   `{"name": "test", "version": "1.0.0"}`,
		"go.mod":         "module test", // This should be detected last
	}

	for filename, content := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0600); err != nil { // #nosec G306 -- test file permissions
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Create detector with temp directory
	output := internal.NewColoredOutput(true)
	detector := &ProjectDetector{
		output:     output,
		currentDir: tempDir,
	}

	characteristics := detector.analyzeProjectFiles()

	// Test that a language is detected (either Go or JavaScript/TypeScript is valid)
	language := characteristics["language"]
	if language != "Go" && language != "JavaScript/TypeScript" {
		t.Errorf("Expected language 'Go' or 'JavaScript/TypeScript', got '%s'", language)
	}

	// Test that appropriate type is detected
	projectType := characteristics["type"]
	validTypes := []string{"Go Module", "Node.js Project"}
	typeValid := false
	for _, validType := range validTypes {
		if projectType == validType {
			typeValid = true
			break
		}
	}
	if !typeValid {
		t.Errorf("Expected type to be one of %v, got '%s'", validTypes, projectType)
	}

	if characteristics["framework"] != "Next.js" {
		t.Errorf("Expected framework 'Next.js', got '%s'", characteristics["framework"])
	}
}

func TestProjectDetector_detectVersionFromPackageJSON(t *testing.T) {
	tempDir := t.TempDir()

	// Create package.json with version
	packageJSON := `{
		"name": "test-package",
		"version": "2.1.0",
		"description": "Test package"
	}`

	packagePath := filepath.Join(tempDir, "package.json")
	if err := os.WriteFile(packagePath, []byte(packageJSON), 0600); err != nil { // #nosec G306 -- test file permissions
		t.Fatalf("Failed to create package.json: %v", err)
	}

	output := internal.NewColoredOutput(true)
	detector := &ProjectDetector{
		output:     output,
		currentDir: tempDir,
	}

	version := detector.detectVersionFromPackageJSON()
	if version != "2.1.0" {
		t.Errorf("Expected version '2.1.0', got '%s'", version)
	}
}

func TestProjectDetector_detectVersionFromFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create VERSION file
	versionContent := "3.2.1\n"
	versionPath := filepath.Join(tempDir, "VERSION")
	if err := os.WriteFile(versionPath, []byte(versionContent), 0600); err != nil { // #nosec G306 -- test file permissions
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	output := internal.NewColoredOutput(true)
	detector := &ProjectDetector{
		output:     output,
		currentDir: tempDir,
	}

	version := detector.detectVersionFromFiles()
	if version != "3.2.1" {
		t.Errorf("Expected version '3.2.1', got '%s'", version)
	}
}

func TestProjectDetector_findActionFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create action files
	actionYML := filepath.Join(tempDir, "action.yml")
	if err := os.WriteFile(
		actionYML,
		[]byte("name: Test Action"),
		0600, // #nosec G306 -- test file permissions
	); err != nil {
		t.Fatalf("Failed to create action.yml: %v", err)
	}

	// Create subdirectory with another action file
	subDir := filepath.Join(tempDir, "subaction")
	if err := os.MkdirAll(subDir, 0750); err != nil { // #nosec G301 -- test directory permissions
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	subActionYAML := filepath.Join(subDir, "action.yaml")
	if err := os.WriteFile(
		subActionYAML,
		[]byte("name: Sub Action"),
		0600, // #nosec G306 -- test file permissions
	); err != nil {
		t.Fatalf("Failed to create sub action.yaml: %v", err)
	}

	output := internal.NewColoredOutput(true)
	detector := &ProjectDetector{
		output:     output,
		currentDir: tempDir,
	}

	// Test non-recursive
	files, err := detector.findActionFiles(tempDir, false)
	if err != nil {
		t.Fatalf("findActionFiles() error = %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 action file, got %d", len(files))
	}

	// Test recursive
	files, err = detector.findActionFiles(tempDir, true)
	if err != nil {
		t.Fatalf("findActionFiles() error = %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 action files, got %d", len(files))
	}
}

func TestProjectDetector_isActionFile(t *testing.T) {
	output := internal.NewColoredOutput(true)
	detector := &ProjectDetector{
		output: output,
	}

	tests := []struct {
		filename string
		expected bool
	}{
		{"action.yml", true},
		{"action.yaml", true},
		{"Action.yml", false},
		{"action.yml.bak", false},
		{"other.yml", false},
		{"readme.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := detector.isActionFile(tt.filename)
			if result != tt.expected {
				t.Errorf("isActionFile(%s) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestProjectDetector_suggestConfiguration(t *testing.T) {
	output := internal.NewColoredOutput(true)
	detector := &ProjectDetector{
		output: output,
	}

	tests := []struct {
		name     string
		settings *DetectedSettings
		expected string
	}{
		{
			name: "composite action",
			settings: &DetectedSettings{
				HasCompositeAction: true,
			},
			expected: "professional",
		},
		{
			name: "with dockerfile",
			settings: &DetectedSettings{
				HasDockerfile: true,
			},
			expected: "github",
		},
		{
			name: "go project",
			settings: &DetectedSettings{
				Language: "Go",
			},
			expected: "minimal",
		},
		{
			name: "with framework",
			settings: &DetectedSettings{
				Framework: "Next.js",
			},
			expected: "github",
		},
		{
			name:     "default case",
			settings: &DetectedSettings{},
			expected: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector.suggestConfiguration(tt.settings)
			if tt.settings.SuggestedTheme != tt.expected {
				t.Errorf("Expected theme %s, got %s", tt.expected, tt.settings.SuggestedTheme)
			}
		})
	}
}
