package wizard

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

func TestProjectDetectorAnalyzeProjectFiles(t *testing.T) {
	t.Parallel()
	// Create temporary directory for testing
	tempDir := t.TempDir()

	// Create test files (go.mod should be processed last to be the final language)
	testFiles := map[string]string{
		"Dockerfile":                   "FROM alpine",
		appconstants.ActionFileNameYML: "name: Test Action",
		"next.config.js":               "module.exports = {}",
		appconstants.PackageJSON:       `{"name": "test", "version": "1.0.0"}`,
		"go.mod":                       "module test", // This should be detected last
	}

	for filename, content := range testFiles {
		testutil.WriteFileInDir(t, tempDir, filename, content)
	}

	// Create detector with temp directory
	output := internal.NewColoredOutput(true)
	detector := &ProjectDetector{
		output:     output,
		currentDir: tempDir,
	}

	characteristics := detector.analyzeProjectFiles()

	// Test that a language is detected (either Go or testutil.TestLangJavaScriptTypeScript is valid)
	language := characteristics["language"]
	if language != "Go" && language != testutil.TestLangJavaScriptTypeScript {
		t.Errorf("Expected language 'Go' or '%s', got '%s'", testutil.TestLangJavaScriptTypeScript, language)
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

func TestProjectDetectorDetectVersionFromPackageJSON(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()

	// Create package.json with version
	packageJSON := `{
		"name": "test-package",
		"version": "2.1.0",
		"description": "Test package"
	}`

	testutil.WriteFileInDir(t, tempDir, appconstants.PackageJSON, packageJSON)

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

func TestProjectDetectorDetectVersionFromFiles(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()

	// Create VERSION file
	versionContent := "3.2.1\n"
	testutil.WriteFileInDir(t, tempDir, "VERSION", versionContent)

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

func TestProjectDetectorFindActionFiles(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()

	// Create action files
	actionYML := filepath.Join(tempDir, appconstants.ActionFileNameYML)
	testutil.WriteTestFile(t, actionYML, "name: Test Action")

	// Create subdirectory with another action file
	subDir := filepath.Join(tempDir, "subaction")
	testutil.CreateTestDir(t, subDir)

	subActionYAML := filepath.Join(subDir, "action.yaml")
	testutil.WriteTestFile(t, subActionYAML, "name: Sub Action")

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

func TestProjectDetectorIsActionFile(t *testing.T) {
	t.Parallel()
	output := internal.NewColoredOutput(true)
	detector := &ProjectDetector{
		output: output,
	}

	tests := []struct {
		filename string
		expected bool
	}{
		{appconstants.ActionFileNameYML, true},
		{"action.yaml", true},
		{"Action.yml", false},
		{"action.yml.bak", false},
		{"other.yml", false},
		{"readme.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			t.Parallel()
			result := detector.isActionFile(tt.filename)
			if result != tt.expected {
				t.Errorf("isActionFile(%s) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestProjectDetectorSuggestConfiguration(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			detector.suggestConfiguration(tt.settings)
			if tt.settings.SuggestedTheme != tt.expected {
				t.Errorf("Expected theme %s, got %s", tt.expected, tt.settings.SuggestedTheme)
			}
		})
	}
}

// TestProjectDetectorSuggestRunsOn tests the runner suggestion logic.
func TestProjectDetectorSuggestRunsOn(t *testing.T) {
	t.Parallel()
	output := internal.NewColoredOutput(true)
	detector := &ProjectDetector{
		output: output,
	}

	tests := []struct {
		name     string
		settings *DetectedSettings
		expected []string
	}{
		{
			name: "javascript/typescript project",
			settings: &DetectedSettings{
				Language:        testutil.TestLangJavaScriptTypeScript,
				SuggestedRunsOn: []string{testutil.RunnerUbuntuLatest},
			},
			expected: []string{
				testutil.RunnerUbuntuLatest,
				testutil.RunnerWindowsLatest,
				testutil.RunnerMacosLatest,
			},
		},
		{
			name: "go project",
			settings: &DetectedSettings{
				Language:        "Go",
				SuggestedRunsOn: []string{testutil.RunnerUbuntuLatest},
			},
			expected: []string{testutil.RunnerUbuntuLatest},
		},
		{
			name: "python project",
			settings: &DetectedSettings{
				Language:        "Python",
				SuggestedRunsOn: []string{testutil.RunnerUbuntuLatest},
			},
			expected: []string{testutil.RunnerUbuntuLatest},
		},
		{
			name: "already has multiple runners",
			settings: &DetectedSettings{
				Language:        testutil.TestLangJavaScriptTypeScript,
				SuggestedRunsOn: []string{testutil.RunnerUbuntuLatest, "custom-runner"},
			},
			expected: []string{testutil.RunnerUbuntuLatest, "custom-runner"},
		},
		{
			name: "unknown language",
			settings: &DetectedSettings{
				Language:        "Rust",
				SuggestedRunsOn: []string{testutil.RunnerUbuntuLatest},
			},
			expected: []string{testutil.RunnerUbuntuLatest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			detector.suggestRunsOn(tt.settings)

			if len(tt.settings.SuggestedRunsOn) != len(tt.expected) {
				t.Errorf("Expected %d runners, got %d", len(tt.expected), len(tt.settings.SuggestedRunsOn))

				return
			}

			for i, expectedRunner := range tt.expected {
				if tt.settings.SuggestedRunsOn[i] != expectedRunner {
					t.Errorf("Expected runner at index %d to be %s, got %s",
						i, expectedRunner, tt.settings.SuggestedRunsOn[i])
				}
			}
		})
	}
}

// assertPermissionsMatch is a helper to validate permissions in tests.
func assertPermissionsMatch(t *testing.T, expected, actual map[string]string) {
	t.Helper()

	if expected == nil && actual != nil {
		t.Errorf("Expected nil permissions, got %v", actual)

		return
	}

	if expected != nil && actual == nil {
		t.Errorf("Expected permissions %v, got nil", expected)

		return
	}

	if expected == nil {
		return
	}

	if len(actual) != len(expected) {
		t.Errorf("Expected %d permissions, got %d", len(expected), len(actual))

		return
	}

	for key, expectedValue := range expected {
		if actualValue, ok := actual[key]; !ok {
			t.Errorf("Expected permission %s not found", key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected permission %s=%s, got %s=%s",
				key, expectedValue, key, actualValue)
		}
	}
}

// TestProjectDetectorSuggestPermissions tests the permissions suggestion logic.
func TestProjectDetectorSuggestPermissions(t *testing.T) {
	t.Parallel()
	output := internal.NewColoredOutput(true)
	detector := &ProjectDetector{
		output: output,
	}

	tests := []struct {
		name     string
		settings *DetectedSettings
		expected map[string]string
	}{
		{
			name: "github action without permissions",
			settings: &DetectedSettings{
				IsGitHubAction:       true,
				SuggestedPermissions: nil,
			},
			expected: map[string]string{
				"contents": "read",
			},
		},
		{
			name: "github action with existing permissions",
			settings: &DetectedSettings{
				IsGitHubAction: true,
				SuggestedPermissions: map[string]string{
					"contents": "write",
					"issues":   "read",
				},
			},
			expected: map[string]string{
				"contents": "write",
				"issues":   "read",
			},
		},
		{
			name: "not a github action",
			settings: &DetectedSettings{
				IsGitHubAction:       false,
				SuggestedPermissions: nil,
			},
			expected: nil,
		},
		{
			name: "github action with empty permissions map",
			settings: &DetectedSettings{
				IsGitHubAction:       true,
				SuggestedPermissions: map[string]string{},
			},
			expected: map[string]string{
				"contents": "read",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			detector.suggestPermissions(tt.settings)
			assertPermissionsMatch(t, tt.expected, tt.settings.SuggestedPermissions)
		})
	}
}

// TestNewProjectDetector tests creating a new project detector.
func TestNewProjectDetector(t *testing.T) {
	t.Parallel()

	output := internal.NewColoredOutput(true)
	detector, err := NewProjectDetector(output)
	if err != nil {
		t.Fatalf("NewProjectDetector() error = %v", err)
	}

	if detector == nil {
		t.Fatal("NewProjectDetector() returned nil")
	}

	if detector.output == nil {
		t.Error("detector.output is nil")
	}

	if detector.currentDir == "" {
		t.Error("detector.currentDir is empty")
	}
}

// TestDetectProjectSettingsIntegration tests the main detection logic.
func TestDetectProjectSettingsIntegration(t *testing.T) {
	// Cannot use t.Parallel() because this test uses t.Chdir()

	// Create a temporary directory with test files
	tempDir := t.TempDir()

	// Create action.yml
	testutil.WriteActionFixture(t, tempDir, testutil.TestFixtureCompositeWithShellStep)

	// Change to temp directory (cleanup automatic via t.Chdir)
	t.Chdir(tempDir)

	output := internal.NewColoredOutput(true)
	detector, err := NewProjectDetector(output)
	if err != nil {
		t.Fatalf("NewProjectDetector() error = %v", err)
	}

	settings, err := detector.DetectProjectSettings()
	if err != nil {
		t.Fatalf("DetectProjectSettings() error = %v", err)
	}

	if settings == nil {
		t.Fatal("DetectProjectSettings() returned nil")
	}

	// Verify action file was detected
	if !settings.IsGitHubAction {
		t.Error("Expected IsGitHubAction to be true")
	}

	if len(settings.ActionFiles) == 0 {
		t.Error("Expected at least one action file to be detected")
	}

	// Verify default values are set
	if len(settings.SuggestedRunsOn) == 0 {
		t.Error("Expected SuggestedRunsOn to have default values")
	}

	if settings.SuggestedPermissions == nil {
		t.Error("Expected SuggestedPermissions to be initialized")
	}
}

// TestDetectRepositoryInfo tests repository info detection.
func TestDetectRepositoryInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		repoRoot string
		wantErr  bool
	}{
		{
			name:     "no git repository",
			repoRoot: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := internal.NewColoredOutput(true)
			detector := &ProjectDetector{
				output:   output,
				repoRoot: tt.repoRoot,
			}

			settings := &DetectedSettings{
				SuggestedPermissions: make(map[string]string),
			}

			err := detector.detectRepositoryInfo(settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("detectRepositoryInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDetectActionFiles tests action file detection.
//
// validateDetectActionFilesResult validates the results of detectActionFiles call.
func validateDetectActionFilesResult(
	t *testing.T,
	settings *DetectedSettings,
	err error,
	wantActionCount int,
	wantErr bool,
) {
	t.Helper()

	if (err != nil) != wantErr {
		t.Errorf("detectActionFiles() error = %v, wantErr %v", err, wantErr)
	}

	if len(settings.ActionFiles) != wantActionCount {
		t.Errorf("Expected %d action files, got %d", wantActionCount, len(settings.ActionFiles))
	}

	if wantActionCount > 0 && !settings.IsGitHubAction {
		t.Error("Expected IsGitHubAction to be true")
	}
}

func TestDetectActionFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		setupFunc       func(t *testing.T, dir string)
		wantActionCount int
		wantErr         bool
	}{
		{
			name: "detects action file",
			setupFunc: func(t *testing.T, dir string) {
				t.Helper()
				content := "name: Test Action\ndescription: Test"
				testutil.WriteFileInDir(t, dir, appconstants.ActionFileNameYML, content)
			},
			wantActionCount: 1,
			wantErr:         false,
		},
		{
			name: "no action files",
			setupFunc: func(t *testing.T, _ string) {
				t.Helper()
				// Don't create any files
			},
			wantActionCount: 0,
			wantErr:         false,
		},
		{
			name: "skips symlink to sensitive file",
			setupFunc: func(t *testing.T, dir string) {
				t.Helper()
				// Create symlink: action.yml -> /etc/passwd
				symlinkPath := filepath.Join(dir, appconstants.ActionFileNameYML)
				err := os.Symlink("/etc/passwd", symlinkPath)
				if err != nil {
					t.Skip("symlink creation not supported on this platform")
				}
			},
			wantActionCount: 0, // Should skip symlinks for security
			wantErr:         false,
		},
		{
			name: "handles directory with .. components safely",
			setupFunc: func(t *testing.T, dir string) {
				t.Helper()
				// Create subdirectory with action.yml
				content := "name: Test\ndescription: Test"
				testutil.CreateNestedAction(t, dir, "subdir", content)
			},
			wantActionCount: 1, // Should find the file safely
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			if tt.setupFunc != nil {
				tt.setupFunc(t, tempDir)
			}

			output := internal.NewColoredOutput(true)
			detector := &ProjectDetector{
				output:     output,
				currentDir: tempDir,
			}

			settings := &DetectedSettings{
				SuggestedPermissions: make(map[string]string),
			}

			err := detector.detectActionFiles(settings)

			validateDetectActionFilesResult(t, settings, err, tt.wantActionCount, tt.wantErr)
		})
	}
}

// TestDetectProjectCharacteristics tests project characteristics detection.
func TestDetectProjectCharacteristics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFunc      func(t *testing.T, dir string)
		wantDockerfile bool
	}{
		{
			name: "detects Dockerfile",
			setupFunc: func(t *testing.T, dir string) {
				t.Helper()
				content := "FROM alpine:latest"
				testutil.WriteFileInDir(t, dir, "Dockerfile", content)
			},
			wantDockerfile: true,
		},
		{
			name: "no Dockerfile",
			setupFunc: func(t *testing.T, _ string) {
				t.Helper()
				// Don't create Dockerfile
			},
			wantDockerfile: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			if tt.setupFunc != nil {
				tt.setupFunc(t, tempDir)
			}

			output := internal.NewColoredOutput(true)
			detector := &ProjectDetector{
				output:     output,
				currentDir: tempDir,
			}

			settings := &DetectedSettings{
				SuggestedPermissions: make(map[string]string),
			}

			detector.detectProjectCharacteristics(settings)

			if settings.HasDockerfile != tt.wantDockerfile {
				t.Errorf("HasDockerfile = %v, want %v", settings.HasDockerfile, tt.wantDockerfile)
			}
		})
	}
}

// TestDetectVersion tests version detection from various sources.
func TestDetectVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFunc func(t *testing.T, dir string)
		want      string
	}{
		{
			name: "detects version from package.json",
			setupFunc: func(t *testing.T, dir string) {
				t.Helper()
				content := `{"version": "1.2.3"}`
				testutil.WriteFileInDir(t, dir, appconstants.PackageJSON, content)
			},
			want: "1.2.3",
		},
		{
			name: "detects version from VERSION file",
			setupFunc: func(t *testing.T, dir string) {
				t.Helper()
				content := "2.0.0\n"
				testutil.WriteFileInDir(t, dir, "VERSION", content)
			},
			want: "2.0.0",
		},
		{
			name: "no version found",
			setupFunc: func(t *testing.T, _ string) {
				t.Helper()
				// Don't create version files
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			if tt.setupFunc != nil {
				tt.setupFunc(t, tempDir)
			}

			output := internal.NewColoredOutput(true)
			detector := &ProjectDetector{
				output:     output,
				currentDir: tempDir,
			}

			version := detector.detectVersion()
			if version != tt.want {
				t.Errorf("detectVersion() = %q, want %q", version, tt.want)
			}
		})
	}
}

// TestDetectVersionFromGitTags tests git tag version detection.
func TestDetectVersionFromGitTags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		repoRoot string
		want     string
	}{
		{
			name:     "no git repository",
			repoRoot: "",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := internal.NewColoredOutput(true)
			detector := &ProjectDetector{
				output:   output,
				repoRoot: tt.repoRoot,
			}

			version := detector.detectVersionFromGitTags()
			if version != tt.want {
				t.Errorf("detectVersionFromGitTags() = %q, want %q", version, tt.want)
			}
		})
	}
}

// TestAnalyzeActionFile tests action file analysis.
func TestAnalyzeActionFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		content   string
		wantErr   bool
		checkFunc func(t *testing.T, settings *DetectedSettings)
	}{
		{
			name:    "analyzes composite action",
			content: testutil.MustReadFixture(testutil.TestFixtureCompositeWithShellStep),
			wantErr: false,
			checkFunc: func(t *testing.T, settings *DetectedSettings) {
				t.Helper()
				if !settings.HasCompositeAction {
					t.Error("Expected HasCompositeAction to be true")
				}
			},
		},
		{
			name:    "handles invalid YAML",
			content: "invalid: yaml: content:",
			wantErr: true,
			checkFunc: func(t *testing.T, _ *DetectedSettings) {
				t.Helper()
				// No specific checks for error case
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			actionPath := testutil.WriteFileInDir(t, tempDir, appconstants.ActionFileNameYML, tt.content)

			output := internal.NewColoredOutput(true)
			detector := &ProjectDetector{
				output:     output,
				currentDir: tempDir,
			}

			settings := &DetectedSettings{
				SuggestedPermissions: make(map[string]string),
			}

			err := detector.analyzeActionFile(actionPath, settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("analyzeActionFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, settings)
			}
		})
	}
}
