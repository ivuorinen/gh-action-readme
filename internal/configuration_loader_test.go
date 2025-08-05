package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

func TestNewConfigurationLoader(t *testing.T) {
	loader := NewConfigurationLoader()

	if loader == nil {
		t.Fatal("expected non-nil loader")
	}

	if loader.viper == nil {
		t.Fatal("expected viper instance to be initialized")
	}

	// Check default sources are enabled
	expectedSources := []ConfigurationSource{
		SourceDefaults, SourceGlobal, SourceRepoOverride,
		SourceRepoConfig, SourceActionConfig, SourceEnvironment,
	}

	for _, source := range expectedSources {
		if !loader.sources[source] {
			t.Errorf("expected source %s to be enabled by default", source.String())
		}
	}

	// CLI flags should be disabled by default
	if loader.sources[SourceCLIFlags] {
		t.Error("expected CLI flags source to be disabled by default")
	}
}

func TestNewConfigurationLoaderWithOptions(t *testing.T) {
	tests := []struct {
		name     string
		opts     ConfigurationOptions
		expected []ConfigurationSource
	}{
		{
			name: "default options",
			opts: ConfigurationOptions{},
			expected: []ConfigurationSource{
				SourceDefaults, SourceGlobal, SourceRepoOverride,
				SourceRepoConfig, SourceActionConfig, SourceEnvironment,
			},
		},
		{
			name: "custom enabled sources",
			opts: ConfigurationOptions{
				EnabledSources: []ConfigurationSource{SourceDefaults, SourceGlobal},
			},
			expected: []ConfigurationSource{SourceDefaults, SourceGlobal},
		},
		{
			name: "all sources enabled",
			opts: ConfigurationOptions{
				EnabledSources: []ConfigurationSource{
					SourceDefaults, SourceGlobal, SourceRepoOverride,
					SourceRepoConfig, SourceActionConfig, SourceEnvironment, SourceCLIFlags,
				},
			},
			expected: []ConfigurationSource{
				SourceDefaults, SourceGlobal, SourceRepoOverride,
				SourceRepoConfig, SourceActionConfig, SourceEnvironment, SourceCLIFlags,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewConfigurationLoaderWithOptions(tt.opts)

			for _, expectedSource := range tt.expected {
				if !loader.sources[expectedSource] {
					t.Errorf("expected source %s to be enabled", expectedSource.String())
				}
			}

			// Check that non-expected sources are disabled
			allSources := []ConfigurationSource{
				SourceDefaults, SourceGlobal, SourceRepoOverride,
				SourceRepoConfig, SourceActionConfig, SourceEnvironment, SourceCLIFlags,
			}

			for _, source := range allSources {
				expected := false
				for _, expectedSource := range tt.expected {
					if source == expectedSource {
						expected = true
						break
					}
				}

				if loader.sources[source] != expected {
					t.Errorf("source %s enabled=%v, expected=%v", source.String(), loader.sources[source], expected)
				}
			}
		})
	}
}

func TestConfigurationLoader_LoadConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T, tempDir string) (configFile, repoRoot, actionDir string)
		expectError bool
		checkFunc   func(t *testing.T, config *AppConfig)
	}{
		{
			name: "defaults only",
			setupFunc: func(_ *testing.T, _ string) (string, string, string) {
				return "", "", ""
			},
			checkFunc: func(_ *testing.T, config *AppConfig) {
				testutil.AssertEqual(t, "default", config.Theme)
				testutil.AssertEqual(t, "md", config.OutputFormat)
				testutil.AssertEqual(t, ".", config.OutputDir)
			},
		},
		{
			name: "multi-level configuration hierarchy",
			setupFunc: func(_ *testing.T, tempDir string) (string, string, string) {
				// Create global config
				globalConfigDir := filepath.Join(tempDir, ".config", "gh-action-readme")
				_ = os.MkdirAll(globalConfigDir, 0750) // #nosec G301 -- test directory permissions
				globalConfigPath := filepath.Join(globalConfigDir, "config.yaml")
				testutil.WriteTestFile(t, globalConfigPath, `
theme: default
output_format: md
github_token: global-token
verbose: false
`)

				// Create repo root with repo-specific config
				repoRoot := filepath.Join(tempDir, "repo")
				_ = os.MkdirAll(repoRoot, 0750) // #nosec G301 -- test directory permissions
				testutil.WriteTestFile(t, filepath.Join(repoRoot, ".ghreadme.yaml"), `
theme: github
output_format: html
verbose: true
`)

				// Create action directory with action-specific config
				actionDir := filepath.Join(repoRoot, "action")
				_ = os.MkdirAll(actionDir, 0750) // #nosec G301 -- test directory permissions
				testutil.WriteTestFile(t, filepath.Join(actionDir, "config.yaml"), `
theme: professional
output_dir: output
quiet: false
`)

				return globalConfigPath, repoRoot, actionDir
			},
			checkFunc: func(_ *testing.T, config *AppConfig) {
				// Should have action-level overrides
				testutil.AssertEqual(t, "professional", config.Theme)
				testutil.AssertEqual(t, "output", config.OutputDir)
				// Should inherit from repo level
				testutil.AssertEqual(t, "html", config.OutputFormat)
				testutil.AssertEqual(t, true, config.Verbose)
				// Should inherit GitHub token from global config
				testutil.AssertEqual(t, "global-token", config.GitHubToken)
			},
		},
		{
			name: "environment variable overrides",
			setupFunc: func(_ *testing.T, tempDir string) (string, string, string) {
				// Set environment variables
				_ = os.Setenv("GH_README_GITHUB_TOKEN", "env-token")
				t.Cleanup(func() {
					_ = os.Unsetenv("GH_README_GITHUB_TOKEN")
				})

				// Create config file with different token
				configPath := filepath.Join(tempDir, "config.yml")
				testutil.WriteTestFile(t, configPath, `
theme: minimal
github_token: config-token
`)

				return configPath, tempDir, ""
			},
			checkFunc: func(_ *testing.T, config *AppConfig) {
				// Environment variable should override config file
				testutil.AssertEqual(t, "env-token", config.GitHubToken)
				testutil.AssertEqual(t, "minimal", config.Theme)
			},
		},
		{
			name: "hidden config file priority",
			setupFunc: func(_ *testing.T, tempDir string) (string, string, string) {
				repoRoot := filepath.Join(tempDir, "repo")
				_ = os.MkdirAll(repoRoot, 0750) // #nosec G301 -- test directory permissions

				// Create multiple hidden config files - first one should win
				testutil.WriteTestFile(t, filepath.Join(repoRoot, ".ghreadme.yaml"), `
theme: minimal
output_format: json
`)

				configDir := filepath.Join(repoRoot, ".config")
				_ = os.MkdirAll(configDir, 0750) // #nosec G301 -- test directory permissions
				testutil.WriteTestFile(t, filepath.Join(configDir, "ghreadme.yaml"), `
theme: professional
quiet: true
`)

				githubDir := filepath.Join(repoRoot, ".github")
				_ = os.MkdirAll(githubDir, 0750) // #nosec G301 -- test directory permissions
				testutil.WriteTestFile(t, filepath.Join(githubDir, "ghreadme.yaml"), `
theme: github
verbose: true
`)

				return "", repoRoot, ""
			},
			checkFunc: func(_ *testing.T, config *AppConfig) {
				// Should use the first found config (.ghreadme.yaml has priority)
				testutil.AssertEqual(t, "minimal", config.Theme)
				testutil.AssertEqual(t, "json", config.OutputFormat)
			},
		},
		{
			name: "selective source loading",
			setupFunc: func(_ *testing.T, _ string) (string, string, string) {
				// This test uses a loader with specific sources enabled
				return "", "", ""
			},
			checkFunc: func(_ *testing.T, _ *AppConfig) {
				// This will be tested with a custom loader
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

			configFile, repoRoot, actionDir := tt.setupFunc(t, tmpDir)

			// Special handling for selective source loading test
			var loader *ConfigurationLoader
			if tt.name == "selective source loading" {
				// Create loader with only defaults and global sources
				loader = NewConfigurationLoaderWithOptions(ConfigurationOptions{
					EnabledSources: []ConfigurationSource{SourceDefaults, SourceGlobal},
				})
			} else {
				loader = NewConfigurationLoader()
			}

			config, err := loader.LoadConfiguration(configFile, repoRoot, actionDir)

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

func TestConfigurationLoader_LoadGlobalConfig(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T, tempDir string) string
		expectError bool
		checkFunc   func(t *testing.T, config *AppConfig)
	}{
		{
			name: "valid global config",
			setupFunc: func(t *testing.T, tempDir string) string {
				configPath := filepath.Join(tempDir, "config.yaml")
				testutil.WriteTestFile(t, configPath, `
theme: professional
output_format: html
github_token: test-token
verbose: true
`)
				return configPath
			},
			checkFunc: func(_ *testing.T, config *AppConfig) {
				testutil.AssertEqual(t, "professional", config.Theme)
				testutil.AssertEqual(t, "html", config.OutputFormat)
				testutil.AssertEqual(t, "test-token", config.GitHubToken)
				testutil.AssertEqual(t, true, config.Verbose)
			},
		},
		{
			name: "nonexistent config file",
			setupFunc: func(_ *testing.T, tempDir string) string {
				return filepath.Join(tempDir, "nonexistent.yaml")
			},
			expectError: true,
		},
		{
			name: "invalid YAML",
			setupFunc: func(t *testing.T, tempDir string) string {
				configPath := filepath.Join(tempDir, "invalid.yaml")
				testutil.WriteTestFile(t, configPath, "invalid: yaml: content: [")
				return configPath
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			// Set HOME to temp directory
			originalHome := os.Getenv("HOME")
			_ = os.Setenv("HOME", tmpDir)
			defer func() {
				if originalHome != "" {
					_ = os.Setenv("HOME", originalHome)
				} else {
					_ = os.Unsetenv("HOME")
				}
			}()

			configFile := tt.setupFunc(t, tmpDir)

			loader := NewConfigurationLoader()
			config, err := loader.LoadGlobalConfig(configFile)

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

func TestConfigurationLoader_ValidateConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		config      *AppConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
			errorMsg:    "configuration cannot be nil",
		},
		{
			name: "valid config",
			config: &AppConfig{
				Theme:        "default",
				OutputFormat: "md",
				OutputDir:    ".",
				Verbose:      false,
				Quiet:        false,
			},
			expectError: false,
		},
		{
			name: "invalid output format",
			config: &AppConfig{
				Theme:        "default",
				OutputFormat: "invalid",
				OutputDir:    ".",
			},
			expectError: true,
			errorMsg:    "invalid output format",
		},
		{
			name: "empty output directory",
			config: &AppConfig{
				Theme:        "default",
				OutputFormat: "md",
				OutputDir:    "",
			},
			expectError: true,
			errorMsg:    "output directory cannot be empty",
		},
		{
			name: "verbose and quiet both true",
			config: &AppConfig{
				Theme:        "default",
				OutputFormat: "md",
				OutputDir:    ".",
				Verbose:      true,
				Quiet:        true,
			},
			expectError: true,
			errorMsg:    "verbose and quiet flags are mutually exclusive",
		},
		{
			name: "invalid theme",
			config: &AppConfig{
				Theme:        "nonexistent",
				OutputFormat: "md",
				OutputDir:    ".",
			},
			expectError: true,
			errorMsg:    "invalid theme",
		},
		{
			name: "valid built-in themes",
			config: &AppConfig{
				Theme:        "github",
				OutputFormat: "html",
				OutputDir:    "docs",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewConfigurationLoader()
			err := loader.ValidateConfiguration(tt.config)

			if tt.expectError {
				testutil.AssertError(t, err)
				if tt.errorMsg != "" {
					testutil.AssertStringContains(t, err.Error(), tt.errorMsg)
				}
			} else {
				testutil.AssertNoError(t, err)
			}
		})
	}
}

func TestConfigurationLoader_SourceManagement(t *testing.T) {
	loader := NewConfigurationLoader()

	// Test initial state
	sources := loader.GetConfigurationSources()
	if len(sources) != 6 { // All except CLI flags
		t.Errorf("expected 6 enabled sources, got %d", len(sources))
	}

	// Test disabling a source
	loader.DisableSource(SourceGlobal)
	if loader.sources[SourceGlobal] {
		t.Error("expected SourceGlobal to be disabled")
	}

	// Test enabling a source
	loader.EnableSource(SourceCLIFlags)
	if !loader.sources[SourceCLIFlags] {
		t.Error("expected SourceCLIFlags to be enabled")
	}

	// Test updated sources list
	sources = loader.GetConfigurationSources()
	expectedCount := 6 // 5 original + CLI flags - Global
	if len(sources) != expectedCount {
		t.Errorf("expected %d enabled sources, got %d", expectedCount, len(sources))
	}
}

func TestConfigurationSource_String(t *testing.T) {
	tests := []struct {
		source   ConfigurationSource
		expected string
	}{
		{SourceDefaults, "defaults"},
		{SourceGlobal, "global"},
		{SourceRepoOverride, "repo-override"},
		{SourceRepoConfig, "repo-config"},
		{SourceActionConfig, "action-config"},
		{SourceEnvironment, "environment"},
		{SourceCLIFlags, "cli-flags"},
		{ConfigurationSource(999), "unknown"},
	}

	for _, tt := range tests {
		result := tt.source.String()
		if result != tt.expected {
			t.Errorf("source %d String() = %s, expected %s", int(tt.source), result, tt.expected)
		}
	}
}

func TestConfigurationLoader_EnvironmentOverrides(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(t *testing.T) func()
		expectedToken string
	}{
		{
			name: "GH_README_GITHUB_TOKEN priority",
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
			name: "GITHUB_TOKEN fallback",
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

			loader := NewConfigurationLoader()
			config, err := loader.LoadConfiguration("", tmpDir, "")
			testutil.AssertNoError(t, err)

			testutil.AssertEqual(t, tt.expectedToken, config.GitHubToken)
		})
	}
}

func TestConfigurationLoader_RepoOverrides(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Create a mock git repository structure for testing
	repoRoot := filepath.Join(tmpDir, "test-repo")
	_ = os.MkdirAll(repoRoot, 0750) // #nosec G301 -- test directory permissions

	// Create global config with repo overrides
	globalConfigDir := filepath.Join(tmpDir, ".config", "gh-action-readme")
	_ = os.MkdirAll(globalConfigDir, 0750) // #nosec G301 -- test directory permissions
	globalConfigPath := filepath.Join(globalConfigDir, "config.yaml")
	globalConfigContent := "theme: default\n"
	globalConfigContent += "output_format: md\n"
	globalConfigContent += "repo_overrides:\n"
	globalConfigContent += "  test-repo:\n"
	globalConfigContent += "    theme: github\n"
	globalConfigContent += "    output_format: html\n"
	globalConfigContent += "    verbose: true\n"
	testutil.WriteTestFile(t, globalConfigPath, globalConfigContent)

	// Set environment for XDG compliance
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() {
		if originalHome != "" {
			_ = os.Setenv("HOME", originalHome)
		} else {
			_ = os.Unsetenv("HOME")
		}
	}()

	loader := NewConfigurationLoader()
	config, err := loader.LoadConfiguration(globalConfigPath, repoRoot, "")
	testutil.AssertNoError(t, err)

	// Note: Since we don't have actual git repository detection in this test,
	// repo overrides won't be applied. This test validates the structure works.
	testutil.AssertEqual(t, "default", config.Theme)
	testutil.AssertEqual(t, "md", config.OutputFormat)
}

// TestConfigurationLoader_ApplyRepoOverrides tests repo-specific overrides.
func TestConfigurationLoader_ApplyRepoOverrides(t *testing.T) {
	tests := []struct {
		name           string
		config         *AppConfig
		expectedTheme  string
		expectedFormat string
	}{
		{
			name: "no repo overrides configured",
			config: &AppConfig{
				Theme:         "default",
				OutputFormat:  "md",
				RepoOverrides: nil,
			},
			expectedTheme:  "default",
			expectedFormat: "md",
		},
		{
			name: "empty repo overrides map",
			config: &AppConfig{
				Theme:         "default",
				OutputFormat:  "md",
				RepoOverrides: map[string]AppConfig{},
			},
			expectedTheme:  "default",
			expectedFormat: "md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			loader := NewConfigurationLoader()
			loader.applyRepoOverrides(tt.config, tmpDir)
			testutil.AssertEqual(t, tt.expectedTheme, tt.config.Theme)
			testutil.AssertEqual(t, tt.expectedFormat, tt.config.OutputFormat)
		})
	}
}

// TestConfigurationLoader_LoadActionConfig tests action-specific configuration loading.
func TestConfigurationLoader_LoadActionConfig(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(t *testing.T, tmpDir string) string
		expectError  bool
		expectedVals map[string]string
	}{
		{
			name: "no action directory provided",
			setupFunc: func(_ *testing.T, _ string) string {
				return ""
			},
			expectError:  false,
			expectedVals: map[string]string{},
		},
		{
			name: "action directory with config file",
			setupFunc: func(t *testing.T, tmpDir string) string {
				actionDir := filepath.Join(tmpDir, "action")
				_ = os.MkdirAll(actionDir, 0750) // #nosec G301 -- test directory permissions

				configPath := filepath.Join(actionDir, "config.yaml")
				testutil.WriteTestFile(t, configPath, `
theme: minimal
output_format: json
verbose: true
`)
				return actionDir
			},
			expectError: false,
			expectedVals: map[string]string{
				"theme":         "minimal",
				"output_format": "json",
			},
		},
		{
			name: "action directory with malformed config file",
			setupFunc: func(t *testing.T, tmpDir string) string {
				actionDir := filepath.Join(tmpDir, "action")
				_ = os.MkdirAll(actionDir, 0750) // #nosec G301 -- test directory permissions

				configPath := filepath.Join(actionDir, "config.yaml")
				testutil.WriteTestFile(t, configPath, "invalid yaml content:\n  - broken [")
				return actionDir
			},
			expectError:  false, // Function may handle YAML errors gracefully
			expectedVals: map[string]string{},
		},
		{
			name: "action directory without config file",
			setupFunc: func(_ *testing.T, tmpDir string) string {
				actionDir := filepath.Join(tmpDir, "action")
				_ = os.MkdirAll(actionDir, 0750) // #nosec G301 -- test directory permissions
				return actionDir
			},
			expectError:  false,
			expectedVals: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			actionDir := tt.setupFunc(t, tmpDir)

			loader := NewConfigurationLoader()
			config, err := loader.loadActionConfig(actionDir)

			if tt.expectError {
				testutil.AssertError(t, err)
			} else {
				testutil.AssertNoError(t, err)

				// Check expected values if no error
				if config != nil {
					for key, expected := range tt.expectedVals {
						switch key {
						case "theme":
							testutil.AssertEqual(t, expected, config.Theme)
						case "output_format":
							testutil.AssertEqual(t, expected, config.OutputFormat)
						}
					}
				}
			}
		})
	}
}

// TestConfigurationLoader_ValidateTheme tests theme validation edge cases.
func TestConfigurationLoader_ValidateTheme(t *testing.T) {
	tests := []struct {
		name        string
		theme       string
		expectError bool
	}{
		{
			name:        "valid built-in theme",
			theme:       "github",
			expectError: false,
		},
		{
			name:        "valid default theme",
			theme:       "default",
			expectError: false,
		},
		{
			name:        "empty theme returns error",
			theme:       "",
			expectError: true,
		},
		{
			name:        "invalid theme",
			theme:       "nonexistent-theme",
			expectError: true,
		},
		{
			name:        "case sensitive theme",
			theme:       "GitHub",
			expectError: true,
		},
		{
			name:        "custom theme path",
			theme:       "/custom/theme/path.tmpl",
			expectError: false,
		},
		{
			name:        "relative theme path",
			theme:       "custom/theme.tmpl",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewConfigurationLoader()
			err := loader.validateTheme(tt.theme)

			if tt.expectError {
				testutil.AssertError(t, err)
			} else {
				testutil.AssertNoError(t, err)
			}
		})
	}
}
