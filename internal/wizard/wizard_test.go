package wizard

import (
	"bufio"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/internal"
)

// testWizard creates a wizard with mocked input for testing.
func testWizard(inputs string) *ConfigWizard {
	// Create a scanner from the input string
	scanner := bufio.NewScanner(strings.NewReader(inputs))

	// Create wizard with quiet output to avoid console spam
	wizard := &ConfigWizard{
		output:  &internal.ColoredOutput{NoColor: true, Quiet: true},
		scanner: scanner,
		config:  internal.DefaultAppConfig(),
	}

	return wizard
}

// Note: Output verification tests are simplified since ColoredOutput is a concrete type
// Tests focus on logic and state changes rather than output messages

// TestPromptWithDefault tests the prompt with default value function.
func TestPromptWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		prompt       string
		defaultValue string
		want         string
	}{
		{
			name:         "user provides value",
			input:        "custom-value\n",
			prompt:       "Enter value",
			defaultValue: "default",
			want:         "custom-value",
		},
		{
			name:         "user accepts default (empty input)",
			input:        "\n",
			prompt:       "Enter value",
			defaultValue: "default",
			want:         "default",
		},
		{
			name:         "user provides empty string with no default",
			input:        "\n",
			prompt:       "Enter value",
			defaultValue: "",
			want:         "",
		},
		{
			name:         "user provides value with whitespace",
			input:        "  value-with-spaces  \n",
			prompt:       "Enter value",
			defaultValue: "default",
			want:         "value-with-spaces",
		},
		{
			name:         "no default provided, user enters value",
			input:        "myvalue\n",
			prompt:       "Enter value",
			defaultValue: "",
			want:         "myvalue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wizard := testWizard(tt.input)
			got := wizard.promptWithDefault(tt.prompt, tt.defaultValue)

			if got != tt.want {
				t.Errorf("promptWithDefault() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestPromptYesNo tests the yes/no prompt function.
func TestPromptYesNo(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		prompt       string
		defaultValue bool
		want         bool
	}{
		{
			name:         "user enters yes",
			input:        "yes\n",
			prompt:       "Continue?",
			defaultValue: false,
			want:         true,
		},
		{
			name:         "user enters y",
			input:        "y\n",
			prompt:       "Continue?",
			defaultValue: false,
			want:         true,
		},
		{
			name:         "user enters no",
			input:        "no\n",
			prompt:       "Continue?",
			defaultValue: true,
			want:         false,
		},
		{
			name:         "user enters n",
			input:        "n\n",
			prompt:       "Continue?",
			defaultValue: true,
			want:         false,
		},
		{
			name:         "user accepts default true",
			input:        "\n",
			prompt:       "Continue?",
			defaultValue: true,
			want:         true,
		},
		{
			name:         "user accepts default false",
			input:        "\n",
			prompt:       "Continue?",
			defaultValue: false,
			want:         false,
		},
		{
			name:         "invalid input then default",
			input:        "maybe\n",
			prompt:       "Continue?",
			defaultValue: true,
			want:         true,
		},
		{
			name:         "case insensitive YES",
			input:        "YES\n",
			prompt:       "Continue?",
			defaultValue: false,
			want:         true,
		},
		{
			name:         "case insensitive NO",
			input:        "NO\n",
			prompt:       "Continue?",
			defaultValue: true,
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wizard := testWizard(tt.input)
			got := wizard.promptYesNo(tt.prompt, tt.defaultValue)

			if got != tt.want {
				t.Errorf("promptYesNo() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestPromptSensitive tests the sensitive input prompt function.
func TestPromptSensitive(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		prompt string
		want   string
	}{
		{
			name:   "user provides token",
			input:  "ghp_1234567890abcdef\n",
			prompt: "Enter token",
			want:   "ghp_1234567890abcdef",
		},
		{
			name:   "user provides empty input",
			input:  "\n",
			prompt: "Enter token",
			want:   "",
		},
		{
			name:   "user provides value with whitespace",
			input:  "  token-value  \n",
			prompt: "Enter token",
			want:   "token-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wizard := testWizard(tt.input)
			got := wizard.promptSensitive(tt.prompt)

			if got != tt.want {
				t.Errorf("promptSensitive() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestConfigureBasicSettings tests basic settings configuration.
//
//nolint:dupl // Similar test structure to TestConfigureTemplateSettings by design
func TestConfigureBasicSettings(t *testing.T) {
	tests := []struct {
		name     string
		inputs   string
		wantOrg  string
		wantRepo string
		wantVer  string
	}{
		{
			name:     "all custom values",
			inputs:   "myorg\nmyrepo\nv1.0.0\n",
			wantOrg:  "myorg",
			wantRepo: "myrepo",
			wantVer:  "v1.0.0",
		},
		{
			name:     "use defaults for org and repo, custom version",
			inputs:   "\n\nv2.0.0\n",
			wantOrg:  "",
			wantRepo: "",
			wantVer:  "v2.0.0",
		},
		{
			name:     "custom org and repo, no version",
			inputs:   "testorg\ntestrepo\n\n",
			wantOrg:  "testorg",
			wantRepo: "testrepo",
			wantVer:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wizard := testWizard(tt.inputs)
			wizard.configureBasicSettings()

			if wizard.config.Organization != tt.wantOrg {
				t.Errorf("Organization = %q, want %q", wizard.config.Organization, tt.wantOrg)
			}
			if wizard.config.Repository != tt.wantRepo {
				t.Errorf("Repository = %q, want %q", wizard.config.Repository, tt.wantRepo)
			}
			if wizard.config.Version != tt.wantVer {
				t.Errorf("Version = %q, want %q", wizard.config.Version, tt.wantVer)
			}
		})
	}
}

// TestConfigureThemeSelection tests theme selection.
func TestConfigureThemeSelection(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantTheme string
	}{
		{
			name:      "select default theme (1)",
			input:     "1\n",
			wantTheme: "default",
		},
		{
			name:      "select github theme (2)",
			input:     "2\n",
			wantTheme: "github",
		},
		{
			name:      "select gitlab theme (3)",
			input:     "3\n",
			wantTheme: "gitlab",
		},
		{
			name:      "select minimal theme (4)",
			input:     "4\n",
			wantTheme: "minimal",
		},
		{
			name:      "select professional theme (5)",
			input:     "5\n",
			wantTheme: "professional",
		},
		{
			name:      "invalid choice defaults to first",
			input:     "99\n",
			wantTheme: "default", // Default config theme
		},
		{
			name:      "empty input uses default",
			input:     "\n",
			wantTheme: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wizard := testWizard(tt.input)
			wizard.configureThemeSelection()

			if wizard.config.Theme != tt.wantTheme {
				t.Errorf("Theme = %q, want %q", wizard.config.Theme, tt.wantTheme)
			}
		})
	}
}

// TestConfigureOutputFormat tests output format selection.
func TestConfigureOutputFormat(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantFormat string
	}{
		{
			name:       "select markdown (1)",
			input:      "1\n",
			wantFormat: "md",
		},
		{
			name:       "select html (2)",
			input:      "2\n",
			wantFormat: "html",
		},
		{
			name:       "select json (3)",
			input:      "3\n",
			wantFormat: "json",
		},
		{
			name:       "select asciidoc (4)",
			input:      "4\n",
			wantFormat: "asciidoc",
		},
		{
			name:       "invalid choice keeps default",
			input:      "99\n",
			wantFormat: "md", // Default format
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wizard := testWizard(tt.input)
			wizard.configureOutputFormat()

			if wizard.config.OutputFormat != tt.wantFormat {
				t.Errorf("OutputFormat = %q, want %q", wizard.config.OutputFormat, tt.wantFormat)
			}
		})
	}
}

// TestConfigureFeatures tests feature configuration.
func TestConfigureFeatures(t *testing.T) {
	tests := []struct {
		name                 string
		inputs               string
		wantAnalyzeDeps      bool
		wantShowSecurityInfo bool
	}{
		{
			name:                 "enable both features",
			inputs:               "y\ny\n",
			wantAnalyzeDeps:      true,
			wantShowSecurityInfo: true,
		},
		{
			name:                 "disable both features",
			inputs:               "n\nn\n",
			wantAnalyzeDeps:      false,
			wantShowSecurityInfo: false,
		},
		{
			name:                 "enable deps, disable security",
			inputs:               "yes\nno\n",
			wantAnalyzeDeps:      true,
			wantShowSecurityInfo: false,
		},
		{
			name:                 "use defaults",
			inputs:               "\n\n",
			wantAnalyzeDeps:      false, // Default is false
			wantShowSecurityInfo: false, // Default is false
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wizard := testWizard(tt.inputs)
			wizard.configureFeatures()

			if wizard.config.AnalyzeDependencies != tt.wantAnalyzeDeps {
				t.Errorf("AnalyzeDependencies = %v, want %v", wizard.config.AnalyzeDependencies, tt.wantAnalyzeDeps)
			}
			if wizard.config.ShowSecurityInfo != tt.wantShowSecurityInfo {
				t.Errorf("ShowSecurityInfo = %v, want %v", wizard.config.ShowSecurityInfo, tt.wantShowSecurityInfo)
			}
		})
	}
}

// TestGetAvailableThemes tests the theme list function.
func TestGetAvailableThemes(t *testing.T) {
	wizard := testWizard("")
	themes := wizard.getAvailableThemes()

	if len(themes) != 5 {
		t.Errorf("getAvailableThemes() returned %d themes, want 5", len(themes))
	}

	// Verify theme names
	expectedThemes := []string{"default", "github", "gitlab", "minimal", "professional"}
	for i, expected := range expectedThemes {
		if themes[i].name != expected {
			t.Errorf("Theme %d = %q, want %q", i, themes[i].name, expected)
		}
	}
}

// TestFindActionFiles tests action file discovery.
func TestFindActionFiles(t *testing.T) {
	wizard := testWizard("")

	t.Run("non-existent directory", func(t *testing.T) {
		files := wizard.findActionFiles("/nonexistent/path")
		if len(files) != 0 {
			t.Errorf("findActionFiles() for non-existent dir = %d files, want 0", len(files))
		}
	})

	t.Run("testdata example-action directory", func(t *testing.T) {
		// Get absolute path to avoid traversal issues
		absPath, err := filepath.Abs("../../testdata/example-action")
		if err != nil {
			t.Fatalf("Failed to get absolute path: %v", err)
		}
		files := wizard.findActionFiles(absPath)
		if len(files) == 0 {
			t.Error("findActionFiles() should find action files in testdata/example-action")
		}
	})

	t.Run("testdata composite-action directory", func(t *testing.T) {
		// Get absolute path to avoid traversal issues
		absPath, err := filepath.Abs("../../testdata/composite-action")
		if err != nil {
			t.Fatalf("Failed to get absolute path: %v", err)
		}
		files := wizard.findActionFiles(absPath)
		if len(files) == 0 {
			t.Error("findActionFiles() should find action files in testdata/composite-action")
		}
	})
}

// TestNewConfigWizard tests wizard initialization.
func TestNewConfigWizard(t *testing.T) {
	output := &internal.ColoredOutput{NoColor: true, Quiet: true}
	wizard := NewConfigWizard(output)

	if wizard == nil {
		t.Fatal("NewConfigWizard() returned nil")
	}

	if wizard.output != output {
		t.Error("NewConfigWizard() did not set output correctly")
	}

	if wizard.scanner == nil {
		t.Error("NewConfigWizard() did not initialize scanner")
	}

	if wizard.config == nil {
		t.Error("NewConfigWizard() did not initialize config")
	}

	// Verify default config values
	if wizard.config.Theme == "" {
		t.Error("NewConfigWizard() config has empty theme")
	}

	if wizard.config.OutputFormat == "" {
		t.Error("NewConfigWizard() config has empty output format")
	}
}

// TestConfigureOutputDirectory tests output directory configuration.
func TestConfigureOutputDirectory(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		initial string
		want    string
	}{
		{
			name:    "custom directory",
			input:   "/custom/output\n",
			initial: ".",
			want:    "/custom/output",
		},
		{
			name:    "use default directory",
			input:   "\n",
			initial: "./docs",
			want:    "./docs",
		},
		{
			name:    "relative path",
			input:   "./output\n",
			initial: ".",
			want:    "./output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wizard := testWizard(tt.input)
			wizard.config.OutputDir = tt.initial
			wizard.configureOutputDirectory()

			if wizard.config.OutputDir != tt.want {
				t.Errorf("OutputDir = %q, want %q", wizard.config.OutputDir, tt.want)
			}
		})
	}
}

// TestConfigureTemplateSettings tests template settings configuration.
//
//nolint:dupl // Similar test structure to TestConfigureBasicSettings by design
func TestConfigureTemplateSettings(t *testing.T) {
	tests := []struct {
		name       string
		inputs     string
		wantTheme  string
		wantFormat string
		wantDir    string
	}{
		{
			name:       "all defaults",
			inputs:     "\n\n\n",
			wantTheme:  "default",
			wantFormat: "md",
			wantDir:    ".",
		},
		{
			name:       "custom theme and format",
			inputs:     "2\n3\n./output\n",
			wantTheme:  "github",
			wantFormat: "json",
			wantDir:    "./output",
		},
		{
			name:       "professional theme html format",
			inputs:     "5\n2\n./docs\n",
			wantTheme:  "professional",
			wantFormat: "html",
			wantDir:    "./docs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wizard := testWizard(tt.inputs)
			wizard.configureTemplateSettings()

			if wizard.config.Theme != tt.wantTheme {
				t.Errorf("Theme = %q, want %q", wizard.config.Theme, tt.wantTheme)
			}
			if wizard.config.OutputFormat != tt.wantFormat {
				t.Errorf("OutputFormat = %q, want %q", wizard.config.OutputFormat, tt.wantFormat)
			}
			if wizard.config.OutputDir != tt.wantDir {
				t.Errorf("OutputDir = %q, want %q", wizard.config.OutputDir, tt.wantDir)
			}
		})
	}
}

// TestConfigureGitHubIntegration tests GitHub integration configuration.
func TestConfigureGitHubIntegration(t *testing.T) {
	tests := []struct {
		name           string
		inputs         string
		existingToken  string
		wantTokenSet   bool
		wantTokenValue string
	}{
		{
			name:           "skip token setup",
			inputs:         "n\n",
			existingToken:  "",
			wantTokenSet:   false,
			wantTokenValue: "",
		},
		{
			name:           "provide valid personal token",
			inputs:         "y\nghp_1234567890abcdefghijklmnopqrstuvwxyz\n",
			existingToken:  "",
			wantTokenSet:   true,
			wantTokenValue: "ghp_1234567890abcdefghijklmnopqrstuvwxyz",
		},
		{
			name:           "provide valid PAT token",
			inputs:         "y\ngithub_pat_1234567890abcdefghijklmnopqrstuvwxyz\n",
			existingToken:  "",
			wantTokenSet:   true,
			wantTokenValue: "github_pat_1234567890abcdefghijklmnopqrstuvwxyz",
		},
		{
			name:           "provide unusual token format",
			inputs:         "y\ntoken_unusual_format\n",
			existingToken:  "",
			wantTokenSet:   true,
			wantTokenValue: "token_unusual_format",
		},
		{
			name:           "empty token after yes",
			inputs:         "y\n\n",
			existingToken:  "",
			wantTokenSet:   false,
			wantTokenValue: "",
		},
		{
			name:           "existing token skips setup",
			inputs:         "",
			existingToken:  "ghp_existing_token",
			wantTokenSet:   true,
			wantTokenValue: "ghp_existing_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wizard := testWizard(tt.inputs)
			if tt.existingToken != "" {
				wizard.config.GitHubToken = tt.existingToken
			}

			wizard.configureGitHubIntegration()

			tokenSet := wizard.config.GitHubToken != ""
			if tokenSet != tt.wantTokenSet {
				t.Errorf("Token set = %v, want %v", tokenSet, tt.wantTokenSet)
			}

			if tt.wantTokenSet && wizard.config.GitHubToken != tt.wantTokenValue {
				t.Errorf("GitHubToken = %q, want %q", wizard.config.GitHubToken, tt.wantTokenValue)
			}
		})
	}
}

// TestShowSummaryAndConfirm tests summary display and confirmation.
func TestShowSummaryAndConfirm(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		config  *internal.AppConfig
		wantErr bool
	}{
		{
			name:  "user confirms with yes",
			input: "y\n",
			config: &internal.AppConfig{
				Organization: "testorg",
				Repository:   "testrepo",
				Theme:        "default",
				OutputFormat: "md",
				OutputDir:    ".",
			},
			wantErr: false,
		},
		{
			name:  "user confirms with Y",
			input: "Y\n",
			config: &internal.AppConfig{
				Organization: "testorg",
				Repository:   "testrepo",
			},
			wantErr: false,
		},
		{
			name:  "user cancels with n",
			input: "n\n",
			config: &internal.AppConfig{
				Organization: "testorg",
				Repository:   "testrepo",
			},
			wantErr: true,
		},
		{
			name:  "user cancels with no",
			input: "no\n",
			config: &internal.AppConfig{
				Organization: "testorg",
				Repository:   "testrepo",
			},
			wantErr: true,
		},
		{
			name:  "user accepts default (yes)",
			input: "\n",
			config: &internal.AppConfig{
				Organization: "testorg",
				Repository:   "testrepo",
			},
			wantErr: false,
		},
		{
			name:  "config with version",
			input: "y\n",
			config: &internal.AppConfig{
				Organization: "testorg",
				Repository:   "testrepo",
				Version:      "v1.0.0",
			},
			wantErr: false,
		},
		{
			name:  "config with features enabled",
			input: "y\n",
			config: &internal.AppConfig{
				Organization:        "testorg",
				Repository:          "testrepo",
				AnalyzeDependencies: true,
				ShowSecurityInfo:    true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wizard := testWizard(tt.input)
			wizard.config = tt.config

			err := wizard.showSummaryAndConfirm()

			if (err != nil) != tt.wantErr {
				t.Errorf("showSummaryAndConfirm() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil {
				// Verify error message contains "canceled"
				if !strings.Contains(err.Error(), "canceled") {
					t.Errorf("showSummaryAndConfirm() error = %v, expected 'canceled' in error message", err)
				}
			}
		})
	}
}

// Test verification helpers for TestRun.

func verifyCompleteWizardFlow(t *testing.T, cfg *internal.AppConfig) {
	t.Helper()
	if cfg.Organization != "myorg" {
		t.Errorf("Organization = %q, want 'myorg'", cfg.Organization)
	}
	if cfg.Repository != "myrepo" {
		t.Errorf("Repository = %q, want 'myrepo'", cfg.Repository)
	}
	if cfg.Version != "v1.0.0" {
		t.Errorf("Version = %q, want 'v1.0.0'", cfg.Version)
	}
	if cfg.Theme != "github" {
		t.Errorf("Theme = %q, want 'github'", cfg.Theme)
	}
	if cfg.OutputFormat != "html" {
		t.Errorf("OutputFormat = %q, want 'html'", cfg.OutputFormat)
	}
	if cfg.OutputDir != "./docs" {
		t.Errorf("OutputDir = %q, want './docs'", cfg.OutputDir)
	}
	if !cfg.AnalyzeDependencies {
		t.Error("AnalyzeDependencies should be true")
	}
	if !cfg.ShowSecurityInfo {
		t.Error("ShowSecurityInfo should be true")
	}
}

func verifyWizardDefaults(t *testing.T, cfg *internal.AppConfig) {
	t.Helper()
	const defaultTheme = "default"
	if cfg.Theme != defaultTheme {
		t.Errorf("Theme = %q, want %q", cfg.Theme, defaultTheme)
	}
	if cfg.OutputFormat != "md" {
		t.Errorf("OutputFormat = %q, want 'md'", cfg.OutputFormat)
	}
}

func verifyGitHubToken(t *testing.T, cfg *internal.AppConfig) {
	t.Helper()
	if cfg.GitHubToken != "ghp_testtoken123456" {
		t.Errorf("GitHubToken = %q, want 'ghp_testtoken123456'", cfg.GitHubToken)
	}
}

func verifyMinimalThemeJSON(t *testing.T, cfg *internal.AppConfig) {
	t.Helper()
	if cfg.Theme != "minimal" {
		t.Errorf("Theme = %q, want 'minimal'", cfg.Theme)
	}
	if cfg.OutputFormat != "json" {
		t.Errorf("OutputFormat = %q, want 'json'", cfg.OutputFormat)
	}
	if cfg.OutputDir != "./output" {
		t.Errorf("OutputDir = %q, want './output'", cfg.OutputDir)
	}
	if cfg.AnalyzeDependencies {
		t.Error("AnalyzeDependencies should be false")
	}
	if cfg.ShowSecurityInfo {
		t.Error("ShowSecurityInfo should be false")
	}
}

func verifyGitLabThemeASCIIDoc(t *testing.T, cfg *internal.AppConfig) {
	t.Helper()
	if cfg.Theme != "gitlab" {
		t.Errorf("Theme = %q, want 'gitlab'", cfg.Theme)
	}
	if cfg.OutputFormat != "asciidoc" {
		t.Errorf("OutputFormat = %q, want 'asciidoc'", cfg.OutputFormat)
	}
	if !cfg.AnalyzeDependencies {
		t.Error("AnalyzeDependencies should be true")
	}
	if cfg.ShowSecurityInfo {
		t.Error("ShowSecurityInfo should be false")
	}
}

func verifyProfessionalThemeAllFeatures(t *testing.T, cfg *internal.AppConfig) {
	t.Helper()
	if cfg.Theme != "professional" {
		t.Errorf("Theme = %q, want 'professional'", cfg.Theme)
	}
	if cfg.OutputFormat != "md" {
		t.Errorf("OutputFormat = %q, want 'md'", cfg.OutputFormat)
	}
	if cfg.OutputDir != "." {
		t.Errorf("OutputDir = %q, want '.'", cfg.OutputDir)
	}
	if !cfg.AnalyzeDependencies {
		t.Error("AnalyzeDependencies should be true")
	}
	if !cfg.ShowSecurityInfo {
		t.Error("ShowSecurityInfo should be true")
	}
	if cfg.GitHubToken != "github_pat_testtoken" {
		t.Errorf("GitHubToken = %q, want 'github_pat_testtoken'", cfg.GitHubToken)
	}
}

// TestRun tests the complete wizard workflow.
func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		inputs  string
		wantErr bool
		verify  func(*testing.T, *internal.AppConfig)
	}{
		{
			name: "complete wizard flow with all custom values",
			inputs: "myorg\nmyrepo\nv1.0.0\n" + // Basic settings
				"2\n" + // GitHub theme
				"2\n" + // HTML format
				"./docs\n" + // Output dir
				"y\ny\n" + // Features: enable both
				"n\n" + // GitHub: skip token
				"y\n", // Confirm
			wantErr: false,
			verify:  verifyCompleteWizardFlow,
		},
		{
			name: "wizard with defaults and confirmation",
			inputs: "\n\n\n" + // Basic: all defaults
				"\n\n\n" + // Template: all defaults
				"\n\n" + // Features: all defaults
				"n\n" + // GitHub: skip
				"y\n", // Confirm
			wantErr: false,
			verify:  verifyWizardDefaults,
		},
		{
			name: "wizard with GitHub token",
			inputs: "\n\n\n" + // Basic: all defaults
				"\n\n\n" + // Template: all defaults
				"\n\n" + // Features: all defaults
				"y\nghp_testtoken123456\n" + // GitHub: set token
				"y\n", // Confirm
			wantErr: false,
			verify:  verifyGitHubToken,
		},
		{
			name: "user cancels at confirmation",
			inputs: "testorg\ntestrepo\n\n" + // Basic settings
				"\n\n\n" + // Template: all defaults
				"\n\n" + // Features: all defaults
				"n\n" + // GitHub: skip
				"n\n", // Cancel at confirmation
			wantErr: true,
			verify:  nil,
		},
		{
			name: "minimal theme with json output",
			inputs: "org\nrepo\n\n" + // Basic
				"4\n3\n./output\n" + // Minimal theme, JSON format
				"n\nn\n" + // Features: disable both
				"n\n" + // GitHub: skip
				"y\n", // Confirm
			wantErr: false,
			verify:  verifyMinimalThemeJSON,
		},
		{
			name: "gitlab theme with asciidoc format",
			inputs: "gitlab-org\nmy-project\nv2.5.0\n" + // Basic
				"3\n4\n./docs\n" + // GitLab theme, AsciiDoc format
				"yes\nno\n" + // Features: deps yes, security no
				"n\n" + // GitHub: skip
				"yes\n", // Confirm with 'yes'
			wantErr: false,
			verify:  verifyGitLabThemeASCIIDoc,
		},
		{
			name: "professional theme with all features",
			inputs: "my-org\nawesome-action\n\n" + // Basic (no version)
				"5\n1\n.\n" + // Professional theme, markdown, current dir
				"y\ny\n" + // Features: both enabled
				"y\ngithub_pat_testtoken\n" + // GitHub: set PAT token
				"y\n", // Confirm
			wantErr: false,
			verify:  verifyProfessionalThemeAllFeatures,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wizard := testWizard(tt.inputs)

			config, err := wizard.Run()

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if tt.wantErr {
				if config != nil {
					t.Error("Run() should return nil config on error")
				}

				return
			}

			if config == nil {
				t.Fatal("Run() returned nil config")
			}

			if tt.verify != nil {
				tt.verify(t, config)
			}
		})
	}
}

// TestDetectProjectSettings tests project settings auto-detection.
func TestDetectProjectSettings(t *testing.T) {
	t.Run("detect in non-git directory", func(t *testing.T) {
		wizard := testWizard("")

		// Should not error even if not in git repo
		err := wizard.detectProjectSettings()

		// This should not fail, just log warnings
		if err != nil {
			// Error is acceptable but shouldn't crash
			t.Logf("detectProjectSettings() error = %v (expected in non-git context)", err)
		}

		// Action directory should be set
		if wizard.actionDir == "" {
			t.Error("detectProjectSettings() did not set actionDir")
		}
	})

	t.Run("sets action directory", func(t *testing.T) {
		wizard := testWizard("")

		_ = wizard.detectProjectSettings()

		if wizard.actionDir == "" {
			t.Error("detectProjectSettings() should set actionDir")
		}
	})

	t.Run("detects repo info when available", func(t *testing.T) {
		wizard := testWizard("")

		// This test runs in the project directory which is a git repo
		err := wizard.detectProjectSettings()

		// Should not error
		if err != nil {
			t.Logf("detectProjectSettings() error = %v", err)
		}

		// Should have detected action directory
		if wizard.actionDir == "" {
			t.Error("actionDir should be set")
		}

		// RepoInfo might be set if we're in a git repo
		if wizard.repoInfo != nil {
			t.Logf("Detected repo info: %+v", wizard.repoInfo)
		}
	})
}

// TestShowSummaryWithTokenFromEnv tests summary with token from environment.
func TestShowSummaryWithTokenFromEnv(t *testing.T) {
	const defaultTheme = "default"

	// Test to improve showSummaryAndConfirm coverage
	wizard := testWizard("y\n")
	wizard.config = &internal.AppConfig{
		Organization:        "test",
		Repository:          "repo",
		Theme:               defaultTheme,
		OutputFormat:        "md",
		OutputDir:           ".",
		AnalyzeDependencies: true,
		ShowSecurityInfo:    false,
	}

	// Set env var to simulate token from environment
	t.Setenv("GITHUB_TOKEN", "test_token_from_env")

	err := wizard.showSummaryAndConfirm()
	if err != nil {
		t.Errorf("showSummaryAndConfirm() unexpected error = %v", err)
	}
}

// TestPromptWithDefaultEdgeCases tests edge cases for promptWithDefault.
func TestPromptWithDefaultEdgeCases(t *testing.T) {
	t.Run("scanner error returns default", func(t *testing.T) {
		// Create a wizard with an input that will cause scanner to return false
		wizard := testWizard("")
		// Scanner will immediately return false since input is exhausted
		result := wizard.promptWithDefault("test", "default")
		if result != "default" {
			t.Errorf("Expected default value when scanner fails, got %q", result)
		}
	})
}

// TestPromptYesNoEdgeCases tests edge cases for promptYesNo.
func TestPromptYesNoEdgeCases(t *testing.T) {
	t.Run("scanner error returns default", func(t *testing.T) {
		wizard := testWizard("")
		// Scanner will immediately return false since input is exhausted
		result := wizard.promptYesNo("test", true)
		if result != true {
			t.Errorf("Expected default true when scanner fails, got %v", result)
		}
	})
}

// TestPromptSensitiveEdgeCases tests edge cases for promptSensitive.
func TestPromptSensitiveEdgeCases(t *testing.T) {
	t.Run("scanner error returns empty string", func(t *testing.T) {
		wizard := testWizard("")
		// Scanner will immediately return false since input is exhausted
		result := wizard.promptSensitive("test")
		if result != "" {
			t.Errorf("Expected empty string when scanner fails, got %q", result)
		}
	})

	t.Run("whitespace is trimmed", func(t *testing.T) {
		wizard := testWizard("  value  \n")
		result := wizard.promptSensitive("test")
		if result != "value" {
			t.Errorf("Expected trimmed value, got %q", result)
		}
	})
}

// TestDisplayThemeOptions tests theme display (verifies no panic).
func TestDisplayThemeOptions(t *testing.T) {
	wizard := testWizard("")
	themes := wizard.getAvailableThemes()

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("displayThemeOptions() panicked: %v", r)
		}
	}()

	wizard.displayThemeOptions(themes)
}

// TestDisplayFormatOptions tests format display (verifies no panic).
func TestDisplayFormatOptions(t *testing.T) {
	wizard := testWizard("")
	formats := []string{"md", "html", "json", "asciidoc"}

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("displayFormatOptions() panicked: %v", r)
		}
	}()

	wizard.displayFormatOptions(formats)
}

// TestConfirmConfiguration tests configuration confirmation.
func TestConfirmConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "user confirms",
			input:   "y\n",
			wantErr: false,
		},
		{
			name:    "user cancels",
			input:   "n\n",
			wantErr: true,
		},
		{
			name:    "user accepts default (yes)",
			input:   "\n",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wizard := testWizard(tt.input)
			err := wizard.confirmConfiguration()

			if (err != nil) != tt.wantErr {
				t.Errorf("confirmConfiguration() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
