package internal

import (
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// TestPermissionParsingMutationResistance provides comprehensive test cases designed
// to catch mutations in the permission parsing logic. These tests target critical
// boundaries, operators, and conditions that are susceptible to mutation.
//
// permissionParsingTestCase defines a test case for permission parsing tests.
type permissionParsingTestCase struct {
	name     string
	yaml     string
	expected map[string]string
	critical bool
}

// buildPermissionParsingTestCases returns all test cases for permission parsing.
// YAML content is loaded from fixture files in testdata/yaml-fixtures/configs/permissions/mutation/.
func buildPermissionParsingTestCases() []permissionParsingTestCase {
	const fixtureDir = "configs/permissions/mutation/"

	return []permissionParsingTestCase{
		{
			name: "off_by_one_indent_two_items",
			yaml: testutil.MustReadFixture(fixtureDir + "off-by-one-indent-two-items.yaml"),
			expected: map[string]string{
				testutil.PermissionContents: testutil.PermissionRead,
				testutil.PermissionIssues:   testutil.PermissionWrite,
			},
			critical: true,
		},
		{
			name: "off_by_one_indent_three_items",
			yaml: testutil.MustReadFixture(fixtureDir + "off-by-one-indent-three-items.yaml"),
			expected: map[string]string{
				testutil.PermissionContents:      testutil.PermissionRead,
				testutil.PermissionIssues:        testutil.PermissionWrite,
				testutil.TestFixturePullRequests: testutil.PermissionRead,
			},
			critical: true,
		},
		{
			name:     "comment_position_at_boundary",
			yaml:     testutil.MustReadFixture(fixtureDir + "comment-position-at-boundary.yaml"),
			expected: map[string]string{testutil.PermissionContents: testutil.PermissionRead},
			critical: true,
		},
		{
			name:     "comment_at_position_zero_parses",
			yaml:     testutil.MustReadFixture(fixtureDir + "comment-at-position-zero-parses.yaml"),
			expected: map[string]string{testutil.PermissionContents: testutil.PermissionRead},
			critical: true,
		},
		{
			name: "dash_prefix_with_spaces",
			yaml: testutil.MustReadFixture(fixtureDir + "dash-prefix-with-spaces.yaml"),
			expected: map[string]string{
				testutil.PermissionContents: testutil.PermissionRead,
				testutil.PermissionIssues:   testutil.PermissionWrite,
			},
			critical: true,
		},
		{
			name: "mixed_dash_and_no_dash",
			yaml: testutil.MustReadFixture(fixtureDir + "mixed-dash-and-no-dash.yaml"),
			expected: map[string]string{
				testutil.PermissionContents: testutil.PermissionRead,
				testutil.PermissionIssues:   testutil.PermissionWrite,
			},
			critical: true,
		},
		{
			name:     "dedent_stops_parsing",
			yaml:     testutil.MustReadFixture(fixtureDir + "dedent-stops-parsing.yaml"),
			expected: map[string]string{testutil.PermissionContents: testutil.PermissionRead},
			critical: true,
		},
		{
			name: "empty_line_in_block_continues",
			yaml: testutil.MustReadFixture(fixtureDir + "empty-line-in-block-continues.yaml"),
			expected: map[string]string{
				testutil.PermissionContents: testutil.PermissionRead,
				testutil.PermissionIssues:   testutil.PermissionWrite,
			},
			critical: false,
		},
		{
			name:     "non_comment_line_stops_parsing",
			yaml:     testutil.MustReadFixture(fixtureDir + "non-comment-line-stops-parsing.yaml"),
			expected: map[string]string{testutil.PermissionContents: testutil.PermissionRead},
			critical: true,
		},
		{
			name:     "exact_expected_indent",
			yaml:     testutil.MustReadFixture(fixtureDir + "exact-expected-indent.yaml"),
			expected: map[string]string{testutil.PermissionContents: testutil.PermissionRead},
			critical: true,
		},
		{
			name:     "colon_in_value_preserved",
			yaml:     testutil.MustReadFixture(fixtureDir + "colon-in-value-preserved.yaml"),
			expected: map[string]string{testutil.PermissionContents: "read:write"},
			critical: true,
		},
		{
			name:     "empty_key_not_parsed",
			yaml:     testutil.MustReadFixture(fixtureDir + "empty-key-not-parsed.yaml"),
			expected: map[string]string{},
			critical: true,
		},
		{
			name:     "empty_value_not_parsed",
			yaml:     testutil.MustReadFixture(fixtureDir + "empty-value-not-parsed.yaml"),
			expected: map[string]string{},
			critical: true,
		},
		{
			name:     "whitespace_only_value_not_parsed",
			yaml:     testutil.MustReadFixture(fixtureDir + "whitespace-only-value-not-parsed.yaml"),
			expected: map[string]string{},
			critical: true,
		},
		{
			name:     "multiple_colons_splits_at_first",
			yaml:     testutil.MustReadFixture(fixtureDir + "multiple-colons-splits-at-first.yaml"),
			expected: map[string]string{"url": "https://example.com:8080"},
			critical: true,
		},
		{
			name:     "inline_comment_removal",
			yaml:     testutil.MustReadFixture(fixtureDir + "inline-comment-removal.yaml"),
			expected: map[string]string{testutil.PermissionContents: testutil.PermissionRead},
			critical: true,
		},
		{
			name:     "inline_comment_at_start_of_value",
			yaml:     testutil.MustReadFixture(fixtureDir + "inline-comment-at-start-of-value.yaml"),
			expected: map[string]string{},
			critical: true,
		},
		{
			name: "deeply_nested_indent",
			yaml: testutil.MustReadFixture(fixtureDir + "deeply-nested-indent.yaml"),
			expected: map[string]string{
				testutil.PermissionContents: testutil.PermissionRead,
				testutil.PermissionIssues:   testutil.PermissionWrite,
			},
			critical: true,
		},
		{
			name:     "minimal_valid_permission",
			yaml:     testutil.MustReadFixture(fixtureDir + "minimal-valid-permission.yaml"),
			expected: map[string]string{"x": "y"},
			critical: true,
		},
		{
			name: "maximum_realistic_permissions",
			yaml: testutil.MustReadFixture(fixtureDir + "maximum-realistic-permissions.yaml"),
			expected: map[string]string{
				testPermissionActions:                    testutil.PermissionWrite,
				appconstants.PermScopeAttestations:       testutil.PermissionWrite,
				appconstants.PermScopeChecks:             testutil.PermissionWrite,
				testutil.PermissionContents:              testutil.PermissionWrite,
				appconstants.PermScopeDeployments:        testutil.PermissionWrite,
				appconstants.PermScopeDiscussions:        testutil.PermissionWrite,
				appconstants.PermScopeIDToken:            testutil.PermissionWrite,
				testutil.PermissionIssues:                testutil.PermissionWrite,
				appconstants.PermScopePackages:           testutil.PermissionWrite,
				appconstants.PermScopePages:              testutil.PermissionWrite,
				appconstants.PermScopePullRequests:       testutil.PermissionWrite,
				appconstants.PermScopeRepositoryProjects: testutil.PermissionWrite,
				appconstants.PermScopeSecurityEvents:     testutil.PermissionWrite,
				appconstants.PermScopeStatuses:           testutil.PermissionWrite,
			},
			critical: false,
		},
	}
}

func TestPermissionParsingMutationResistance(t *testing.T) {
	tests := buildPermissionParsingTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPermissionParsingCase(t, tt.yaml, tt.expected)
		})
	}
}

func testPermissionParsingCase(t *testing.T, yaml string, expected map[string]string) {
	t.Helper()

	// Create temporary file with test YAML
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "action.yml")

	testutil.WriteTestFile(t, testFile, yaml)

	// Parse permissions
	result, err := parsePermissionsFromComments(testFile)
	if err != nil {
		t.Fatalf("parsePermissionsFromComments() error = %v", err)
	}

	// Verify expected permissions
	if len(result) != len(expected) {
		t.Errorf("got %d permissions, want %d", len(result), len(expected))
		t.Logf("got: %v", result)
		t.Logf("want: %v", expected)
	}

	for key, expectedValue := range expected {
		gotValue, exists := result[key]
		if !exists {
			t.Errorf(testutil.TestFixtureMissingPermKey, key)

			continue
		}
		if gotValue != expectedValue {
			t.Errorf("permission %q: got value %q, want %q", key, gotValue, expectedValue)
		}
	}

	// Check for unexpected keys
	for key := range result {
		if _, expected := expected[key]; !expected {
			t.Errorf("unexpected permission key %q with value %q", key, result[key])
		}
	}
}

// TestMergePermissionsMutationResistance tests the permission merging logic
// for mutations in nil checks, map operations, and precedence logic.
//

func TestMergePermissionsMutationResistance(t *testing.T) {
	tests := []struct {
		name         string
		yamlPerms    map[string]string
		commentPerms map[string]string
		expected     map[string]string
		critical     bool
		description  string
	}{
		{
			name:         "nil_yaml_nil_comment",
			yamlPerms:    nil,
			commentPerms: nil,
			expected:     nil,
			critical:     true,
			description:  "Both nil should stay nil (nil check critical)",
		},
		{
			name:         "nil_yaml_with_comment",
			yamlPerms:    nil,
			commentPerms: map[string]string{testutil.PermissionContents: testutil.PermissionRead},
			expected:     map[string]string{testutil.PermissionContents: testutil.PermissionRead},
			critical:     true,
			description:  "Nil YAML replaced by comment perms (first condition)",
		},
		{
			name:         "yaml_with_nil_comment",
			yamlPerms:    map[string]string{testutil.PermissionContents: testutil.PermissionWrite},
			commentPerms: nil,
			expected:     map[string]string{testutil.PermissionContents: testutil.PermissionWrite},
			critical:     true,
			description:  "Nil comment keeps YAML perms (second condition)",
		},
		{
			name:         "empty_yaml_empty_comment",
			yamlPerms:    map[string]string{},
			commentPerms: map[string]string{},
			expected:     map[string]string{},
			critical:     true,
			description:  "Both empty should stay empty",
		},
		{
			name:         "yaml_overrides_comment_same_key",
			yamlPerms:    map[string]string{testutil.PermissionContents: testutil.PermissionWrite},
			commentPerms: map[string]string{testutil.PermissionContents: testutil.PermissionRead},
			expected:     map[string]string{testutil.PermissionContents: testutil.PermissionWrite},
			critical:     true,
			description:  "YAML value wins conflict (exists check critical)",
		},
		{
			name:         "non_conflicting_keys_merged",
			yamlPerms:    map[string]string{testutil.PermissionContents: testutil.PermissionWrite},
			commentPerms: map[string]string{testutil.PermissionIssues: testutil.PermissionRead},
			expected: map[string]string{
				testutil.PermissionContents: testutil.PermissionWrite,
				testutil.PermissionIssues:   testutil.PermissionRead,
			},
			critical:    true,
			description: "Non-conflicting keys both included",
		},
		{
			name: "multiple_yaml_override_multiple_comment",
			yamlPerms: map[string]string{
				testutil.PermissionContents: testutil.PermissionWrite,
				testutil.PermissionIssues:   testutil.PermissionWrite,
			},
			commentPerms: map[string]string{
				testutil.PermissionContents:      testutil.PermissionRead,
				testutil.TestFixturePullRequests: testutil.PermissionRead,
			},
			expected: map[string]string{
				testutil.PermissionContents:      testutil.PermissionWrite, // YAML wins
				testutil.PermissionIssues:        testutil.PermissionWrite, // Only in YAML
				testutil.TestFixturePullRequests: testutil.PermissionRead,  // Only in comment
			},
			critical:    true,
			description: "Complex merge with conflicts and unique keys",
		},
		{
			name:         "single_key_conflict",
			yamlPerms:    map[string]string{"x": "a"},
			commentPerms: map[string]string{"x": "b"},
			expected:     map[string]string{"x": "a"},
			critical:     true,
			description:  "Minimal conflict test (YAML precedence)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testMergePermissionsCase(t, tt.yamlPerms, tt.commentPerms, tt.expected, tt.description)
		})
	}
}

func testMergePermissionsCase(
	t *testing.T,
	yamlPerms, commentPerms, expected map[string]string,
	description string,
) {
	t.Helper()

	// Create ActionYML with test permissions
	action := &ActionYML{
		Permissions: copyStringMap(yamlPerms),
	}

	// Copy commentPerms to avoid mutation during test
	commentPermsCopy := copyStringMap(commentPerms)

	// Perform merge
	mergePermissions(action, commentPermsCopy)

	// Verify result
	assertPermissionsMatch(t, action.Permissions, expected, description)
}

// copyStringMap creates a deep copy of a string map, returning nil for nil input.
func copyStringMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	result := make(map[string]string, len(input))
	for k, v := range input {
		result[k] = v
	}

	return result
}

// assertPermissionsMatch verifies that got permissions match expected permissions.
func assertPermissionsMatch(
	t *testing.T,
	got, want map[string]string,
	description string,
) {
	t.Helper()

	if want == nil {
		if got != nil {
			t.Errorf("expected nil permissions, got %v", got)
		}

		return
	}

	if got == nil {
		t.Errorf("expected non-nil permissions %v, got nil", want)

		return
	}

	if len(got) != len(want) {
		t.Errorf("got %d permissions, want %d", len(got), len(want))
		t.Logf("got: %v", got)
		t.Logf("want: %v", want)
	}

	for key, expectedValue := range want {
		gotValue, exists := got[key]
		if !exists {
			t.Errorf(testutil.TestFixtureMissingPermKey, key)

			continue
		}
		if gotValue != expectedValue {
			t.Errorf("permission %q: got %q, want %q (description: %s)",
				key, gotValue, expectedValue, description)
		}
	}

	for key := range got {
		if _, expected := want[key]; !expected {
			t.Errorf("unexpected permission key %q", key)
		}
	}
}

// permissionLineTestCase defines a test case for parsePermissionLine tests.
type permissionLineTestCase struct {
	name        string
	content     string
	expectKey   string
	expectValue string
	expectOk    bool
	critical    bool
	description string
}

// parseFailCase creates a test case expecting parse failure with empty results.
func parseFailCase(name, content, description string) permissionLineTestCase {
	return permissionLineTestCase{
		name:        name,
		content:     content,
		expectKey:   "",
		expectValue: "",
		expectOk:    false,
		critical:    true,
		description: description,
	}
}

// TestParsePermissionLineMutationResistance tests string manipulation boundaries
// in permission line parsing that are susceptible to mutation.
//

func TestParsePermissionLineMutationResistance(t *testing.T) {
	tests := []permissionLineTestCase{
		{
			name:        "basic_key_value",
			content:     testutil.TestFixtureContentsRead,
			expectKey:   testutil.PermissionContents,
			expectValue: testutil.PermissionRead,
			expectOk:    true,
			critical:    true,
			description: "Basic parsing",
		},
		{
			name:        "with_leading_dash",
			content:     "- contents: read",
			expectKey:   testutil.PermissionContents,
			expectValue: testutil.PermissionRead,
			expectOk:    true,
			critical:    true,
			description: "TrimPrefix(\"-\") critical",
		},
		{
			name:        "with_inline_comment_at_position_1",
			content:     "contents: r#comment",
			expectKey:   testutil.PermissionContents,
			expectValue: "r",
			expectOk:    true,
			critical:    true,
			description: "Index() > 0 boundary (idx=10)",
		},
		// Failure test cases with empty expected results
		parseFailCase(
			"inline_comment_at_position_0_of_value",
			"contents: #read",
			"Index() at position 0 in value (should fail parse)",
		),
		{
			name:        "comment_in_middle_of_line",
			content:     "contents: read  # Required",
			expectKey:   testutil.PermissionContents,
			expectValue: testutil.PermissionRead,
			expectOk:    true,
			critical:    true,
			description: "Comment removal before parse",
		},
		parseFailCase("no_colon", "contents read", "len(parts) == 2 check"),
		{
			name:        "multiple_colons",
			content:     "url: https://example.com:8080",
			expectKey:   "url",
			expectValue: "https://example.com:8080",
			expectOk:    true,
			critical:    true,
			description: "SplitN with n=2 preserves colons in value",
		},
		parseFailCase("empty_key", ": value", "key != \"\" check critical"),
		parseFailCase("empty_value", "key:", "value != \"\" check critical"),
		parseFailCase("whitespace_key", "  : value", "TrimSpace on key critical"),
		parseFailCase("whitespace_value", "key:   ", "TrimSpace on value critical"),
		{
			name:        "single_char_key_value",
			content:     "a: b",
			expectKey:   "a",
			expectValue: "b",
			expectOk:    true,
			critical:    true,
			description: "Minimal valid case",
		},
		{
			name:        "colon_in_key_should_not_happen",
			content:     "key:name: value",
			expectKey:   "key",
			expectValue: "name: value",
			expectOk:    true,
			critical:    false,
			description: "First colon splits (malformed input)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testParsePermissionLineCase(
				t,
				tt.content,
				tt.expectKey,
				tt.expectValue,
				tt.expectOk,
				tt.description,
			)
		})
	}
}

func testParsePermissionLineCase(
	t *testing.T,
	content, expectKey, expectValue string,
	expectOk bool,
	description string,
) {
	t.Helper()

	key, value, ok := parsePermissionLine(content)

	if ok != expectOk {
		t.Errorf("ok: got %v, want %v (description: %s)", ok, expectOk, description)
	}

	if ok {
		if key != expectKey {
			t.Errorf("key: got %q, want %q (description: %s)", key, expectKey, description)
		}
		if value != expectValue {
			t.Errorf("value: got %q, want %q (description: %s)", value, expectValue, description)
		}
	}
}

// TestProcessPermissionEntryMutationResistance tests indentation logic that is
// highly susceptible to off-by-one mutations.
//

func TestProcessPermissionEntryMutationResistance(t *testing.T) {
	tests := []struct {
		name              string
		line              string
		content           string
		initialExpected   int
		expectBreak       bool
		expectPermissions map[string]string
		critical          bool
		description       string
	}{
		{
			name:              "first_item_sets_indent",
			line:              "#   contents: read",
			content:           testutil.TestFixtureContentsRead,
			initialExpected:   -1,
			expectBreak:       false,
			expectPermissions: map[string]string{testutil.PermissionContents: testutil.PermissionRead},
			critical:          true,
			description:       "*expectedItemIndent == -1 check",
		},
		{
			name:              "same_indent_continues",
			line:              "#   issues: write",
			content:           testutil.TestFixtureIssuesWrite,
			initialExpected:   3,
			expectBreak:       false,
			expectPermissions: map[string]string{testutil.PermissionIssues: testutil.PermissionWrite},
			critical:          true,
			description:       "contentIndent == expectedItemIndent",
		},
		{
			name:              "dedent_by_one_breaks",
			line:              "#  issues: write",
			content:           testutil.TestFixtureIssuesWrite,
			initialExpected:   3,
			expectBreak:       true,
			expectPermissions: map[string]string{},
			critical:          true,
			description:       "contentIndent < expectedItemIndent (2 < 3)",
		},
		{
			name:              "dedent_by_two_breaks",
			line:              "# issues: write",
			content:           testutil.TestFixtureIssuesWrite,
			initialExpected:   3,
			expectBreak:       true,
			expectPermissions: map[string]string{},
			critical:          true,
			description:       "contentIndent < expectedItemIndent (0 < 3)",
		},
		{
			name:              "indent_more_continues",
			line:              "#     issues: write",
			content:           testutil.TestFixtureIssuesWrite,
			initialExpected:   3,
			expectBreak:       false,
			expectPermissions: map[string]string{testutil.PermissionIssues: testutil.PermissionWrite},
			critical:          false,
			description:       "More indent allowed (unusual but valid)",
		},
		{
			name:              "zero_indent_with_zero_expected",
			line:              "# contents: read",
			content:           testutil.TestFixtureContentsRead,
			initialExpected:   0,
			expectBreak:       false,
			expectPermissions: map[string]string{testutil.PermissionContents: testutil.PermissionRead},
			critical:          true,
			description:       "Boundary: 0 == 0",
		},
		{
			name:              "large_indent_value",
			line:              "#          contents: read",
			content:           testutil.TestFixtureContentsRead,
			initialExpected:   -1,
			expectBreak:       false,
			expectPermissions: map[string]string{testutil.PermissionContents: testutil.PermissionRead},
			critical:          false,
			description:       "Large indent value (10 spaces)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testProcessPermissionEntryCase(
				t,
				tt.line,
				tt.content,
				tt.initialExpected,
				tt.expectBreak,
				tt.expectPermissions,
				tt.description,
			)
		})
	}
}

func testProcessPermissionEntryCase(
	t *testing.T,
	line, content string,
	initialExpected int,
	expectBreak bool,
	expectPermissions map[string]string,
	description string,
) {
	t.Helper()

	permissions := make(map[string]string)
	expectedIndent := initialExpected

	shouldBreak := processPermissionEntry(line, content, &expectedIndent, permissions)

	if shouldBreak != expectBreak {
		t.Errorf("shouldBreak: got %v, want %v (description: %s)",
			shouldBreak, expectBreak, description)
	}

	if len(permissions) != len(expectPermissions) {
		t.Errorf("got %d permissions, want %d (description: %s)",
			len(permissions), len(expectPermissions), description)
	}

	for key, expectedValue := range expectPermissions {
		gotValue, exists := permissions[key]
		if !exists {
			t.Errorf(testutil.TestFixtureMissingPermKey, key)

			continue
		}
		if gotValue != expectedValue {
			t.Errorf("permission %q: got %q, want %q", key, gotValue, expectedValue)
		}
	}

	// Verify expected indent was set if it was -1
	if initialExpected == -1 && len(expectPermissions) > 0 {
		if expectedIndent == -1 {
			t.Error("expectedIndent should have been set from -1")
		}
	}
}
