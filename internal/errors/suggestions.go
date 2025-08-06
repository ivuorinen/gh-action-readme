package errors

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// GetSuggestions returns context-aware suggestions for the given error code.
func GetSuggestions(code ErrorCode, context map[string]string) []string {
	if handler := getSuggestionHandler(code); handler != nil {
		return handler(context)
	}

	return getDefaultSuggestions()
}

// getSuggestionHandler returns the appropriate suggestion function for the error code.
func getSuggestionHandler(code ErrorCode) func(map[string]string) []string {
	handlers := map[ErrorCode]func(map[string]string) []string{
		ErrCodeFileNotFound:       getFileNotFoundSuggestions,
		ErrCodePermission:         getPermissionSuggestions,
		ErrCodeInvalidYAML:        getInvalidYAMLSuggestions,
		ErrCodeInvalidAction:      getInvalidActionSuggestions,
		ErrCodeNoActionFiles:      getNoActionFilesSuggestions,
		ErrCodeGitHubAPI:          getGitHubAPISuggestions,
		ErrCodeConfiguration:      getConfigurationSuggestions,
		ErrCodeValidation:         getValidationSuggestions,
		ErrCodeTemplateRender:     getTemplateSuggestions,
		ErrCodeFileWrite:          getFileWriteSuggestions,
		ErrCodeDependencyAnalysis: getDependencyAnalysisSuggestions,
		ErrCodeCacheAccess:        getCacheAccessSuggestions,
	}

	// Special cases for handlers without context
	switch code {
	case ErrCodeGitHubRateLimit:
		return func(_ map[string]string) []string { return getGitHubRateLimitSuggestions() }
	case ErrCodeGitHubAuth:
		return func(_ map[string]string) []string { return getGitHubAuthSuggestions() }
	case ErrCodeFileNotFound, ErrCodePermission, ErrCodeInvalidYAML, ErrCodeInvalidAction,
		ErrCodeNoActionFiles, ErrCodeGitHubAPI, ErrCodeConfiguration, ErrCodeValidation,
		ErrCodeTemplateRender, ErrCodeFileWrite, ErrCodeDependencyAnalysis, ErrCodeCacheAccess,
		ErrCodeUnknown:
		// These cases are handled by the map above
	}

	return handlers[code]
}

// getDefaultSuggestions returns generic suggestions for unknown errors.
func getDefaultSuggestions() []string {
	return []string{
		"Check the error message for more details",
		"Run with --verbose flag for additional debugging information",
		"Visit the project documentation for help",
	}
}

func getFileNotFoundSuggestions(context map[string]string) []string {
	suggestions := []string{}

	if path, ok := context["path"]; ok {
		suggestions = append(suggestions,
			"Check if the file exists: "+path,
			"Verify the file path is correct",
		)

		// Check if it might be a case sensitivity issue
		dir := filepath.Dir(path)
		if _, err := os.Stat(dir); err == nil {
			suggestions = append(suggestions,
				"Check for case sensitivity in the filename",
				"Try: ls -la "+dir,
			)
		}

		// Suggest common file names if looking for action files
		if strings.Contains(path, "action") {
			suggestions = append(suggestions,
				"Common action file names: action.yml, action.yaml",
				"Check if the file is in a subdirectory",
			)
		}
	}

	suggestions = append(suggestions,
		"Use --recursive flag to search in subdirectories",
		"Ensure you have read permissions for the directory",
	)

	return suggestions
}

func getPermissionSuggestions(context map[string]string) []string {
	suggestions := []string{}

	if path, ok := context["path"]; ok {
		suggestions = append(suggestions,
			"Check file permissions: ls -la "+path,
			"Try changing permissions: chmod 644 "+path,
		)

		// Check if it's a directory
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			suggestions = append(suggestions,
				"For directories, try: chmod 755 "+path,
			)
		}
	}

	// Add OS-specific suggestions
	switch runtime.GOOS {
	case "windows":
		suggestions = append(suggestions,
			"Run the command prompt as Administrator",
			"Check Windows file permissions in Properties > Security",
		)
	default:
		suggestions = append(suggestions,
			"Check file ownership with: ls -la",
			"You may need to use sudo for system directories",
			"Ensure the parent directory has write permissions",
		)
	}

	return suggestions
}

func getInvalidYAMLSuggestions(context map[string]string) []string {
	suggestions := []string{
		"Check YAML indentation (use spaces, not tabs)",
		"Ensure all strings with special characters are quoted",
		"Verify brackets and braces are properly closed",
		"Use a YAML validator: https://yaml-online-parser.appspot.com/",
	}

	if line, ok := context["line"]; ok {
		suggestions = append([]string{
			fmt.Sprintf("Error near line %s - check indentation and syntax", line),
		}, suggestions...)
	}

	if err, ok := context["error"]; ok {
		if strings.Contains(err, "tab") {
			suggestions = append([]string{
				"YAML files must use spaces for indentation, not tabs",
				"Replace all tabs with spaces (usually 2 or 4 spaces)",
			}, suggestions...)
		}
	}

	suggestions = append(suggestions,
		"Common YAML issues: missing colons, incorrect nesting, invalid characters",
		"Example valid action.yml structure available in documentation",
	)

	return suggestions
}

func getInvalidActionSuggestions(context map[string]string) []string {
	suggestions := []string{
		"Ensure the action file has required fields: name, description",
		"Check that 'runs' section is properly configured",
		"Verify the action type is valid (composite, docker, or javascript)",
	}

	if missingFields, ok := context["missing_fields"]; ok {
		suggestions = append([]string{
			"Missing required fields: " + missingFields,
		}, suggestions...)
	}

	if invalidField, ok := context["invalid_field"]; ok {
		suggestions = append(suggestions,
			fmt.Sprintf("Invalid field '%s' - check spelling and placement", invalidField),
		)
	}

	suggestions = append(suggestions,
		"Refer to GitHub Actions documentation for action.yml schema",
		"Use 'gh-action-readme schema' to see the expected format",
		"Validate against the schema with 'gh-action-readme validate'",
	)

	return suggestions
}

func getNoActionFilesSuggestions(context map[string]string) []string {
	suggestions := []string{
		"Ensure you're in the correct directory",
		"Look for files named 'action.yml' or 'action.yaml'",
		"Use --recursive flag to search subdirectories",
	}

	if dir, ok := context["directory"]; ok {
		suggestions = append(suggestions,
			"Current directory: "+dir,
			fmt.Sprintf("Try: find %s -name 'action.y*ml' -type f", dir),
		)
	}

	suggestions = append(suggestions,
		"GitHub Actions must have an action.yml or action.yaml file",
		"Check if the file has a different extension (.yaml vs .yml)",
		"Example: gh-action-readme gen --recursive",
	)

	return suggestions
}

func getGitHubAPISuggestions(context map[string]string) []string {
	suggestions := []string{
		"Check your internet connection",
		"Verify GitHub's API status: https://www.githubstatus.com/",
		"Ensure your GitHub token has the necessary permissions",
	}

	if statusCode, ok := context["status_code"]; ok {
		switch statusCode {
		case "401":
			suggestions = append([]string{
				"Authentication failed - check your GitHub token",
				"Token may be expired or revoked",
			}, suggestions...)
		case "403":
			suggestions = append([]string{
				"Access forbidden - check token permissions",
				"You may have hit the rate limit",
			}, suggestions...)
		case "404":
			suggestions = append([]string{
				"Repository or resource not found",
				"Check if the repository is private and token has access",
			}, suggestions...)
		}
	}

	return suggestions
}

func getGitHubRateLimitSuggestions() []string {
	return []string{
		"GitHub API rate limit exceeded",
		"Authenticate with a GitHub token to increase limits",
		"Set GITHUB_TOKEN environment variable: export GITHUB_TOKEN=your_token",
		"For GitHub CLI: gh auth login",
		"Rate limits reset every hour",
		"Consider using caching to reduce API calls",
		"Use --quiet mode to reduce API usage for non-critical features",
	}
}

func getGitHubAuthSuggestions() []string {
	return []string{
		"Set GitHub token: export GITHUB_TOKEN=your_personal_access_token",
		"Or use GitHub CLI: gh auth login",
		"Create a token at: https://github.com/settings/tokens",
		"Token needs 'repo' scope for private repositories",
		"For public repos only, 'public_repo' scope is sufficient",
		"Check if token is set: echo $GITHUB_TOKEN",
		"Ensure token hasn't expired",
	}
}

func getConfigurationSuggestions(context map[string]string) []string {
	suggestions := []string{
		"Check configuration file syntax",
		"Ensure configuration file exists",
		"Use 'gh-action-readme config init' to create default config",
		"Valid config locations: .gh-action-readme.yml, ~/.config/gh-action-readme/config.yaml",
	}

	if configPath, ok := context["config_path"]; ok {
		suggestions = append(suggestions,
			"Config path: "+configPath,
			"Check if file exists: ls -la "+configPath,
		)
	}

	if err, ok := context["error"]; ok {
		if strings.Contains(err, "permission") {
			suggestions = append(suggestions,
				"Check file permissions for config file",
				"Ensure parent directory is writable",
			)
		}
	}

	return suggestions
}

func getValidationSuggestions(context map[string]string) []string {
	suggestions := []string{
		"Review validation errors for specific issues",
		"Check required fields are present",
		"Ensure field values match expected types",
		"Use 'gh-action-readme schema' to see valid structure",
	}

	if fields, ok := context["invalid_fields"]; ok {
		suggestions = append([]string{
			"Invalid fields: " + fields,
			"Check spelling and nesting of these fields",
		}, suggestions...)
	}

	if validationType, ok := context["validation_type"]; ok {
		switch validationType {
		case "required":
			suggestions = append(suggestions,
				"Add missing required fields to your action.yml",
				"Required fields typically include: name, description, runs",
			)
		case "type":
			suggestions = append(suggestions,
				"Ensure field values match expected types",
				"Strings should be quoted, booleans should be true/false",
			)
		}
	}

	return suggestions
}

func getTemplateSuggestions(context map[string]string) []string {
	suggestions := []string{
		"Check template syntax",
		"Ensure all template variables are defined",
		"Verify custom template path is correct",
	}

	if templatePath, ok := context["template_path"]; ok {
		suggestions = append(suggestions,
			"Template path: "+templatePath,
			"Ensure template file exists and is readable",
		)
	}

	if theme, ok := context["theme"]; ok {
		suggestions = append(suggestions,
			"Current theme: "+theme,
			"Try using a different theme: --theme github",
			"Available themes: default, github, gitlab, minimal, professional",
		)
	}

	return suggestions
}

func getFileWriteSuggestions(context map[string]string) []string {
	suggestions := []string{
		"Check if the output directory exists",
		"Ensure you have write permissions",
		"Verify disk space is available",
	}

	if outputPath, ok := context["output_path"]; ok {
		dir := filepath.Dir(outputPath)
		suggestions = append(suggestions,
			"Output directory: "+dir,
			"Check permissions: ls -la "+dir,
			"Create directory if needed: mkdir -p "+dir,
		)

		// Check if file already exists
		if _, err := os.Stat(outputPath); err == nil {
			suggestions = append(suggestions,
				"File already exists - it will be overwritten",
				"Back up existing file if needed",
			)
		}
	}

	return suggestions
}

func getDependencyAnalysisSuggestions(context map[string]string) []string {
	suggestions := []string{
		"Ensure GitHub token is set for dependency analysis",
		"Check that the action file contains valid dependencies",
		"Verify network connectivity to GitHub",
	}

	if action, ok := context["action"]; ok {
		suggestions = append(suggestions,
			"Analyzing action: "+action,
			"Only composite actions have analyzable dependencies",
		)
	}

	return append(suggestions,
		"Dependency analysis requires 'uses' statements in composite actions",
		"Example: uses: actions/checkout@v4",
	)
}

func getCacheAccessSuggestions(context map[string]string) []string {
	suggestions := []string{
		"Check cache directory permissions",
		"Ensure cache directory exists",
		"Try clearing cache: gh-action-readme cache clear",
		"Default cache location: ~/.cache/gh-action-readme",
	}

	if cachePath, ok := context["cache_path"]; ok {
		suggestions = append(suggestions,
			"Cache path: "+cachePath,
			"Check permissions: ls -la "+cachePath,
			"You can disable cache temporarily with environment variables",
		)
	}

	return suggestions
}
