package wizard

import (
	"bufio"
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
	// This test would require filesystem setup
	// For now, testing the logic with a non-existent directory
	wizard := testWizard("")

	// Test with non-existent directory
	files := wizard.findActionFiles("/nonexistent/path")
	if len(files) != 0 {
		t.Errorf("findActionFiles() for non-existent dir = %d files, want 0", len(files))
	}
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
