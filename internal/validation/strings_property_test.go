package validation

import (
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestFormatUsesStatementProperties verifies properties of uses statement formatting.
//
//nolint:gocyclo // Property-based test with multiple properties
func TestFormatUsesStatementProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property 1: Result always contains exactly one @ symbol (if non-empty)
	properties.Property("uses statement has exactly one @ symbol when non-empty",
		prop.ForAll(
			func(org string, repo string, version string) bool {
				result := FormatUsesStatement(org, repo, version)

				if result == "" {
					return true // Empty result is acceptable for empty inputs
				}

				count := strings.Count(result, "@")

				return count == 1
			},
			gen.AlphaString(),
			gen.AlphaString(),
			gen.AlphaString(),
		),
	)

	// Property 2: Non-empty org and repo produce non-empty result
	properties.Property("non-empty org and repo produce non-empty result",
		prop.ForAll(
			func(org string, repo string, version string) bool {
				// Only test when org and repo are non-empty
				if org == "" || repo == "" {
					return true
				}

				result := FormatUsesStatement(org, repo, version)

				return result != ""
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString(),
		),
	)

	// Property 3: Result starts with org/repo pattern
	properties.Property("uses statement starts with org/repo when both non-empty",
		prop.ForAll(
			func(org string, repo string, version string) bool {
				if org == "" || repo == "" {
					return true
				}

				result := FormatUsesStatement(org, repo, version)
				expected := org + "/" + repo

				return strings.HasPrefix(result, expected)
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString(),
		),
	)

	// Property 4: Empty org or repo always produces empty result
	properties.Property("empty org or repo produces empty result",
		prop.ForAll(
			func(org string, repo string, version string) bool {
				if org == "" || repo == "" {
					result := FormatUsesStatement(org, repo, version)

					return result == ""
				}

				return true // Skip when both non-empty
			},
			gen.AlphaString(),
			gen.AlphaString(),
			gen.AlphaString(),
		),
	)

	// Property 5: Version part always has @ prefix in result
	properties.Property("version part in result always has @ prefix",
		prop.ForAll(
			func(org string, repo string, version string) bool {
				if org == "" || repo == "" {
					return true
				}

				result := FormatUsesStatement(org, repo, version)

				// Find the @ symbol
				atIndex := strings.Index(result, "@")
				if atIndex == -1 {
					return false // Should always have @
				}

				// Everything after org/repo should start with @
				expectedPrefix := org + "/" + repo + "@"

				return strings.HasPrefix(result, expectedPrefix)
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString(),
		),
	)

	properties.TestingRun(t)
}

// TestStringNormalizationProperties verifies idempotency and whitespace properties.
func TestStringNormalizationProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property 1: Idempotency - normalizing twice produces same result as once
	properties.Property("normalization is idempotent",
		prop.ForAll(
			func(input string) bool {
				n1 := TrimAndNormalize(input)
				n2 := TrimAndNormalize(n1)

				return n1 == n2
			},
			gen.AnyString(),
		),
	)

	// Property 2: No consecutive spaces in output
	properties.Property("normalized string has no consecutive spaces",
		prop.ForAll(
			func(input string) bool {
				result := TrimAndNormalize(input)

				return !strings.Contains(result, "  ")
			},
			gen.AnyString(),
		),
	)

	// Property 3: No leading whitespace
	properties.Property("normalized string has no leading whitespace",
		prop.ForAll(
			func(input string) bool {
				result := TrimAndNormalize(input)

				if result == "" {
					return true
				}

				return !strings.HasPrefix(result, " ") &&
					!strings.HasPrefix(result, "\t") &&
					!strings.HasPrefix(result, "\n")
			},
			gen.AnyString(),
		),
	)

	// Property 4: No trailing whitespace
	properties.Property("normalized string has no trailing whitespace",
		prop.ForAll(
			func(input string) bool {
				result := TrimAndNormalize(input)

				if result == "" {
					return true
				}

				return !strings.HasSuffix(result, " ") &&
					!strings.HasSuffix(result, "\t") &&
					!strings.HasSuffix(result, "\n")
			},
			gen.AnyString(),
		),
	)

	// Property 5: All-whitespace input becomes empty
	properties.Property("whitespace-only input becomes empty",
		prop.ForAll(
			func() bool {
				// Generate whitespace-only strings
				whitespaceOnly := "   \t\n\r   "
				result := TrimAndNormalize(whitespaceOnly)

				return result == ""
			},
		),
	)

	properties.TestingRun(t)
}

// TestVersionCleaningProperties verifies version string cleaning properties.
// versionCleaningIdempotentProperty verifies cleaning twice produces same result.
func versionCleaningIdempotentProperty(version string) bool {
	v1 := CleanVersionString(version)
	v2 := CleanVersionString(v1)

	return v1 == v2
}

// versionRemovesSingleVProperty verifies single 'v' is removed.
func versionRemovesSingleVProperty(version string) bool {
	result := CleanVersionString(version)
	if result == "" {
		return true
	}

	trimmed := strings.TrimSpace(version)
	if strings.HasPrefix(trimmed, "v") && !strings.HasPrefix(trimmed, "vv") {
		return !strings.HasPrefix(result, "v")
	}

	return true
}

// versionHasNoBoundaryWhitespaceProperty verifies no leading/trailing whitespace.
func versionHasNoBoundaryWhitespaceProperty(version string) bool {
	result := CleanVersionString(version)
	if result == "" {
		return true
	}

	return !strings.HasPrefix(result, " ") &&
		!strings.HasSuffix(result, " ") &&
		!strings.HasPrefix(result, "\t") &&
		!strings.HasSuffix(result, "\t")
}

// whitespaceOnlyVersionBecomesEmptyProperty verifies whitespace-only inputs become empty.
func whitespaceOnlyVersionBecomesEmptyProperty() bool {
	whitespaceInputs := []string{"   ", "\t\t", "\n", " \t\n "}
	for _, input := range whitespaceInputs {
		result := CleanVersionString(input)
		if result != "" {
			return false
		}
	}

	return true
}

// nonVContentPreservedProperty verifies non-v content is preserved and trimmed.
func nonVContentPreservedProperty(content string) bool {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" || strings.HasPrefix(trimmed, "v") {
		return true // Skip these cases
	}

	result := CleanVersionString(content)

	return result == trimmed
}

func TestVersionCleaningProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property 1: Idempotency - cleaning twice produces same result
	properties.Property("version cleaning is idempotent",
		prop.ForAll(versionCleaningIdempotentProperty, gen.AnyString()),
	)

	// Property 2: Result never starts with single 'v' (TrimPrefix removes only one)
	properties.Property("cleaned version removes single leading v",
		prop.ForAll(versionRemovesSingleVProperty, gen.AnyString()),
	)

	// Property 3: No leading/trailing whitespace in result
	properties.Property("cleaned version has no boundary whitespace",
		prop.ForAll(versionHasNoBoundaryWhitespaceProperty, gen.AnyString()),
	)

	// Property 4: Whitespace-only input becomes empty
	properties.Property("whitespace-only version becomes empty",
		prop.ForAll(whitespaceOnlyVersionBecomesEmptyProperty),
	)

	// Property 5: Preserves non-v content and trims whitespace
	properties.Property("non-v content is preserved",
		prop.ForAll(
			nonVContentPreservedProperty,
			gen.OneGenOf(
				gen.AlphaString(),
				gen.AlphaString().Map(func(s string) string { return "  " + s }),
				gen.AlphaString().Map(func(s string) string { return s + "  " }),
				gen.AlphaString().Map(func(s string) string { return "  " + s + "  " }),
				gen.AlphaString().Map(func(s string) string { return "\t" + s + "\n" }),
			),
		),
	)

	properties.TestingRun(t)
}

// TestSanitizeActionNameProperties verifies action name sanitization properties.
func TestSanitizeActionNameProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property 1: Result is always lowercase
	properties.Property("sanitized name is always lowercase",
		prop.ForAll(
			func(name string) bool {
				result := SanitizeActionName(name)

				return result == strings.ToLower(result)
			},
			gen.AnyString(),
		),
	)

	// Property 2: No spaces in result
	properties.Property("sanitized name has no spaces",
		prop.ForAll(
			func(name string) bool {
				result := SanitizeActionName(name)

				return !strings.Contains(result, " ")
			},
			gen.AnyString(),
		),
	)

	// Property 3: Idempotency
	properties.Property("sanitization is idempotent",
		prop.ForAll(
			func(name string) bool {
				s1 := SanitizeActionName(name)
				s2 := SanitizeActionName(s1)

				return s1 == s2
			},
			gen.AnyString(),
		),
	)

	// Property 4: Whitespace-only input becomes empty
	properties.Property("whitespace-only input becomes empty",
		prop.ForAll(
			func() bool {
				whitespaceInputs := []string{"   ", "\t\t", "  \n  "}
				for _, input := range whitespaceInputs {
					result := SanitizeActionName(input)
					if result != "" {
						return false
					}
				}

				return true
			},
		),
	)

	// Property 5: Spaces become hyphens
	properties.Property("spaces are converted to hyphens",
		prop.ForAll(
			func(word1 string, word2 string) bool {
				// Only test when words are non-empty and don't contain spaces
				if word1 == "" || word2 == "" ||
					strings.Contains(word1, " ") ||
					strings.Contains(word2, " ") {
					return true
				}

				input := word1 + " " + word2
				result := SanitizeActionName(input)

				// Result should contain a hyphen where the space was
				expectedPart1 := strings.ToLower(word1)
				expectedPart2 := strings.ToLower(word2)
				expected := expectedPart1 + "-" + expectedPart2

				return result == expected
			},
			gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && !strings.Contains(s, " ") }),
			gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && !strings.Contains(s, " ") }),
		),
	)

	properties.TestingRun(t)
}

// TestParseGitHubURLProperties verifies URL parsing properties.
//
//nolint:gocyclo // Property-based test with multiple properties
func TestParseGitHubURLProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property 1: Empty inputs produce empty outputs
	properties.Property("empty URL produces empty org and repo",
		prop.ForAll(
			func() bool {
				org, repo := ParseGitHubURL("")

				return org == "" && repo == ""
			},
		),
	)

	// Property 2: Valid org/repo format always parses
	properties.Property("simple org/repo format always parses correctly",
		prop.ForAll(
			func(org string, repo string) bool {
				if org == "" || repo == "" || strings.Contains(org, "/") || strings.Contains(repo, "/") {
					return true // Skip invalid inputs
				}

				url := org + "/" + repo
				gotOrg, gotRepo := ParseGitHubURL(url)

				return gotOrg == org && gotRepo == repo
			},
			gen.AlphaString().SuchThat(func(s string) bool {
				return len(s) > 0 && !strings.Contains(s, "/") && !strings.Contains(s, ".")
			}),
			gen.AlphaString().SuchThat(func(s string) bool {
				return len(s) > 0 && !strings.Contains(s, "/")
			}),
		),
	)

	// Property 3: Result org/repo never contain slashes
	properties.Property("parsed org and repo never contain slashes",
		prop.ForAll(
			func(url string) bool {
				org, repo := ParseGitHubURL(url)

				return !strings.Contains(org, "/") && !strings.Contains(repo, "/")
			},
			gen.AnyString(),
		),
	)

	// Property 4: Invalid input produces empty result
	properties.Property("URLs without slash produce empty result",
		prop.ForAll(
			func(url string) bool {
				if strings.Contains(url, "/") {
					return true // Skip URLs with slashes
				}

				org, repo := ParseGitHubURL(url)
				// URLs without slashes should not parse (unless they match github.com pattern)
				if strings.Contains(url, "github.com") {
					return true // github.com URLs might parse differently
				}

				return org == "" && repo == ""
			},
			gen.AlphaString(),
		),
	)

	// Property 5: Both org and repo empty, or both non-empty
	properties.Property("org and repo are both empty or both non-empty",
		prop.ForAll(
			func(url string) bool {
				org, repo := ParseGitHubURL(url)
				// Either both empty, or both non-empty
				return (org == "" && repo == "") || (org != "" && repo != "")
			},
			gen.AnyString(),
		),
	)

	properties.TestingRun(t)
}
