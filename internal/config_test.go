package internal

import (
	"net/http"
	"path/filepath"
	"testing"

	"github.com/google/go-github/v74/github"

	"github.com/ivuorinen/gh-action-readme/appconstants"
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
				Theme:        testutil.TestThemeDefault,
				OutputFormat: "md",
				OutputDir:    ".",
				Template:     testutil.TestTemplateWithPrefix,
				Schema:       "schemas/schema.json",
				Verbose:      false,
				Quiet:        false,
				GitHubToken:  "",
			},
		},
		{
			name:       "custom config file",
			configFile: testutil.TestFileCustomConfig,
			setupFunc: func(t *testing.T, tempDir string) {
				t.Helper()
				testutil.WriteFileInDir(
					t,
					tempDir,
					testutil.TestFileCustomConfig,
					testutil.MustReadFixture(testutil.TestFixtureProfessionalConfig),
				)
			},
			expected: &AppConfig{
				Theme:        testutil.TestThemeProfessional,
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
			configFile: testutil.TestPathConfigYML,
			setupFunc: func(t *testing.T, tempDir string) {
				t.Helper()
				testutil.WriteFileInDir(t, tempDir, testutil.TestPathConfigYML, "invalid: yaml: content: [")
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
			tmpDir, cleanup := testutil.SetupTestEnvironment(t)
			defer cleanup()

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
				// Clear environment variables to ensure config file values are used
				t.Setenv(appconstants.EnvGitHubTokenStandard, "")
				t.Setenv(appconstants.EnvGitHubToken, "")

				// Create global config
				globalConfigDir := filepath.Join(tempDir, testutil.TestDirDotConfig, testutil.TestBinaryName)
				globalConfigPath := WriteConfigFixture(t, globalConfigDir, testutil.TestConfigGlobalDefault)

				// Create repo root with repo-specific config
				repoRoot := filepath.Join(tempDir, "repo")
				testutil.WriteFileInDir(t, repoRoot, testutil.TestFileGHReadmeYAML,
					string(testutil.MustReadFixture(testutil.TestConfigRepoSimple)))

				// Create current directory with action-specific config
				currentDir := filepath.Join(repoRoot, "action")
				WriteConfigFixture(t, currentDir, testutil.TestConfigActionSimple)

				return globalConfigPath, repoRoot, currentDir
			},
			checkFunc: func(t *testing.T, config *AppConfig) {
				t.Helper()
				// Should have action-level overrides
				testutil.AssertEqual(t, testutil.TestThemeProfessional, config.Theme)
				testutil.AssertEqual(t, "output", config.OutputDir)
				// Should inherit from repo level
				testutil.AssertEqual(t, "html", config.OutputFormat)
				// Should inherit GitHub token from global config
				testutil.AssertEqual(t, testutil.TestTokenStd, config.GitHubToken)
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
				testutil.WriteFileInDir(t, tempDir, testutil.TestPathConfigYML, `
theme: minimal
github_token: config-token
`)
				configPath := filepath.Join(tempDir, testutil.TestPathConfigYML)

				return configPath, tempDir, tempDir
			},
			checkFunc: func(t *testing.T, config *AppConfig) {
				t.Helper()
				// Environment variable should override config file
				testutil.AssertEqual(t, "env-token", config.GitHubToken)
				testutil.AssertEqual(t, testutil.TestThemeMinimal, config.Theme)
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
				configDir := filepath.Join(xdgConfigHome, testutil.TestBinaryName)
				configPath := WriteConfigFixture(t, configDir, testutil.TestConfigGitHubVerbose)

				return configPath, tempDir, tempDir
			},
			checkFunc: func(t *testing.T, config *AppConfig) {
				t.Helper()
				testutil.AssertEqual(t, testutil.TestThemeGitHub, config.Theme)
				testutil.AssertEqual(t, true, config.Verbose)
			},
		},
		{
			name: "hidden config file discovery",
			setupFunc: func(t *testing.T, tempDir string) (string, string, string) {
				t.Helper()
				repoRoot := filepath.Join(tempDir, "repo")

				// Create multiple hidden config files
				testutil.WriteFileInDir(t, repoRoot, testutil.TestFileGHReadmeYAML,
					string(testutil.MustReadFixture(testutil.TestConfigMinimalTheme)))

				configDir := filepath.Join(repoRoot, testutil.TestDirDotConfig)
				testutil.WriteFileInDir(t, configDir, "ghreadme.yaml",
					string(testutil.MustReadFixture(testutil.TestConfigProfessionalQuiet)))

				githubDir := filepath.Join(repoRoot, ".github")
				testutil.WriteFileInDir(t, githubDir, "ghreadme.yaml",
					string(testutil.MustReadFixture(testutil.TestConfigGitHubVerbose)))

				return "", repoRoot, repoRoot
			},
			checkFunc: func(t *testing.T, config *AppConfig) {
				t.Helper()
				// Should use the first found config (.ghreadme.yaml has priority)
				testutil.AssertEqual(t, testutil.TestThemeMinimal, config.Theme)
				testutil.AssertEqual(t, "json", config.OutputFormat)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.SetupTestEnvironment(t)
			defer cleanup()

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
			contains: testutil.TestBinaryName,
		},
		{
			name: "HOME fallback",
			setupFunc: func(t *testing.T, tempDir string) {
				t.Helper()
				t.Setenv("XDG_CONFIG_HOME", "")
				t.Setenv("HOME", tempDir)
			},
			contains: testutil.TestDirDotConfig,
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
	_, cleanup := testutil.SetupTestEnvironment(t)
	defer cleanup()

	err := WriteDefaultConfig()
	testutil.AssertNoError(t, err)

	// Check that config file was created
	configPath, _ := GetConfigPath()
	t.Logf("Expected config path: %s", configPath)
	testutil.AssertFileExists(t, configPath)

	// Verify config file content
	config, err := InitConfig(configPath)
	testutil.AssertNoError(t, err)

	// Should have default values
	testutil.AssertEqual(t, testutil.TestThemeDefault, config.Theme)
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
			theme:        testutil.TestThemeDefault,
			expectError:  false,
			shouldExist:  true,
			expectedPath: testutil.TestTemplateWithPrefix,
		},
		{
			name:         "github theme",
			theme:        testutil.TestThemeGitHub,
			expectError:  false,
			shouldExist:  true,
			expectedPath: "templates/themes/github/readme.tmpl",
		},
		{
			name:         "gitlab theme",
			theme:        testutil.TestThemeGitLab,
			expectError:  false,
			shouldExist:  true,
			expectedPath: "templates/themes/gitlab/readme.tmpl",
		},
		{
			name:         "minimal theme",
			theme:        testutil.TestThemeMinimal,
			expectError:  false,
			shouldExist:  true,
			expectedPath: "templates/themes/minimal/readme.tmpl",
		},
		{
			name:         "professional theme",
			theme:        testutil.TestThemeProfessional,
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

	globalConfigDir := filepath.Join(tmpDir, testutil.TestDirDotConfig, testutil.TestBinaryName)
	WriteConfigFixture(t, globalConfigDir, testutil.TestConfigGlobalBaseToken)

	repoRoot := filepath.Join(tmpDir, "repo")
	testutil.WriteFileInDir(t, repoRoot, testutil.TestFileGHReadmeYAML,
		string(testutil.MustReadFixture(testutil.TestConfigRepoVerbose)))

	// Set HOME and XDG_CONFIG_HOME to temp directory
	testutil.SetupConfigEnvironment(t, tmpDir)

	// Use the specific config file path instead of relying on XDG discovery
	configPath := filepath.Join(
		tmpDir,
		testutil.TestDirDotConfig,
		testutil.TestBinaryName,
		testutil.TestFileConfigYAML,
	)
	config, err := LoadConfiguration(configPath, repoRoot, repoRoot)
	testutil.AssertNoError(t, err)

	// Should have merged values
	testutil.AssertEqual(t, testutil.TestThemeGitHub, config.Theme) // from repo config
	testutil.AssertEqual(t, "html", config.OutputFormat)            // from repo config
	testutil.AssertEqual(t, true, config.Verbose)                   // from repo config
	testutil.AssertEqual(t, "base-token", config.GitHubToken)       // from global config
	testutil.AssertEqual(t, "schemas/schema.json", config.Schema)   // default value
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
			stdEnvToken:   testutil.TestTokenStd,
			configToken:   testutil.TestTokenConfig,
			expectedToken: "tool-token",
		},
		{
			name:          "standard env var when tool env not set",
			toolEnvToken:  "",
			stdEnvToken:   testutil.TestTokenStd,
			configToken:   testutil.TestTokenConfig,
			expectedToken: testutil.TestTokenStd,
		},
		{
			name:          "config token when env vars not set",
			toolEnvToken:  "",
			stdEnvToken:   "",
			configToken:   testutil.TestTokenConfig,
			expectedToken: testutil.TestTokenConfig,
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
			configToken:   testutil.TestTokenConfig,
			expectedToken: testutil.TestTokenConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.toolEnvToken != "" {
				t.Setenv(appconstants.EnvGitHubToken, tt.toolEnvToken)
			} else {
				t.Setenv(appconstants.EnvGitHubToken, "")
			}
			if tt.stdEnvToken != "" {
				t.Setenv(appconstants.EnvGitHubTokenStandard, tt.stdEnvToken)
			} else {
				t.Setenv(appconstants.EnvGitHubTokenStandard, "")
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
		createMapMergeTest(
			"merge permissions into empty dst",
			nil,
			map[string]string{"read": "read", "write": "write"},
			map[string]string{"read": "read", "write": "write"},
			true, // isPermissions
		),
		createMapMergeTest(
			"merge permissions into existing dst",
			map[string]string{"read": "existing"},
			map[string]string{"read": "new", "write": "write"},
			map[string]string{"read": "new", "write": "write"},
			true, // isPermissions
		),
		createMapMergeTest(
			"merge variables into empty dst",
			nil,
			map[string]string{"VAR1": "value1", "VAR2": "value2"},
			map[string]string{"VAR1": "value1", "VAR2": "value2"},
			false, // isPermissions
		),
		createMapMergeTest(
			"merge variables into existing dst",
			map[string]string{"VAR1": "existing"},
			map[string]string{"VAR1": "new", "VAR2": "value2"},
			map[string]string{"VAR1": "new", "VAR2": "value2"},
			false, // isPermissions
		),
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
			src:      &AppConfig{RunsOn: []string{testutil.RunnerUbuntuLatest, testutil.RunnerWindowsLatest}},
			expected: []string{testutil.RunnerUbuntuLatest, testutil.RunnerWindowsLatest},
		},
		{
			name:     "merge runsOn replaces existing dst",
			dst:      &AppConfig{RunsOn: []string{"macos-latest"}},
			src:      &AppConfig{RunsOn: []string{testutil.RunnerUbuntuLatest, testutil.RunnerWindowsLatest}},
			expected: []string{testutil.RunnerUbuntuLatest, testutil.RunnerWindowsLatest},
		},
		{
			name:     "empty src does not affect dst",
			dst:      &AppConfig{RunsOn: []string{testutil.RunnerUbuntuLatest}},
			src:      &AppConfig{},
			expected: []string{testutil.RunnerUbuntuLatest},
		},
		{
			name:     "empty src slice does not affect dst",
			dst:      &AppConfig{RunsOn: []string{testutil.RunnerUbuntuLatest}},
			src:      &AppConfig{RunsOn: []string{}},
			expected: []string{testutil.RunnerUbuntuLatest},
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

// assertBooleanConfigFields is a helper that checks all boolean fields in AppConfig.
func assertBooleanConfigFields(t *testing.T, got, want *AppConfig) {
	t.Helper()

	fields := []struct {
		name    string
		gotVal  bool
		wantVal bool
	}{
		{"AnalyzeDependencies", got.AnalyzeDependencies, want.AnalyzeDependencies},
		{"ShowSecurityInfo", got.ShowSecurityInfo, want.ShowSecurityInfo},
		{"Verbose", got.Verbose, want.Verbose},
		{"Quiet", got.Quiet, want.Quiet},
		{"UseDefaultBranch", got.UseDefaultBranch, want.UseDefaultBranch},
	}

	for _, field := range fields {
		if field.gotVal != field.wantVal {
			t.Errorf("%s = %v, want %v", field.name, field.gotVal, field.wantVal)
		}
	}
}

// TestMergeBooleanFields tests merging boolean configuration fields.
func TestMergeBooleanFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		dst  *AppConfig
		src  *AppConfig
		want *AppConfig
	}{
		createBoolFieldMergeTest(
			"merge all true values",
			boolFields{false, false, false, false, false},
			boolFields{true, true, true, true, true},
			boolFields{true, true, true, true, true},
		),
		createBoolFieldMergeTest(
			"merge only some true values",
			boolFields{false, true, false, true, false},
			boolFields{true, false, true, false, false},
			boolFields{true, true, true, true, false},
		),
		createBoolFieldMergeTest(
			"merge with all source false",
			boolFields{true, true, true, true, true},
			boolFields{false, false, false, false, false},
			boolFields{true, true, true, true, true},
		),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mergeBooleanFields(tt.dst, tt.src)

			assertBooleanConfigFields(t, tt.dst, tt.want)
		})
	}
}

// TestMergeSecurityFields tests merging security-sensitive configuration fields.
func TestMergeSecurityFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		dst         *AppConfig
		src         *AppConfig
		allowTokens bool
		want        *AppConfig
	}{
		createTokenMergeTest(
			"allow tokens - merge token",
			"",
			"ghp_test_token",
			"ghp_test_token",
			true,
		),
		createTokenMergeTest(
			"disallow tokens - do not merge token",
			"",
			"ghp_test_token",
			"",
			false,
		),
		createTokenMergeTest(
			"allow tokens - do not overwrite with empty",
			"ghp_existing_token",
			"",
			"ghp_existing_token",
			true,
		),
		createTokenMergeTest(
			"allow tokens - overwrite existing token",
			"ghp_old_token",
			"ghp_new_token",
			"ghp_new_token",
			true,
		),
		{
			name: "allow tokens - merge repo overrides into nil dst",
			dst: &AppConfig{
				RepoOverrides: nil,
			},
			src: &AppConfig{
				RepoOverrides: map[string]AppConfig{
					testutil.OrgRepo: {Organization: testutil.OrgName, Repository: testutil.RepoName},
				},
			},
			allowTokens: true,
			want: &AppConfig{
				RepoOverrides: map[string]AppConfig{
					testutil.OrgRepo: {Organization: testutil.OrgName, Repository: testutil.RepoName},
				},
			},
		},
		{
			name: "allow tokens - merge repo overrides into existing dst",
			dst: &AppConfig{
				RepoOverrides: map[string]AppConfig{
					testutil.ExistingRepo: {Organization: testutil.ExistingOrgName, Repository: testutil.RepoName},
				},
			},
			src: &AppConfig{
				RepoOverrides: map[string]AppConfig{
					testutil.NewRepo: {Organization: testutil.NewOrgName, Repository: testutil.RepoName},
				},
			},
			allowTokens: true,
			want: &AppConfig{
				RepoOverrides: map[string]AppConfig{
					testutil.ExistingRepo: {Organization: testutil.ExistingOrgName, Repository: testutil.RepoName},
					testutil.NewRepo:      {Organization: testutil.NewOrgName, Repository: testutil.RepoName},
				},
			},
		},
		{
			name: "disallow tokens - do not merge repo overrides",
			dst: &AppConfig{
				RepoOverrides: nil,
			},
			src: &AppConfig{
				RepoOverrides: map[string]AppConfig{
					testutil.OrgRepo: {Organization: testutil.OrgName, Repository: testutil.RepoName},
				},
			},
			allowTokens: false,
			want: &AppConfig{
				RepoOverrides: nil,
			},
		},
		{
			name: "allow tokens - empty source repo overrides",
			dst: &AppConfig{
				RepoOverrides: map[string]AppConfig{
					testutil.ExistingRepo: {Organization: testutil.ExistingOrgName, Repository: testutil.RepoName},
				},
			},
			src: &AppConfig{
				RepoOverrides: map[string]AppConfig{},
			},
			allowTokens: true,
			want: &AppConfig{
				RepoOverrides: map[string]AppConfig{
					testutil.ExistingRepo: {Organization: testutil.ExistingOrgName, Repository: testutil.RepoName},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mergeSecurityFields(tt.dst, tt.src, tt.allowTokens)

			if tt.dst.GitHubToken != tt.want.GitHubToken {
				t.Errorf("GitHubToken = %q, want %q",
					tt.dst.GitHubToken, tt.want.GitHubToken)
			}

			assertRepoOverrides(t, tt.dst.RepoOverrides, tt.want.RepoOverrides)
		})
	}
}

// assertRepoOverrides validates that RepoOverrides match expectations.
func assertRepoOverrides(t *testing.T, got, want map[string]AppConfig) {
	t.Helper()

	if want == nil {
		if got != nil {
			t.Errorf("RepoOverrides = %v, want nil", got)
		}

		return
	}

	if got == nil {
		t.Error("RepoOverrides is nil, want non-nil")

		return
	}

	for key, wantVal := range want {
		gotVal, exists := got[key]
		if !exists {
			t.Errorf("RepoOverrides missing key %q", key)
		} else if gotVal.Organization != wantVal.Organization ||
			gotVal.Repository != wantVal.Repository {
			t.Errorf("RepoOverrides[%q] = %+v, want %+v",
				key, gotVal, wantVal)
		}
	}

	if len(got) != len(want) {
		t.Errorf("RepoOverrides length = %d, want %d", len(got), len(want))
	}
}

// assertGitHubClientValid checks that a GitHub client is properly initialized.
func assertGitHubClientValid(t *testing.T, client *GitHubClient, expectedToken string) {
	t.Helper()
	if client == nil {
		t.Error("expected non-nil client")

		return
	}
	if client.Client == nil {
		t.Error("expected non-nil GitHub client")
	}
	if client.Token != expectedToken {
		t.Errorf("expected token %q, got %q", expectedToken, client.Token)
	}
}

// TestNewGitHubClient_EdgeCases tests GitHub client initialization edge cases.
func TestNewGitHubClientEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		token       string
		expectError bool
		description string
	}{
		{
			name:        "valid classic GitHub token",
			token:       "ghp_1234567890abcdefghijklmnopqrstuvwxyzABCD",
			expectError: false,
			description: "Should create client with valid classic token",
		},
		{
			name:        "valid fine-grained PAT",
			token:       "github_pat_11AAAAAA0AAAAaAaaAaaaAaa_AaAAaAAaAAAaAAAAAaAAaAAaAaAAaAAAAaAAAAAAAAaAAaAAaAaaAA",
			expectError: false,
			description: "Should create client with fine-grained token",
		},
		{
			name:        "empty token",
			token:       "",
			expectError: false,
			description: "Should create client without authentication",
		},
		{
			name:        "short token",
			token:       "ghp_short",
			expectError: false,
			description: "Should create client even with unusual token format",
		},
		{
			name:        "token with special characters",
			token:       "test-token_123",
			expectError: false,
			description: "Should handle tokens with various characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client, err := NewGitHubClient(tt.token)

			if tt.expectError {
				testutil.AssertError(t, err)

				return
			}

			testutil.AssertNoError(t, err)
			assertGitHubClientValid(t, client, tt.token)
		})
	}
}

// TestValidateGitHubClientCreation tests raw GitHub client creation validation.
// This test demonstrates the use of the assertGitHubClient helper for
// validating github.Client instances with different configurations.
func TestValidateGitHubClientCreation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) (*github.Client, error)
		expectError bool
		description string
	}{
		{
			name: "successful client creation with nil transport",
			setupFunc: func(t *testing.T) (*github.Client, error) {
				t.Helper()
				// Valid client creation - github.NewClient handles nil gracefully
				return github.NewClient(nil), nil
			},
			expectError: false,
			description: "Should create valid GitHub client with default transport",
		},
		{
			name: "successful client creation with custom HTTP client",
			setupFunc: func(t *testing.T) (*github.Client, error) {
				t.Helper()
				// Create client with custom HTTP client for testing
				mockHTTPClient := &http.Client{
					Transport: &testutil.MockTransport{},
				}
				return github.NewClient(mockHTTPClient), nil
			},
			expectError: false,
			description: "Should create valid GitHub client with custom transport",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client, err := tt.setupFunc(t)

			// Use the assertGitHubClient helper to validate the result
			assertGitHubClient(t, client, err, tt.expectError)
		})
	}
}

// runTemplatePathTest runs a template path test with setup and validation.
func runTemplatePathTest(
	t *testing.T,
	setupFunc func(*testing.T) (string, func()),
	checkFunc func(*testing.T, string),
) {
	t.Helper()
	templatePath, cleanup := setupFunc(t)
	defer cleanup()
	result := resolveTemplatePath(templatePath)
	if checkFunc != nil {
		checkFunc(t, result)
	}
}

// TestResolveTemplatePath_EdgeCases tests template path resolution edge cases.
func TestResolveTemplatePathEdgeCases(t *testing.T) {
	// Note: Cannot use t.Parallel() because one subtest uses t.Chdir()

	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) (templatePath string, cleanup func())
		checkFunc   func(t *testing.T, result string)
		description string
	}{
		{
			name: "absolute path - return as-is",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir, cleanup := testutil.TempDir(t)
				testutil.WriteFileInDir(t, tmpDir, "template.tmpl", "test template")
				absPath := filepath.Join(tmpDir, "template.tmpl")

				return absPath, cleanup
			},
			checkFunc: func(t *testing.T, result string) {
				t.Helper()
				if !filepath.IsAbs(result) {
					t.Errorf("expected absolute path, got: %s", result)
				}
			},
			description: "Absolute paths should be returned unchanged",
		},
		{
			name: "embedded template - available",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				// Use a path we know is embedded
				return testutil.TestTemplateReadme, func() { /* No cleanup needed for embedded templates */ }
			},
			checkFunc: func(t *testing.T, result string) {
				t.Helper()
				if result != testutil.TestTemplateReadme {
					t.Errorf("expected %q, got: %s", testutil.TestTemplateReadme, result)
				}
			},
			description: "Embedded templates should return original path",
		},
		{
			name: "embedded template with templates/ prefix",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()

				return testutil.TestTemplateWithPrefix, func() { /* No cleanup needed for embedded templates */ }
			},
			checkFunc: func(t *testing.T, result string) {
				t.Helper()
				if result != testutil.TestTemplateWithPrefix {
					t.Errorf("expected %q, got: %s", testutil.TestTemplateWithPrefix, result)
				}
			},
			description: "Embedded templates with prefix should return original path",
		},
		{
			name: "filesystem template - exists in current dir",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir, cleanup := testutil.TempDir(t)
				// Create template in current directory
				templateName := "custom-template.tmpl"
				testutil.WriteFileInDir(t, tmpDir, templateName, "custom template")

				// Change to tmpDir
				t.Chdir(tmpDir)

				return templateName, cleanup
			},
			checkFunc: func(t *testing.T, result string) {
				t.Helper()
				if result == "" {
					t.Error(testutil.TestMsgExpectedNonEmpty)
				}
			},
			description: "Templates in current directory should be found",
		},
		{
			name: "non-existent template - fallback to original path",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()

				return "nonexistent-template.tmpl", func() { /* No cleanup needed for non-existent template test */ }
			},
			checkFunc: func(t *testing.T, result string) {
				t.Helper()
				if result != "nonexistent-template.tmpl" {
					t.Errorf("expected original path, got: %s", result)
				}
			},
			description: "Non-existent templates should return original path",
		},
		{
			name: "empty path",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()

				return "", func() { /* No cleanup needed for empty path test */ }
			},
			checkFunc: func(t *testing.T, _ string) {
				t.Helper()
				// Empty path may return binary directory or empty string
				// depending on whether GetBinaryDir succeeds
				// Just verify it doesn't crash
			},
			description: "Empty path should not crash",
		},
		{
			name: "relative path with subdirectory",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()

				return "themes/github/readme.tmpl", func() { /* No cleanup needed for relative path test */ }
			},
			checkFunc: func(t *testing.T, result string) {
				t.Helper()
				// Should return the path (either embedded or fallback)
				if result == "" {
					t.Error(testutil.TestMsgExpectedNonEmpty)
				}
			},
			description: "Relative paths with subdirectories should be resolved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Cannot use t.Parallel() because one subtest uses t.Chdir()
			runTemplatePathTest(t, tt.setupFunc, tt.checkFunc)
		})
	}
}

// TestDetectRepositoryName_EdgeCases tests repository name detection edge cases.
func TestDetectRepositoryNameEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFunc      func(t *testing.T) string
		expectedResult string
		description    string
	}{
		{
			name: "empty repo root",
			setupFunc: func(t *testing.T) string {
				t.Helper()

				return ""
			},
			expectedResult: "",
			description:    "Empty repo root should return empty string",
		},
		{
			name: "non-existent directory",
			setupFunc: func(t *testing.T) string {
				t.Helper()

				return "/nonexistent/path/to/repo"
			},
			expectedResult: "",
			description:    "Non-existent directory should return empty string",
		},
		{
			name: "directory without git",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir, _ := testutil.TempDir(t)

				return tmpDir
			},
			expectedResult: "",
			description:    "Directory without .git should return empty string",
		},
		createGitRemoteTestCase(
			"valid git repository with GitHub remote",
			`[remote "origin"]
	url = https://github.com/testorg/testrepo.git
	fetch = +refs/heads/*:refs/remotes/origin/*
`,
			"testorg/testrepo",
			"Valid GitHub repo should return org/repo",
		),
		createGitRemoteTestCase(
			"git repository with SSH remote",
			`[remote "origin"]
	url = git@github.com:sshorg/sshrepo.git
	fetch = +refs/heads/*:refs/remotes/origin/*
`,
			"sshorg/sshrepo",
			"SSH remote should be parsed correctly",
		),
		createGitRemoteTestCase(
			"git repository without remote",
			"", // No config content
			"",
			"Repository without remote should return empty string",
		),
		createGitRemoteTestCase(
			"git repository with non-GitHub remote",
			`[remote "origin"]
	url = https://gitlab.com/glorg/glrepo.git
	fetch = +refs/heads/*:refs/remotes/origin/*
`,
			"",
			"Non-GitHub remote should return empty string",
		),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repoRoot := tt.setupFunc(t)
			result := DetectRepositoryName(repoRoot)

			if result != tt.expectedResult {
				t.Errorf("DetectRepositoryName() = %q, want %q (test: %s)",
					result, tt.expectedResult, tt.description)
			}
		})
	}
}

// TestLoadConfiguration_EdgeCases tests configuration loading edge cases.
func TestLoadConfigurationEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) (configFile, repoRoot, currentDir string)
		expectError bool
		checkFunc   func(t *testing.T, config *AppConfig)
		description string
	}{
		{
			name: "empty config file path with defaults",
			setupFunc: func(t *testing.T) (string, string, string) {
				t.Helper()
				tmpDir, _ := testutil.TempDir(t)
				testutil.SetupConfigEnvironment(t, tmpDir)

				return "", tmpDir, tmpDir
			},
			expectError: false,
			checkFunc: func(t *testing.T, config *AppConfig) {
				t.Helper()
				if config == nil {
					t.Fatal(testutil.TestMsgExpectedNonNilConfig)
				}
				// Should have default values
				if config.Theme == "" {
					t.Error("expected non-empty theme (default)")
				}
			},
			description: "Empty config file should load defaults",
		},
		{
			name: "all paths empty",
			setupFunc: func(t *testing.T) (string, string, string) {
				t.Helper()
				tmpDir, _ := testutil.TempDir(t)
				t.Setenv("HOME", tmpDir)

				return "", "", ""
			},
			expectError: false,
			checkFunc: func(t *testing.T, config *AppConfig) {
				t.Helper()
				if config == nil {
					t.Fatal(testutil.TestMsgExpectedNonNilConfig)
				}
			},
			description: "All empty paths should still return config",
		},
		{
			name: "config file with minimal values",
			setupFunc: func(t *testing.T) (string, string, string) {
				t.Helper()
				tmpDir, _ := testutil.TempDir(t)
				testutil.WriteFileInDir(t, tmpDir, testutil.TestFileConfigYAML, "theme: minimal\n")
				configPath := filepath.Join(tmpDir, testutil.TestFileConfigYAML)

				return configPath, tmpDir, tmpDir
			},
			expectError: false,
			checkFunc: func(t *testing.T, config *AppConfig) {
				t.Helper()
				testutil.AssertEqual(t, testutil.TestThemeMinimal, config.Theme)
			},
			description: "Minimal config should merge with defaults",
		},
		{
			name: "invalid config file path",
			setupFunc: func(t *testing.T) (string, string, string) {
				t.Helper()
				tmpDir, _ := testutil.TempDir(t)

				return filepath.Join(tmpDir, "nonexistent.yaml"), tmpDir, tmpDir
			},
			expectError: true,
			description: "Invalid config file path should error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, repoRoot, currentDir := tt.setupFunc(t)

			config, err := LoadConfiguration(configFile, repoRoot, currentDir)

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

// TestInitConfig_EdgeCases tests config initialization edge cases.
func TestInitConfigEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) string
		expectError bool
		checkFunc   func(t *testing.T, config *AppConfig)
		description string
	}{
		{
			name: "empty config file path - use default",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir, _ := testutil.TempDir(t)
				testutil.SetupConfigEnvironment(t, tmpDir)

				return ""
			},
			expectError: false,
			checkFunc: func(t *testing.T, config *AppConfig) {
				t.Helper()
				if config == nil {
					t.Fatal(testutil.TestMsgExpectedNonNilConfig)
				}
				// Should have default values
				testutil.AssertEqual(t, testutil.TestThemeDefault, config.Theme)
			},
			description: "Empty path should use default config",
		},
		{
			name: "config file with empty values",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir, _ := testutil.TempDir(t)
				testutil.WriteFileInDir(t, tmpDir, "empty.yaml", "---\n")
				configPath := filepath.Join(tmpDir, "empty.yaml")

				return configPath
			},
			expectError: false,
			checkFunc: func(t *testing.T, config *AppConfig) {
				t.Helper()
				// Should still have default values filled in
				if config.Theme == "" {
					t.Error("expected non-empty theme from defaults")
				}
			},
			description: "Empty config should be filled with defaults",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := tt.setupFunc(t)

			config, err := InitConfig(configPath)

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
