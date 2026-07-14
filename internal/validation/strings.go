package validation

import (
	"regexp"
	"strings"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

// reUnsafeActionNameChar matches any character not allowed in a single
// URL/path-safe action-name segment (keeps a-z, 0-9, '.', '_', '-').
var reUnsafeActionNameChar = regexp.MustCompile(`[^a-z0-9._-]`)

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
