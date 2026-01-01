package internal

import (
	"os"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

// loadGitHubTokenFromEnv retrieves the GitHub token from environment variables.
// It checks both the tool-specific environment variable (GHREADME_GITHUB_TOKEN)
// and the standard GitHub environment variable (GITHUB_TOKEN) in that order.
// Returns an empty string if no token is found.
func loadGitHubTokenFromEnv() string {
	// Priority 1: Tool-specific env var
	if token := os.Getenv(appconstants.EnvGitHubToken); token != "" {
		return token
	}

	// Priority 2: Standard GitHub env var
	if token := os.Getenv(appconstants.EnvGitHubTokenStandard); token != "" {
		return token
	}

	return ""
}
