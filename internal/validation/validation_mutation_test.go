package validation

import (
	"strings"
	"testing"
)

// TestIsCommitSHAMutationResistance tests SHA validation for boundary mutations.
// Critical mutations to catch:
// - len(version) >= 7 changed to > 7 or >= 8
// - Regex pattern changes (e.g., + to *, removal of quantifiers).
func TestIsCommitSHAMutationResistance(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		want        bool
		critical    bool
		description string
	}{
		// Boundary: len >= 7
		{
			name:        "boundary_7_chars_valid",
			version:     "abc1234",
			want:        true,
			critical:    true,
			description: "Exactly 7 chars (boundary for >= 7)",
		},
		{
			name:        "boundary_6_chars_invalid",
			version:     "abc123",
			want:        false,
			critical:    true,
			description: "6 chars should fail (< 7)",
		},
		{
			name:        "boundary_8_chars_valid",
			version:     "abc12345",
			want:        true,
			critical:    false,
			description: "8 chars valid",
		},

		// Boundary: full SHA (40 chars)
		{
			name:        "boundary_40_chars_valid",
			version:     strings.Repeat("a", 40),
			want:        true,
			critical:    true,
			description: "Full 40-char SHA",
		},
		{
			name:        "boundary_39_chars_valid_short_sha",
			version:     strings.Repeat("a", 39),
			want:        true,
			critical:    false,
			description: "39 chars still valid as short SHA",
		},
		{
			name:        "boundary_41_chars_invalid_too_long",
			version:     strings.Repeat("a", 41),
			want:        false,
			critical:    true,
			description: "41 chars exceeds SHA length",
		},

		// Hex character validation (regex critical)
		{
			name:        "all_hex_chars_valid",
			version:     "abcdef0123456789",
			want:        true,
			critical:    false,
			description: "All hex chars",
		},
		{
			name:        "uppercase_hex_invalid",
			version:     "ABCDEF0",
			want:        false,
			critical:    true,
			description: "Uppercase hex chars (regex only accepts [a-f], not [A-F])",
		},
		{
			name:        "mixed_case_hex_invalid",
			version:     "AbCdEf0",
			want:        false,
			critical:    true,
			description: "Mixed case hex (regex only accepts lowercase)",
		},
		{
			name:        "non_hex_char_g_invalid",
			version:     "abcdefg",
			want:        false,
			critical:    true,
			description: "Contains 'g' (not hex)",
		},
		{
			name:        "non_hex_char_z_invalid",
			version:     "abcdefz",
			want:        false,
			critical:    true,
			description: "Contains 'z' (not hex)",
		},
		{
			name:        "special_char_invalid",
			version:     "abc-def",
			want:        false,
			critical:    true,
			description: "Contains dash",
		},

		// Empty/whitespace
		{
			name:        "empty_string_invalid",
			version:     "",
			want:        false,
			critical:    true,
			description: "Empty string (len < 7)",
		},
		{
			name:        "whitespace_invalid",
			version:     "   ",
			want:        false,
			critical:    false,
			description: "Whitespace only",
		},

		// Real-world SHA examples
		{
			name:        "real_short_sha",
			version:     "abc1234",
			want:        true,
			critical:    false,
			description: "Realistic 7-char short SHA",
		},
		{
			name:        "real_full_sha",
			version:     "1234567890abcdef1234567890abcdef12345678",
			want:        true,
			critical:    false,
			description: "Realistic full SHA",
		},
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
	tests := []struct {
		name        string
		version     string
		want        bool
		critical    bool
		description string
	}{
		// Basic semver
		{
			name:        "basic_semver",
			version:     "1.2.3",
			want:        true,
			critical:    false,
			description: "Basic X.Y.Z",
		},
		{
			name:        "basic_semver_with_v",
			version:     "v1.2.3",
			want:        true,
			critical:    true,
			description: "v prefix optional (v? quantifier)",
		},

		// Missing parts (should fail)
		{
			name:        "missing_patch_invalid",
			version:     "1.2",
			want:        false,
			critical:    true,
			description: "Missing patch version",
		},
		{
			name:        "missing_minor_patch_invalid",
			version:     "1",
			want:        false,
			critical:    true,
			description: "Only major version",
		},
		{
			name:        "extra_parts_invalid",
			version:     "1.2.3.4",
			want:        false,
			critical:    true,
			description: "Too many parts (no $ anchor would allow this)",
		},

		// Prerelease versions (optional part)
		{
			name:        "prerelease_alpha",
			version:     "1.2.3-alpha",
			want:        true,
			critical:    true,
			description: "Prerelease part (- with ? quantifier)",
		},
		{
			name:        "prerelease_alpha_1",
			version:     "1.2.3-alpha.1",
			want:        true,
			critical:    true,
			description: "Prerelease with dot",
		},
		{
			name:        "prerelease_multiple_parts",
			version:     "1.2.3-alpha.beta.1",
			want:        true,
			critical:    false,
			description: "Multiple prerelease parts",
		},
		{
			name:        "empty_prerelease_invalid",
			version:     "1.2.3-",
			want:        false,
			critical:    true,
			description: "Dash with no prerelease (+ requires content)",
		},

		// Build metadata (optional part)
		{
			name:        "build_metadata",
			version:     "1.2.3+build.123",
			want:        true,
			critical:    true,
			description: "Build metadata (+ with ? quantifier)",
		},
		{
			name:        "empty_build_invalid",
			version:     "1.2.3+",
			want:        false,
			critical:    true,
			description: "Plus with no build metadata",
		},
		{
			name:        "build_metadata_only_numbers",
			version:     "1.2.3+20130313144700",
			want:        true,
			critical:    false,
			description: "Build with only numbers",
		},

		// Combined prerelease and build
		{
			name:        "prerelease_and_build",
			version:     "1.2.3-alpha+build.123",
			want:        true,
			critical:    false,
			description: "Both prerelease and build",
		},

		// Zero versions
		{
			name:        "zero_version",
			version:     "0.0.0",
			want:        true,
			critical:    false,
			description: "All zeros valid",
		},
		{
			name:        "zero_major",
			version:     "0.1.2",
			want:        true,
			critical:    false,
			description: "Zero major valid",
		},

		// Large numbers
		{
			name:        "large_numbers",
			version:     "100.200.300",
			want:        true,
			critical:    false,
			description: "Multi-digit versions",
		},

		// Invalid formats
		{
			name:        "no_dots_invalid",
			version:     "123",
			want:        false,
			critical:    true,
			description: "No dots",
		},
		{
			name:        "letters_in_version_invalid",
			version:     "a.b.c",
			want:        false,
			critical:    true,
			description: "Letters in version numbers",
		},
		{
			name:        "leading_zero_technically_valid",
			version:     "01.02.03",
			want:        true,
			critical:    false,
			description: "Leading zeros (regex allows)",
		},

		// v prefix edge cases
		{
			name:        "double_v_invalid",
			version:     "vv1.2.3",
			want:        false,
			critical:    true,
			description: "Double v prefix (v? means 0 or 1)",
		},
		{
			name:        "uppercase_V_invalid",
			version:     "V1.2.3",
			want:        false,
			critical:    true,
			description: "Uppercase V not allowed",
		},

		// Whitespace
		{
			name:        "leading_whitespace_invalid",
			version:     " 1.2.3",
			want:        false,
			critical:    true,
			description: "Leading space (^ anchor)",
		},
		{
			name:        "trailing_whitespace_invalid",
			version:     "1.2.3 ",
			want:        false,
			critical:    true,
			description: "Trailing space ($ anchor)",
		},
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
	tests := []struct {
		name        string
		version     string
		want        bool
		critical    bool
		description string
	}{
		// Semantic version cases (first part of ||)
		{
			name:        "semver_is_pinned",
			version:     "v1.2.3",
			want:        true,
			critical:    true,
			description: "Semver satisfies first condition",
		},
		{
			name:        "semver_no_v_is_pinned",
			version:     "1.2.3",
			want:        true,
			critical:    true,
			description: "Semver without v",
		},

		// Full SHA cases (second part of ||)
		{
			name:        "full_40_char_sha_is_pinned",
			version:     strings.Repeat("a", 40),
			want:        true,
			critical:    true,
			description: "40-char SHA satisfies: IsCommitSHA() && len == 40",
		},
		{
			name:        "39_char_sha_not_pinned",
			version:     strings.Repeat("a", 39),
			want:        false,
			critical:    true,
			description: "39-char SHA fails: len != 40 (critical boundary)",
		},
		{
			name:        "41_char_not_sha_not_pinned",
			version:     strings.Repeat("a", 41),
			want:        false,
			critical:    true,
			description: "41 chars: not valid SHA && len != 40",
		},

		// Short SHA cases (should not be pinned)
		{
			name:        "7_char_sha_not_pinned",
			version:     "abcdef0",
			want:        false,
			critical:    true,
			description: "7-char SHA: IsCommitSHA() true but len != 40",
		},
		{
			name:        "20_char_sha_not_pinned",
			version:     strings.Repeat("a", 20),
			want:        false,
			critical:    true,
			description: "20-char SHA: IsCommitSHA() true but len != 40",
		},

		// Major-only versions (not pinned)
		{
			name:        "major_only_not_pinned",
			version:     "v1",
			want:        false,
			critical:    true,
			description: "v1 not semver, not pinned",
		},
		{
			name:        "major_minor_not_pinned",
			version:     "v1.2",
			want:        false,
			critical:    true,
			description: "v1.2 not semver (missing patch), not pinned",
		},

		// Branch names (not pinned)
		{
			name:        "branch_main_not_pinned",
			version:     "main",
			want:        false,
			critical:    true,
			description: "Branch name: not semver, not SHA",
		},
		{
			name:        "branch_develop_not_pinned",
			version:     "develop",
			want:        false,
			critical:    false,
			description: "Branch name: not semver, not SHA",
		},

		// Edge cases with prerelease/build
		{
			name:        "semver_with_prerelease_pinned",
			version:     "1.2.3-alpha",
			want:        true,
			critical:    false,
			description: "Semver with prerelease still pinned",
		},
		{
			name:        "semver_with_build_pinned",
			version:     "1.2.3+build",
			want:        true,
			critical:    false,
			description: "Semver with build metadata still pinned",
		},

		// Empty/invalid
		{
			name:        "empty_not_pinned",
			version:     "",
			want:        false,
			critical:    true,
			description: "Empty string: not semver, not SHA",
		},

		// Operator mutation detection tests
		{
			name:        "exactly_40_boundary",
			version:     strings.Repeat("a", 40),
			want:        true,
			critical:    true,
			description: "Exactly 40: tests == boundary (not !=, <, >, <=, >=)",
		},
		{
			name:        "40_char_non_hex_not_sha",
			version:     strings.Repeat("z", 40),
			want:        false,
			critical:    true,
			description: "40 chars but not hex: IsCommitSHA() false, so && fails",
		},
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
			version:     "v1.2.3",
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
			version:     "v1",
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
