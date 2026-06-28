package validation

import (
	"regexp"
	"strings"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

var (
	// Repository segment is [^/]+? (non-greedy, allows dots so names like
	// "my.repo" are preserved) with an explicit optional ".git" suffix stripped
	// and any trailing path ignored. The optional (?::\d+)? tolerates a host port
	// (e.g. ssh://git@ssh.github.com:443/org/repo) so it is not captured as the org.
	// The leading (?:^|[@/.]) anchors the host so only a real github.com host (or a
	// subdomain like ssh.github.com) matches — without it "notgithub.com/org/repo"
	// would match on the github.com suffix. Mirrors git.reGitHubURL.
	reGitHubURLFull   = regexp.MustCompile(`(?:^|[@/.])github\.com(?::\d+)?[:/]([^/]+)/([^/]+?)(?:\.git)?(?:/.*)?$`)
	reGitHubURLSimple = regexp.MustCompile(`^([^/]+)/([^/]+?)(?:\.git)?$`)
	reWhitespace      = regexp.MustCompile(`\s+`)
	// reUnsafeActionNameChar matches any character not allowed in a single
	// URL/path-safe action-name segment (keeps a-z, 0-9, '.', '_', '-').
	reUnsafeActionNameChar = regexp.MustCompile(`[^a-z0-9._-]`)
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
	// Lowercase, trim surrounding whitespace, then replace every character that is
	// not a letter, digit, '.', '_' or '-' (spaces, '/', '&', tabs, …) with a
	// hyphen, so the result is a single URL/path-safe segment. Without this a name
	// like "My/Sub Action" produced "my/sub-action", injecting an extra path
	// segment into the generated uses: statement and repository URL. Each unsafe
	// char maps to one hyphen (runs are not collapsed) to preserve prior behavior.
	lowered := strings.ToLower(strings.TrimSpace(name))

	return reUnsafeActionNameChar.ReplaceAllString(lowered, "-")
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
