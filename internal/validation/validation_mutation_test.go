package validation

import (
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

// Test case helpers - reduce duplication in table-driven tests

// shaTestCase represents a SHA validation test case.
type shaTestCase struct {
	name        string
	version     string
	want        bool
	critical    bool
	description string
}

// makeSHATestCase constructs a SHA test case.
func makeSHATestCase(name, version string, want, critical bool, desc string) shaTestCase {
	return shaTestCase{
		name:        name,
		version:     version,
		want:        want,
		critical:    critical,
		description: desc,
	}
}

// semverTestCase represents a semantic version validation test case.
type semverTestCase struct {
	name        string
	version     string
	want        bool
	critical    bool
	description string
}

// makeSemverTestCase constructs a semantic version test case.
func makeSemverTestCase(name, version string, want, critical bool, desc string) semverTestCase {
	return semverTestCase{
		name:        name,
		version:     version,
		want:        want,
		critical:    critical,
		description: desc,
	}
}

// pinnedTestCase represents a version pinning test case.
type pinnedTestCase struct {
	name        string
	version     string
	want        bool
	critical    bool
	description string
}

// makePinnedTestCase constructs a version pinning test case.
func makePinnedTestCase(name, version string, want, critical bool, desc string) pinnedTestCase {
	return pinnedTestCase{
		name:        name,
		version:     version,
		want:        want,
		critical:    critical,
		description: desc,
	}
}

// TestIsCommitSHAMutationResistance tests SHA validation for boundary mutations.
// Critical mutations to catch:
// - len(version) >= 7 changed to > 7 or >= 8
// - Regex pattern changes (e.g., + to *, removal of quantifiers).
func TestIsCommitSHAMutationResistance(t *testing.T) {
	tests := []shaTestCase{
		// Boundary: len >= 7
		makeSHATestCase("boundary_7_chars_valid", "abc1234", true, true, "Exactly 7 chars (boundary for >= 7)"),
		makeSHATestCase("boundary_6_chars_invalid", "abc123", false, true, "6 chars should fail (< 7)"),
		makeSHATestCase("boundary_8_chars_valid", "abc12345", true, false, "8 chars valid"),

		// Boundary: full SHA (40 chars)
		makeSHATestCase("boundary_40_chars_valid", strings.Repeat("a", 40), true, true, "Full 40-char SHA"),
		makeSHATestCase(
			"boundary_39_chars_valid_short_sha",
			strings.Repeat("a", 39),
			true,
			false,
			"39 chars still valid as short SHA",
		),
		makeSHATestCase(
			"boundary_41_chars_invalid_too_long",
			strings.Repeat("a", 41),
			false,
			true,
			"41 chars exceeds SHA length",
		),

		// Hex character validation (regex critical)
		makeSHATestCase("all_hex_chars_valid", "abcdef0123456789", true, false, "All hex chars"),
		makeSHATestCase(
			"uppercase_hex_invalid",
			"ABCDEF0",
			false,
			true,
			"Uppercase hex chars (regex only accepts [a-f], not [A-F])",
		),
		makeSHATestCase(
			"mixed_case_hex_invalid",
			"AbCdEf0",
			false,
			true,
			"Mixed case hex (regex only accepts lowercase)",
		),
		makeSHATestCase("non_hex_char_g_invalid", "abcdefg", false, true, "Contains 'g' (not hex)"),
		makeSHATestCase("non_hex_char_z_invalid", "abcdefz", false, true, "Contains 'z' (not hex)"),
		makeSHATestCase("special_char_invalid", "abc-def", false, true, "Contains dash"),

		// Empty/whitespace
		makeSHATestCase("empty_string_invalid", "", false, true, "Empty string (len < 7)"),
		makeSHATestCase("whitespace_invalid", "   ", false, false, "Whitespace only"),

		// Real-world SHA examples
		makeSHATestCase("real_short_sha", "abc1234", true, false, "Realistic 7-char short SHA"),
		makeSHATestCase("real_full_sha", "1234567890abcdef1234567890abcdef12345678", true, false, "Realistic full SHA"),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCommitSHA(tt.version)
			if got != tt.want {
				t.Errorf("IsCommitSHA(%q) = %v, want %v (description: %s)",
					tt.version, got, tt.want, tt.description)
			}
		})
	}
}

// TestIsSemanticVersionMutationResistance tests semver validation for regex mutations.
// Critical mutations to catch:
// - Quantifier changes (? to *, + to *, removal of ?)
// - Part removal (prerelease, build metadata)
// - Anchor removal (^ or $).
func TestIsSemanticVersionMutationResistance(t *testing.T) {
	tests := []semverTestCase{
		// Basic semver
		makeSemverTestCase("basic_semver", "1.2.3", true, false, "Basic X.Y.Z"),
		makeSemverTestCase(
			"basic_semver_with_v",
			testutil.TestVersionSemantic,
			true,
			true,
			"v prefix optional (v? quantifier)",
		),

		// Missing parts (should fail)
		makeSemverTestCase("missing_patch_invalid", "1.2", false, true, "Missing patch version"),
		makeSemverTestCase("missing_minor_patch_invalid", "1", false, true, "Only major version"),
		makeSemverTestCase(
			"extra_parts_invalid",
			testutil.MutationSemverInvalidExtraParts,
			false,
			true,
			"Too many parts (no $ anchor would allow this)",
		),

		// Prerelease versions (optional part)
		makeSemverTestCase("prerelease_alpha", "1.2.3-alpha", true, true, "Prerelease part (- with ? quantifier)"),
		makeSemverTestCase("prerelease_alpha_1", "1.2.3-alpha.1", true, true, "Prerelease with dot"),
		makeSemverTestCase("prerelease_multiple_parts", "1.2.3-alpha.beta.1", true, false, "Multiple prerelease parts"),
		makeSemverTestCase(
			"empty_prerelease_invalid",
			testutil.MutationSemverEmptyPrerelease,
			false,
			true,
			"Dash with no prerelease (+ requires content)",
		),

		// Build metadata (optional part)
		makeSemverTestCase("build_metadata", "1.2.3+build.123", true, true, "Build metadata (+ with ? quantifier)"),
		makeSemverTestCase("empty_build_invalid", "1.2.3+", false, true, "Plus with no build metadata"),
		makeSemverTestCase(
			"build_metadata_only_numbers",
			testutil.MutationSemverBuildOnlyNumbers,
			true,
			false,
			"Build with only numbers",
		),

		// Combined prerelease and build
		makeSemverTestCase("prerelease_and_build", "1.2.3-alpha+build.123", true, false, "Both prerelease and build"),

		// Zero versions
		makeSemverTestCase("zero_version", "0.0.0", true, false, "All zeros valid"),
		makeSemverTestCase("zero_major", "0.1.2", true, false, "Zero major valid"),

		// Large numbers
		makeSemverTestCase("large_numbers", "100.200.300", true, false, "Multi-digit versions"),

		// Invalid formats
		makeSemverTestCase("no_dots_invalid", "123", false, true, "No dots"),
		makeSemverTestCase("letters_in_version_invalid", "a.b.c", false, true, "Letters in version numbers"),
		makeSemverTestCase("leading_zero_technically_valid", "01.02.03", true, false, "Leading zeros (regex allows)"),

		// v prefix edge cases
		makeSemverTestCase(
			"double_v_invalid",
			testutil.MutationSemverDoubleV,
			false,
			true,
			"Double v prefix (v? means 0 or 1)",
		),
		makeSemverTestCase(
			"uppercase_V_invalid",
			testutil.MutationSemverUppercaseV,
			false,
			true,
			"Uppercase V not allowed",
		),

		// Whitespace
		makeSemverTestCase(
			"leading_whitespace_invalid",
			testutil.MutationSemverLeadingSpace,
			false,
			true,
			"Leading space (^ anchor)",
		),
		makeSemverTestCase(
			"trailing_whitespace_invalid",
			testutil.MutationSemverTrailingSpace,
			false,
			true,
			"Trailing space ($ anchor)",
		),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSemanticVersion(tt.version)
			if got != tt.want {
				t.Errorf("IsSemanticVersion(%q) = %v, want %v (description: %s)",
					tt.version, got, tt.want, tt.description)
			}
		})
	}
}

// TestIsVersionPinnedMutationResistance tests version pinning logic for operator mutations.
// Critical mutations to catch:
// - || changed to && (complete logic inversion)
// - && changed to || in SHA check
// - == 40 changed to != 40, > 40, < 40, >= 40, <= 40
// - Removal of IsSemanticVersion() or IsCommitSHA() calls.
func TestIsVersionPinnedMutationResistance(t *testing.T) {
	tests := []pinnedTestCase{
		// Semantic version cases (first part of ||)
		makePinnedTestCase(
			"semver_is_pinned",
			testutil.TestVersionSemantic,
			true,
			true,
			"Semver satisfies first condition",
		),
		makePinnedTestCase("semver_no_v_is_pinned", "1.2.3", true, true, "Semver without v"),

		// Full SHA cases (second part of ||)
		makePinnedTestCase(
			"full_40_char_sha_is_pinned",
			strings.Repeat("a", 40),
			true,
			true,
			"40-char SHA satisfies: IsCommitSHA() && len == 40",
		),
		makePinnedTestCase(
			"39_char_sha_not_pinned",
			strings.Repeat("a", 39),
			false,
			true,
			"39-char SHA fails: len != 40 (critical boundary)",
		),
		makePinnedTestCase(
			"41_char_not_sha_not_pinned",
			strings.Repeat("a", 41),
			false,
			true,
			"41 chars: not valid SHA && len != 40",
		),

		// Short SHA cases (should not be pinned)
		makePinnedTestCase(
			"7_char_sha_not_pinned",
			"abcdef0",
			false,
			true,
			"7-char SHA: IsCommitSHA() true but len != 40",
		),
		makePinnedTestCase(
			"20_char_sha_not_pinned",
			strings.Repeat("a", 20),
			false,
			true,
			"20-char SHA: IsCommitSHA() true but len != 40",
		),

		// Major-only versions (not pinned)
		makePinnedTestCase(
			"major_only_not_pinned",
			appconstants.VersionTagV1,
			false,
			true,
			"v1 not semver, not pinned",
		),
		makePinnedTestCase(
			"major_minor_not_pinned",
			"v1.2",
			false,
			true,
			"v1.2 not semver (missing patch), not pinned",
		),

		// Branch names (not pinned)
		makePinnedTestCase("branch_main_not_pinned", "main", false, true, "Branch name: not semver, not SHA"),
		makePinnedTestCase("branch_develop_not_pinned", "develop", false, false, "Branch name: not semver, not SHA"),

		// Edge cases with prerelease/build
		makePinnedTestCase(
			"semver_with_prerelease_pinned",
			"1.2.3-alpha",
			true,
			false,
			"Semver with prerelease still pinned",
		),
		makePinnedTestCase(
			"semver_with_build_pinned",
			"1.2.3+build",
			true,
			false,
			"Semver with build metadata still pinned",
		),

		// Empty/invalid
		makePinnedTestCase("empty_not_pinned", "", false, true, "Empty string: not semver, not SHA"),

		// Operator mutation detection tests
		makePinnedTestCase(
			"exactly_40_boundary",
			strings.Repeat("a", 40),
			true,
			true,
			"Exactly 40: tests == boundary (not !=, <, >, <=, >=)",
		),
		makePinnedTestCase(
			"40_char_non_hex_not_sha",
			strings.Repeat("z", 40),
			false,
			true,
			"40 chars but not hex: IsCommitSHA() false, so && fails",
		),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsVersionPinned(tt.version)
			if got != tt.want {
				t.Errorf("IsVersionPinned(%q) = %v, want %v (description: %s)",
					tt.version, got, tt.want, tt.description)
			}
		})
	}
}

// TestVersionValidationLogicCombinations tests the interaction between validation
// functions to catch mutations in boolean logic.
func TestVersionValidationLogicCombinations(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		isSHA       bool
		isSemver    bool
		isPinned    bool
		description string
	}{
		{
			name:        "full_sha_all_true",
			version:     strings.Repeat("a", 40),
			isSHA:       true,
			isSemver:    false,
			isPinned:    true,
			description: "40-char SHA: SHA && pinned, not semver",
		},
		{
			name:        "short_sha_not_pinned",
			version:     "abcdef0",
			isSHA:       true,
			isSemver:    false,
			isPinned:    false,
			description: "7-char SHA: SHA but not pinned",
		},
		{
			name:        "semver_all_relevant_true",
			version:     testutil.TestVersionSemantic,
			isSHA:       false,
			isSemver:    true,
			isPinned:    true,
			description: "Semver: not SHA, is semver, is pinned",
		},
		{
			name:        "branch_all_false",
			version:     "main",
			isSHA:       false,
			isSemver:    false,
			isPinned:    false,
			description: "Branch: nothing true",
		},
		{
			name:        "v1_not_semver_not_pinned",
			version:     appconstants.VersionTagV1,
			isSHA:       false,
			isSemver:    false,
			isPinned:    false,
			description: "Major-only: not valid semver",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSHA := IsCommitSHA(tt.version)
			gotSemver := IsSemanticVersion(tt.version)
			gotPinned := IsVersionPinned(tt.version)

			if gotSHA != tt.isSHA {
				t.Errorf("IsCommitSHA(%q) = %v, want %v", tt.version, gotSHA, tt.isSHA)
			}
			if gotSemver != tt.isSemver {
				t.Errorf("IsSemanticVersion(%q) = %v, want %v", tt.version, gotSemver, tt.isSemver)
			}
			if gotPinned != tt.isPinned {
				t.Errorf("IsVersionPinned(%q) = %v, want %v (description: %s)",
					tt.version, gotPinned, tt.isPinned, tt.description)
			}
		})
	}
}
