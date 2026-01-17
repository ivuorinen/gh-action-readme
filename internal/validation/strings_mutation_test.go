package validation

import (
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

// urlTestCase represents a single URL parsing test case.
type urlTestCase struct {
	name        string
	url         string
	wantOrg     string
	wantRepo    string
	critical    bool
	description string
}

// makeURLTestCase creates a URL test case with fewer lines of code.
func makeURLTestCase(name, url, org, repo string, critical bool, desc string) urlTestCase {
	return urlTestCase{
		name:        name,
		url:         url,
		wantOrg:     org,
		wantRepo:    repo,
		critical:    critical,
		description: desc,
	}
}

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

// TestParseGitHubURLMutationResistance tests URL parsing for regex and boundary mutations.
// Critical mutations to catch:
// - Pattern order changes (SSH vs HTTPS precedence)
// - len(matches) >= 3 changed to > 3, == 3, etc.
// - Return statement modifications (returning wrong indices).
func TestParseGitHubURLMutationResistance(t *testing.T) {
	tests := []urlTestCase{
		// HTTPS URLs
		makeURLTestCase(
			"https_standard",
			testutil.MutationURLHTTPS,
			testutil.MutationOrgOctocat,
			testutil.MutationRepoHelloWorld,
			false,
			"Standard HTTPS URL",
		),
		makeURLTestCase(
			"https_with_git_extension",
			testutil.MutationURLHTTPSGit,
			testutil.MutationOrgOctocat,
			testutil.MutationRepoHelloWorld,
			true,
			".git extension handled by (?:\\.git)? regex",
		),

		// SSH URLs
		makeURLTestCase(
			"ssh_standard",
			testutil.MutationURLSSH,
			testutil.MutationOrgOctocat,
			testutil.MutationRepoHelloWorld,
			true,
			"SSH URL with colon separator ([:/] pattern)",
		),
		makeURLTestCase(
			"ssh_with_git_extension",
			testutil.MutationURLSSHGit,
			testutil.MutationOrgOctocat,
			testutil.MutationRepoHelloWorld,
			true,
			"SSH with .git",
		),

		// Simple format
		makeURLTestCase(
			"simple_org_repo",
			testutil.MutationURLSimple,
			testutil.MutationOrgOctocat,
			testutil.MutationRepoHelloWorld,
			true,
			"Simple org/repo format (second pattern)",
		),

		// Edge cases with special characters
		makeURLTestCase(
			"org_with_dash",
			testutil.MutationURLSetupNode,
			testutil.MutationOrgActions,
			testutil.MutationRepoSetupNode,
			false,
			"Hyphen in repo name",
		),
		makeURLTestCase("org_with_number", "org123/repo456", "org123", "repo456", false, "Numbers in org/repo"),

		// Boundary: len(matches) >= 3
		makeURLTestCase(
			"exactly_3_matches",
			"a/b",
			"a",
			"b",
			true,
			"Minimal valid: exactly 3 matches (full, org, repo)",
		),

		// Invalid URLs (should return empty)
		makeURLTestCase(
			"no_slash_invalid",
			"octocatHello-World",
			testutil.MutationStrEmpty,
			testutil.MutationStrEmpty,
			true,
			"No slash separator",
		),
		makeURLTestCase(
			"empty_string",
			testutil.MutationStrEmpty,
			testutil.MutationStrEmpty,
			testutil.MutationStrEmpty,
			true,
			"Empty string",
		),
		makeURLTestCase(
			"only_org",
			"octocat/",
			testutil.MutationStrEmpty,
			testutil.MutationStrEmpty,
			true,
			"Trailing slash, no repo",
		),
		makeURLTestCase(
			"only_repo",
			"/Hello-World",
			testutil.MutationStrEmpty,
			testutil.MutationStrEmpty,
			true,
			"Leading slash, no org",
		),

		// Pattern precedence tests
		makeURLTestCase(
			"github_com_in_middle",
			testutil.MutationURLGitHubReadme,
			testutil.MutationOrgIvuorinen,
			testutil.MutationRepoGhActionReadme,
			false,
			"First pattern should match",
		),

		// Regex capture group tests
		makeURLTestCase(
			"multiple_slashes",
			"octocat/Hello-World/extra",
			testutil.MutationStrEmpty,
			testutil.MutationStrEmpty,
			false,
			"Extra path segments invalid for simple format",
		),

		// .git extension edge cases
		makeURLTestCase(
			"double_git_extension",
			"octocat/Hello-World.git.git",
			testutil.MutationStrEmpty,
			testutil.MutationStrEmpty,
			true,
			"Dots not allowed in repo name by [^/.] pattern",
		),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOrg, gotRepo := ParseGitHubURL(tt.url)

			if gotOrg != tt.wantOrg {
				t.Errorf("ParseGitHubURL(%q) org = %q, want %q (description: %s)",
					tt.url, gotOrg, tt.wantOrg, tt.description)
			}
			if gotRepo != tt.wantRepo {
				t.Errorf("ParseGitHubURL(%q) repo = %q, want %q (description: %s)",
					tt.url, gotRepo, tt.wantRepo, tt.description)
			}
		})
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
			"hello-world",
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
			"hello-world",
			true,
			"Both lowercase and space replacement",
		),
		makeSanitizeTestCase(
			"leading_trailing_spaces_uppercase",
			"  HELLO WORLD  ",
			"hello-world",
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

// TestTrimAndNormalizeMutationResistance tests whitespace normalization.
// Critical mutations to catch:
// - Regex quantifier changes (\s+ to \s*, \s, etc.)
// - TrimSpace removal or reordering
// - ReplaceAllString to different methods.
func TestTrimAndNormalizeMutationResistance(t *testing.T) {
	tests := []sanitizeTestCase{
		// Leading/trailing whitespace
		makeSanitizeTestCase("leading_whitespace", "  hello", "hello", true, "TrimSpace removes leading"),
		makeSanitizeTestCase("trailing_whitespace", "hello  ", "hello", true, "TrimSpace removes trailing"),
		makeSanitizeTestCase("both_sides_whitespace", "  hello  ", "hello", true, "TrimSpace removes both sides"),

		// Internal whitespace normalization
		makeSanitizeTestCase(
			"double_space",
			testutil.ValidationHelloWorld,
			testutil.ValidationHelloWorld,
			true,
			"Double space to single (\\s+ pattern)",
		),
		makeSanitizeTestCase(
			"triple_space",
			"hello   world",
			testutil.ValidationHelloWorld,
			true,
			"Triple space to single",
		),
		makeSanitizeTestCase(
			"many_spaces",
			"hello          world",
			testutil.ValidationHelloWorld,
			true,
			"Many spaces to single (+ quantifier)",
		),

		// Mixed whitespace types
		makeSanitizeTestCase(
			"tab_character",
			"hello\tworld",
			testutil.ValidationHelloWorld,
			true,
			"Tab normalized to space (\\s includes tabs)",
		),
		makeSanitizeTestCase(
			"newline_character",
			"hello\nworld",
			testutil.ValidationHelloWorld,
			true,
			"Newline normalized to space (\\s includes newlines)",
		),
		makeSanitizeTestCase(
			"carriage_return",
			"hello\rworld",
			testutil.ValidationHelloWorld,
			true,
			"CR normalized to space",
		),
		makeSanitizeTestCase(
			"mixed_whitespace",
			"hello \t\n world",
			testutil.ValidationHelloWorld,
			true,
			"Mixed whitespace types to single space",
		),

		// Combined leading/trailing and internal
		makeSanitizeTestCase(
			"all_whitespace_issues",
			"  hello   world  ",
			testutil.ValidationHelloWorld,
			true,
			"Trim + normalize internal",
		),

		// Edge cases
		makeSanitizeTestCase(
			"empty_string",
			testutil.MutationStrEmpty,
			testutil.MutationStrEmpty,
			true,
			testutil.MutationDescEmptyInput,
		),
		makeSanitizeTestCase("only_spaces", "     ", testutil.MutationStrEmpty, true, "Only spaces (trimmed to empty)"),
		makeSanitizeTestCase(
			"only_whitespace_mixed",
			" \t\n\r ",
			testutil.MutationStrEmpty,
			true,
			"Only various whitespace types",
		),
		makeSanitizeTestCase("no_whitespace", "hello", "hello", false, "No whitespace to normalize"),
		makeSanitizeTestCase(
			"single_space_valid",
			testutil.ValidationHelloWorld,
			testutil.ValidationHelloWorld,
			false,
			"Already normalized",
		),

		// Multiple words
		makeSanitizeTestCase(
			"three_words_excess_spaces",
			"one   two   three",
			"one two three",
			false,
			"Three words with excess spaces",
		),

		// Unicode whitespace
		makeSanitizeTestCase(
			"regular_space",
			testutil.ValidationHelloWorld,
			testutil.ValidationHelloWorld,
			false,
			"Regular ASCII space",
		),

		// Quantifier verification (\s+ means one or more)
		makeSanitizeTestCase("single_space_between", "a b", "a b", true, "Single space not collapsed (need + for >1)"),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TrimAndNormalize(tt.input)
			if got != tt.want {
				t.Errorf("TrimAndNormalize(%q) = %q, want %q (description: %s)",
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

// TestCleanVersionStringMutationResistance tests version cleaning for operation order.
// Critical mutations to catch:
// - TrimSpace removal
// - TrimPrefix removal or wrong prefix
// - Operation order (trim then prefix vs prefix then trim).
func TestCleanVersionStringMutationResistance(t *testing.T) {
	tests := []sanitizeTestCase{
		// v prefix removal
		makeSanitizeTestCase("v_prefix_removed", "v1.2.3", "1.2.3", true, "TrimPrefix(\"v\") applied"),
		makeSanitizeTestCase("no_v_prefix_unchanged", "1.2.3", "1.2.3", true, "No v prefix to remove"),

		// Whitespace handling
		makeSanitizeTestCase("leading_whitespace", "  v1.2.3", "1.2.3", true, "TrimSpace before TrimPrefix"),
		makeSanitizeTestCase("trailing_whitespace", "v1.2.3  ", "1.2.3", true, "TrimSpace applied"),
		makeSanitizeTestCase("both_whitespace_and_v", "  v1.2.3  ", "1.2.3", true, "Both TrimSpace and TrimPrefix"),

		// Operation order critical
		makeSanitizeTestCase(
			"whitespace_before_v",
			" v1.2.3",
			"1.2.3",
			true,
			"TrimSpace must happen before TrimPrefix",
		),

		// Edge cases
		makeSanitizeTestCase("only_v", "v", testutil.MutationStrEmpty, true, "Just v becomes empty"),
		makeSanitizeTestCase(
			"empty_string",
			testutil.MutationStrEmpty,
			testutil.MutationStrEmpty,
			true,
			testutil.MutationDescEmptyInput,
		),
		makeSanitizeTestCase("only_whitespace", "   ", testutil.MutationStrEmpty, true, "Only spaces"),

		// Multiple v's
		makeSanitizeTestCase(
			"double_v",
			"vv1.2.3",
			"v1.2.3",
			true,
			"Only first v removed (TrimPrefix, not ReplaceAll)",
		),

		// No changes needed
		makeSanitizeTestCase("already_clean", "1.2.3", "1.2.3", false, "Already clean"),

		// Real-world versions
		makeSanitizeTestCase("semver_with_v", testutil.MutationVersionV2, "2.5.1", false, "Realistic semver"),
		makeSanitizeTestCase("semver_no_v", "2.5.1", "2.5.1", false, "Realistic semver without v"),

		// Whitespace variations
		makeSanitizeTestCase("tab_character", "\tv1.2.3", "1.2.3", true, "Tab handled by TrimSpace"),
		makeSanitizeTestCase("newline", "v1.2.3\n", "1.2.3", true, "Newline handled by TrimSpace"),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanVersionString(tt.input)
			if got != tt.want {
				t.Errorf("CleanVersionString(%q) = %q, want %q (description: %s)",
					tt.input, got, tt.want, tt.description)
			}
		})
	}
}
