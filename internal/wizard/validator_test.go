package wizard

import (
	"testing"

	"github.com/ivuorinen/gh-action-readme/internal"
)

func TestConfigValidator_ValidateConfig(t *testing.T) {
	output := internal.NewColoredOutput(true) // quiet mode for testing
	validator := NewConfigValidator(output)

	tests := []struct {
		name           string
		config         *internal.AppConfig
		expectValid    bool
		expectErrors   int
		expectWarnings int
	}{
		{
			name: "valid config",
			config: &internal.AppConfig{
				Organization:        "testorg",
				Repository:          "testrepo",
				Version:             "1.0.0",
				Theme:               "github",
				OutputFormat:        "md",
				OutputDir:           ".",
				AnalyzeDependencies: true,
				ShowSecurityInfo:    false,
				RunsOn:              []string{"ubuntu-latest"},
				Permissions:         map[string]string{"contents": "read"},
			},
			expectValid:    true,
			expectErrors:   0,
			expectWarnings: 0,
		},
		{
			name: "invalid theme and format",
			config: &internal.AppConfig{
				Organization: "testorg",
				Repository:   "testrepo",
				Theme:        "invalid-theme",
				OutputFormat: "invalid-format",
				OutputDir:    ".",
			},
			expectValid:  false,
			expectErrors: 2, // theme + format
		},
		{
			name: "empty required fields",
			config: &internal.AppConfig{
				Theme:        "default",
				OutputFormat: "md",
				OutputDir:    "",
			},
			expectValid:  false,
			expectErrors: 1, // output_dir
		},
		{
			name: "invalid permissions",
			config: &internal.AppConfig{
				Organization: "testorg",
				Repository:   "testrepo",
				Theme:        "github",
				OutputFormat: "md",
				OutputDir:    ".",
				Permissions:  map[string]string{"contents": "invalid-value"},
			},
			expectValid:  false,
			expectErrors: 1, // invalid permission value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateConfig(tt.config)

			if result.Valid != tt.expectValid {
				t.Errorf("ValidateConfig() valid = %v, want %v", result.Valid, tt.expectValid)
			}

			if len(result.Errors) != tt.expectErrors {
				t.Errorf("ValidateConfig() errors = %d, want %d", len(result.Errors), tt.expectErrors)
			}

			if tt.expectWarnings > 0 && len(result.Warnings) < tt.expectWarnings {
				t.Errorf("ValidateConfig() warnings = %d, want at least %d", len(result.Warnings), tt.expectWarnings)
			}
		})
	}
}

func TestConfigValidator_ValidateField(t *testing.T) {
	output := internal.NewColoredOutput(true)
	validator := NewConfigValidator(output)

	tests := []struct {
		name        string
		fieldName   string
		value       string
		expectValid bool
	}{
		{"valid organization", "organization", "testorg", true},
		{"invalid organization", "organization", "test@org", false},
		{"valid repository", "repository", "test-repo", true},
		{"invalid repository", "repository", "test repo", false},
		{"valid version", "version", "1.0.0", true},
		{"invalid version", "version", "not-a-version", true}, // warning only
		{"valid theme", "theme", "github", true},
		{"invalid theme", "theme", "nonexistent", false},
		{"valid format", "output_format", "json", true},
		{"invalid format", "output_format", "xml", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateField(tt.fieldName, tt.value)

			if result.Valid != tt.expectValid {
				t.Errorf("ValidateField() valid = %v, want %v", result.Valid, tt.expectValid)
			}
		})
	}
}

func TestConfigValidator_isValidGitHubName(t *testing.T) {
	output := internal.NewColoredOutput(true)
	validator := NewConfigValidator(output)

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid name", "test-org", true},
		{"valid name with numbers", "test123", true},
		{"valid name with underscore", "test_org", true},
		{"empty name", "", false},
		{"name with spaces", "test org", false},
		{"name starting with hyphen", "-test", false},
		{"name ending with hyphen", "test-", false},
		{"name with special chars", "test@org", false},
		{"very long name", "this-is-a-very-long-organization-name-that-exceeds-the-limit", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.isValidGitHubName(tt.input)
			if got != tt.want {
				t.Errorf("isValidGitHubName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestConfigValidator_isValidSemanticVersion(t *testing.T) {
	output := internal.NewColoredOutput(true)
	validator := NewConfigValidator(output)

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid version", "1.0.0", true},
		{"valid version with pre-release", "1.0.0-alpha", true},
		{"valid version with build", "1.0.0+build.1", true},
		{"valid version full", "1.0.0-alpha.1+build.2", true},
		{"invalid version", "1.0", false},
		{"invalid version with letters", "v1.0.0", false},
		{"invalid version format", "1.0.0.0", false},
		{"empty version", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.isValidSemanticVersion(tt.input)
			if got != tt.want {
				t.Errorf("isValidSemanticVersion(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestConfigValidator_isValidGitHubToken(t *testing.T) {
	output := internal.NewColoredOutput(true)
	validator := NewConfigValidator(output)

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"classic token", "ghp_1234567890abcdef1234567890abcdef12345678", true},
		{"fine-grained token", "github_pat_1234567890abcdef", true},
		{"app token", "ghs_1234567890abcdef", true},
		{"oauth token", "gho_1234567890abcdef", true},
		{"user token", "ghu_1234567890abcdef", true},
		{"refresh token", "ghr_1234567890abcdef", true},
		{"invalid token", "invalid_token", false},
		{"empty token", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.isValidGitHubToken(tt.input)
			if got != tt.want {
				t.Errorf("isValidGitHubToken(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestConfigValidator_isValidVariableName(t *testing.T) {
	output := internal.NewColoredOutput(true)
	validator := NewConfigValidator(output)

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid name", "MY_VAR", true},
		{"valid name with underscore", "_MY_VAR", true},
		{"valid name lowercase", "my_var", true},
		{"valid name mixed", "My_Var_123", true},
		{"invalid name with spaces", "MY VAR", false},
		{"invalid name with hyphen", "MY-VAR", false},
		{"invalid name starting with number", "123_VAR", false},
		{"invalid name with special chars", "MY@VAR", false},
		{"empty name", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.isValidVariableName(tt.input)
			if got != tt.want {
				t.Errorf("isValidVariableName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
