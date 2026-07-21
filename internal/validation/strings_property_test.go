package validation

import (
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestFormatUsesStatementProperties verifies properties of uses statement formatting.
func TestFormatUsesStatementProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)
	registerUsesStatementProperties(properties)
	properties.TestingRun(t)
}

// registerUsesStatementProperties registers all uses statement property tests.
func registerUsesStatementProperties(properties *gopter.Properties) {
	registerUsesStatementAtSymbolProperty(properties)
	registerUsesStatementNonEmptyProperty(properties)
	registerUsesStatementPrefixProperty(properties)
	registerUsesStatementEmptyInputProperty(properties)
	registerUsesStatementVersionPrefixProperty(properties)
}

// registerUsesStatementAtSymbolProperty tests that result contains exactly one @ symbol.
func registerUsesStatementAtSymbolProperty(properties *gopter.Properties) {
	properties.Property("uses statement has exactly one @ symbol when non-empty",
		prop.ForAll(
			func(org, repo, version string) bool {
				result := FormatUsesStatement(org, repo, version)
				if result == "" {
					return true
				}

				return strings.Count(result, "@") == 1
			},
			gen.AlphaString(),
			gen.AlphaString(),
			gen.AlphaString(),
		),
	)
}

// registerUsesStatementNonEmptyProperty tests non-empty inputs produce non-empty result.
func registerUsesStatementNonEmptyProperty(properties *gopter.Properties) {
	properties.Property("non-empty org and repo produce non-empty result",
		prop.ForAll(
			func(org, repo, version string) bool {
				if org == "" || repo == "" {
					return true
				}

				return FormatUsesStatement(org, repo, version) != ""
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString(),
		),
	)
}

// registerUsesStatementPrefixProperty tests result starts with org/repo pattern.
func registerUsesStatementPrefixProperty(properties *gopter.Properties) {
	properties.Property("uses statement starts with org/repo when both non-empty",
		prop.ForAll(
			func(org, repo, version string) bool {
				if org == "" || repo == "" {
					return true
				}
				result := FormatUsesStatement(org, repo, version)

				return strings.HasPrefix(result, org+"/"+repo)
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString(),
		),
	)
}

// registerUsesStatementEmptyInputProperty tests empty inputs produce empty result.
func registerUsesStatementEmptyInputProperty(properties *gopter.Properties) {
	properties.Property("empty org or repo produces empty result",
		prop.ForAll(
			func(org, repo, version string) bool {
				if org == "" || repo == "" {
					return FormatUsesStatement(org, repo, version) == ""
				}

				return true
			},
			gen.AlphaString(),
			gen.AlphaString(),
			gen.AlphaString(),
		),
	)
}

// registerUsesStatementVersionPrefixProperty tests version part has @ prefix.
func registerUsesStatementVersionPrefixProperty(properties *gopter.Properties) {
	properties.Property("version part in result always has @ prefix",
		prop.ForAll(
			func(org, repo, version string) bool {
				if org == "" || repo == "" {
					return true
				}
				result := FormatUsesStatement(org, repo, version)
				atIndex := strings.Index(result, "@")
				if atIndex == -1 {
					return false
				}

				return strings.HasPrefix(result, org+"/"+repo+"@")
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString(),
		),
	)
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
