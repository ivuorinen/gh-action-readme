package wizard

import (
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// newTestValidator creates a ConfigValidator for testing with quiet output.
// Reduces duplication across validator tests.
func newTestValidator() *ConfigValidator {
	output := internal.NewColoredOutput(true)

	return NewConfigValidator(output)
}

func TestConfigValidatorValidateConfig(t *testing.T) {
	t.Parallel()
	validator := newTestValidator()

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
			t.Parallel()
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

func TestConfigValidatorValidateField(t *testing.T) {
	t.Parallel()
	validator := newTestValidator()

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
			t.Parallel()
			result := validator.ValidateField(tt.fieldName, tt.value)

			if result.Valid != tt.expectValid {
				t.Errorf("ValidateField() valid = %v, want %v", result.Valid, tt.expectValid)
			}
		})
	}
}

func TestConfigValidatorIsValidGitHubName(t *testing.T) {
	t.Parallel()
	validator := newTestValidator()

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
			t.Parallel()
			got := validator.isValidGitHubName(tt.input)
			if got != tt.want {
				t.Errorf("isValidGitHubName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestConfigValidatorIsValidSemanticVersion(t *testing.T) {
	t.Parallel()
	validator := newTestValidator()

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
			t.Parallel()
			got := validator.isValidSemanticVersion(tt.input)
			if got != tt.want {
				t.Errorf("isValidSemanticVersion(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestConfigValidatorIsValidGitHubToken(t *testing.T) {
	t.Parallel()
	validator := newTestValidator()

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
			t.Parallel()
			got := validator.isValidGitHubToken(tt.input)
			if got != tt.want {
				t.Errorf("isValidGitHubToken(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestConfigValidatorIsValidVariableName(t *testing.T) {
	t.Parallel()
	validator := newTestValidator()

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
			t.Parallel()
			got := validator.isValidVariableName(tt.input)
			if got != tt.want {
				t.Errorf("isValidVariableName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestDisplayValidationResult tests DisplayValidationResult UI output.
func TestDisplayValidationResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		result         *ValidationResult
		expectSuccess  bool
		expectErrors   bool
		expectWarnings bool
	}{
		{
			name: "valid configuration",
			result: &ValidationResult{
				Valid:       true,
				Errors:      []ValidationError{},
				Warnings:    []ValidationWarning{},
				Suggestions: []string{},
			},
			expectSuccess: true,
		},
		{
			name: "configuration with errors",
			result: &ValidationResult{
				Valid: false,
				Errors: []ValidationError{
					{Field: "organization", Message: testutil.TestMsgCannotBeEmpty, Value: ""},
					{Field: "repository", Message: "invalid format", Value: "test"},
				},
				Warnings:    []ValidationWarning{},
				Suggestions: []string{},
			},
			expectErrors: true,
		},
		{
			name: "configuration with warnings",
			result: &ValidationResult{
				Valid:  true,
				Errors: []ValidationError{},
				Warnings: []ValidationWarning{
					{Field: "github_token", Message: "should be stored securely"},
					{Field: "output_dir", Message: "directory does not exist"},
				},
				Suggestions: []string{},
			},
			expectSuccess:  true,
			expectWarnings: true,
		},
		{
			name: "configuration with errors and warnings",
			result: &ValidationResult{
				Valid: false,
				Errors: []ValidationError{
					{Field: "organization", Message: testutil.TestMsgCannotBeEmpty, Value: ""},
				},
				Warnings: []ValidationWarning{
					{Field: "github_token", Message: "should be stored securely"},
				},
				Suggestions: []string{},
			},
			expectErrors:   true,
			expectWarnings: true,
		},
		{
			name: "configuration with suggestions",
			result: &ValidationResult{
				Valid: false,
				Errors: []ValidationError{
					{Field: "organization", Message: testutil.TestMsgCannotBeEmpty, Value: ""},
				},
				Warnings: []ValidationWarning{},
				Suggestions: []string{
					"Check the organization name",
					"Verify repository settings",
				},
			},
			expectErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			validator := newTestValidator()

			// This should not panic
			validator.DisplayValidationResult(tt.result)
		})
	}
}

// TestValidateGitHubToken_AllTokenTypes tests all GitHub token formats.
func TestValidateGitHubToken_AllTokenTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		token          string
		expectWarnings int // Number of warnings expected
	}{
		{
			name:           "classic token (ghp_)",
			token:          "ghp_1234567890abcdefghijklmnopqrstuvwxyzABCD",
			expectWarnings: 1, // Security warning only
		},
		{
			name:           "fine-grained token (github_pat_)",
			token:          "github_pat_11AAAAAA0AAAAaAaaAaaaAaa_AaAAaAAaAAAaAAAAAaAAaAAaAaAAaAAAAaAAAAAAAAaAAaAAaAaaAA",
			expectWarnings: 1, // Security warning only
		},
		{
			name:           "OAuth token (gho_)",
			token:          "gho_1234567890abcdefghijklmnopqrstuvwxyzABCD",
			expectWarnings: 1, // Security warning only
		},
		{
			name:           "user token (ghu_)",
			token:          "ghu_1234567890abcdefghijklmnopqrstuvwxyzABCD",
			expectWarnings: 1, // Security warning only
		},
		{
			name:           "server token (ghs_)",
			token:          "ghs_1234567890abcdefghijklmnopqrstuvwxyzABCD",
			expectWarnings: 1, // Security warning only
		},
		{
			name:           "refresh token (ghr_)",
			token:          "ghr_1234567890abcdefghijklmnopqrstuvwxyzABCD",
			expectWarnings: 1, // Security warning only
		},
		{
			name:           "invalid token format",
			token:          "invalid_token_format",
			expectWarnings: 2, // Format warning + security warning
		},
		{
			name:           "empty token",
			token:          "",
			expectWarnings: 0, // Empty is valid (optional)
		},
		{
			name:           "token too short",
			token:          "ghp_abc",
			expectWarnings: 1, // Just security warning (format is valid prefix)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			validator := newTestValidator()

			result := &ValidationResult{Valid: true}
			validator.validateGitHubToken(tt.token, result)

			if len(result.Warnings) != tt.expectWarnings {
				t.Errorf("validateGitHubToken(%q) warnings = %d, want %d",
					tt.token, len(result.Warnings), tt.expectWarnings)
			}
		})
	}
}

// TestValidateVariables_ReservedNames tests all reserved variable names.
func TestValidateVariables_ReservedNames(t *testing.T) {
	t.Parallel()

	reservedNames := []string{
		"GITHUB_TOKEN",
		"GITHUB_ACTOR",
		"GITHUB_REPOSITORY",
		"GITHUB_SHA",
	}

	for _, reserved := range reservedNames {
		t.Run(reserved, func(t *testing.T) {
			t.Parallel()
			validator := newTestValidator()

			variables := map[string]string{
				reserved: "some-value",
			}

			result := &ValidationResult{Valid: true}
			validator.validateVariables(variables, result)

			// Should have warning about reserved name
			found := false
			for _, warning := range result.Warnings {
				if strings.Contains(warning.Message, "conflicts with GitHub environment variable") &&
					strings.Contains(warning.Message, reserved) {
					found = true

					break
				}
			}
			if !found {
				t.Errorf("expected warning about reserved variable name %s", reserved)
			}
		})
	}
}

// TestValidateVariables_InvalidFormats tests variable name validation.
func TestValidateVariables_InvalidFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		varName     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid uppercase with underscores",
			varName:     "MY_CUSTOM_VAR",
			expectError: false,
		},
		{
			name:        "valid single word",
			varName:     "VERSION",
			expectError: false,
		},
		{
			name:        "lowercase allowed",
			varName:     "my_var",
			expectError: false,
		},
		{
			name:        "starts with number",
			varName:     "1_VAR",
			expectError: true,
			errorMsg:    testutil.TestMsgInvalidVariableName,
		},
		{
			name:        "contains hyphen",
			varName:     "MY-VAR",
			expectError: true,
			errorMsg:    testutil.TestMsgInvalidVariableName,
		},
		{
			name:        "contains space",
			varName:     "MY VAR",
			expectError: true,
			errorMsg:    testutil.TestMsgInvalidVariableName,
		},
		{
			name:        "contains special characters",
			varName:     "MY_VAR!",
			expectError: true,
			errorMsg:    testutil.TestMsgInvalidVariableName,
		},
		{
			name:        "empty variable name",
			varName:     "",
			expectError: true,
			errorMsg:    testutil.TestMsgInvalidVariableName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			validator := newTestValidator()

			variables := map[string]string{
				tt.varName: "value",
			}

			result := &ValidationResult{Valid: true}
			validator.validateVariables(variables, result)

			assertValidationError(t, result, "variables", tt.expectError, tt.errorMsg)
		})
	}
}

// TestValidateOutputDir_Paths tests various path formats.
func TestValidateOutputDir_Paths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		outputDir   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "absolute path existing",
			outputDir:   "/tmp",
			expectError: false,
		},
		{
			name:        "relative path current dir",
			outputDir:   ".",
			expectError: false,
		},
		{
			name:        "empty path",
			outputDir:   "",
			expectError: true,
			errorMsg:    testutil.TestMsgCannotBeEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			validator := newTestValidator()

			result := &ValidationResult{Valid: true}
			validator.validateOutputDir(tt.outputDir, result)

			assertValidationError(t, result, "output_dir", tt.expectError, tt.errorMsg)
		})
	}
}
