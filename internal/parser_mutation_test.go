package internal

import (
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

// TestPermissionParsingMutationResistance provides comprehensive test cases designed
// to catch mutations in the permission parsing logic. These tests target critical
// boundaries, operators, and conditions that are susceptible to mutation.
//
//nolint:gocyclo // Comprehensive mutation test cases require multiple scenarios
func TestPermissionParsingMutationResistance(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected map[string]string
		critical bool // Critical mutations should fail tests
	}{
		{
			name: "off_by_one_indent_two_items",
			yaml: `# permissions:
#   contents: read
#   issues: write`,
			expected: map[string]string{"contents": "read", "issues": "write"},
			critical: true, // Indent comparison operators critical
		},
		{
			name: "off_by_one_indent_three_items",
			yaml: `# permissions:
#   contents: read
#   issues: write
#   pull-requests: read`,
			expected: map[string]string{
				"contents":                       "read",
				"issues":                         "write",
				testutil.TestFixturePullRequests: "read",
			},
			critical: true,
		},
		{
			name: "comment_position_at_boundary",
			yaml: `# permissions:
#   contents: read  # inline comment at position > 0`,
			expected: map[string]string{"contents": "read"},
			critical: true, // String.Index() boundary critical (> 0 vs >= 0)
		},
		{
			name: "comment_at_position_zero_parses",
			yaml: `# permissions:
#contents: read`,
			expected: map[string]string{"contents": "read"},
			critical: true, // Zero indent is valid (sets expectedItemIndent to 0)
		},
		{
			name: "dash_prefix_with_spaces",
			yaml: `# permissions:
#   - contents: read
#   - issues: write`,
			expected: map[string]string{"contents": "read", "issues": "write"},
			critical: true, // Dash removal logic
		},
		{
			name: "mixed_dash_and_no_dash",
			yaml: `# permissions:
#   - contents: read
#   issues: write`,
			expected: map[string]string{"contents": "read", "issues": "write"},
			critical: true,
		},
		{
			name: "dedent_stops_parsing",
			yaml: `# permissions:
#   contents: read
# This line is dedented and should stop parsing
#   issues: write`,
			expected: map[string]string{"contents": "read"},
			critical: true, // Dedent detection critical (< vs <=)
		},
		{
			name: "empty_line_in_block_continues",
			yaml: `# permissions:
#   contents: read
#
#   issues: write`,
			expected: map[string]string{"contents": "read", "issues": "write"},
			critical: false, // Empty lines should be handled
		},
		{
			name: "non_comment_line_stops_parsing",
			yaml: `# permissions:
#   contents: read
name: Test Action
#   issues: write`,
			expected: map[string]string{"contents": "read"},
			critical: true, // HasPrefix("#") check critical
		},
		{
			name: "exact_expected_indent",
			yaml: `# permissions:
#   contents: read`,
			expected: map[string]string{"contents": "read"},
			critical: true, // First item sets expected indent
		},
		{
			name: "colon_in_value_preserved",
			yaml: `# permissions:
#   contents: read:write`,
			expected: map[string]string{"contents": "read:write"},
			critical: true, // SplitN(content, ":", 2) critical
		},
		{
			name: "empty_key_not_parsed",
			yaml: `# permissions:
#   : read`,
			expected: map[string]string{},
			critical: true, // key != "" check critical
		},
		{
			name: "empty_value_not_parsed",
			yaml: `# permissions:
#   contents:`,
			expected: map[string]string{},
			critical: true, // value != "" check critical
		},
		{
			name: "whitespace_only_value_not_parsed",
			yaml: `# permissions:
#   contents:   `,
			expected: map[string]string{},
			critical: true, // TrimSpace on value critical
		},
		{
			name: "multiple_colons_splits_at_first",
			yaml: `# permissions:
#   url: https://example.com:8080`,
			expected: map[string]string{"url": "https://example.com:8080"},
			critical: true, // SplitN behavior with n=2
		},
		{
			name: "inline_comment_removal",
			yaml: `# permissions:
#   contents: read  # Required for checkout`,
			expected: map[string]string{"contents": "read"},
			critical: true, // Index(content, "#") > 0 boundary
		},
		{
			name: "inline_comment_at_start_of_value",
			yaml: `# permissions:
#   contents: #read`,
			expected: map[string]string{},
			critical: true, // idx > 0 (not >= 0) prevents this
		},
		{
			name: "deeply_nested_indent",
			yaml: `# permissions:
#     contents: read
#     issues: write`,
			expected: map[string]string{"contents": "read", "issues": "write"},
			critical: true, // Indent calculation with more spaces
		},
		{
			name: "minimal_valid_permission",
			yaml: `# permissions:
#   x: y`,
			expected: map[string]string{"x": "y"},
			critical: true, // Minimal case for key/value
		},
		{
			name: "maximum_realistic_permissions",
			yaml: `# permissions:
#   actions: write
#   attestations: write
#   checks: write
#   contents: write
#   deployments: write
#   discussions: write
#   id-token: write
#   issues: write
#   packages: write
#   pages: write
#   pull-requests: write
#   repository-projects: write
#   security-events: write
#   statuses: write`,
			expected: map[string]string{
				"actions":                        "write",
				"attestations":                   "write",
				"checks":                         "write",
				"contents":                       "write",
				"deployments":                    "write",
				"discussions":                    "write",
				"id-token":                       "write",
				"issues":                         "write",
				"packages":                       "write",
				"pages":                          "write",
				testutil.TestFixturePullRequests: "write",
				"repository-projects":            "write",
				"security-events":                "write",
				"statuses":                       "write",
			},
			critical: false, // Stress test for map operations
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPermissionParsingCase(t, tt.yaml, tt.expected)
		})
	}
}

func testPermissionParsingCase(t *testing.T, yaml string, expected map[string]string) {
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
//nolint:gocyclo // Table-driven test with many test cases
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
			commentPerms: map[string]string{"contents": "read"},
			expected:     map[string]string{"contents": "read"},
			critical:     true,
			description:  "Nil YAML replaced by comment perms (first condition)",
		},
		{
			name:         "yaml_with_nil_comment",
			yamlPerms:    map[string]string{"contents": "write"},
			commentPerms: nil,
			expected:     map[string]string{"contents": "write"},
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
			yamlPerms:    map[string]string{"contents": "write"},
			commentPerms: map[string]string{"contents": "read"},
			expected:     map[string]string{"contents": "write"},
			critical:     true,
			description:  "YAML value wins conflict (exists check critical)",
		},
		{
			name:         "non_conflicting_keys_merged",
			yamlPerms:    map[string]string{"contents": "write"},
			commentPerms: map[string]string{"issues": "read"},
			expected:     map[string]string{"contents": "write", "issues": "read"},
			critical:     true,
			description:  "Non-conflicting keys both included",
		},
		{
			name: "multiple_yaml_override_multiple_comment",
			yamlPerms: map[string]string{
				"contents": "write",
				"issues":   "write",
			},
			commentPerms: map[string]string{
				"contents":                       "read",
				testutil.TestFixturePullRequests: "read",
			},
			expected: map[string]string{
				"contents":                       "write", // YAML wins
				"issues":                         "write", // Only in YAML
				testutil.TestFixturePullRequests: "read",  // Only in comment
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

func testMergePermissionsCase(t *testing.T, yamlPerms, commentPerms, expected map[string]string, description string) {
	// Create ActionYML with test permissions
	action := &ActionYML{}

	// Copy yamlPerms to avoid test pollution
	if yamlPerms != nil {
		action.Permissions = make(map[string]string)
		for k, v := range yamlPerms {
			action.Permissions[k] = v
		}
	}

	// Copy commentPerms to avoid mutation during test
	var commentPermsCopy map[string]string
	if commentPerms != nil {
		commentPermsCopy = make(map[string]string)
		for k, v := range commentPerms {
			commentPermsCopy[k] = v
		}
	}

	// Perform merge
	mergePermissions(action, commentPermsCopy)

	// Verify result
	if expected == nil {
		if action.Permissions != nil {
			t.Errorf("expected nil permissions, got %v", action.Permissions)
		}

		return
	}

	if action.Permissions == nil {
		t.Errorf("expected non-nil permissions %v, got nil", expected)

		return
	}

	if len(action.Permissions) != len(expected) {
		t.Errorf("got %d permissions, want %d", len(action.Permissions), len(expected))
		t.Logf("got: %v", action.Permissions)
		t.Logf("want: %v", expected)
	}

	for key, expectedValue := range expected {
		gotValue, exists := action.Permissions[key]
		if !exists {
			t.Errorf(testutil.TestFixtureMissingPermKey, key)

			continue
		}
		if gotValue != expectedValue {
			t.Errorf("permission %q: got %q, want %q (description: %s)",
				key, gotValue, expectedValue, description)
		}
	}

	for key := range action.Permissions {
		if _, expected := expected[key]; !expected {
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
//nolint:gocyclo // Comprehensive mutation test cases require multiple scenarios
func TestParsePermissionLineMutationResistance(t *testing.T) {
	tests := []permissionLineTestCase{
		{
			name:        "basic_key_value",
			content:     testutil.TestFixtureContentsRead,
			expectKey:   "contents",
			expectValue: "read",
			expectOk:    true,
			critical:    true,
			description: "Basic parsing",
		},
		{
			name:        "with_leading_dash",
			content:     "- contents: read",
			expectKey:   "contents",
			expectValue: "read",
			expectOk:    true,
			critical:    true,
			description: "TrimPrefix(\"-\") critical",
		},
		{
			name:        "with_inline_comment_at_position_1",
			content:     "contents: r#comment",
			expectKey:   "contents",
			expectValue: "r",
			expectOk:    true,
			critical:    true,
			description: "Index() > 0 boundary (idx=10)",
		},
		// Failure test cases with empty expected results
		parseFailCase("inline_comment_at_position_0_of_value", "contents: #read", "Index() at position in value (should fail parse)"),
		{
			name:        "comment_in_middle_of_line",
			content:     "contents: read  # Required",
			expectKey:   "contents",
			expectValue: "read",
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
			testParsePermissionLineCase(t, tt.content, tt.expectKey, tt.expectValue, tt.expectOk, tt.description)
		})
	}
}

func testParsePermissionLineCase(t *testing.T, content, expectKey, expectValue string, expectOk bool, description string) {
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
//nolint:gocyclo // Comprehensive mutation test cases require multiple scenarios
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
			expectPermissions: map[string]string{"contents": "read"},
			critical:          true,
			description:       "*expectedItemIndent == -1 check",
		},
		{
			name:              "same_indent_continues",
			line:              "#   issues: write",
			content:           testutil.TestFixtureIssuesWrite,
			initialExpected:   3,
			expectBreak:       false,
			expectPermissions: map[string]string{"issues": "write"},
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
			expectPermissions: map[string]string{"issues": "write"},
			critical:          false,
			description:       "More indent allowed (unusual but valid)",
		},
		{
			name:              "zero_indent_with_zero_expected",
			line:              "# contents: read",
			content:           testutil.TestFixtureContentsRead,
			initialExpected:   0,
			expectBreak:       false,
			expectPermissions: map[string]string{"contents": "read"},
			critical:          true,
			description:       "Boundary: 0 == 0",
		},
		{
			name:              "large_indent_value",
			line:              "#          contents: read",
			content:           testutil.TestFixtureContentsRead,
			initialExpected:   -1,
			expectBreak:       false,
			expectPermissions: map[string]string{"contents": "read"},
			critical:          false,
			description:       "Large indent value (10 spaces)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testProcessPermissionEntryCase(t, tt.line, tt.content, tt.initialExpected, tt.expectBreak, tt.expectPermissions, tt.description)
		})
	}
}

func testProcessPermissionEntryCase(t *testing.T, line, content string, initialExpected int, expectBreak bool, expectPermissions map[string]string, description string) {
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
