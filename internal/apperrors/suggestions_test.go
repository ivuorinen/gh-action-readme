package apperrors

import (
	"runtime"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

func TestGetSuggestions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		code     appconstants.ErrorCode
		context  map[string]string
		contains []string
	}{
		{
			name:    "file not found with path",
			code:    appconstants.ErrCodeFileNotFound,
			context: testutil.ContextWithPath("/path/to/action.yml"),
			contains: []string{
				"Check if the file exists: /path/to/action.yml",
				"Verify the file path is correct",
				"--recursive flag",
			},
		},
		{
			name:    "file not found action file",
			code:    appconstants.ErrCodeFileNotFound,
			context: testutil.ContextWithPath("/project/action.yml"),
			contains: []string{
				"Common action file names: action.yml, action.yaml",
				"Check if the file is in a subdirectory",
			},
		},
		{
			name:    "permission denied",
			code:    appconstants.ErrCodePermission,
			context: testutil.ContextWithPath("/restricted/file.txt"),
			contains: []string{
				"Check file permissions: ls -la /restricted/file.txt",
				"chmod 644 /restricted/file.txt",
			},
		},
		{
			name:    "invalid YAML with line number",
			code:    appconstants.ErrCodeInvalidYAML,
			context: testutil.ContextWithLine("25"),
			contains: []string{
				"Error near line 25",
				"Check YAML indentation",
				"use spaces, not tabs",
				"YAML validator",
			},
		},
		{
			name:    "invalid YAML with tab error",
			code:    appconstants.ErrCodeInvalidYAML,
			context: testutil.ContextWithError("found character that cannot start any token (tab)"),
			contains: []string{
				"YAML files must use spaces for indentation, not tabs",
				"Replace all tabs with spaces",
			},
		},
		{
			name:    "invalid action with missing fields",
			code:    appconstants.ErrCodeInvalidAction,
			context: testutil.ContextWithMissingFields("name, description"),
			contains: []string{
				"Missing required fields: name, description",
				"required fields: name, description",
				"gh-action-readme schema",
			},
		},
		{
			name:    testutil.TestCaseNameNoActionFiles,
			code:    appconstants.ErrCodeNoActionFiles,
			context: testutil.ContextWithDirectory("/project"),
			contains: []string{
				"Current directory: /project",
				"find /project -name 'action.y*ml'",
				"--recursive flag",
				"action.yml or action.yaml",
			},
		},
		{
			name:    "GitHub API 401 error",
			code:    appconstants.ErrCodeGitHubAPI,
			context: testutil.ContextWithStatusCode("401"),
			contains: []string{
				"Authentication failed",
				"check your GitHub token",
				"Token may be expired",
			},
		},
		{
			name:    "GitHub API 403 error",
			code:    appconstants.ErrCodeGitHubAPI,
			context: testutil.ContextWithStatusCode("403"),
			contains: []string{
				"Access forbidden",
				"check token permissions",
				"rate limit",
			},
		},
		{
			name:    "GitHub API 404 error",
			code:    appconstants.ErrCodeGitHubAPI,
			context: testutil.ContextWithStatusCode("404"),
			contains: []string{
				"Repository or resource not found",
				"repository is private",
			},
		},
		{
			name:    "GitHub rate limit",
			code:    appconstants.ErrCodeGitHubRateLimit,
			context: testutil.EmptyContext(),
			contains: []string{
				"rate limit exceeded",
				"GITHUB_TOKEN",
				"gh auth login",
				"Rate limits reset every hour",
			},
		},
		{
			name:    "GitHub auth",
			code:    appconstants.ErrCodeGitHubAuth,
			context: testutil.EmptyContext(),
			contains: []string{
				"export GITHUB_TOKEN",
				"gh auth login",
				"https://github.com/settings/tokens",
				"'repo' scope",
			},
		},
		{
			name:    "configuration error with path",
			code:    appconstants.ErrCodeConfiguration,
			context: testutil.ContextWithConfigPath("~/.config/gh-action-readme/config.yaml"),
			contains: []string{
				"Config path: ~/.config/gh-action-readme/config.yaml",
				"ls -la ~/.config/gh-action-readme/config.yaml",
				"gh-action-readme config init",
			},
		},
		{
			name:    "validation error with invalid fields",
			code:    appconstants.ErrCodeValidation,
			context: testutil.ContextWithField("invalid_fields", "runs.using, inputs.test"),
			contains: []string{
				"Invalid fields: runs.using, inputs.test",
				"Check spelling and nesting",
				"gh-action-readme schema",
			},
		},
		{
			name:    "template error with theme",
			code:    appconstants.ErrCodeTemplateRender,
			context: testutil.ContextWithField("theme", "custom"),
			contains: []string{
				"Current theme: custom",
				"Try using a different theme",
				"Available themes:",
			},
		},
		{
			name:    "file write error with output path",
			code:    appconstants.ErrCodeFileWrite,
			context: testutil.ContextWithField("output_path", "/output/README.md"),
			contains: []string{
				"Output directory: /output",
				"Check permissions: ls -la /output",
				"mkdir -p /output",
			},
		},
		{
			name:    "dependency analysis error",
			code:    appconstants.ErrCodeDependencyAnalysis,
			context: testutil.ContextWithField("action", "my-action"),
			contains: []string{
				"Analyzing action: my-action",
				"GitHub token is set",
				"composite actions",
			},
		},
		{
			name:    "cache access error",
			code:    appconstants.ErrCodeCacheAccess,
			context: testutil.ContextWithField("cache_path", "~/.cache/gh-action-readme"),
			contains: []string{
				"Cache path: ~/.cache/gh-action-readme",
				"gh-action-readme cache clear",
				"permissions: ls -la ~/.cache/gh-action-readme",
			},
		},
		{
			name:    "unknown error code",
			code:    "UNKNOWN_TEST_CODE",
			context: testutil.EmptyContext(),
			contains: []string{
				"Check the error message",
				"--verbose flag",
				"project documentation",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			suggestions := GetSuggestions(tt.code, tt.context)
			testutil.AssertSliceContainsAll(t, suggestions, tt.contains)
		})
	}
}

func TestGetPermissionSuggestionsOSSpecific(t *testing.T) {
	t.Parallel()

	context := testutil.ContextWithPath("/test/file")
	suggestions := getPermissionSuggestions(context)

	switch runtime.GOOS {
	case "windows":
		testutil.AssertSliceContainsAll(t, suggestions, []string{"Administrator", "Windows file permissions"})
	default:
		testutil.AssertSliceContainsAll(t, suggestions, []string{"sudo", "ls -la"})
	}
}

func TestGetSuggestionsEmptyContext(t *testing.T) {
	t.Parallel()

	// Test that all error codes work with empty context
	errorCodes := []appconstants.ErrorCode{
		appconstants.ErrCodeFileNotFound,
		appconstants.ErrCodePermission,
		appconstants.ErrCodeInvalidYAML,
		appconstants.ErrCodeInvalidAction,
		appconstants.ErrCodeNoActionFiles,
		appconstants.ErrCodeGitHubAPI,
		appconstants.ErrCodeGitHubRateLimit,
		appconstants.ErrCodeGitHubAuth,
		appconstants.ErrCodeConfiguration,
		appconstants.ErrCodeValidation,
		appconstants.ErrCodeTemplateRender,
		appconstants.ErrCodeFileWrite,
		appconstants.ErrCodeDependencyAnalysis,
		appconstants.ErrCodeCacheAccess,
	}

	for _, code := range errorCodes {
		t.Run(string(code), func(t *testing.T) {
			t.Parallel()

			suggestions := GetSuggestions(code, testutil.EmptyContext())
			if len(suggestions) == 0 {
				t.Errorf("GetSuggestions(%s, {}) returned empty slice", code)
			}
		})
	}
}

func TestGetFileNotFoundSuggestionsActionFile(t *testing.T) {
	t.Parallel()

	context := testutil.ContextWithPath("/project/action.yml")

	suggestions := getFileNotFoundSuggestions(context)
	testutil.AssertSliceContainsAll(t, suggestions, []string{"action.yml, action.yaml", "subdirectory"})
}

func TestGetInvalidYAMLSuggestionsTabError(t *testing.T) {
	t.Parallel()

	context := testutil.ContextWithError("found character that cannot start any token, tab character")

	suggestions := getInvalidYAMLSuggestions(context)
	testutil.AssertSliceContainsAll(t, suggestions, []string{"tabs with spaces"})
}

func TestGetGitHubAPISuggestionsStatusCodes(t *testing.T) {
	t.Parallel()

	statusCodes := map[string]string{
		"401": "Authentication failed",
		"403": "Access forbidden",
		"404": "not found",
	}

	for code, expectedText := range statusCodes {
		t.Run("status_"+code, func(t *testing.T) {
			t.Parallel()

			context := testutil.ContextWithStatusCode(code)
			suggestions := getGitHubAPISuggestions(context)
			testutil.AssertSliceContainsAll(t, suggestions, []string{expectedText})
		})
	}
}

// TestGetValidationSuggestions tests the getValidationSuggestions function.
func TestGetValidationSuggestions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		context          map[string]string
		expectedContains []string
	}{
		{
			name:    "basic validation suggestions",
			context: map[string]string{},
			expectedContains: []string{
				"Review validation errors",
				"Check required fields",
				"Use 'gh-action-readme schema' to see valid structure",
			},
		},
		{
			name:    "with invalid_fields context",
			context: testutil.ContextWithField("invalid_fields", "runs.using, description"),
			expectedContains: []string{
				"Invalid fields: runs.using, description",
				"Check spelling and nesting",
			},
		},
		{
			name:    "with validation_type required",
			context: testutil.ContextWithField("validation_type", "required"),
			expectedContains: []string{
				"Add missing required fields",
				"name, description, runs",
			},
		},
		{
			name:    "with validation_type type",
			context: testutil.ContextWithField("validation_type", "type"),
			expectedContains: []string{
				"Ensure field values match expected types",
				"Strings should be quoted",
			},
		},
		{
			name: "with both invalid_fields and validation_type",
			context: testutil.MergeContexts(
				testutil.ContextWithField("invalid_fields", "name"),
				testutil.ContextWithField("validation_type", "required"),
			),
			expectedContains: []string{
				"Invalid fields: name",
				"Add missing required fields",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			suggestions := getValidationSuggestions(tt.context)
			testutil.AssertSliceContainsAll(t, suggestions, tt.expectedContains)
		})
	}
}

// TestGetConfigurationSuggestions tests the getConfigurationSuggestions function.
func TestGetConfigurationSuggestions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		context          map[string]string
		expectedContains []string
	}{
		{
			name:    "basic configuration suggestions",
			context: map[string]string{},
			expectedContains: []string{
				"Check configuration file syntax",
				"Ensure configuration file exists",
				"Use 'gh-action-readme config init'",
			},
		},
		{
			name:    "with config_path context",
			context: testutil.ContextWithConfigPath("/path/to/config.yaml"),
			expectedContains: []string{
				"Config path: /path/to/config.yaml",
				"Check if file exists: ls -la /path/to/config.yaml",
			},
		},
		{
			name:    "with permission error in context",
			context: testutil.ContextWithError("permission denied"),
			expectedContains: []string{
				"Check file permissions for config file",
				"Ensure parent directory is writable",
			},
		},
		{
			name: "with both config_path and permission error",
			context: testutil.MergeContexts(
				testutil.ContextWithConfigPath("/restricted/config.yaml"),
				testutil.ContextWithError("permission denied while reading"),
			),
			expectedContains: []string{
				"Config path: /restricted/config.yaml",
				"Check file permissions for config file",
			},
		},
		{
			name:    testutil.TestCaseNamePathTraversal,
			context: testutil.ContextWithConfigPath("../../../etc/passwd"),
			expectedContains: []string{
				"Check configuration file syntax",
				"Ensure configuration file exists",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			suggestions := getConfigurationSuggestions(tt.context)
			testutil.AssertSliceContainsAll(t, suggestions, tt.expectedContains)
		})
	}
}

// TestGetTemplateSuggestions tests the getTemplateSuggestions function.
func TestGetTemplateSuggestions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		context          map[string]string
		expectedContains []string
	}{
		{
			name:    "basic template suggestions",
			context: map[string]string{},
			expectedContains: []string{
				"Check template syntax",
				"Ensure all template variables are defined",
				"Verify custom template path is correct",
			},
		},
		{
			name:    "with template_path context",
			context: testutil.ContextWithField("template_path", "/path/to/custom-template.tmpl"),
			expectedContains: []string{
				"Template path: /path/to/custom-template.tmpl",
				"Ensure template file exists and is readable",
			},
		},
		{
			name:    "with theme context",
			context: testutil.ContextWithField("theme", "custom-theme"),
			expectedContains: []string{
				"Current theme: custom-theme",
				"Try using a different theme: --theme github",
				"Available themes: default, github, gitlab, minimal, professional",
			},
		},
		{
			name: "with both template_path and theme",
			context: testutil.MergeContexts(
				testutil.ContextWithField("template_path", "/custom/template.tmpl"),
				testutil.ContextWithField("theme", "github"),
			),
			expectedContains: []string{
				"Template path: /custom/template.tmpl",
				"Current theme: github",
			},
		},
		{
			name:    testutil.TestCaseNamePathTraversal,
			context: testutil.ContextWithField("template_path", "../../../../../../etc/passwd"),
			expectedContains: []string{
				"Check template syntax",
				"Ensure all template variables are defined",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			suggestions := getTemplateSuggestions(tt.context)
			testutil.AssertSliceContainsAll(t, suggestions, tt.expectedContains)
		})
	}
}
