package validation

import (
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

// sanitizeTestCase represents a string sanitization test case.
type sanitizeTestCase struct {
	name        string
	input       string
	want        string
	critical    bool
	description string
}

// makeSanitizeTestCase creates a sanitize test case with fewer lines of code.
func makeSanitizeTestCase(name, input, want string, critical bool, desc string) sanitizeTestCase {
	return sanitizeTestCase{
		name:        name,
		input:       input,
		want:        want,
		critical:    critical,
		description: desc,
	}
}

// formatTestCase represents a uses statement formatting test case.
type formatTestCase struct {
	name        string
	org         string
	repo        string
	version     string
	want        string
	critical    bool
	description string
}

// makeFormatTestCase creates a format test case with fewer lines of code.
func makeFormatTestCase(name, org, repo, version, want string, critical bool, desc string) formatTestCase {
	return formatTestCase{
		name:        name,
		org:         org,
		repo:        repo,
		version:     version,
		want:        want,
		critical:    critical,
		description: desc,
	}
}

// TestSanitizeActionNameMutationResistance tests string transformation order and operations.
// Critical mutations to catch:
// - Order of operations (TrimSpace, ReplaceAll, ToLower)
// - ReplaceAll vs Replace (all occurrences vs first)
// - Wrong replacement string.
func TestSanitizeActionNameMutationResistance(t *testing.T) {
	tests := []sanitizeTestCase{
		// Basic transformations
		makeSanitizeTestCase("lowercase_conversion", "UPPERCASE", "uppercase", true, "ToLower applied"),
		makeSanitizeTestCase(
			"space_to_dash",
			testutil.ValidationHelloWorld,
			testutil.MutationStrHelloWorldDash,
			true,
			"ReplaceAll spaces with dashes",
		),
		makeSanitizeTestCase("trim_spaces", "  hello  ", "hello", true, "TrimSpace applied"),

		// Multiple spaces (ReplaceAll vs Replace critical)
		makeSanitizeTestCase(
			"multiple_spaces_all_replaced",
			"hello  world  test",
			"hello--world--test",
			true,
			"All spaces replaced (ReplaceAll, not Replace)",
		),
		makeSanitizeTestCase("three_consecutive_spaces", "a   b", "a---b", true, "Each space replaced individually"),

		// Operation order tests
		makeSanitizeTestCase(
			"uppercase_with_spaces",
			"HELLO WORLD",
			testutil.MutationStrHelloWorldDash,
			true,
			"Both lowercase and space replacement",
		),
		makeSanitizeTestCase(
			"leading_trailing_spaces_uppercase",
			"  HELLO WORLD  ",
			testutil.MutationStrHelloWorldDash,
			true,
			"All transformations: trim, replace, lowercase",
		),

		// Edge cases
		makeSanitizeTestCase(
			"empty_string",
			testutil.MutationStrEmpty,
			testutil.MutationStrEmpty,
			true,
			testutil.MutationDescEmptyInput,
		),
		makeSanitizeTestCase("only_spaces", "   ", testutil.MutationStrEmpty, true, "Only spaces (trimmed to empty)"),
		makeSanitizeTestCase(
			"no_changes_needed",
			"already-sanitized",
			"already-sanitized",
			false,
			"Already in correct format",
		),

		// Special characters
		makeSanitizeTestCase(
			"mixed_case_with_hyphens",
			testutil.MutationStrSetupNode,
			"setup-node",
			false,
			"Existing hyphens preserved",
		),
		makeSanitizeTestCase("underscore_preserved", "hello_world", "hello_world", false, "Underscores not replaced"),
		makeSanitizeTestCase("numbers_preserved", "Action 123", "action-123", false, "Numbers preserved"),

		// Real-world action names
		makeSanitizeTestCase(
			"checkout_action",
			testutil.MutationStrCheckoutCode,
			testutil.MutationStrCheckoutCodeDash,
			false,
			"Realistic action name",
		),
		makeSanitizeTestCase(
			"setup_go_action",
			testutil.MutationStrSetupGoEnvironment,
			testutil.MutationStrSetupGoEnvironmentD,
			false,
			"Multi-word action name",
		),

		// Single character
		makeSanitizeTestCase("single_char", "A", "a", false, "Single character"),
		makeSanitizeTestCase("single_space", " ", testutil.MutationStrEmpty, true, "Single space (trimmed)"),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeActionName(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeActionName(%q) = %q, want %q (description: %s)",
					tt.input, got, tt.want, tt.description)
			}
		})
	}
}

// TestFormatUsesStatementMutationResistance tests uses statement formatting logic.
// Critical mutations to catch:
// - Empty string checks (org == "" changed to !=, etc.)
// - || changed to && in empty check
// - HasPrefix negation (! added/removed)
// - String concatenation order
// - Default version "v1" changed.
func TestFormatUsesStatementMutationResistance(t *testing.T) {
	tests := []formatTestCase{
		// Basic formatting
		makeFormatTestCase(
			"basic_with_version",
			testutil.MutationOrgActions,
			testutil.ValidationCheckout,
			testutil.ValidationCheckoutV3,
			testutil.MutationUsesActionsCheckout,
			false,
			"Standard format",
		),

		// Empty checks (critical)
		makeFormatTestCase(
			"empty_org_returns_empty",
			testutil.MutationStrEmpty,
			testutil.ValidationCheckout,
			testutil.ValidationCheckoutV3,
			testutil.MutationStrEmpty,
			true,
			"org == \"\" check",
		),
		makeFormatTestCase(
			"empty_repo_returns_empty",
			testutil.MutationOrgActions,
			testutil.MutationStrEmpty,
			testutil.ValidationCheckoutV3,
			testutil.MutationStrEmpty,
			true,
			"repo == \"\" check",
		),
		makeFormatTestCase(
			"both_empty_returns_empty",
			testutil.MutationStrEmpty,
			testutil.MutationStrEmpty,
			testutil.ValidationCheckoutV3,
			testutil.MutationStrEmpty,
			true,
			"org == \"\" || repo == \"\" (|| operator critical)",
		),

		// Default version (critical)
		makeFormatTestCase(
			"empty_version_defaults_v1",
			testutil.MutationOrgActions,
			testutil.ValidationCheckout,
			testutil.MutationStrEmpty,
			testutil.MutationUsesActionsCheckoutV1,
			true,
			"version == \"\" defaults to \"v1\"",
		),

		// @ prefix handling (critical)
		makeFormatTestCase(
			"version_without_at",
			testutil.MutationOrgActions,
			testutil.ValidationCheckout,
			testutil.ValidationCheckoutV3,
			testutil.MutationUsesActionsCheckout,
			true,
			"@ added when not present (!HasPrefix check)",
		),
		makeFormatTestCase(
			"version_with_at",
			testutil.MutationOrgActions,
			testutil.ValidationCheckout,
			"@v3",
			testutil.MutationUsesActionsCheckout,
			true,
			"@ not duplicated (HasPrefix check)",
		),
		makeFormatTestCase(
			"double_at_if_hasprefix_fails",
			testutil.MutationOrgActions,
			testutil.ValidationCheckout,
			"@@v3",
			"actions/checkout@@v3",
			false,
			"Malformed input with double @",
		),

		// String concatenation order
		makeFormatTestCase(
			"concatenation_order",
			"org",
			"repo",
			"ver",
			testutil.MutationUsesOrgRepo,
			true,
			"Correct concatenation: org + \"/\" + repo + version",
		),

		// Edge cases
		makeFormatTestCase("single_char_org_repo", "a", "b", "c", "a/b@c", false, "Minimal valid input"),
		makeFormatTestCase(
			"branch_name_version",
			testutil.MutationOrgActions,
			testutil.ValidationCheckout,
			"main",
			"actions/checkout@main",
			false,
			"Branch name as version",
		),
		makeFormatTestCase(
			"sha_version",
			testutil.MutationOrgActions,
			testutil.ValidationCheckout,
			"abc1234567890def",
			"actions/checkout@abc1234567890def",
			false,
			"SHA as version",
		),

		// Whitespace in inputs
		makeFormatTestCase(
			"org_with_spaces_not_trimmed",
			" actions ",
			testutil.ValidationCheckout,
			testutil.ValidationCheckoutV3,
			" actions /checkout@v3",
			false,
			"Spaces preserved (no TrimSpace in function)",
		),

		// Special characters
		makeFormatTestCase(
			"hyphen_in_repo",
			testutil.MutationOrgActions,
			testutil.MutationRepoSetupNode,
			testutil.ValidationCheckoutV3,
			"actions/setup-node@v3",
			false,
			"Hyphen in repo name",
		),
		makeFormatTestCase(
			"at_in_version_position",
			testutil.MutationOrgActions,
			testutil.ValidationCheckout,
			"@v3",
			testutil.MutationUsesActionsCheckout,
			true,
			"Existing @ not duplicated",
		),

		// Boolean operator mutation detection
		makeFormatTestCase(
			"non_empty_org_empty_repo",
			testutil.MutationOrgActions,
			testutil.MutationStrEmpty,
			testutil.ValidationCheckoutV3,
			testutil.MutationStrEmpty,
			true,
			"|| means either empty returns \"\" (not &&)",
		),
		makeFormatTestCase(
			"empty_org_non_empty_repo",
			testutil.MutationStrEmpty,
			testutil.ValidationCheckout,
			testutil.ValidationCheckoutV3,
			testutil.MutationStrEmpty,
			true,
			"|| means either empty returns \"\" (not &&)",
		),

		// Default version with @ handling
		makeFormatTestCase(
			"empty_version_gets_at_prefix",
			testutil.MutationOrgActions,
			testutil.ValidationCheckout,
			testutil.MutationStrEmpty,
			testutil.MutationUsesActionsCheckoutV1,
			true,
			"Empty version: default \"v1\" then @ added",
		),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatUsesStatement(tt.org, tt.repo, tt.version)
			if got != tt.want {
				t.Errorf("FormatUsesStatement(%q, %q, %q) = %q, want %q (description: %s)",
					tt.org, tt.repo, tt.version, got, tt.want, tt.description)
			}
		})
	}
}
