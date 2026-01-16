package validation

import (
	"testing"
)

// TestParseGitHubURLMutationResistance tests URL parsing for regex and boundary mutations.
// Critical mutations to catch:
// - Pattern order changes (SSH vs HTTPS precedence)
// - len(matches) >= 3 changed to > 3, == 3, etc.
// - Return statement modifications (returning wrong indices).
func TestParseGitHubURLMutationResistance(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		wantOrg     string
		wantRepo    string
		critical    bool
		description string
	}{
		// HTTPS URLs
		{
			name:        "https_standard",
			url:         "https://github.com/octocat/Hello-World",
			wantOrg:     "octocat",
			wantRepo:    "Hello-World",
			critical:    false,
			description: "Standard HTTPS URL",
		},
		{
			name:        "https_with_git_extension",
			url:         "https://github.com/octocat/Hello-World.git",
			wantOrg:     "octocat",
			wantRepo:    "Hello-World",
			critical:    true,
			description: ".git extension handled by (?:\\.git)? regex",
		},

		// SSH URLs
		{
			name:        "ssh_standard",
			url:         "git@github.com:octocat/Hello-World",
			wantOrg:     "octocat",
			wantRepo:    "Hello-World",
			critical:    true,
			description: "SSH URL with colon separator ([:/] pattern)",
		},
		{
			name:        "ssh_with_git_extension",
			url:         "git@github.com:octocat/Hello-World.git",
			wantOrg:     "octocat",
			wantRepo:    "Hello-World",
			critical:    true,
			description: "SSH with .git",
		},

		// Simple format
		{
			name:        "simple_org_repo",
			url:         "octocat/Hello-World",
			wantOrg:     "octocat",
			wantRepo:    "Hello-World",
			critical:    true,
			description: "Simple org/repo format (second pattern)",
		},

		// Edge cases with special characters
		{
			name:        "org_with_dash",
			url:         "actions/setup-node",
			wantOrg:     "actions",
			wantRepo:    "setup-node",
			critical:    false,
			description: "Hyphen in repo name",
		},
		{
			name:        "org_with_number",
			url:         "org123/repo456",
			wantOrg:     "org123",
			wantRepo:    "repo456",
			critical:    false,
			description: "Numbers in org/repo",
		},

		// Boundary: len(matches) >= 3
		{
			name:        "exactly_3_matches",
			url:         "a/b",
			wantOrg:     "a",
			wantRepo:    "b",
			critical:    true,
			description: "Minimal valid: exactly 3 matches (full, org, repo)",
		},

		// Invalid URLs (should return empty)
		{
			name:        "no_slash_invalid",
			url:         "octocatHello-World",
			wantOrg:     "",
			wantRepo:    "",
			critical:    true,
			description: "No slash separator",
		},
		{
			name:        "empty_string",
			url:         "",
			wantOrg:     "",
			wantRepo:    "",
			critical:    true,
			description: "Empty string",
		},
		{
			name:        "only_org",
			url:         "octocat/",
			wantOrg:     "",
			wantRepo:    "",
			critical:    true,
			description: "Trailing slash, no repo",
		},
		{
			name:        "only_repo",
			url:         "/Hello-World",
			wantOrg:     "",
			wantRepo:    "",
			critical:    true,
			description: "Leading slash, no org",
		},

		// Pattern precedence tests
		{
			name:        "github_com_in_middle",
			url:         "https://github.com/ivuorinen/gh-action-readme",
			wantOrg:     "ivuorinen",
			wantRepo:    "gh-action-readme",
			critical:    false,
			description: "First pattern should match",
		},

		// Regex capture group tests
		{
			name:        "multiple_slashes",
			url:         "octocat/Hello-World/extra",
			wantOrg:     "",
			wantRepo:    "",
			critical:    false,
			description: "Extra path segments invalid for simple format",
		},

		// .git extension edge cases
		{
			name:        "double_git_extension",
			url:         "octocat/Hello-World.git.git",
			wantOrg:     "",
			wantRepo:    "",
			critical:    true,
			description: "Dots not allowed in repo name by [^/.] pattern",
		},
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
	tests := []struct {
		name        string
		input       string
		want        string
		critical    bool
		description string
	}{
		// Basic transformations
		{
			name:        "lowercase_conversion",
			input:       "UPPERCASE",
			want:        "uppercase",
			critical:    true,
			description: "ToLower applied",
		},
		{
			name:        "space_to_dash",
			input:       "hello world",
			want:        "hello-world",
			critical:    true,
			description: "ReplaceAll spaces with dashes",
		},
		{
			name:        "trim_spaces",
			input:       "  hello  ",
			want:        "hello",
			critical:    true,
			description: "TrimSpace applied",
		},

		// Multiple spaces (ReplaceAll vs Replace critical)
		{
			name:        "multiple_spaces_all_replaced",
			input:       "hello  world  test",
			want:        "hello--world--test",
			critical:    true,
			description: "All spaces replaced (ReplaceAll, not Replace)",
		},
		{
			name:        "three_consecutive_spaces",
			input:       "a   b",
			want:        "a---b",
			critical:    true,
			description: "Each space replaced individually",
		},

		// Operation order tests
		{
			name:        "uppercase_with_spaces",
			input:       "HELLO WORLD",
			want:        "hello-world",
			critical:    true,
			description: "Both lowercase and space replacement",
		},
		{
			name:        "leading_trailing_spaces_uppercase",
			input:       "  HELLO WORLD  ",
			want:        "hello-world",
			critical:    true,
			description: "All transformations: trim, replace, lowercase",
		},

		// Edge cases
		{
			name:        "empty_string",
			input:       "",
			want:        "",
			critical:    true,
			description: "Empty input",
		},
		{
			name:        "only_spaces",
			input:       "   ",
			want:        "",
			critical:    true,
			description: "Only spaces (trimmed to empty)",
		},
		{
			name:        "no_changes_needed",
			input:       "already-sanitized",
			want:        "already-sanitized",
			critical:    false,
			description: "Already in correct format",
		},

		// Special characters
		{
			name:        "mixed_case_with_hyphens",
			input:       "Setup-Node",
			want:        "setup-node",
			critical:    false,
			description: "Existing hyphens preserved",
		},
		{
			name:        "underscore_preserved",
			input:       "hello_world",
			want:        "hello_world",
			critical:    false,
			description: "Underscores not replaced",
		},
		{
			name:        "numbers_preserved",
			input:       "Action 123",
			want:        "action-123",
			critical:    false,
			description: "Numbers preserved",
		},

		// Real-world action names
		{
			name:        "checkout_action",
			input:       "Checkout Code",
			want:        "checkout-code",
			critical:    false,
			description: "Realistic action name",
		},
		{
			name:        "setup_go_action",
			input:       "Setup Go Environment",
			want:        "setup-go-environment",
			critical:    false,
			description: "Multi-word action name",
		},

		// Single character
		{
			name:        "single_char",
			input:       "A",
			want:        "a",
			critical:    false,
			description: "Single character",
		},
		{
			name:        "single_space",
			input:       " ",
			want:        "",
			critical:    true,
			description: "Single space (trimmed)",
		},
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
	tests := []struct {
		name        string
		input       string
		want        string
		critical    bool
		description string
	}{
		// Leading/trailing whitespace
		{
			name:        "leading_whitespace",
			input:       "  hello",
			want:        "hello",
			critical:    true,
			description: "TrimSpace removes leading",
		},
		{
			name:        "trailing_whitespace",
			input:       "hello  ",
			want:        "hello",
			critical:    true,
			description: "TrimSpace removes trailing",
		},
		{
			name:        "both_sides_whitespace",
			input:       "  hello  ",
			want:        "hello",
			critical:    true,
			description: "TrimSpace removes both sides",
		},

		// Internal whitespace normalization
		{
			name:        "double_space",
			input:       "hello  world",
			want:        "hello world",
			critical:    true,
			description: "Double space to single (\\s+ pattern)",
		},
		{
			name:        "triple_space",
			input:       "hello   world",
			want:        "hello world",
			critical:    true,
			description: "Triple space to single",
		},
		{
			name:        "many_spaces",
			input:       "hello          world",
			want:        "hello world",
			critical:    true,
			description: "Many spaces to single (+ quantifier)",
		},

		// Mixed whitespace types
		{
			name:        "tab_character",
			input:       "hello\tworld",
			want:        "hello world",
			critical:    true,
			description: "Tab normalized to space (\\s includes tabs)",
		},
		{
			name:        "newline_character",
			input:       "hello\nworld",
			want:        "hello world",
			critical:    true,
			description: "Newline normalized to space (\\s includes newlines)",
		},
		{
			name:        "carriage_return",
			input:       "hello\rworld",
			want:        "hello world",
			critical:    true,
			description: "CR normalized to space",
		},
		{
			name:        "mixed_whitespace",
			input:       "hello \t\n world",
			want:        "hello world",
			critical:    true,
			description: "Mixed whitespace types to single space",
		},

		// Combined leading/trailing and internal
		{
			name:        "all_whitespace_issues",
			input:       "  hello   world  ",
			want:        "hello world",
			critical:    true,
			description: "Trim + normalize internal",
		},

		// Edge cases
		{
			name:        "empty_string",
			input:       "",
			want:        "",
			critical:    true,
			description: "Empty input",
		},
		{
			name:        "only_spaces",
			input:       "     ",
			want:        "",
			critical:    true,
			description: "Only spaces (trimmed to empty)",
		},
		{
			name:        "only_whitespace_mixed",
			input:       " \t\n\r ",
			want:        "",
			critical:    true,
			description: "Only various whitespace types",
		},
		{
			name:        "no_whitespace",
			input:       "hello",
			want:        "hello",
			critical:    false,
			description: "No whitespace to normalize",
		},
		{
			name:        "single_space_valid",
			input:       "hello world",
			want:        "hello world",
			critical:    false,
			description: "Already normalized",
		},

		// Multiple words
		{
			name:        "three_words_excess_spaces",
			input:       "one   two   three",
			want:        "one two three",
			critical:    false,
			description: "Three words with excess spaces",
		},

		// Unicode whitespace
		{
			name:        "regular_space",
			input:       "hello world",
			want:        "hello world",
			critical:    false,
			description: "Regular ASCII space",
		},

		// Quantifier verification (\s+ means one or more)
		{
			name:        "single_space_between",
			input:       "a b",
			want:        "a b",
			critical:    true,
			description: "Single space not collapsed (need + for >1)",
		},
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
	tests := []struct {
		name        string
		org         string
		repo        string
		version     string
		want        string
		critical    bool
		description string
	}{
		// Basic formatting
		{
			name:        "basic_with_version",
			org:         "actions",
			repo:        "checkout",
			version:     "v3",
			want:        "actions/checkout@v3",
			critical:    false,
			description: "Standard format",
		},

		// Empty checks (critical)
		{
			name:        "empty_org_returns_empty",
			org:         "",
			repo:        "checkout",
			version:     "v3",
			want:        "",
			critical:    true,
			description: "org == \"\" check",
		},
		{
			name:        "empty_repo_returns_empty",
			org:         "actions",
			repo:        "",
			version:     "v3",
			want:        "",
			critical:    true,
			description: "repo == \"\" check",
		},
		{
			name:        "both_empty_returns_empty",
			org:         "",
			repo:        "",
			version:     "v3",
			want:        "",
			critical:    true,
			description: "org == \"\" || repo == \"\" (|| operator critical)",
		},

		// Default version (critical)
		{
			name:        "empty_version_defaults_v1",
			org:         "actions",
			repo:        "checkout",
			version:     "",
			want:        "actions/checkout@v1",
			critical:    true,
			description: "version == \"\" defaults to \"v1\"",
		},

		// @ prefix handling (critical)
		{
			name:        "version_without_at",
			org:         "actions",
			repo:        "checkout",
			version:     "v3",
			want:        "actions/checkout@v3",
			critical:    true,
			description: "@ added when not present (!HasPrefix check)",
		},
		{
			name:        "version_with_at",
			org:         "actions",
			repo:        "checkout",
			version:     "@v3",
			want:        "actions/checkout@v3",
			critical:    true,
			description: "@ not duplicated (HasPrefix check)",
		},
		{
			name:        "double_at_if_hasprefix_fails",
			org:         "actions",
			repo:        "checkout",
			version:     "@@v3",
			want:        "actions/checkout@@v3",
			critical:    false,
			description: "Malformed input with double @",
		},

		// String concatenation order
		{
			name:        "concatenation_order",
			org:         "org",
			repo:        "repo",
			version:     "ver",
			want:        "org/repo@ver",
			critical:    true,
			description: "Correct concatenation: org + \"/\" + repo + version",
		},

		// Edge cases
		{
			name:        "single_char_org_repo",
			org:         "a",
			repo:        "b",
			version:     "c",
			want:        "a/b@c",
			critical:    false,
			description: "Minimal valid input",
		},
		{
			name:        "branch_name_version",
			org:         "actions",
			repo:        "checkout",
			version:     "main",
			want:        "actions/checkout@main",
			critical:    false,
			description: "Branch name as version",
		},
		{
			name:        "sha_version",
			org:         "actions",
			repo:        "checkout",
			version:     "abc1234567890def",
			want:        "actions/checkout@abc1234567890def",
			critical:    false,
			description: "SHA as version",
		},

		// Whitespace in inputs
		{
			name:        "org_with_spaces_not_trimmed",
			org:         " actions ",
			repo:        "checkout",
			version:     "v3",
			want:        " actions /checkout@v3",
			critical:    false,
			description: "Spaces preserved (no TrimSpace in function)",
		},

		// Special characters
		{
			name:        "hyphen_in_repo",
			org:         "actions",
			repo:        "setup-node",
			version:     "v3",
			want:        "actions/setup-node@v3",
			critical:    false,
			description: "Hyphen in repo name",
		},
		{
			name:        "at_in_version_position",
			org:         "actions",
			repo:        "checkout",
			version:     "@v3",
			want:        "actions/checkout@v3",
			critical:    true,
			description: "Existing @ not duplicated",
		},

		// Boolean operator mutation detection
		{
			name:        "non_empty_org_empty_repo",
			org:         "actions",
			repo:        "",
			version:     "v3",
			want:        "",
			critical:    true,
			description: "|| means either empty returns \"\" (not &&)",
		},
		{
			name:        "empty_org_non_empty_repo",
			org:         "",
			repo:        "checkout",
			version:     "v3",
			want:        "",
			critical:    true,
			description: "|| means either empty returns \"\" (not &&)",
		},

		// Default version with @ handling
		{
			name:        "empty_version_gets_at_prefix",
			org:         "actions",
			repo:        "checkout",
			version:     "",
			want:        "actions/checkout@v1",
			critical:    true,
			description: "Empty version: default \"v1\" then @ added",
		},
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
	tests := []struct {
		name        string
		input       string
		want        string
		critical    bool
		description string
	}{
		// v prefix removal
		{
			name:        "v_prefix_removed",
			input:       "v1.2.3",
			want:        "1.2.3",
			critical:    true,
			description: "TrimPrefix(\"v\") applied",
		},
		{
			name:        "no_v_prefix_unchanged",
			input:       "1.2.3",
			want:        "1.2.3",
			critical:    true,
			description: "No v prefix to remove",
		},

		// Whitespace handling
		{
			name:        "leading_whitespace",
			input:       "  v1.2.3",
			want:        "1.2.3",
			critical:    true,
			description: "TrimSpace before TrimPrefix",
		},
		{
			name:        "trailing_whitespace",
			input:       "v1.2.3  ",
			want:        "1.2.3",
			critical:    true,
			description: "TrimSpace applied",
		},
		{
			name:        "both_whitespace_and_v",
			input:       "  v1.2.3  ",
			want:        "1.2.3",
			critical:    true,
			description: "Both TrimSpace and TrimPrefix",
		},

		// Operation order critical
		{
			name:        "whitespace_before_v",
			input:       " v1.2.3",
			want:        "1.2.3",
			critical:    true,
			description: "TrimSpace must happen before TrimPrefix",
		},

		// Edge cases
		{
			name:        "only_v",
			input:       "v",
			want:        "",
			critical:    true,
			description: "Just v becomes empty",
		},
		{
			name:        "empty_string",
			input:       "",
			want:        "",
			critical:    true,
			description: "Empty input",
		},
		{
			name:        "only_whitespace",
			input:       "   ",
			want:        "",
			critical:    true,
			description: "Only spaces",
		},

		// Multiple v's
		{
			name:        "double_v",
			input:       "vv1.2.3",
			want:        "v1.2.3",
			critical:    true,
			description: "Only first v removed (TrimPrefix, not ReplaceAll)",
		},

		// No changes needed
		{
			name:        "already_clean",
			input:       "1.2.3",
			want:        "1.2.3",
			critical:    false,
			description: "Already clean",
		},

		// Real-world versions
		{
			name:        "semver_with_v",
			input:       "v2.5.1",
			want:        "2.5.1",
			critical:    false,
			description: "Realistic semver",
		},
		{
			name:        "semver_no_v",
			input:       "2.5.1",
			want:        "2.5.1",
			critical:    false,
			description: "Realistic semver without v",
		},

		// Whitespace variations
		{
			name:        "tab_character",
			input:       "\tv1.2.3",
			want:        "1.2.3",
			critical:    true,
			description: "Tab handled by TrimSpace",
		},
		{
			name:        "newline",
			input:       "v1.2.3\n",
			want:        "1.2.3",
			critical:    true,
			description: "Newline handled by TrimSpace",
		},
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
