package errors

import (
	"runtime"
	"strings"
	"testing"
)

func TestGetSuggestions(t *testing.T) {
	tests := []struct {
		name     string
		code     ErrorCode
		context  map[string]string
		contains []string
	}{
		{
			name: "file not found with path",
			code: ErrCodeFileNotFound,
			context: map[string]string{
				"path": "/path/to/action.yml",
			},
			contains: []string{
				"Check if the file exists: /path/to/action.yml",
				"Verify the file path is correct",
				"--recursive flag",
			},
		},
		{
			name: "file not found action file",
			code: ErrCodeFileNotFound,
			context: map[string]string{
				"path": "/project/action.yml",
			},
			contains: []string{
				"Common action file names: action.yml, action.yaml",
				"Check if the file is in a subdirectory",
			},
		},
		{
			name: "permission denied",
			code: ErrCodePermission,
			context: map[string]string{
				"path": "/restricted/file.txt",
			},
			contains: []string{
				"Check file permissions: ls -la /restricted/file.txt",
				"chmod 644 /restricted/file.txt",
			},
		},
		{
			name: "invalid YAML with line number",
			code: ErrCodeInvalidYAML,
			context: map[string]string{
				"line": "25",
			},
			contains: []string{
				"Error near line 25",
				"Check YAML indentation",
				"use spaces, not tabs",
				"YAML validator",
			},
		},
		{
			name: "invalid YAML with tab error",
			code: ErrCodeInvalidYAML,
			context: map[string]string{
				"error": "found character that cannot start any token (tab)",
			},
			contains: []string{
				"YAML files must use spaces for indentation, not tabs",
				"Replace all tabs with spaces",
			},
		},
		{
			name: "invalid action with missing fields",
			code: ErrCodeInvalidAction,
			context: map[string]string{
				"missing_fields": "name, description",
			},
			contains: []string{
				"Missing required fields: name, description",
				"required fields: name, description",
				"gh-action-readme schema",
			},
		},
		{
			name: "no action files",
			code: ErrCodeNoActionFiles,
			context: map[string]string{
				"directory": "/project",
			},
			contains: []string{
				"Current directory: /project",
				"find /project -name 'action.y*ml'",
				"--recursive flag",
				"action.yml or action.yaml",
			},
		},
		{
			name: "GitHub API 401 error",
			code: ErrCodeGitHubAPI,
			context: map[string]string{
				"status_code": "401",
			},
			contains: []string{
				"Authentication failed",
				"check your GitHub token",
				"Token may be expired",
			},
		},
		{
			name: "GitHub API 403 error",
			code: ErrCodeGitHubAPI,
			context: map[string]string{
				"status_code": "403",
			},
			contains: []string{
				"Access forbidden",
				"check token permissions",
				"rate limit",
			},
		},
		{
			name: "GitHub API 404 error",
			code: ErrCodeGitHubAPI,
			context: map[string]string{
				"status_code": "404",
			},
			contains: []string{
				"Repository or resource not found",
				"repository is private",
			},
		},
		{
			name:    "GitHub rate limit",
			code:    ErrCodeGitHubRateLimit,
			context: map[string]string{},
			contains: []string{
				"rate limit exceeded",
				"GITHUB_TOKEN",
				"gh auth login",
				"Rate limits reset every hour",
			},
		},
		{
			name:    "GitHub auth",
			code:    ErrCodeGitHubAuth,
			context: map[string]string{},
			contains: []string{
				"export GITHUB_TOKEN",
				"gh auth login",
				"https://github.com/settings/tokens",
				"'repo' scope",
			},
		},
		{
			name: "configuration error with path",
			code: ErrCodeConfiguration,
			context: map[string]string{
				"config_path": "~/.config/gh-action-readme/config.yaml",
			},
			contains: []string{
				"Config path: ~/.config/gh-action-readme/config.yaml",
				"ls -la ~/.config/gh-action-readme/config.yaml",
				"gh-action-readme config init",
			},
		},
		{
			name: "validation error with invalid fields",
			code: ErrCodeValidation,
			context: map[string]string{
				"invalid_fields": "runs.using, inputs.test",
			},
			contains: []string{
				"Invalid fields: runs.using, inputs.test",
				"Check spelling and nesting",
				"gh-action-readme schema",
			},
		},
		{
			name: "template error with theme",
			code: ErrCodeTemplateRender,
			context: map[string]string{
				"theme": "custom",
			},
			contains: []string{
				"Current theme: custom",
				"Try using a different theme",
				"Available themes:",
			},
		},
		{
			name: "file write error with output path",
			code: ErrCodeFileWrite,
			context: map[string]string{
				"output_path": "/output/README.md",
			},
			contains: []string{
				"Output directory: /output",
				"Check permissions: ls -la /output",
				"mkdir -p /output",
			},
		},
		{
			name: "dependency analysis error",
			code: ErrCodeDependencyAnalysis,
			context: map[string]string{
				"action": "my-action",
			},
			contains: []string{
				"Analyzing action: my-action",
				"GitHub token is set",
				"composite actions",
			},
		},
		{
			name: "cache access error",
			code: ErrCodeCacheAccess,
			context: map[string]string{
				"cache_path": "~/.cache/gh-action-readme",
			},
			contains: []string{
				"Cache path: ~/.cache/gh-action-readme",
				"gh-action-readme cache clear",
				"permissions: ls -la ~/.cache/gh-action-readme",
			},
		},
		{
			name:    "unknown error code",
			code:    "UNKNOWN_TEST_CODE",
			context: map[string]string{},
			contains: []string{
				"Check the error message",
				"--verbose flag",
				"project documentation",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := GetSuggestions(tt.code, tt.context)

			if len(suggestions) == 0 {
				t.Error("GetSuggestions() returned empty slice")
				return
			}

			allSuggestions := strings.Join(suggestions, " ")
			for _, expected := range tt.contains {
				if !strings.Contains(allSuggestions, expected) {
					t.Errorf(
						"GetSuggestions() missing expected content:\nExpected to contain: %q\nSuggestions:\n%s",
						expected,
						strings.Join(suggestions, "\n"),
					)
				}
			}
		})
	}
}

func TestGetPermissionSuggestions_OSSpecific(t *testing.T) {
	context := map[string]string{"path": "/test/file"}
	suggestions := getPermissionSuggestions(context)

	allSuggestions := strings.Join(suggestions, " ")

	switch runtime.GOOS {
	case "windows":
		if !strings.Contains(allSuggestions, "Administrator") {
			t.Error("Windows-specific suggestions should mention Administrator")
		}
		if !strings.Contains(allSuggestions, "Windows file permissions") {
			t.Error("Windows-specific suggestions should mention Windows file permissions")
		}
	default:
		if !strings.Contains(allSuggestions, "sudo") {
			t.Error("Unix-specific suggestions should mention sudo")
		}
		if !strings.Contains(allSuggestions, "ls -la") {
			t.Error("Unix-specific suggestions should mention ls -la")
		}
	}
}

func TestGetSuggestions_EmptyContext(t *testing.T) {
	// Test that all error codes work with empty context
	errorCodes := []ErrorCode{
		ErrCodeFileNotFound,
		ErrCodePermission,
		ErrCodeInvalidYAML,
		ErrCodeInvalidAction,
		ErrCodeNoActionFiles,
		ErrCodeGitHubAPI,
		ErrCodeGitHubRateLimit,
		ErrCodeGitHubAuth,
		ErrCodeConfiguration,
		ErrCodeValidation,
		ErrCodeTemplateRender,
		ErrCodeFileWrite,
		ErrCodeDependencyAnalysis,
		ErrCodeCacheAccess,
	}

	for _, code := range errorCodes {
		t.Run(string(code), func(t *testing.T) {
			suggestions := GetSuggestions(code, map[string]string{})
			if len(suggestions) == 0 {
				t.Errorf("GetSuggestions(%s, {}) returned empty slice", code)
			}
		})
	}
}

func TestGetFileNotFoundSuggestions_ActionFile(t *testing.T) {
	context := map[string]string{
		"path": "/project/action.yml",
	}

	suggestions := getFileNotFoundSuggestions(context)
	allSuggestions := strings.Join(suggestions, " ")

	// Should suggest common action file names when path contains "action"
	if !strings.Contains(allSuggestions, "action.yml, action.yaml") {
		t.Error("Should suggest common action file names for action file paths")
	}

	if !strings.Contains(allSuggestions, "subdirectory") {
		t.Error("Should suggest checking subdirectories for action files")
	}
}

func TestGetInvalidYAMLSuggestions_TabError(t *testing.T) {
	context := map[string]string{
		"error": "found character that cannot start any token, tab character",
	}

	suggestions := getInvalidYAMLSuggestions(context)
	allSuggestions := strings.Join(suggestions, " ")

	// Should prioritize tab-specific suggestions when error mentions tabs
	if !strings.Contains(allSuggestions, "tabs with spaces") {
		t.Error("Should provide tab-specific suggestions when error mentions tabs")
	}
}

func TestGetGitHubAPISuggestions_StatusCodes(t *testing.T) {
	statusCodes := map[string]string{
		"401": "Authentication failed",
		"403": "Access forbidden",
		"404": "not found",
	}

	for code, expectedText := range statusCodes {
		t.Run("status_"+code, func(t *testing.T) {
			context := map[string]string{"status_code": code}
			suggestions := getGitHubAPISuggestions(context)
			allSuggestions := strings.Join(suggestions, " ")

			if !strings.Contains(allSuggestions, expectedText) {
				t.Errorf("Status code %s suggestions should contain %q", code, expectedText)
			}
		})
	}
}
