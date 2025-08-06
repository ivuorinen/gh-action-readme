package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

func TestInitConfig(t *testing.T) {

	tests := []struct {
		name        string
		configFile  string
		setupFunc   func(t *testing.T, tempDir string)
		expectError bool
		expected    *AppConfig
	}{
		{
			name:       "default config when no file exists",
			configFile: "",
			setupFunc:  nil,
			expected: &AppConfig{
				Theme:        "default",
				OutputFormat: "md",
				OutputDir:    ".",
				Template:     "templates/readme.tmpl",
				Schema:       "schemas/schema.json",
				Verbose:      false,
				Quiet:        false,
				GitHubToken:  "",
			},
		},
		{
			name:       "custom config file",
			configFile: "custom-config.yml",
			setupFunc: func(t *testing.T, tempDir string) {
				t.Helper()
				configPath := filepath.Join(tempDir, "custom-config.yml")
				testutil.WriteTestFile(t, configPath, testutil.MustReadFixture("professional-config.yml"))
			},
			expected: &AppConfig{
				Theme:        "professional",
				OutputFormat: "html",
				OutputDir:    "docs",
				Template:     "custom-template.tmpl",
				Schema:       "custom-schema.json",
				Verbose:      true,
				Quiet:        false,
				GitHubToken:  "test-token-from-config",
			},
		},
		{
			name:       "invalid config file",
			configFile: "config.yml",
			setupFunc: func(t *testing.T, tempDir string) {
				t.Helper()
				configPath := filepath.Join(tempDir, "config.yml")
				testutil.WriteTestFile(t, configPath, "invalid: yaml: content: [")
			},
			expectError: true,
		},
		{
			name:        "nonexistent config file",
			configFile:  "nonexistent.yml",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			// Set XDG_CONFIG_HOME to our temp directory
			t.Setenv("XDG_CONFIG_HOME", tmpDir)
			t.Setenv("HOME", tmpDir)

			if tt.setupFunc != nil {
				tt.setupFunc(t, tmpDir)
			}

			// Set config file path if specified
			configPath := ""
			if tt.configFile != "" {
				configPath = filepath.Join(tmpDir, tt.configFile)
			}

			config, err := InitConfig(configPath)

			if tt.expectError {
				testutil.AssertError(t, err)

				return
			}

			testutil.AssertNoError(t, err)

			// Verify config values
			if tt.expected != nil {
				testutil.AssertEqual(t, tt.expected.Theme, config.Theme)
				testutil.AssertEqual(t, tt.expected.OutputFormat, config.OutputFormat)
				testutil.AssertEqual(t, tt.expected.OutputDir, config.OutputDir)
				testutil.AssertEqual(t, tt.expected.Template, config.Template)
				testutil.AssertEqual(t, tt.expected.Schema, config.Schema)
				testutil.AssertEqual(t, tt.expected.Verbose, config.Verbose)
				testutil.AssertEqual(t, tt.expected.Quiet, config.Quiet)
				testutil.AssertEqual(t, tt.expected.GitHubToken, config.GitHubToken)
			}
		})
	}
}

func TestLoadConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T, tempDir string) (configFile, repoRoot, currentDir string)
		expectError bool
		checkFunc   func(t *testing.T, config *AppConfig)
	}{
		{
			name: "multi-level config hierarchy",
			setupFunc: func(t *testing.T, tempDir string) (string, string, string) {
				t.Helper()
				// Create global config
				globalConfigDir := filepath.Join(tempDir, ".config", "gh-action-readme")
				_ = os.MkdirAll(globalConfigDir, 0750) // #nosec G301 -- test directory permissions
				globalConfigPath := filepath.Join(globalConfigDir, "config.yaml")
				testutil.WriteTestFile(t, globalConfigPath, `
theme: default
output_format: md
github_token: global-token
`)

				// Create repo root with repo-specific config
				repoRoot := filepath.Join(tempDir, "repo")
				_ = os.MkdirAll(repoRoot, 0750) // #nosec G301 -- test directory permissions
				testutil.WriteTestFile(t, filepath.Join(repoRoot, ".ghreadme.yaml"), `
theme: github
output_format: html
`)

				// Create current directory with action-specific config
				currentDir := filepath.Join(repoRoot, "action")
				_ = os.MkdirAll(currentDir, 0750) // #nosec G301 -- test directory permissions
				testutil.WriteTestFile(t, filepath.Join(currentDir, "config.yaml"), `
theme: professional
output_dir: output
`)

				return globalConfigPath, repoRoot, currentDir
			},
			checkFunc: func(t *testing.T, config *AppConfig) {
				t.Helper()
				// Should have action-level overrides
				testutil.AssertEqual(t, "professional", config.Theme)
				testutil.AssertEqual(t, "output", config.OutputDir)
				// Should inherit from repo level
				testutil.AssertEqual(t, "html", config.OutputFormat)
				// Should inherit GitHub token from global config
				testutil.AssertEqual(t, "global-token", config.GitHubToken)
			},
		},
		{
			name: "environment variable overrides",
			setupFunc: func(t *testing.T, tempDir string) (string, string, string) {
				t.Helper()
				// Set environment variables
				t.Setenv("GH_README_GITHUB_TOKEN", "env-token")
				t.Setenv("GITHUB_TOKEN", "fallback-token")

				// Create config file
				configPath := filepath.Join(tempDir, "config.yml")
				testutil.WriteTestFile(t, configPath, `
theme: minimal
github_token: config-token
`)

				return configPath, tempDir, tempDir
			},
			checkFunc: func(t *testing.T, config *AppConfig) {
				t.Helper()
				// Environment variable should override config file
				testutil.AssertEqual(t, "env-token", config.GitHubToken)
				testutil.AssertEqual(t, "minimal", config.Theme)
			},
		},
		{
			name: "XDG compliance",
			setupFunc: func(t *testing.T, tempDir string) (string, string, string) {
				t.Helper()
				// Set XDG environment variables
				xdgConfigHome := filepath.Join(tempDir, "xdg-config")
				t.Setenv("XDG_CONFIG_HOME", xdgConfigHome)

				// Create XDG-compliant config
				configDir := filepath.Join(xdgConfigHome, "gh-action-readme")
				_ = os.MkdirAll(configDir, 0750) // #nosec G301 -- test directory permissions
				configPath := filepath.Join(configDir, "config.yaml")
				testutil.WriteTestFile(t, configPath, `
theme: github
verbose: true
`)

				return configPath, tempDir, tempDir
			},
			checkFunc: func(t *testing.T, config *AppConfig) {
				t.Helper()
				testutil.AssertEqual(t, "github", config.Theme)
				testutil.AssertEqual(t, true, config.Verbose)
			},
		},
		{
			name: "hidden config file discovery",
			setupFunc: func(t *testing.T, tempDir string) (string, string, string) {
				t.Helper()
				repoRoot := filepath.Join(tempDir, "repo")
				_ = os.MkdirAll(repoRoot, 0750) // #nosec G301 -- test directory permissions

				// Create multiple hidden config files
				testutil.WriteTestFile(t, filepath.Join(repoRoot, ".ghreadme.yaml"), `
theme: minimal
output_format: json
`)

				testutil.WriteTestFile(t, filepath.Join(repoRoot, ".config", "ghreadme.yaml"), `
theme: professional
quiet: true
`)

				testutil.WriteTestFile(t, filepath.Join(repoRoot, ".github", "ghreadme.yaml"), `
theme: github
verbose: true
`)

				return "", repoRoot, repoRoot
			},
			checkFunc: func(t *testing.T, config *AppConfig) {
				t.Helper()
				// Should use the first found config (.ghreadme.yaml has priority)
				testutil.AssertEqual(t, "minimal", config.Theme)
				testutil.AssertEqual(t, "json", config.OutputFormat)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			// Set HOME to temp directory for fallback
			t.Setenv("HOME", tmpDir)

			configFile, repoRoot, currentDir := tt.setupFunc(t, tmpDir)

			config, err := LoadConfiguration(configFile, repoRoot, currentDir)

			if tt.expectError {
				testutil.AssertError(t, err)

				return
			}

			testutil.AssertNoError(t, err)

			if tt.checkFunc != nil {
				tt.checkFunc(t, config)
			}
		})
	}
}

func TestGetConfigPath(t *testing.T) {

	tests := []struct {
		name      string
		setupFunc func(t *testing.T, tempDir string)
		contains  string
	}{
		{
			name: "XDG_CONFIG_HOME set",
			setupFunc: func(t *testing.T, tempDir string) {
				t.Helper()
				t.Setenv("XDG_CONFIG_HOME", tempDir)
				t.Setenv("HOME", "")
			},
			contains: "gh-action-readme",
		},
		{
			name: "HOME fallback",
			setupFunc: func(t *testing.T, tempDir string) {
				t.Helper()
				t.Setenv("XDG_CONFIG_HOME", "")
				t.Setenv("HOME", tempDir)
			},
			contains: ".config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			tt.setupFunc(t, tmpDir)

			path, err := GetConfigPath()
			testutil.AssertNoError(t, err)

			if !filepath.IsAbs(path) {
				t.Errorf("expected absolute path, got: %s", path)
			}

			testutil.AssertStringContains(t, path, tt.contains)
		})
	}
}

func TestWriteDefaultConfig(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Set XDG_CONFIG_HOME to our temp directory
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	err := WriteDefaultConfig()
	testutil.AssertNoError(t, err)

	// Check that config file was created
	configPath, _ := GetConfigPath()
	t.Logf("Expected config path: %s", configPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("config file was not created at: %s", configPath)
		// List what files were actually created
		if files, err := os.ReadDir(tmpDir); err == nil {
			t.Logf("Files in tmpDir: %v", files)
		}
	}

	// Verify config file content
	config, err := InitConfig(configPath)
	testutil.AssertNoError(t, err)

	// Should have default values
	testutil.AssertEqual(t, "default", config.Theme)
	testutil.AssertEqual(t, "md", config.OutputFormat)
	testutil.AssertEqual(t, ".", config.OutputDir)
}

func TestResolveThemeTemplate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		theme        string
		expectError  bool
		shouldExist  bool
		expectedPath string
	}{
		{
			name:         "default theme",
			theme:        "default",
			expectError:  false,
			shouldExist:  true,
			expectedPath: "templates/readme.tmpl",
		},
		{
			name:         "github theme",
			theme:        "github",
			expectError:  false,
			shouldExist:  true,
			expectedPath: "templates/themes/github/readme.tmpl",
		},
		{
			name:         "gitlab theme",
			theme:        "gitlab",
			expectError:  false,
			shouldExist:  true,
			expectedPath: "templates/themes/gitlab/readme.tmpl",
		},
		{
			name:         "minimal theme",
			theme:        "minimal",
			expectError:  false,
			shouldExist:  true,
			expectedPath: "templates/themes/minimal/readme.tmpl",
		},
		{
			name:         "professional theme",
			theme:        "professional",
			expectError:  false,
			shouldExist:  true,
			expectedPath: "templates/themes/professional/readme.tmpl",
		},
		{
			name:        "unknown theme",
			theme:       "nonexistent",
			expectError: true,
		},
		{
			name:        "empty theme",
			theme:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := resolveThemeTemplate(tt.theme)

			if tt.expectError {
				if path != "" {
					t.Errorf("expected empty path on error, got: %s", path)
				}

				return
			}

			if path == "" {
				t.Error("expected non-empty path")
			}

			if tt.expectedPath != "" {
				testutil.AssertStringContains(t, path, tt.expectedPath)
			}

			// Note: We can't check file existence here because template files
			// might not be present in the test environment
		})
	}
}

func TestConfigTokenHierarchy(t *testing.T) {
	tests := testutil.GetGitHubTokenHierarchyTests()

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			cleanup := tt.SetupFunc(t)
			defer cleanup()

			tmpDir, tmpCleanup := testutil.TempDir(t)
			defer tmpCleanup()

			// Use default config
			config, err := LoadConfiguration("", tmpDir, tmpDir)
			testutil.AssertNoError(t, err)

			testutil.AssertEqual(t, tt.ExpectedToken, config.GitHubToken)
		})
	}
}

func TestConfigMerging(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Test config merging by creating config files and seeing the result

	globalConfigDir := filepath.Join(tmpDir, ".config", "gh-action-readme")
	_ = os.MkdirAll(globalConfigDir, 0750) // #nosec G301 -- test directory permissions
	testutil.WriteTestFile(t, filepath.Join(globalConfigDir, "config.yaml"), `
theme: default
output_format: md
github_token: base-token
verbose: false
`)

	repoRoot := filepath.Join(tmpDir, "repo")
	_ = os.MkdirAll(repoRoot, 0750) // #nosec G301 -- test directory permissions
	testutil.WriteTestFile(t, filepath.Join(repoRoot, ".ghreadme.yaml"), `
theme: github
output_format: html
verbose: true
`)

	// Set HOME and XDG_CONFIG_HOME to temp directory
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))

	// Use the specific config file path instead of relying on XDG discovery
	configPath := filepath.Join(tmpDir, ".config", "gh-action-readme", "config.yaml")
	config, err := LoadConfiguration(configPath, repoRoot, repoRoot)
	testutil.AssertNoError(t, err)

	// Should have merged values
	testutil.AssertEqual(t, "github", config.Theme)               // from repo config
	testutil.AssertEqual(t, "html", config.OutputFormat)          // from repo config
	testutil.AssertEqual(t, true, config.Verbose)                 // from repo config
	testutil.AssertEqual(t, "base-token", config.GitHubToken)     // from global config
	testutil.AssertEqual(t, "schemas/schema.json", config.Schema) // default value
}

// TestGetGitHubToken tests GitHub token resolution with different priority levels.
func TestGetGitHubToken(t *testing.T) {

	tests := []struct {
		name          string
		toolEnvToken  string
		stdEnvToken   string
		configToken   string
		expectedToken string
	}{
		{
			name:          "tool-specific env var has highest priority",
			toolEnvToken:  "tool-token",
			stdEnvToken:   "std-token",
			configToken:   "config-token",
			expectedToken: "tool-token",
		},
		{
			name:          "standard env var when tool env not set",
			toolEnvToken:  "",
			stdEnvToken:   "std-token",
			configToken:   "config-token",
			expectedToken: "std-token",
		},
		{
			name:          "config token when env vars not set",
			toolEnvToken:  "",
			stdEnvToken:   "",
			configToken:   "config-token",
			expectedToken: "config-token",
		},
		{
			name:          "empty string when nothing set",
			toolEnvToken:  "",
			stdEnvToken:   "",
			configToken:   "",
			expectedToken: "",
		},
		{
			name:          "empty env var does not override config",
			toolEnvToken:  "",
			stdEnvToken:   "",
			configToken:   "config-token",
			expectedToken: "config-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.toolEnvToken != "" {
				t.Setenv(EnvGitHubToken, tt.toolEnvToken)
			} else {
				t.Setenv(EnvGitHubToken, "")
			}
			if tt.stdEnvToken != "" {
				t.Setenv(EnvGitHubTokenStandard, tt.stdEnvToken)
			} else {
				t.Setenv(EnvGitHubTokenStandard, "")
			}

			config := &AppConfig{GitHubToken: tt.configToken}
			result := GetGitHubToken(config)

			testutil.AssertEqual(t, tt.expectedToken, result)
		})
	}
}

// TestMergeMapFields tests the merging of map fields in configuration.
func TestMergeMapFields(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		dst      *AppConfig
		src      *AppConfig
		expected *AppConfig
	}{
		{
			name: "merge permissions into empty dst",
			dst:  &AppConfig{},
			src: &AppConfig{
				Permissions: map[string]string{"read": "read", "write": "write"},
			},
			expected: &AppConfig{
				Permissions: map[string]string{"read": "read", "write": "write"},
			},
		},
		{
			name: "merge permissions into existing dst",
			dst: &AppConfig{
				Permissions: map[string]string{"read": "existing"},
			},
			src: &AppConfig{
				Permissions: map[string]string{"read": "new", "write": "write"},
			},
			expected: &AppConfig{
				Permissions: map[string]string{"read": "new", "write": "write"},
			},
		},
		{
			name: "merge variables into empty dst",
			dst:  &AppConfig{},
			src: &AppConfig{
				Variables: map[string]string{"VAR1": "value1", "VAR2": "value2"},
			},
			expected: &AppConfig{
				Variables: map[string]string{"VAR1": "value1", "VAR2": "value2"},
			},
		},
		{
			name: "merge variables into existing dst",
			dst: &AppConfig{
				Variables: map[string]string{"VAR1": "existing"},
			},
			src: &AppConfig{
				Variables: map[string]string{"VAR1": "new", "VAR2": "value2"},
			},
			expected: &AppConfig{
				Variables: map[string]string{"VAR1": "new", "VAR2": "value2"},
			},
		},
		{
			name: "merge both permissions and variables",
			dst: &AppConfig{
				Permissions: map[string]string{"read": "existing"},
			},
			src: &AppConfig{
				Permissions: map[string]string{"write": "write"},
				Variables:   map[string]string{"VAR1": "value1"},
			},
			expected: &AppConfig{
				Permissions: map[string]string{"read": "existing", "write": "write"},
				Variables:   map[string]string{"VAR1": "value1"},
			},
		},
		{
			name: "empty src does not affect dst",
			dst: &AppConfig{
				Permissions: map[string]string{"read": "read"},
				Variables:   map[string]string{"VAR1": "value1"},
			},
			src: &AppConfig{},
			expected: &AppConfig{
				Permissions: map[string]string{"read": "read"},
				Variables:   map[string]string{"VAR1": "value1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Deep copy dst to avoid modifying test data
			dst := &AppConfig{}
			if tt.dst.Permissions != nil {
				dst.Permissions = make(map[string]string)
				for k, v := range tt.dst.Permissions {
					dst.Permissions[k] = v
				}
			}
			if tt.dst.Variables != nil {
				dst.Variables = make(map[string]string)
				for k, v := range tt.dst.Variables {
					dst.Variables[k] = v
				}
			}

			mergeMapFields(dst, tt.src)

			testutil.AssertEqual(t, tt.expected.Permissions, dst.Permissions)
			testutil.AssertEqual(t, tt.expected.Variables, dst.Variables)
		})
	}
}

// TestMergeSliceFields tests the merging of slice fields in configuration.
func TestMergeSliceFields(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		dst      *AppConfig
		src      *AppConfig
		expected []string
	}{
		{
			name:     "merge runsOn into empty dst",
			dst:      &AppConfig{},
			src:      &AppConfig{RunsOn: []string{"ubuntu-latest", "windows-latest"}},
			expected: []string{"ubuntu-latest", "windows-latest"},
		},
		{
			name:     "merge runsOn replaces existing dst",
			dst:      &AppConfig{RunsOn: []string{"macos-latest"}},
			src:      &AppConfig{RunsOn: []string{"ubuntu-latest", "windows-latest"}},
			expected: []string{"ubuntu-latest", "windows-latest"},
		},
		{
			name:     "empty src does not affect dst",
			dst:      &AppConfig{RunsOn: []string{"ubuntu-latest"}},
			src:      &AppConfig{},
			expected: []string{"ubuntu-latest"},
		},
		{
			name:     "empty src slice does not affect dst",
			dst:      &AppConfig{RunsOn: []string{"ubuntu-latest"}},
			src:      &AppConfig{RunsOn: []string{}},
			expected: []string{"ubuntu-latest"},
		},
		{
			name:     "single item slice",
			dst:      &AppConfig{},
			src:      &AppConfig{RunsOn: []string{"self-hosted"}},
			expected: []string{"self-hosted"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mergeSliceFields(tt.dst, tt.src)

			// Compare slices manually since they can't be compared directly
			if len(tt.expected) != len(tt.dst.RunsOn) {
				t.Errorf("expected slice length %d, got %d", len(tt.expected), len(tt.dst.RunsOn))

				return
			}
			for i, expected := range tt.expected {
				if i >= len(tt.dst.RunsOn) || tt.dst.RunsOn[i] != expected {
					t.Errorf("expected %v, got %v", tt.expected, tt.dst.RunsOn)

					return
				}
			}
		})
	}
}
