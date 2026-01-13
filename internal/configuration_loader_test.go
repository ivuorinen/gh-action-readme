package internal

import (
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

func TestNewConfigurationLoader(t *testing.T) {
	t.Parallel()

	loader := NewConfigurationLoader()

	if loader == nil {
		t.Fatal("expected non-nil loader")
	}

	sources := loader.GetConfigurationSources()
	if len(sources) == 0 {
		t.Error("expected non-empty configuration sources")
	}

	// Verify all sources are enabled by default
	expectedSources := []appconstants.ConfigurationSource{
		appconstants.SourceDefaults,
		appconstants.SourceGlobal,
		appconstants.SourceRepoOverride,
		appconstants.SourceRepoConfig,
		appconstants.SourceActionConfig,
		appconstants.SourceEnvironment,
	}

	for _, source := range expectedSources {
		found := false
		for _, s := range sources {
			if s == source {
				found = true

				break
			}
		}
		if !found {
			t.Errorf("expected source %s to be enabled by default", source)
		}
	}
}

func TestNewConfigurationLoaderWithOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		opts  ConfigurationOptions
		check func(t *testing.T, loader *ConfigurationLoader)
	}{
		{
			name: "custom config file",
			opts: ConfigurationOptions{
				ConfigFile:  "/tmp/custom-config.yaml",
				AllowTokens: true,
			},
			check: func(t *testing.T, loader *ConfigurationLoader) {
				t.Helper()
				if loader == nil {
					t.Fatal("expected non-nil loader")
				}
			},
		},
		{
			name: "disabled sources",
			opts: ConfigurationOptions{
				EnabledSources: []appconstants.ConfigurationSource{
					appconstants.SourceDefaults,
				},
			},
			check: func(t *testing.T, loader *ConfigurationLoader) {
				t.Helper()
				sources := loader.GetConfigurationSources()
				if len(sources) != 1 {
					t.Errorf("expected 1 source, got %d", len(sources))
				}
			},
		},
		{
			name: "empty enabled sources - all enabled",
			opts: ConfigurationOptions{
				EnabledSources: []appconstants.ConfigurationSource{},
			},
			check: func(t *testing.T, loader *ConfigurationLoader) {
				t.Helper()
				sources := loader.GetConfigurationSources()
				if len(sources) < 2 {
					t.Errorf("expected all sources enabled, got %d", len(sources))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			loader := NewConfigurationLoaderWithOptions(tt.opts)

			if tt.check != nil {
				tt.check(t, loader)
			}
		})
	}
}

func TestConfigurationLoader_LoadConfiguration(t *testing.T) {
	// Note: Cannot use t.Parallel() because subtests use t.Setenv()

	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) (configFile, repoRoot, actionDir string)
		expectError bool
		checkFunc   func(t *testing.T, config *AppConfig)
		description string
	}{
		{
			name: "all empty paths - use defaults",
			setupFunc: func(t *testing.T) (string, string, string) {
				t.Helper()
				tmpDir, _ := testutil.TempDir(t)
				testutil.SetupConfigEnvironment(t, tmpDir)

				return "", "", ""
			},
			expectError: false,
			checkFunc: func(t *testing.T, config *AppConfig) {
				t.Helper()
				if config == nil {
					t.Fatal("expected non-nil config")
				}
				if config.Theme == "" {
					t.Error("expected default theme")
				}
			},
			description: "Should load defaults when all paths empty",
		},
		{
			name: "global config only",
			setupFunc: func(t *testing.T) (string, string, string) {
				t.Helper()
				tmpDir, _ := testutil.TempDir(t)
				configPath := filepath.Join(tmpDir, "config.yaml")
				testutil.WriteTestFile(t, configPath, `
theme: github
output_format: html
`)

				return configPath, "", ""
			},
			expectError: false,
			checkFunc: func(t *testing.T, config *AppConfig) {
				t.Helper()
				testutil.AssertEqual(t, testutil.TestThemeGitHub, config.Theme)
				testutil.AssertEqual(t, "html", config.OutputFormat)
			},
			description: "Should load global config only",
		},
		{
			name: "repo config overrides global",
			setupFunc: func(t *testing.T) (string, string, string) {
				t.Helper()
				tmpDir, _ := testutil.TempDir(t)

				// Global config
				globalPath := filepath.Join(tmpDir, "global.yaml")
				testutil.WriteTestFile(t, globalPath, `
theme: default
output_format: md
`)

				// Repo config
				repoRoot := filepath.Join(tmpDir, "repo")
				testutil.WriteFileInDir(t, repoRoot, ".ghreadme.yaml", `
theme: minimal
`)

				return globalPath, repoRoot, ""
			},
			expectError: false,
			checkFunc: func(t *testing.T, config *AppConfig) {
				t.Helper()
				testutil.AssertEqual(t, testutil.TestThemeMinimal, config.Theme)
				testutil.AssertEqual(t, "md", config.OutputFormat) // From global
			},
			description: "Repo config should override global",
		},
		{
			name: "action config has highest priority",
			setupFunc: func(t *testing.T) (string, string, string) {
				t.Helper()
				tmpDir, _ := testutil.TempDir(t)

				// Global config
				globalPath := filepath.Join(tmpDir, "global.yaml")
				testutil.WriteTestFile(t, globalPath, `
theme: default
output_format: md
`)

				// Repo config
				repoRoot := filepath.Join(tmpDir, "repo")
				testutil.WriteFileInDir(t, repoRoot, ".ghreadme.yaml", `
theme: minimal
`)

				// Action config
				actionDir := filepath.Join(repoRoot, "action")
				testutil.WriteFileInDir(t, actionDir, "config.yaml", `
theme: professional
`)

				return globalPath, repoRoot, actionDir
			},
			expectError: false,
			checkFunc: func(t *testing.T, config *AppConfig) {
				t.Helper()
				testutil.AssertEqual(t, testutil.TestThemeProfessional, config.Theme)
			},
			description: "Action config should have highest priority",
		},
		{
			name: "invalid global config file",
			setupFunc: func(t *testing.T) (string, string, string) {
				t.Helper()
				tmpDir, _ := testutil.TempDir(t)
				configPath := filepath.Join(tmpDir, "bad.yaml")
				testutil.WriteTestFile(t, configPath, `{invalid yaml: [[`)

				return configPath, "", ""
			},
			expectError: true,
			description: "Should error on invalid global config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, repoRoot, actionDir := tt.setupFunc(t)

			loader := NewConfigurationLoader()
			config, err := loader.LoadConfiguration(configFile, repoRoot, actionDir)

			if tt.expectError {
				testutil.AssertError(t, err)
			} else {
				testutil.AssertNoError(t, err)

				if tt.checkFunc != nil {
					tt.checkFunc(t, config)
				}
			}
		})
	}
}

func TestConfigurationLoader_LoadGlobalConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) string
		expectError bool
		checkFunc   func(t *testing.T, config *AppConfig)
		description string
	}{
		{
			name: "valid global config",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir, _ := testutil.TempDir(t)
				configPath := filepath.Join(tmpDir, "config.yaml")
				testutil.WriteTestFile(t, configPath, `
theme: github
output_format: html
verbose: true
`)

				return configPath
			},
			expectError: false,
			checkFunc: func(t *testing.T, config *AppConfig) {
				t.Helper()
				testutil.AssertEqual(t, testutil.TestThemeGitHub, config.Theme)
				testutil.AssertEqual(t, "html", config.OutputFormat)
				testutil.AssertEqual(t, true, config.Verbose)
			},
			description: "Should load valid global config",
		},
		{
			name: "empty config file",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir, _ := testutil.TempDir(t)
				configPath := filepath.Join(tmpDir, "empty.yaml")
				testutil.WriteTestFile(t, configPath, "---\n")

				return configPath
			},
			expectError: false,
			checkFunc: func(t *testing.T, config *AppConfig) {
				t.Helper()
				if config == nil {
					t.Fatal("expected non-nil config")
				}
			},
			description: "Empty config should not error",
		},
		{
			name: "config file does not exist",
			setupFunc: func(t *testing.T) string {
				t.Helper()

				return "/nonexistent/config.yaml"
			},
			expectError: true,
			description: "Non-existent config should error",
		},
		{
			name: "malformed YAML",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir, _ := testutil.TempDir(t)
				configPath := filepath.Join(tmpDir, "bad.yaml")
				testutil.WriteTestFile(t, configPath, `{{{invalid}}}`)

				return configPath
			},
			expectError: true,
			description: "Malformed YAML should error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			configFile := tt.setupFunc(t)

			loader := NewConfigurationLoader()
			config, err := loader.LoadGlobalConfig(configFile)

			if tt.expectError {
				testutil.AssertError(t, err)
			} else {
				testutil.AssertNoError(t, err)

				if tt.checkFunc != nil {
					tt.checkFunc(t, config)
				}
			}
		})
	}
}

func TestConfigurationLoader_ValidateConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *AppConfig
		expectError bool
		description string
	}{
		{
			name: "valid configuration",
			config: &AppConfig{
				Theme:        testutil.TestThemeDefault,
				OutputFormat: "md",
				OutputDir:    ".",
			},
			expectError: false,
			description: "Valid config should pass",
		},
		{
			name: "invalid theme",
			config: &AppConfig{
				Theme:        "invalid-theme",
				OutputFormat: "md",
			},
			expectError: true,
			description: "Invalid theme should error",
		},
		{
			name: "empty theme",
			config: &AppConfig{
				Theme:        "",
				OutputFormat: "md",
			},
			expectError: true,
			description: "Empty theme should error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			loader := NewConfigurationLoader()
			err := loader.ValidateConfiguration(tt.config)

			if tt.expectError {
				testutil.AssertError(t, err)
			} else {
				testutil.AssertNoError(t, err)
			}
		})
	}
}

func TestConfigurationLoader_SourceManagement(t *testing.T) {
	t.Parallel()

	loader := NewConfigurationLoader()

	// Initially, all sources should be enabled
	sources := loader.GetConfigurationSources()
	if len(sources) < 4 {
		t.Errorf("expected at least 4 sources initially, got %d", len(sources))
	}

	// Disable a source
	loader.DisableSource(appconstants.SourceRepoConfig)

	// Verify it's disabled
	sources = loader.GetConfigurationSources()
	for _, source := range sources {
		if source == appconstants.SourceRepoConfig {
			t.Error("expected SourceRepoConfig to be disabled")
		}
	}

	// Re-enable the source
	loader.EnableSource(appconstants.SourceRepoConfig)

	// Verify it's enabled again
	sources = loader.GetConfigurationSources()
	found := false
	for _, source := range sources {
		if source == appconstants.SourceRepoConfig {
			found = true

			break
		}
	}
	if !found {
		t.Error("expected SourceRepoConfig to be re-enabled")
	}
}

func TestConfigurationSourceString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		source   appconstants.ConfigurationSource
		expected string
	}{
		{appconstants.SourceDefaults, "defaults"},
		{appconstants.SourceGlobal, "global"},
		{appconstants.SourceRepoOverride, "repo-override"},
		{appconstants.SourceRepoConfig, "repo-config"},
		{appconstants.SourceActionConfig, "action-config"},
		{appconstants.SourceEnvironment, "environment"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()

			result := tt.source.String()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestConfigurationLoader_EnvironmentOverrides(t *testing.T) {
	tests := []struct {
		name        string
		setupEnv    func(t *testing.T)
		setupConfig func(t *testing.T) *AppConfig
		checkFunc   func(t *testing.T, config *AppConfig)
		description string
	}{
		{
			name: "GH_README_GITHUB_TOKEN overrides config",
			setupEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv(appconstants.EnvGitHubToken, testutil.TestTokenEnv)
			},
			setupConfig: func(t *testing.T) *AppConfig {
				t.Helper()

				return &AppConfig{
					GitHubToken: testutil.TestTokenConfig,
				}
			},
			checkFunc: func(t *testing.T, config *AppConfig) {
				t.Helper()
				testutil.AssertEqual(t, testutil.TestTokenEnv, config.GitHubToken)
			},
			description: "Environment variable should override config token",
		},
		{
			name: "GITHUB_TOKEN fallback",
			setupEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv(appconstants.EnvGitHubToken, "")
				t.Setenv(appconstants.EnvGitHubTokenStandard, "standard-token")
			},
			setupConfig: func(t *testing.T) *AppConfig {
				t.Helper()

				return &AppConfig{}
			},
			checkFunc: func(t *testing.T, config *AppConfig) {
				t.Helper()
				testutil.AssertEqual(t, "standard-token", config.GitHubToken)
			},
			description: "Should use GITHUB_TOKEN when GH_README_GITHUB_TOKEN not set",
		},
		{
			name: "config token used when no env vars",
			setupEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv(appconstants.EnvGitHubToken, "")
				t.Setenv(appconstants.EnvGitHubTokenStandard, "")
			},
			setupConfig: func(t *testing.T) *AppConfig {
				t.Helper()

				return &AppConfig{
					GitHubToken: testutil.TestTokenConfig,
				}
			},
			checkFunc: func(t *testing.T, config *AppConfig) {
				t.Helper()
				testutil.AssertEqual(t, testutil.TestTokenConfig, config.GitHubToken)
			},
			description: "Should preserve config token when no env vars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv(t)
			config := tt.setupConfig(t)

			loader := NewConfigurationLoader()
			loader.loadEnvironmentStep(config)

			if tt.checkFunc != nil {
				tt.checkFunc(t, config)
			}
		})
	}
}

func TestConfigurationLoader_ApplyRepoOverrides(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(t *testing.T) (config *AppConfig, repoRoot string)
		expectedTheme  string
		expectedFormat string
		description    string
	}{
		{
			name: "matching repo override applied",
			setupFunc: func(t *testing.T) (*AppConfig, string) {
				t.Helper()
				tmpDir, _ := testutil.TempDir(t)

				testutil.CreateGitRepoWithRemote(t, tmpDir, "https://github.com/test/repo.git")

				config := &AppConfig{
					Theme:        testutil.TestThemeDefault,
					OutputFormat: "md",
					RepoOverrides: map[string]AppConfig{
						"test/repo": {
							Theme:        testutil.TestThemeProfessional,
							OutputFormat: "html",
						},
					},
				}

				return config, tmpDir
			},
			expectedTheme:  testutil.TestThemeProfessional,
			expectedFormat: "html",
			description:    "Matching repo override should be applied",
		},
		{
			name: "no override when repo doesn't match",
			setupFunc: func(t *testing.T) (*AppConfig, string) {
				t.Helper()
				tmpDir, _ := testutil.TempDir(t)

				testutil.CreateGitRepoWithRemote(t, tmpDir, "https://github.com/different/repo.git")

				config := &AppConfig{
					Theme:        testutil.TestThemeDefault,
					OutputFormat: "md",
					RepoOverrides: map[string]AppConfig{
						"test/repo": {
							Theme:        testutil.TestThemeProfessional,
							OutputFormat: "html",
						},
					},
				}

				return config, tmpDir
			},
			expectedTheme:  testutil.TestThemeDefault,
			expectedFormat: "md",
			description:    "No override when repo doesn't match",
		},
		{
			name: "no override when no git repository",
			setupFunc: func(t *testing.T) (*AppConfig, string) {
				t.Helper()
				tmpDir, _ := testutil.TempDir(t)

				config := &AppConfig{
					Theme:        testutil.TestThemeDefault,
					OutputFormat: "md",
					RepoOverrides: map[string]AppConfig{
						"test/repo": {
							Theme:        testutil.TestThemeProfessional,
							OutputFormat: "html",
						},
					},
				}

				return config, tmpDir
			},
			expectedTheme:  testutil.TestThemeDefault,
			expectedFormat: "md",
			description:    "No override when not a git repository",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, repoRoot := tt.setupFunc(t)

			loader := NewConfigurationLoader()
			loader.applyRepoOverrides(config, repoRoot)

			// Verify expected values
			testutil.AssertEqual(t, tt.expectedTheme, config.Theme)
			testutil.AssertEqual(t, tt.expectedFormat, config.OutputFormat)
		})
	}
}

func TestConfigurationLoader_LoadActionConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) string
		expectError bool
		checkFunc   func(t *testing.T, config *AppConfig)
		description string
	}{
		{
			name: "valid action config",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir, _ := testutil.TempDir(t)
				testutil.WriteFileInDir(t, tmpDir, "config.yaml", `
theme: minimal
output_dir: dist
`)

				return tmpDir
			},
			expectError: false,
			checkFunc: func(t *testing.T, config *AppConfig) {
				t.Helper()
				testutil.AssertEqual(t, testutil.TestThemeMinimal, config.Theme)
				testutil.AssertEqual(t, "dist", config.OutputDir)
			},
			description: "Should load action config",
		},
		{
			name: "no action config file",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir, _ := testutil.TempDir(t)

				return tmpDir
			},
			expectError: false,
			checkFunc: func(t *testing.T, _ *AppConfig) {
				t.Helper()
				// Empty config is okay
			},
			description: "Missing action config should not error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actionDir := tt.setupFunc(t)

			loader := NewConfigurationLoader()
			config, err := loader.loadActionConfig(actionDir)

			if tt.expectError {
				testutil.AssertError(t, err)
			} else {
				testutil.AssertNoError(t, err)

				if tt.checkFunc != nil {
					tt.checkFunc(t, config)
				}
			}
		})
	}
}

func TestConfigurationLoader_ValidateTheme(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		theme       string
		expectError bool
	}{
		{
			name:        "valid theme - default",
			theme:       testutil.TestThemeDefault,
			expectError: false,
		},
		{
			name:        "valid theme - github",
			theme:       testutil.TestThemeGitHub,
			expectError: false,
		},
		{
			name:        "valid theme - minimal",
			theme:       testutil.TestThemeMinimal,
			expectError: false,
		},
		{
			name:        "invalid theme",
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

			loader := NewConfigurationLoader()
			config := &AppConfig{
				Theme: tt.theme,
			}

			err := loader.validateTheme(config.Theme)

			if tt.expectError {
				testutil.AssertError(t, err)
			} else {
				testutil.AssertNoError(t, err)
			}
		})
	}
}

func TestConfigurationLoaderApplyRepoOverridesWithRepoRoot(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(t *testing.T) (config *AppConfig, repoRoot string)
		expectedTheme  string
		expectedFormat string
		description    string
	}{
		{
			name: "override applied with valid repo root",
			setupFunc: func(t *testing.T) (*AppConfig, string) {
				t.Helper()
				tmpDir, _ := testutil.TempDir(t)

				testutil.CreateGitRepoWithRemote(t, tmpDir, "https://github.com/myorg/myrepo.git")

				config := &AppConfig{
					Theme:        testutil.TestThemeDefault,
					OutputFormat: "md",
					RepoOverrides: map[string]AppConfig{
						"myorg/myrepo": {
							Theme:        testutil.TestThemeGitHub,
							OutputFormat: "json",
						},
					},
				}

				return config, tmpDir
			},
			expectedTheme:  "github",
			expectedFormat: "json",
			description:    "Should apply repo override for detected repository",
		},
		{
			name: "no override with empty repo root",
			setupFunc: func(t *testing.T) (*AppConfig, string) {
				t.Helper()

				config := &AppConfig{
					Theme:        testutil.TestThemeDefault,
					OutputFormat: "md",
					RepoOverrides: map[string]AppConfig{
						"myorg/myrepo": {
							Theme:        testutil.TestThemeGitHub,
							OutputFormat: "json",
						},
					},
				}

				return config, ""
			},
			expectedTheme:  testutil.TestThemeDefault,
			expectedFormat: "md",
			description:    "Should not apply override when repo root is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, repoRoot := tt.setupFunc(t)

			loader := NewConfigurationLoader()
			loader.applyRepoOverrides(config, repoRoot)

			// Verify expected values
			testutil.AssertEqual(t, tt.expectedTheme, config.Theme)
			testutil.AssertEqual(t, tt.expectedFormat, config.OutputFormat)
		})
	}
}
