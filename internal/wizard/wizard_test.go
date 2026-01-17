package wizard

import (
	"bufio"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/testutil"
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
			prompt:       testutil.WizardPromptEnter,
			defaultValue: appconstants.ThemeDefault,
			want:         "custom-value",
		},
		{
			name:         "user accepts default (empty input)",
			input:        "\n",
			prompt:       testutil.WizardPromptEnter,
			defaultValue: appconstants.ThemeDefault,
			want:         appconstants.ThemeDefault,
		},
		{
			name:         "user provides empty string with no default",
			input:        "\n",
			prompt:       testutil.WizardPromptEnter,
			defaultValue: "",
			want:         "",
		},
		{
			name:         testutil.TestCaseNameUserWhitespace,
			input:        "  value-with-spaces  \n",
			prompt:       testutil.WizardPromptEnter,
			defaultValue: appconstants.ThemeDefault,
			want:         "value-with-spaces",
		},
		{
			name:         "no default provided, user enters value",
			input:        "myvalue\n",
			prompt:       testutil.WizardPromptEnter,
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
			prompt:       testutil.WizardPromptContinue,
			defaultValue: false,
			want:         true,
		},
		{
			name:         "user enters y",
			input:        testutil.WizardInputYes,
			prompt:       testutil.WizardPromptContinue,
			defaultValue: false,
			want:         true,
		},
		{
			name:         "user enters no",
			input:        "no\n",
			prompt:       testutil.WizardPromptContinue,
			defaultValue: true,
			want:         false,
		},
		{
			name:         "user enters n",
			input:        testutil.WizardInputNo,
			prompt:       testutil.WizardPromptContinue,
			defaultValue: true,
			want:         false,
		},
		{
			name:         "user accepts default true",
			input:        "\n",
			prompt:       testutil.WizardPromptContinue,
			defaultValue: true,
			want:         true,
		},
		{
			name:         "user accepts default false",
			input:        "\n",
			prompt:       testutil.WizardPromptContinue,
			defaultValue: false,
			want:         false,
		},
		{
			name:         "invalid input then default",
			input:        "maybe\n",
			prompt:       testutil.WizardPromptContinue,
			defaultValue: true,
			want:         true,
		},
		{
			name:         "case insensitive YES",
			input:        "YES\n",
			prompt:       testutil.WizardPromptContinue,
			defaultValue: false,
			want:         true,
		},
		{
			name:         "case insensitive NO",
			input:        "NO\n",
			prompt:       testutil.WizardPromptContinue,
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
			prompt: testutil.WizardInputEnterToken,
			want:   "ghp_1234567890abcdef",
		},
		{
			name:   "user provides empty input",
			input:  "\n",
			prompt: testutil.WizardInputEnterToken,
			want:   "",
		},
		{
			name:   testutil.TestCaseNameUserWhitespace,
			input:  "  token-value  \n",
			prompt: testutil.WizardInputEnterToken,
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
			wantVer:  testutil.TestVersion,
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
			wantOrg:  testutil.WizardOrgTest,
			wantRepo: testutil.WizardRepoTest,
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
			wantTheme: appconstants.ThemeDefault,
		},
		{
			name:      "select github theme (2)",
			input:     "2\n",
			wantTheme: appconstants.ThemeGitHub,
		},
		{
			name:      "select gitlab theme (3)",
			input:     "3\n",
			wantTheme: appconstants.ThemeGitLab,
		},
		{
			name:      "select minimal theme (4)",
			input:     "4\n",
			wantTheme: appconstants.ThemeMinimal,
		},
		{
			name:      "select professional theme (5)",
			input:     "5\n",
			wantTheme: appconstants.ThemeProfessional,
		},
		{
			name:      "invalid choice defaults to first",
			input:     "99\n",
			wantTheme: appconstants.ThemeDefault, // Default config theme
		},
		{
			name:      "empty input uses default",
			input:     "\n",
			wantTheme: appconstants.ThemeDefault,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wizard := testWizard(tt.input)
			wizard.configureThemeSelection()

			if wizard.config.Theme != tt.wantTheme {
				t.Errorf(testutil.TestMsgThemeFormat, wizard.config.Theme, tt.wantTheme)
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
			wantFormat: appconstants.OutputFormatMarkdown,
		},
		{
			name:       "select html (2)",
			input:      "2\n",
			wantFormat: appconstants.OutputFormatHTML,
		},
		{
			name:       "select json (3)",
			input:      "3\n",
			wantFormat: appconstants.OutputFormatJSON,
		},
		{
			name:       "select asciidoc (4)",
			input:      "4\n",
			wantFormat: "asciidoc",
		},
		{
			name:       "invalid choice keeps default",
			input:      "99\n",
			wantFormat: appconstants.OutputFormatMarkdown, // Default format
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
			inputs:               testutil.WizardInputYesNewline,
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
	expectedThemes := []string{
		appconstants.ThemeDefault,
		appconstants.ThemeGitHub,
		appconstants.ThemeGitLab,
		appconstants.ThemeMinimal,
		appconstants.ThemeProfessional,
	}
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
			initial: testutil.TestDirDocs,
			want:    testutil.TestDirDocs,
		},
		{
			name:    testutil.TestCaseNameRelativePath,
			input:   testutil.TestDirOutput + "\n",
			initial: ".",
			want:    testutil.TestDirOutput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wizard := testWizard(tt.input)
			wizard.config.OutputDir = tt.initial
			wizard.configureOutputDirectory()

			if wizard.config.OutputDir != tt.want {
				t.Errorf(testutil.ErrOutputDirMismatch, wizard.config.OutputDir, tt.want)
			}
		})
	}
}

// TestConfigureTemplateSettings tests template settings configuration.
//

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
			inputs:     testutil.WizardInputThreeNewlines,
			wantTheme:  appconstants.ThemeDefault,
			wantFormat: appconstants.OutputFormatMarkdown,
			wantDir:    ".",
		},
		{
			name:       "custom theme and format",
			inputs:     "2\n3\n" + testutil.TestDirOutput + "\n",
			wantTheme:  appconstants.ThemeGitHub,
			wantFormat: appconstants.OutputFormatJSON,
			wantDir:    testutil.TestDirOutput,
		},
		{
			name:       "professional theme html format",
			inputs:     "5\n2\n" + testutil.TestDirDocs + "\n",
			wantTheme:  appconstants.ThemeProfessional,
			wantFormat: appconstants.OutputFormatHTML,
			wantDir:    testutil.TestDirDocs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wizard := testWizard(tt.inputs)
			wizard.configureTemplateSettings()

			if wizard.config.Theme != tt.wantTheme {
				t.Errorf(testutil.TestMsgThemeFormat, wizard.config.Theme, tt.wantTheme)
			}
			if wizard.config.OutputFormat != tt.wantFormat {
				t.Errorf("OutputFormat = %q, want %q", wizard.config.OutputFormat, tt.wantFormat)
			}
			if wizard.config.OutputDir != tt.wantDir {
				t.Errorf(testutil.ErrOutputDirMismatch, wizard.config.OutputDir, tt.wantDir)
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
			inputs:         testutil.WizardInputNo,
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
			input: testutil.WizardInputYes,
			config: &internal.AppConfig{
				Organization: testutil.WizardOrgTest,
				Repository:   testutil.WizardRepoTest,
				Theme:        appconstants.ThemeDefault,
				OutputFormat: appconstants.OutputFormatMarkdown,
				OutputDir:    ".",
			},
			wantErr: false,
		},
		{
			name:  "user confirms with Y",
			input: "Y\n",
			config: &internal.AppConfig{
				Organization: testutil.WizardOrgTest,
				Repository:   testutil.WizardRepoTest,
			},
			wantErr: false,
		},
		{
			name:  "user cancels with n",
			input: testutil.WizardInputNo,
			config: &internal.AppConfig{
				Organization: testutil.WizardOrgTest,
				Repository:   testutil.WizardRepoTest,
			},
			wantErr: true,
		},
		{
			name:  "user cancels with no",
			input: "no\n",
			config: &internal.AppConfig{
				Organization: testutil.WizardOrgTest,
				Repository:   testutil.WizardRepoTest,
			},
			wantErr: true,
		},
		{
			name:  testutil.TestCaseNameUserAcceptDefault,
			input: "\n",
			config: &internal.AppConfig{
				Organization: testutil.WizardOrgTest,
				Repository:   testutil.WizardRepoTest,
			},
			wantErr: false,
		},
		{
			name:  "config with version",
			input: testutil.WizardInputYes,
			config: &internal.AppConfig{
				Organization: testutil.WizardOrgTest,
				Repository:   testutil.WizardRepoTest,
				Version:      testutil.TestVersion,
			},
			wantErr: false,
		},
		{
			name:  "config with features enabled",
			input: testutil.WizardInputYes,
			config: &internal.AppConfig{
				Organization:        testutil.WizardOrgTest,
				Repository:          testutil.WizardRepoTest,
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
	if cfg.Version != testutil.TestVersion {
		t.Errorf("Version = %q, want 'v1.0.0'", cfg.Version)
	}
	if cfg.Theme != appconstants.ThemeGitHub {
		t.Errorf("Theme = %q, want 'github'", cfg.Theme)
	}
	if cfg.OutputFormat != appconstants.OutputFormatHTML {
		t.Errorf("OutputFormat = %q, want 'html'", cfg.OutputFormat)
	}
	if cfg.OutputDir != testutil.TestDirDocs {
		t.Errorf(testutil.ErrOutputDirMismatch, cfg.OutputDir, testutil.TestDirDocs)
	}
	if !cfg.AnalyzeDependencies {
		t.Error(testutil.TestMsgAnalyzeDepsTrue)
	}
	if !cfg.ShowSecurityInfo {
		t.Error("ShowSecurityInfo should be true")
	}
}

func verifyWizardDefaults(t *testing.T, cfg *internal.AppConfig) {
	t.Helper()
	const defaultTheme = appconstants.ThemeDefault
	if cfg.Theme != defaultTheme {
		t.Errorf(testutil.TestMsgThemeFormat, cfg.Theme, defaultTheme)
	}
	if cfg.OutputFormat != appconstants.OutputFormatMarkdown {
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
	if cfg.Theme != appconstants.ThemeMinimal {
		t.Errorf("Theme = %q, want 'minimal'", cfg.Theme)
	}
	if cfg.OutputFormat != appconstants.OutputFormatJSON {
		t.Errorf("OutputFormat = %q, want 'json'", cfg.OutputFormat)
	}
	if cfg.OutputDir != testutil.TestDirOutput {
		t.Errorf(testutil.ErrOutputDirMismatch, cfg.OutputDir, testutil.TestDirOutput)
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
	if cfg.Theme != appconstants.ThemeGitLab {
		t.Errorf("Theme = %q, want 'gitlab'", cfg.Theme)
	}
	if cfg.OutputFormat != "asciidoc" {
		t.Errorf("OutputFormat = %q, want 'asciidoc'", cfg.OutputFormat)
	}
	if !cfg.AnalyzeDependencies {
		t.Error(testutil.TestMsgAnalyzeDepsTrue)
	}
	if cfg.ShowSecurityInfo {
		t.Error("ShowSecurityInfo should be false")
	}
}

func verifyProfessionalThemeAllFeatures(t *testing.T, cfg *internal.AppConfig) {
	t.Helper()
	if cfg.Theme != appconstants.ThemeProfessional {
		t.Errorf("Theme = %q, want 'professional'", cfg.Theme)
	}
	if cfg.OutputFormat != appconstants.OutputFormatMarkdown {
		t.Errorf("OutputFormat = %q, want 'md'", cfg.OutputFormat)
	}
	if cfg.OutputDir != "." {
		t.Errorf("OutputDir = %q, want '.'", cfg.OutputDir)
	}
	if !cfg.AnalyzeDependencies {
		t.Error(testutil.TestMsgAnalyzeDepsTrue)
	}
	if !cfg.ShowSecurityInfo {
		t.Error("ShowSecurityInfo should be true")
	}
	if cfg.GitHubToken != "github_pat_testtoken" {
		t.Errorf("GitHubToken = %q, want 'github_pat_testtoken'", cfg.GitHubToken)
	}
}

// TestRun tests the complete wizard workflow.
// verifyWizardTestResult validates the result of a wizard Run() call.
func verifyWizardTestResult(
	t *testing.T,
	err error,
	wantErr bool,
	config *internal.AppConfig,
	verify func(*testing.T, *internal.AppConfig),
) {
	t.Helper()

	if (err != nil) != wantErr {
		t.Errorf("Run() error = %v, wantErr %v", err, wantErr)

		return
	}

	if wantErr {
		if config != nil {
			t.Error("Run() should return nil config on error")
		}

		return
	}

	if config == nil {
		t.Fatal("Run() returned nil config")
	}

	if verify != nil {
		verify(t, config)
	}
}

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
				testutil.TestDirDocs + "\n" + // Output dir
				testutil.WizardInputYesNewline + // Features: enable both
				testutil.WizardInputNo + // GitHub: skip token
				testutil.WizardInputYes, // Confirm
			wantErr: false,
			verify:  verifyCompleteWizardFlow,
		},
		{
			name: "wizard with defaults and confirmation",
			inputs: testutil.WizardInputThreeNewlines + // Basic: all defaults
				testutil.WizardInputThreeNewlines + // Template: all defaults
				"\n\n" + // Features: all defaults
				testutil.WizardInputNo + // GitHub: skip
				testutil.WizardInputYes, // Confirm
			wantErr: false,
			verify:  verifyWizardDefaults,
		},
		{
			name: "wizard with GitHub token",
			inputs: testutil.WizardInputThreeNewlines + // Basic: all defaults
				testutil.WizardInputThreeNewlines + // Template: all defaults
				"\n\n" + // Features: all defaults
				"y\nghp_testtoken123456\n" + // GitHub: set token
				testutil.WizardInputYes, // Confirm
			wantErr: false,
			verify:  verifyGitHubToken,
		},
		{
			name: "user cancels at confirmation",
			inputs: "testorg\ntestrepo\n\n" + // Basic settings
				testutil.WizardInputThreeNewlines + // Template: all defaults
				"\n\n" + // Features: all defaults
				testutil.WizardInputNo + // GitHub: skip
				testutil.WizardInputNo, // Cancel at confirmation
			wantErr: true,
			verify:  nil,
		},
		{
			name: "minimal theme with json output",
			inputs: "org\nrepo\n\n" + // Basic
				"4\n3\n" + testutil.TestDirOutput + "\n" + // Minimal theme, JSON format
				"n\nn\n" + // Features: disable both
				testutil.WizardInputNo + // GitHub: skip
				testutil.WizardInputYes, // Confirm
			wantErr: false,
			verify:  verifyMinimalThemeJSON,
		},
		{
			name: "gitlab theme with asciidoc format",
			inputs: "gitlab-org\nmy-project\nv2.5.0\n" + // Basic
				"3\n4\n" + testutil.TestDirDocs + "\n" + // GitLab theme, AsciiDoc format
				"yes\nno\n" + // Features: deps yes, security no
				testutil.WizardInputNo + // GitHub: skip
				"yes\n", // Confirm with 'yes'
			wantErr: false,
			verify:  verifyGitLabThemeASCIIDoc,
		},
		{
			name: "professional theme with all features",
			inputs: "my-org\nawesome-action\n\n" + // Basic (no version)
				"5\n1\n.\n" + // Professional theme, markdown, current dir
				testutil.WizardInputYesNewline + // Features: both enabled
				"y\ngithub_pat_testtoken\n" + // GitHub: set PAT token
				testutil.WizardInputYes, // Confirm
			wantErr: false,
			verify:  verifyProfessionalThemeAllFeatures,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wizard := testWizard(tt.inputs)

			config, err := wizard.Run()

			verifyWizardTestResult(t, err, tt.wantErr, config, tt.verify)
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
	const defaultTheme = appconstants.ThemeDefault

	// Test to improve showSummaryAndConfirm coverage
	wizard := testWizard(testutil.WizardInputYes)
	wizard.config = &internal.AppConfig{
		Organization:        "test",
		Repository:          "repo",
		Theme:               defaultTheme,
		OutputFormat:        appconstants.OutputFormatMarkdown,
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
		result := wizard.promptWithDefault("test", appconstants.ThemeDefault)
		if result != appconstants.ThemeDefault {
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
	formats := []string{
		appconstants.OutputFormatMarkdown,
		appconstants.OutputFormatHTML,
		appconstants.OutputFormatJSON,
		"asciidoc",
	}

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
			input:   testutil.WizardInputYes,
			wantErr: false,
		},
		{
			name:    "user cancels",
			input:   testutil.WizardInputNo,
			wantErr: true,
		},
		{
			name:    testutil.TestCaseNameUserAcceptDefault,
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
