// Package wizard provides configuration validation functionality.
package wizard

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal"
)

// ValidationResult represents the result of configuration validation.
type ValidationResult struct {
	Valid       bool
	Errors      []ValidationError
	Warnings    []ValidationWarning
	Suggestions []string
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string
	Message string
	Value   string
}

// ValidationWarning represents a validation warning.
type ValidationWarning struct {
	Field   string
	Message string
	Value   string
}

// validPermissionsMap defines valid GitHub Actions permissions and their allowed values.
var validPermissionsMap = map[string][]string{
	"actions":             {"read", "write"},
	"checks":              {"read", "write"},
	"contents":            {"read", "write"},
	"deployments":         {"read", "write"},
	"id-token":            {"write"},
	"issues":              {"read", "write"},
	"discussions":         {"read", "write"},
	"packages":            {"read", "write"},
	"pull-requests":       {"read", "write"},
	"repository-projects": {"read", "write"},
	"security-events":     {"read", "write"},
	"statuses":            {"read", "write"},
}

// ConfigValidator handles configuration validation with immediate feedback.
type ConfigValidator struct {
	output *internal.ColoredOutput
}

// NewConfigValidator creates a new configuration validator.
func NewConfigValidator(output *internal.ColoredOutput) *ConfigValidator {
	return &ConfigValidator{
		output: output,
	}
}

// ValidateConfig validates a complete configuration and returns detailed results.
func (v *ConfigValidator) ValidateConfig(config *internal.AppConfig) *ValidationResult {
	result := &ValidationResult{
		Valid:       true,
		Errors:      []ValidationError{},
		Warnings:    []ValidationWarning{},
		Suggestions: []string{},
	}

	// Validate each field
	v.validateOrganization(config.Organization, result)
	v.validateRepository(config.Repository, result)
	v.validateVersion(config.Version, result)
	v.validateTheme(config.Theme, result)
	v.validateOutputFormat(config.OutputFormat, result)
	v.validateOutputDir(config.OutputDir, result)
	v.validateGitHubToken(config.GitHubToken, result)
	v.validatePermissions(config.Permissions, result)
	v.validateRunsOn(config.RunsOn, result)
	v.validateVariables(config.Variables, result)

	// Set overall validity
	result.Valid = len(result.Errors) == 0

	return result
}

// ValidateField validates a single field and provides immediate feedback.
func (v *ConfigValidator) ValidateField(fieldName, value string) *ValidationResult {
	result := &ValidationResult{
		Valid:       true,
		Errors:      []ValidationError{},
		Warnings:    []ValidationWarning{},
		Suggestions: []string{},
	}

	switch fieldName {
	case "organization":
		v.validateOrganization(value, result)
	case "repository":
		v.validateRepository(value, result)
	case "version":
		v.validateVersion(value, result)
	case appconstants.ConfigKeyTheme:
		v.validateTheme(value, result)
	case appconstants.ConfigKeyOutputFormat:
		v.validateOutputFormat(value, result)
	case appconstants.ConfigKeyOutputDir:
		v.validateOutputDir(value, result)
	case "github_token":
		v.validateGitHubToken(value, result)
	default:
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   fieldName,
			Message: "Unknown field",
			Value:   value,
		})
	}

	result.Valid = len(result.Errors) == 0

	return result
}

// DisplayValidationResult displays validation results to the user.
func (v *ConfigValidator) DisplayValidationResult(result *ValidationResult) {
	if result.Valid {
		v.output.Success("✅ Configuration is valid")
	} else {
		v.output.Error("❌ Configuration has errors")
	}

	// Display errors
	for _, err := range result.Errors {
		v.output.Error("  • %s: %s (value: %s)", err.Field, err.Message, err.Value)
	}

	// Display warnings
	for _, warning := range result.Warnings {
		v.output.Warning("  ⚠️  %s: %s", warning.Field, warning.Message)
	}

	// Display suggestions
	if len(result.Suggestions) > 0 {
		v.output.Info(appconstants.SectionSuggestions)
		for _, suggestion := range result.Suggestions {
			v.output.Printf("  💡 %s", suggestion)
		}
	}
}

// validateOrganization validates the organization field.
func (v *ConfigValidator) validateOrganization(org string, result *ValidationResult) {
	v.validateFieldWithEmptyCheck(
		"organization",
		org,
		v.isValidGitHubName,
		"Organization is empty - will use auto-detected value",
		"Invalid organization name format",
		"Organization names can only contain alphanumeric characters and hyphens",
		result,
	)
}

// validateRepository validates the repository field.
func (v *ConfigValidator) validateRepository(repo string, result *ValidationResult) {
	v.validateFieldWithEmptyCheck(
		"repository",
		repo,
		v.isValidGitHubName,
		"Repository is empty - will use auto-detected value",
		"Invalid repository name format",
		"Repository names can only contain alphanumeric characters, hyphens, and underscores",
		result,
	)
}

// addWarningWithSuggestion is a helper to add a warning and suggestion together.
func addWarningWithSuggestion(result *ValidationResult, field, message, value, suggestion string) {
	result.Warnings = append(result.Warnings, ValidationWarning{
		Field:   field,
		Message: message,
		Value:   value,
	})
	result.Suggestions = append(result.Suggestions, suggestion)
}

// validateVersion validates the version field.
func (v *ConfigValidator) validateVersion(version string, result *ValidationResult) {
	if version == "" {
		// Empty version is valid
		return
	}

	// Check if it follows semantic versioning
	if !v.isValidSemanticVersion(version) {
		addWarningWithSuggestion(result,
			"version",
			"Version does not follow semantic versioning (x.y.z)",
			version,
			"Consider using semantic versioning format (e.g., 1.0.0)")
	}
}

// validateTheme validates the theme field.
func (v *ConfigValidator) validateTheme(theme string, result *ValidationResult) {
	validThemes := []string{
		appconstants.ThemeDefault,
		appconstants.ThemeGitHub,
		appconstants.ThemeGitLab,
		appconstants.ThemeMinimal,
		appconstants.ThemeProfessional,
	}

	v.validateFieldInList("theme", theme, validThemes, "Invalid theme", result)
}

// validateOutputFormat validates the output format field.
func (v *ConfigValidator) validateOutputFormat(format string, result *ValidationResult) {
	validFormats := appconstants.GetSupportedOutputFormats()

	v.validateFieldInList("output_format", format, validFormats, "Invalid output format", result)
}

// validateOutputDir validates the output directory field.
func (v *ConfigValidator) validateOutputDir(dir string, result *ValidationResult) {
	if dir == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "output_dir",
			Message: "Output directory cannot be empty",
			Value:   dir,
		})

		return
	}

	// Check if directory exists or can be created
	if !filepath.IsAbs(dir) {
		// Relative path - check if parent exists
		parent := filepath.Dir(dir)
		if parent != "." {
			if _, err := os.Stat(parent); os.IsNotExist(err) {
				addWarningWithSuggestion(result,
					"output_dir",
					"Parent directory does not exist",
					dir,
					"Ensure the parent directory exists or will be created")
			}
		}
	} else {
		// Absolute path - check if it exists
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			addWarningWithSuggestion(result,
				"output_dir",
				"Directory does not exist",
				dir,
				"Directory will be created if it doesn't exist")
		}
	}
}

// validateGitHubToken validates the GitHub token field.
func (v *ConfigValidator) validateGitHubToken(token string, result *ValidationResult) {
	if token == "" {
		// Empty token is valid (optional)
		return
	}

	// Check token format
	if !v.isValidGitHubToken(token) {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "github_token",
			Message: "Token format looks unusual",
			Value:   "[REDACTED]",
		})
		result.Suggestions = append(result.Suggestions,
			"GitHub tokens usually start with 'ghp_' or 'github_pat_'")
	}

	// Security warning
	result.Warnings = append(result.Warnings, ValidationWarning{
		Field:   "github_token",
		Message: "Tokens should be stored securely in environment variables",
		Value:   "[REDACTED]",
	})
	result.Suggestions = append(result.Suggestions,
		"Consider using GITHUB_TOKEN environment variable instead")
}

// validatePermissionValue validates a single permission value and updates the result.
func (v *ConfigValidator) validatePermissionValue(
	permission, value string,
	validValues []string,
	result *ValidationResult,
) {
	if !v.isValueInList(value, validValues) {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "permissions." + permission,
			Message: "Invalid permission value",
			Value:   value,
		})
		result.Suggestions = append(result.Suggestions,
			fmt.Sprintf("Valid values for %s: %s", permission, strings.Join(validValues, ", ")))
	}
}

// validatePermissions validates the permissions field.
func (v *ConfigValidator) validatePermissions(permissions map[string]string, result *ValidationResult) {
	if len(permissions) == 0 {
		return
	}

	for permission, value := range permissions {
		// Check if permission is valid
		validValues, permissionExists := validPermissionsMap[permission]
		if !permissionExists {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:   "permissions",
				Message: "Unknown permission: " + permission,
				Value:   value,
			})

			continue
		}

		// Check if value is valid
		v.validatePermissionValue(permission, value, validValues, result)
	}
}

// validateRunsOn validates the runs-on field.
func (v *ConfigValidator) validateRunsOn(runsOn []string, result *ValidationResult) {
	if len(runsOn) == 0 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "runs_on",
			Message: "No runners specified",
			Value:   "[]",
		})
		result.Suggestions = append(result.Suggestions,
			"Consider specifying at least one runner (e.g., ubuntu-latest)")

		return
	}

	validRunners := []string{
		appconstants.RunnerUbuntuLatest, "ubuntu-22.04", "ubuntu-20.04",
		appconstants.RunnerWindowsLatest, "windows-2022", "windows-2019",
		appconstants.RunnerMacosLatest, "macos-13", "macos-12", "macos-11",
	}

	for _, runner := range runsOn {
		// Check if it's a GitHub-hosted runner
		isValid := v.isValueInList(runner, validRunners)

		// If not a standard runner, it might be self-hosted
		if !isValid {
			if !strings.HasPrefix(runner, "self-hosted") {
				addWarningWithSuggestion(result,
					"runs_on",
					"Unknown runner: "+runner,
					runner,
					"Ensure the runner is available in your GitHub organization")
			}
		}
	}
}

// validateVariables validates custom variables.
func (v *ConfigValidator) validateVariables(variables map[string]string, result *ValidationResult) {
	if len(variables) == 0 {
		return
	}

	for key, value := range variables {
		// Check for reserved variable names
		reservedNames := []string{"GITHUB_TOKEN", "GITHUB_ACTOR", "GITHUB_REPOSITORY", "GITHUB_SHA"}
		for _, reserved := range reservedNames {
			if strings.EqualFold(key, reserved) {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Field:   "variables",
					Message: "Variable name conflicts with GitHub environment variable: " + key,
					Value:   value,
				})

				break
			}
		}

		// Check for valid variable name format
		if !v.isValidVariableName(key) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "variables",
				Message: "Invalid variable name: " + key,
				Value:   value,
			})
			result.Suggestions = append(result.Suggestions,
				"Variable names should contain only letters, numbers, and underscores")
		}
	}
}

// isValueInList checks if a value exists in a list of valid options.
func (v *ConfigValidator) isValueInList(value string, validOptions []string) bool {
	return slices.Contains(validOptions, value)
}

// isValidGitHubName checks if a name follows GitHub naming rules.
func (v *ConfigValidator) isValidGitHubName(name string) bool {
	if len(name) == 0 || len(name) > 39 {
		return false
	}

	// GitHub names can contain alphanumeric characters and hyphens
	// Cannot start or end with hyphen
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9]([a-zA-Z0-9\-_]*[a-zA-Z0-9])?$`, name)

	return matched
}

// isValidSemanticVersion checks if a version follows semantic versioning.
func (v *ConfigValidator) isValidSemanticVersion(version string) bool {
	// Basic semantic version pattern: x.y.z with optional pre-release and build metadata
	pattern := `^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)` +
		`(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?` +
		`(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`
	matched, _ := regexp.MatchString(pattern, version)

	return matched
}

// isValidGitHubToken checks if a token follows GitHub token format.
func (v *ConfigValidator) isValidGitHubToken(token string) bool {
	// GitHub personal access tokens start with ghp_ or github_pat_
	// Classic tokens are 40 characters after the prefix
	// Fine-grained tokens have different formats
	return strings.HasPrefix(token, appconstants.TokenPrefixGitHubPersonal) ||
		strings.HasPrefix(token, appconstants.TokenPrefixGitHubPAT) ||
		strings.HasPrefix(token, "gho_") ||
		strings.HasPrefix(token, "ghu_") ||
		strings.HasPrefix(token, "ghs_") ||
		strings.HasPrefix(token, "ghr_")
}

// isValidVariableName checks if a variable name is valid.
func (v *ConfigValidator) isValidVariableName(name string) bool {
	if len(name) == 0 {
		return false
	}

	// Variable names should start with letter or underscore
	// and contain only letters, numbers, and underscores
	matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, name)

	return matched
}
