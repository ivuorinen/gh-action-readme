package validation

import (
	"regexp"
	"strings"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

var (
	reGitHubURLFull   = regexp.MustCompile(`github\.com[:/]([^/]+)/([^/.]+)(?:\.git)?`)
	reGitHubURLSimple = regexp.MustCompile(`^([^/]+)/([^/.]+)$`)
	reWhitespace      = regexp.MustCompile(`\s+`)
)

// CleanVersionString removes common prefixes and normalizes version strings.
func CleanVersionString(version string) string {
	cleaned := strings.TrimSpace(version)

	return strings.TrimPrefix(cleaned, "v")
}

// ParseGitHubURL extracts organization and repository from a GitHub URL.
func ParseGitHubURL(url string) (organization, repository string) {
	for _, re := range []*regexp.Regexp{reGitHubURLFull, reGitHubURLSimple} {
		matches := re.FindStringSubmatch(url)
		if len(matches) >= 3 {
			return matches[1], matches[2]
		}
	}

	return "", ""
}

// SanitizeActionName converts action name to a URL-friendly format.
func SanitizeActionName(name string) string {
	// Convert to lowercase and replace spaces with hyphens
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(name), " ", "-"))
}

// TrimAndNormalize removes extra whitespace and normalizes strings.
func TrimAndNormalize(input string) string {
	return reWhitespace.ReplaceAllString(strings.TrimSpace(input), " ")
}

// FormatUsesStatement creates a properly formatted GitHub Action uses statement.
func FormatUsesStatement(org, repo, version string) string {
	if org == "" || repo == "" {
		return ""
	}

	if version == "" {
		version = appconstants.VersionTagV1
	}

	// Ensure version starts with @
	if !strings.HasPrefix(version, "@") {
		version = "@" + version
	}

	return org + "/" + repo + version
}
