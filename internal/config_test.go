package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

func TestInitConfig(t *testing.T) {
	// Save original environment
	originalXDGConfig := os.Getenv("XDG_CONFIG_HOME")
	originalHome := os.Getenv("HOME")
	defer func() {
		if originalXDGConfig != "" {
			_ = os.Setenv("XDG_CONFIG_HOME", originalXDGConfig)
		} else {
			_ = os.Unsetenv("XDG_CONFIG_HOME")
		}
		if originalHome != "" {
			_ = os.Setenv("HOME", originalHome)
		}
	}()

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
				Template:     "",
				Schema:       "schemas/action.schema.json",
				Verbose:      false,
				Quiet:        false,
				GitHubToken:  "",
			},
		},
		{
			name:       "custom config file",
			configFile: "custom-config.yml",
			setupFunc: func(t *testing.T, tempDir string) {
				configPath := filepath.Join(tempDir, "custom-config.yml")
				testutil.WriteTestFile(t, configPath, testutil.CustomConfigYAML)
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
			_ = os.Setenv("XDG_CONFIG_HOME", tmpDir)
			_ = os.Setenv("HOME", tmpDir)

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
				// Create global config
				globalConfigDir := filepath.Join(tempDir, ".config", "gh-action-readme")
				_ = os.MkdirAll(globalConfigDir, 0755)
				testutil.WriteTestFile(t, filepath.Join(globalConfigDir, "config.yml"), `
theme: default
output_format: md
github_token: global-token
`)

				// Create repo root with repo-specific config
				repoRoot := filepath.Join(tempDir, "repo")
				_ = os.MkdirAll(repoRoot, 0755)
				testutil.WriteTestFile(t, filepath.Join(repoRoot, ".ghreadme.yaml"), `
theme: github
output_format: html
`)

				// Create current directory with action-specific config
				currentDir := filepath.Join(repoRoot, "action")
				_ = os.MkdirAll(currentDir, 0755)
				testutil.WriteTestFile(t, filepath.Join(currentDir, ".ghreadme.yaml"), `
theme: professional
output_dir: output
`)

				return "", repoRoot, currentDir
			},
			checkFunc: func(t *testing.T, config *AppConfig) {
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
				// Set environment variables
				_ = os.Setenv("GH_README_GITHUB_TOKEN", "env-token")
				_ = os.Setenv("GITHUB_TOKEN", "fallback-token")

				// Create config file
				configPath := filepath.Join(tempDir, "config.yml")
				testutil.WriteTestFile(t, configPath, `
theme: minimal
github_token: config-token
`)

				t.Cleanup(func() {
					_ = os.Unsetenv("GH_README_GITHUB_TOKEN")
					_ = os.Unsetenv("GITHUB_TOKEN")
				})

				return configPath, tempDir, tempDir
			},
			checkFunc: func(t *testing.T, config *AppConfig) {
				// Environment variable should override config file
				testutil.AssertEqual(t, "env-token", config.GitHubToken)
				testutil.AssertEqual(t, "minimal", config.Theme)
			},
		},
		{
			name: "XDG compliance",
			setupFunc: func(t *testing.T, tempDir string) (string, string, string) {
				// Set XDG environment variables
				xdgConfigHome := filepath.Join(tempDir, "xdg-config")
				_ = os.Setenv("XDG_CONFIG_HOME", xdgConfigHome)

				// Create XDG-compliant config
				configDir := filepath.Join(xdgConfigHome, "gh-action-readme")
				_ = os.MkdirAll(configDir, 0755)
				testutil.WriteTestFile(t, filepath.Join(configDir, "config.yml"), `
theme: github
verbose: true
`)

				t.Cleanup(func() {
					_ = os.Unsetenv("XDG_CONFIG_HOME")
				})

				return "", tempDir, tempDir
			},
			checkFunc: func(t *testing.T, config *AppConfig) {
				testutil.AssertEqual(t, "github", config.Theme)
				testutil.AssertEqual(t, true, config.Verbose)
			},
		},
		{
			name: "hidden config file discovery",
			setupFunc: func(t *testing.T, tempDir string) (string, string, string) {
				repoRoot := filepath.Join(tempDir, "repo")
				_ = os.MkdirAll(repoRoot, 0755)

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
			originalHome := os.Getenv("HOME")
			_ = os.Setenv("HOME", tmpDir)
			defer func() {
				if originalHome != "" {
					_ = os.Setenv("HOME", originalHome)
				} else {
					_ = os.Unsetenv("HOME")
				}
			}()

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
	// Save original environment
	originalXDGConfig := os.Getenv("XDG_CONFIG_HOME")
	originalHome := os.Getenv("HOME")
	defer func() {
		if originalXDGConfig != "" {
			_ = os.Setenv("XDG_CONFIG_HOME", originalXDGConfig)
		} else {
			_ = os.Unsetenv("XDG_CONFIG_HOME")
		}
		if originalHome != "" {
			_ = os.Setenv("HOME", originalHome)
		}
	}()

	tests := []struct {
		name      string
		setupFunc func(t *testing.T, tempDir string)
		contains  string
	}{
		{
			name: "XDG_CONFIG_HOME set",
			setupFunc: func(_ *testing.T, tempDir string) {
				_ = os.Setenv("XDG_CONFIG_HOME", tempDir)
				_ = os.Unsetenv("HOME")
			},
			contains: "gh-action-readme",
		},
		{
			name: "HOME fallback",
			setupFunc: func(_ *testing.T, tempDir string) {
				_ = os.Unsetenv("XDG_CONFIG_HOME")
				_ = os.Setenv("HOME", tempDir)
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
	originalXDGConfig := os.Getenv("XDG_CONFIG_HOME")
	_ = os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() {
		if originalXDGConfig != "" {
			_ = os.Setenv("XDG_CONFIG_HOME", originalXDGConfig)
		} else {
			_ = os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	err := WriteDefaultConfig()
	testutil.AssertNoError(t, err)

	// Check that config file was created
	configPath, _ := GetConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("config file was not created at: %s", configPath)
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
	tests := []struct {
		name          string
		setupFunc     func(t *testing.T) func()
		expectedToken string
	}{
		{
			name: "GH_README_GITHUB_TOKEN has highest priority",
			setupFunc: func(_ *testing.T) func() {
				_ = os.Setenv("GH_README_GITHUB_TOKEN", "priority-token")
				_ = os.Setenv("GITHUB_TOKEN", "fallback-token")
				return func() {
					_ = os.Unsetenv("GH_README_GITHUB_TOKEN")
					_ = os.Unsetenv("GITHUB_TOKEN")
				}
			},
			expectedToken: "priority-token",
		},
		{
			name: "GITHUB_TOKEN as fallback",
			setupFunc: func(_ *testing.T) func() {
				_ = os.Unsetenv("GH_README_GITHUB_TOKEN")
				_ = os.Setenv("GITHUB_TOKEN", "fallback-token")
				return func() {
					_ = os.Unsetenv("GITHUB_TOKEN")
				}
			},
			expectedToken: "fallback-token",
		},
		{
			name: "no environment variables",
			setupFunc: func(_ *testing.T) func() {
				_ = os.Unsetenv("GH_README_GITHUB_TOKEN")
				_ = os.Unsetenv("GITHUB_TOKEN")
				return func() {}
			},
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setupFunc(t)
			defer cleanup()

			tmpDir, tmpCleanup := testutil.TempDir(t)
			defer tmpCleanup()

			// Use default config
			config, err := LoadConfiguration("", tmpDir, tmpDir)
			testutil.AssertNoError(t, err)

			testutil.AssertEqual(t, tt.expectedToken, config.GitHubToken)
		})
	}
}

func TestConfigMerging(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Test config merging by creating config files and seeing the result

	globalConfigDir := filepath.Join(tmpDir, ".config", "gh-action-readme")
	_ = os.MkdirAll(globalConfigDir, 0755)
	testutil.WriteTestFile(t, filepath.Join(globalConfigDir, "config.yml"), `
theme: default
output_format: md
github_token: base-token
verbose: false
`)

	repoRoot := filepath.Join(tmpDir, "repo")
	_ = os.MkdirAll(repoRoot, 0755)
	testutil.WriteTestFile(t, filepath.Join(repoRoot, ".ghreadme.yaml"), `
theme: github
output_format: html
verbose: true
`)

	// Set HOME to temp directory
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() {
		if originalHome != "" {
			_ = os.Setenv("HOME", originalHome)
		}
	}()

	config, err := LoadConfiguration("", repoRoot, repoRoot)
	testutil.AssertNoError(t, err)

	// Should have merged values
	testutil.AssertEqual(t, "github", config.Theme)                      // from repo config
	testutil.AssertEqual(t, "html", config.OutputFormat)                 // from repo config
	testutil.AssertEqual(t, true, config.Verbose)                        // from repo config
	testutil.AssertEqual(t, "base-token", config.GitHubToken)            // from global config
	testutil.AssertEqual(t, "schemas/action.schema.json", config.Schema) // default value
}
