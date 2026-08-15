package internal

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

const (
	testPermissionActions = appconstants.PermScopeActions
)

// parseActionFromContent creates a temporary action.yml file with the given content and parses it.
func parseActionFromContent(t *testing.T, content string) (*ActionYML, error) {
	t.Helper()

	actionPath := testutil.CreateTempActionFile(t, content)

	return ParseActionYML(actionPath)
}

// TestDiscoverActionFilesIsCaseSensitive pins the discovery contract: GitHub only
// loads lowercase action.yml / action.yaml, so a mixed-case file must be found by
// NEITHER the recursive walker nor the non-recursive lookup. The two paths
// previously disagreed — the walker lowercased the name, the direct lookup did not.
func TestDiscoverActionFilesIsCaseSensitive(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"Action.YML", "ACTION.YAML", "Action.yaml"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			testutil.WriteActionFixtureAs(t, tmpDir, name, testutil.TestFixtureMinimalComposite)

			recursive, err := DiscoverActionFiles(tmpDir, true, nil)
			if err != nil {
				t.Fatalf("recursive discovery: %v", err)
			}
			if len(recursive) != 0 {
				t.Errorf("recursive discovery matched %q, want no matches: %v", name, recursive)
			}

			nonRecursive := DiscoverActionFilesNonRecursive(tmpDir)
			if len(nonRecursive) != 0 {
				t.Errorf("non-recursive discovery matched %q, want no matches: %v", name, nonRecursive)
			}
		})
	}
}

// validateDiscoveredFiles checks if discovered files match expected count and paths.
func validateDiscoveredFiles(t *testing.T, files []string, wantCount int, wantPaths []string) {
	t.Helper()

	if len(files) != wantCount {
		t.Errorf("DiscoverActionFiles() returned %d files, want %d", len(files), wantCount)
		t.Logf("Got files: %v", files)
		t.Logf("Want files: %v", wantPaths)
	}

	// Check that all expected files are present
	fileMap := make(map[string]bool)
	for _, f := range files {
		fileMap[f] = true
	}

	for _, wantPath := range wantPaths {
		if !fileMap[wantPath] {
			t.Errorf("Expected file %s not found in results", wantPath)
		}
	}
}

// TestShouldIgnoreDirectory tests the directory filtering logic.
func TestShouldIgnoreDirectory(t *testing.T) {
	tests := []struct {
		name        string
		dirName     string
		ignoredDirs []string
		want        bool
	}{
		{
			name:        "exact match - node_modules",
			dirName:     appconstants.DirNodeModules,
			ignoredDirs: []string{appconstants.DirNodeModules, appconstants.DirVendor},
			want:        true,
		},
		{
			name:        "exact match - vendor",
			dirName:     appconstants.DirVendor,
			ignoredDirs: []string{appconstants.DirNodeModules, appconstants.DirVendor},
			want:        true,
		},
		{
			name:        testutil.TestCaseNameNoMatch,
			dirName:     "src",
			ignoredDirs: []string{appconstants.DirNodeModules, appconstants.DirVendor},
			want:        false,
		},
		{
			name:        "empty ignore list",
			dirName:     appconstants.DirNodeModules,
			ignoredDirs: []string{},
			want:        false,
		},
		{
			name:        "dot prefix match - .git",
			dirName:     appconstants.DirGit,
			ignoredDirs: []string{appconstants.DirGit},
			want:        true,
		},
		{
			name:        ".github not ignored when only .git in list",
			dirName:     appconstants.DirGitHub,
			ignoredDirs: []string{appconstants.DirGit},
			want:        false,
		},
		{
			name:        ".gitlab not ignored when only .git in list",
			dirName:     appconstants.DirGitLab,
			ignoredDirs: []string{appconstants.DirGit},
			want:        false,
		},
		{
			name:        ".github ignored when .github is explicitly in list",
			dirName:     appconstants.DirGitHub,
			ignoredDirs: []string{appconstants.DirGitHub},
			want:        true,
		},
		{
			name:        "dot prefix no match",
			dirName:     ".config",
			ignoredDirs: []string{appconstants.DirGit},
			want:        false,
		},
		{
			name:        "case sensitive - NODE_MODULES vs node_modules",
			dirName:     "NODE_MODULES",
			ignoredDirs: []string{appconstants.DirNodeModules},
			want:        false,
		},
		{
			name:        "partial name not matched",
			dirName:     "my_vendor",
			ignoredDirs: []string{appconstants.DirVendor},
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldIgnoreDirectory(tt.dirName, tt.ignoredDirs)
			if got != tt.want {
				t.Errorf("shouldIgnoreDirectory(%q, %v) = %v, want %v",
					tt.dirName, tt.ignoredDirs, got, tt.want)
			}
		})
	}
}

// TestDiscoverActionFilesWithIgnoredDirectories tests file discovery with directory filtering.
func TestDiscoverActionFilesWithIgnoredDirectories(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()

	// Create directory structure:
	// tmpDir/
	//   action.yml (should be found)
	//   node_modules/
	//     action.yml (should be ignored)
	//   vendor/
	//     action.yml (should be ignored)
	//   .git/
	//     action.yml (should be ignored)
	//   src/
	//     action.yml (should be found)

	// Create root action.yml
	rootAction := testutil.WriteFileInDir(t, tmpDir, appconstants.ActionFileNameYML, testutil.TestYAMLRoot)

	// Create directories with action.yml files
	_, nodeModulesAction := testutil.CreateNestedAction(
		t,
		tmpDir,
		appconstants.DirNodeModules,
		testutil.TestYAMLNodeModules,
	)
	_, vendorAction := testutil.CreateNestedAction(t, tmpDir, appconstants.DirVendor, testutil.TestYAMLVendor)
	_, gitAction := testutil.CreateNestedAction(t, tmpDir, appconstants.DirGit, testutil.TestYAMLGit)
	_, srcAction := testutil.CreateNestedAction(t, tmpDir, "src", testutil.TestYAMLSrc)

	tests := []struct {
		name        string
		ignoredDirs []string
		wantCount   int
		wantPaths   []string
	}{
		{
			name:        "with default ignore list",
			ignoredDirs: []string{appconstants.DirGit, appconstants.DirNodeModules, appconstants.DirVendor},
			wantCount:   2,
			wantPaths:   []string{rootAction, srcAction},
		},
		{
			name:        "with empty ignore list",
			ignoredDirs: []string{},
			wantCount:   5,
			wantPaths:   []string{rootAction, gitAction, nodeModulesAction, srcAction, vendorAction},
		},
		{
			name:        "ignore only node_modules",
			ignoredDirs: []string{appconstants.DirNodeModules},
			wantCount:   4,
			wantPaths:   []string{rootAction, gitAction, srcAction, vendorAction},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := DiscoverActionFiles(tmpDir, true, tt.ignoredDirs)
			if err != nil {
				t.Fatalf(testutil.ErrDiscoverActionFiles(), err)
			}

			validateDiscoveredFiles(t, files, tt.wantCount, tt.wantPaths)
		})
	}
}

// TestDiscoverActionFilesNestedIgnoredDirs tests that subdirectories of ignored dirs are skipped.
func TestDiscoverActionFilesNestedIgnoredDirs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure:
	// tmpDir/
	//   node_modules/
	//     deep/
	//       nested/
	//         action.yml (should be ignored)

	nodeModulesDir := testutil.CreateTestSubdir(t, tmpDir, appconstants.DirNodeModules, "deep", "nested")

	testutil.WriteFileInDir(t, nodeModulesDir, appconstants.ActionFileNameYML, testutil.TestYAMLNested)

	files, err := DiscoverActionFiles(tmpDir, true, []string{appconstants.DirNodeModules})
	if err != nil {
		t.Fatalf(testutil.ErrDiscoverActionFiles(), err)
	}

	if len(files) != 0 {
		t.Errorf("DiscoverActionFiles() returned %d files, want 0 (nested dirs should be skipped)", len(files))
		t.Logf("Got files: %v", files)
	}
}

// TestDiscoverActionFilesNonRecursive tests that non-recursive mode ignores the filter.
func TestDiscoverActionFilesNonRecursive(t *testing.T) {
	tmpDir := t.TempDir()

	// Create action.yml in root
	rootAction := testutil.WriteFileInDir(t, tmpDir, appconstants.ActionFileNameYML, testutil.TestYAMLRoot)

	// Create subdirectory (should not be searched in non-recursive mode)
	subDir := filepath.Join(tmpDir, "sub")
	if err := os.Mkdir(subDir, appconstants.FilePermDir); err != nil {
		t.Fatalf(testutil.ErrCreateDir("sub"), err)
	}
	testutil.WriteFileInDir(t, subDir, appconstants.ActionFileNameYML, testutil.TestYAMLSub)

	files, err := DiscoverActionFiles(tmpDir, false, []string{})
	if err != nil {
		t.Fatalf(testutil.ErrDiscoverActionFiles(), err)
	}

	if len(files) != 1 {
		t.Errorf("DiscoverActionFiles() non-recursive returned %d files, want 1", len(files))
	}

	if len(files) > 0 && files[0] != rootAction {
		t.Errorf("DiscoverActionFiles() = %v, want %v", files[0], rootAction)
	}
}

// TestParsePermissionsFromComments tests parsing permissions from header comments.
func TestParsePermissionsFromComments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		want    map[string]string
		wantErr bool
	}{
		{
			name:    "single permission with dash format",
			content: string(testutil.MustReadFixture(testutil.TestFixturePermissionsDashSingle)),
			want: map[string]string{
				testutil.PermissionContents: testutil.PermissionRead,
			},
			wantErr: false,
		},
		{
			name:    "multiple permissions",
			content: string(testutil.MustReadFixture(testutil.TestFixturePermissionsDashMultiple)),
			want: map[string]string{
				testutil.PermissionContents:        testutil.PermissionRead,
				testutil.PermissionIssues:          testutil.PermissionWrite,
				appconstants.PermScopePullRequests: testutil.PermissionWrite,
			},
			wantErr: false,
		},
		{
			name:    "permissions without dash",
			content: string(testutil.MustReadFixture(testutil.TestFixturePermissionsObject)),
			want: map[string]string{
				testutil.PermissionContents: testutil.PermissionRead,
				testutil.PermissionIssues:   testutil.PermissionWrite,
			},
			wantErr: false,
		},
		{
			name:    "no permissions block",
			content: string(testutil.MustReadFixture(testutil.TestFixturePermissionsNone)),
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name:    "permissions with inline comments",
			content: string(testutil.MustReadFixture(testutil.TestFixturePermissionsInlineComments)),
			want: map[string]string{
				testutil.PermissionContents: testutil.PermissionRead,
				testutil.PermissionIssues:   testutil.PermissionWrite,
			},
			wantErr: false,
		},
		{
			name:    "empty permissions block",
			content: string(testutil.MustReadFixture(testutil.TestFixturePermissionsEmpty)),
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name:    "permissions with mixed formats",
			content: string(testutil.MustReadFixture(testutil.TestFixturePermissionsMixed)),
			want: map[string]string{
				testutil.PermissionContents: testutil.PermissionRead,
				testutil.PermissionIssues:   testutil.PermissionWrite,
			},
			wantErr: false,
		},
		{
			name:    "tab-indented permission comment block",
			content: string(testutil.MustReadFixture(testutil.TestFixturePermissionsTabIndented)),
			want: map[string]string{
				testutil.PermissionContents: testutil.PermissionRead,
				testutil.PermissionIssues:   testutil.PermissionWrite,
			},
			wantErr: false,
		},
		{
			name:    "two permission blocks split by a dedented prose comment",
			content: string(testutil.MustReadFixture(testutil.TestFixturePermissionsTwoBlocksWithProse)),
			want: map[string]string{
				testutil.PermissionContents: testutil.PermissionRead,
				testutil.PermissionIssues:   testutil.PermissionWrite,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actionPath := testutil.CreateTempActionFile(t, tt.content)
			hc, err := scanHeaderComments(actionPath)
			got := hc.Permissions

			if (err != nil) != tt.wantErr {
				t.Errorf("scanHeaderComments() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("scanHeaderComments().Permissions = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseActionYMLWithCommentPermissions tests that ParseActionYML includes comment permissions.
func TestParseActionYMLWithCommentPermissions(t *testing.T) {
	t.Parallel()

	content := testutil.TestPermissionsHeader +
		"#   - contents: read\n" +
		testutil.TestActionNameLine +
		testutil.TestDescriptionLine +
		testutil.TestRunsLine +
		testutil.TestCompositeUsing +
		testutil.TestStepsEmpty

	action, err := parseActionFromContent(t, content)
	if err != nil {
		t.Fatalf(testutil.TestErrorFormat, err)
	}

	if action.Permissions == nil {
		t.Fatal("Expected permissions to be parsed from comments")
	}

	if action.Permissions[testutil.PermissionContents] != "read" {
		t.Errorf("Expected contents: read, got %v", action.Permissions)
	}
}

// TestParseActionYMLYAMLPermissionsOverrideComments tests that YAML permissions override comments.
func TestParseActionYMLYAMLPermissionsOverrideComments(t *testing.T) {
	t.Parallel()

	content := testutil.TestPermissionsHeader +
		"#   - contents: read\n" +
		"#   - issues: write\n" +
		testutil.TestActionNameLine +
		testutil.TestDescriptionLine +
		"permissions:\n" +
		"  contents: write  # YAML override\n" +
		testutil.TestRunsLine +
		testutil.TestCompositeUsing +
		testutil.TestStepsEmpty

	action, err := parseActionFromContent(t, content)
	if err != nil {
		t.Fatalf(testutil.TestErrorFormat, err)
	}

	// YAML should override comment
	if action.Permissions[testutil.PermissionContents] != testutil.PermissionWrite {
		t.Errorf(
			"Expected YAML permissions to override comment permissions, got contents: %v",
			action.Permissions[testutil.PermissionContents],
		)
	}

	// Comment permission should be merged in
	if action.Permissions["issues"] != testutil.PermissionWrite {
		t.Errorf(
			"Expected comment permissions to be merged with YAML permissions, got issues: %v",
			action.Permissions["issues"],
		)
	}
}

// TestParseActionYMLOnlyYAMLPermissions tests parsing when only YAML permissions exist.
func TestParseActionYMLOnlyYAMLPermissions(t *testing.T) {
	t.Parallel()

	content := testutil.TestActionNameLine +
		testutil.TestDescriptionLine +
		"permissions:\n" +
		"  contents: read\n" +
		"  issues: write\n" +
		testutil.TestRunsLine +
		testutil.TestCompositeUsing +
		testutil.TestStepsEmpty

	action, err := parseActionFromContent(t, content)
	if err != nil {
		t.Fatalf(testutil.TestErrorFormat, err)
	}

	if action.Permissions == nil {
		t.Fatal("Expected permissions to be parsed from YAML")
	}

	if action.Permissions[testutil.PermissionContents] != "read" {
		t.Errorf("Expected contents: read, got %v", action.Permissions)
	}

	if action.Permissions["issues"] != testutil.PermissionWrite {
		t.Errorf("Expected issues: write, got %v", action.Permissions)
	}
}

// TestParseActionYMLNoPermissions tests parsing when no permissions exist.
func TestParseActionYMLNoPermissions(t *testing.T) {
	t.Parallel()

	content := testutil.TestActionNameLine +
		testutil.TestDescriptionLine +
		testutil.TestRunsLine +
		testutil.TestCompositeUsing +
		testutil.TestStepsEmpty

	action, err := parseActionFromContent(t, content)
	if err != nil {
		t.Fatalf(testutil.TestErrorFormat, err)
	}

	if action.Permissions != nil {
		t.Errorf("Expected no permissions, got %v", action.Permissions)
	}
}

// TestParseActionYMLScalarPermissions verifies N122: the scalar `permissions:
// read-all` shorthand parses into an `all` scope instead of failing the whole file
// with "string was used where mapping is expected".
func TestParseActionYMLScalarPermissions(t *testing.T) {
	t.Parallel()

	action, err := parseActionFromContent(
		t,
		string(testutil.MustReadFixture(testutil.TestFixturePermissionsScalarReadAll)),
	)
	if err != nil {
		t.Fatalf("ParseActionYML must accept the scalar permissions form, got: %v", err)
	}

	if action.Permissions["all"] != testutil.PermissionRead {
		t.Errorf("permissions: read-all should parse to all: read, got %v", action.Permissions)
	}
}

// TestParseActionYMLScalarNoneOverridesComment verifies that a scalar
// `permissions: none` is authoritative: a header-comment permission block must
// not leak past it (YAML wins over comments).
func TestParseActionYMLScalarNoneOverridesComment(t *testing.T) {
	t.Parallel()

	action, err := parseActionFromContent(
		t,
		string(testutil.MustReadFixture(testutil.TestFixturePermissionsScalarNoneWithComment)),
	)
	if err != nil {
		t.Fatalf(testutil.TestErrorFormat, err)
	}

	if len(action.Permissions) != 0 {
		t.Errorf("permissions: none must yield no permissions, but comment leaked: %#v", action.Permissions)
	}
}

// TestParseActionYMLScalarReadAllOverridesComment verifies that a scalar
// `permissions: read-all` is authoritative: comment-declared per-scope
// permissions must not be merged into the global scalar result.
func TestParseActionYMLScalarReadAllOverridesComment(t *testing.T) {
	t.Parallel()

	action, err := parseActionFromContent(
		t,
		string(testutil.MustReadFixture(testutil.TestFixturePermissionsScalarReadAllWithComment)),
	)
	if err != nil {
		t.Fatalf(testutil.TestErrorFormat, err)
	}

	if len(action.Permissions) != 1 || action.Permissions["all"] != testutil.PermissionRead {
		t.Errorf("permissions: read-all must yield only {all: read}, got %#v", action.Permissions)
	}
}

// TestParsePermissionsProseNotParsed verifies that a prose line with a colon
// inside the permissions comment block is not recorded as a permission.
func TestParsePermissionsProseNotParsed(t *testing.T) {
	t.Parallel()

	action, err := parseActionFromContent(
		t,
		string(testutil.MustReadFixture(testutil.TestFixturePermissionsProseInBlock)),
	)
	if err != nil {
		t.Fatalf(testutil.TestErrorFormat, err)
	}

	if _, ok := action.Permissions["Note"]; ok {
		t.Errorf("prose line was parsed as a permission: %#v", action.Permissions)
	}
	if action.Permissions[testutil.PermissionContents] != testutil.PermissionRead {
		t.Errorf("valid permission before the prose line should still parse: %#v", action.Permissions)
	}
}

// TestParseActionYMLFlexBoolRequired verifies N142: quoted and word-form boolean
// values for `required` parse instead of failing the whole file.
func TestParseActionYMLFlexBoolRequired(t *testing.T) {
	t.Parallel()

	action, err := parseActionFromContent(t, testutil.MustReadFixture(testutil.TestFixtureRequiredQuoted))
	if err != nil {
		t.Fatalf("ParseActionYML must accept quoted/word boolean required, got: %v", err)
	}

	if !bool(action.Inputs["token"].Required) {
		t.Error(`required: "true" should parse as true`)
	}
	if !bool(action.Inputs["flag"].Required) {
		t.Error("required: yes should parse as true")
	}
	if bool(action.Inputs["optional"].Required) {
		t.Error("required: false should parse as false")
	}
	if !bool(action.Inputs["num_true"].Required) {
		t.Error("required: 1 (numeric scalar) should parse as true")
	}
	if bool(action.Inputs["num_false"].Required) {
		t.Error("required: 0 (numeric scalar) should parse as false")
	}
}

// TestParsePermissionsFromCommentsWithBOM verifies N143: a UTF-8 BOM at the start
// of the file does not abort the comment-permission scan.
func TestParsePermissionsFromCommentsWithBOM(t *testing.T) {
	t.Parallel()

	// The fixture is stored BOM-free (a committed BOM fails editorconfig's
	// charset=utf-8 check); prepend the BOM here to exercise the parser's BOM path.
	const bom = "\ufeff"
	action, err := parseActionFromContent(t, bom+testutil.MustReadFixture(testutil.TestFixturePermissionsWithBOM))
	if err != nil {
		t.Fatalf(testutil.TestErrorFormat, err)
	}

	if action.Permissions[testutil.PermissionContents] != testutil.PermissionRead ||
		action.Permissions[testutil.PermissionIssues] != testutil.PermissionWrite {
		t.Errorf("BOM-prefixed permission comments not parsed: %#v", action.Permissions)
	}
}

// TestParseActionYMLMalformedYAML tests parsing with malformed YAML.
func TestParseActionYMLMalformedYAML(t *testing.T) {
	t.Parallel()

	content := testutil.TestActionNameLine +
		testutil.TestDescriptionLine +
		"invalid-yaml: [\n" + // Unclosed bracket
		"  - item"

	_, err := parseActionFromContent(t, content)
	if err == nil {
		t.Error("Expected error for malformed YAML, got nil")
	}
}

// TestParseActionYMLEmptyFile tests parsing an empty file.
func TestParseActionYMLEmptyFile(t *testing.T) {
	t.Parallel()

	actionPath := testutil.CreateTempActionFile(t, "")
	_, err := ParseActionYML(actionPath)
	// Empty file should return EOF error from YAML parser
	if err == nil {
		t.Error("Expected EOF error for empty file, got nil")
	}
}

// TestParsePermissionLineEdgeCases tests edge cases in permission line parsing.
func TestParsePermissionLineEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantKey   string
		wantValue string
		wantOK    bool
	}{
		{
			name:      "comment at start is parsed",
			input:     "#contents: read",
			wantKey:   "#contents",
			wantValue: testutil.PermissionRead,
			wantOK:    true,
		},
		{
			name:      "empty value after colon",
			input:     "contents:",
			wantKey:   "",
			wantValue: "",
			wantOK:    false,
		},
		{
			name:      "only spaces after colon",
			input:     "contents:   ",
			wantKey:   "",
			wantValue: "",
			wantOK:    false,
		},
		{
			name:      "valid with inline comment",
			input:     "contents: read # required",
			wantKey:   testutil.PermissionContents,
			wantValue: testutil.PermissionRead,
			wantOK:    true,
		},
		{
			name:      "valid with leading dash",
			input:     "- issues: write",
			wantKey:   "issues",
			wantValue: testutil.PermissionWrite,
			wantOK:    true,
		},
		{
			name:      "no colon",
			input:     "invalid permission line",
			wantKey:   "",
			wantValue: "",
			wantOK:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, value, ok := parsePermissionLine(tt.input)

			if ok != tt.wantOK {
				t.Errorf("parsePermissionLine() ok = %v, want %v", ok, tt.wantOK)
			}

			if key != tt.wantKey {
				t.Errorf("parsePermissionLine() key = %q, want %q", key, tt.wantKey)
			}

			if value != tt.wantValue {
				t.Errorf("parsePermissionLine() value = %q, want %q", value, tt.wantValue)
			}
		})
	}
}

// TestProcessPermissionEntryIndentationEdgeCases tests indentation scenarios.
func TestProcessPermissionEntryIndentationEdgeCases(t *testing.T) {
	tests := []struct {
		name               string
		line               string
		content            string
		initialIndent      int
		wantBreak          bool
		wantPermissionsLen int
	}{
		{
			name:               "first item sets indent",
			line:               testutil.TestContentsRead,
			content:            "contents: read",
			initialIndent:      -1,
			wantBreak:          false,
			wantPermissionsLen: 1,
		},
		{
			name:               "dedented breaks",
			line:               "# contents: read",
			content:            "contents: read",
			initialIndent:      2,
			wantBreak:          true,
			wantPermissionsLen: 0,
		},
		{
			name:               "same indent continues",
			line:               "#   issues: write",
			content:            "issues: write",
			initialIndent:      3,
			wantBreak:          false,
			wantPermissionsLen: 1,
		},
		{
			name:               "invalid format skipped",
			line:               "#   invalid-line-no-colon",
			content:            "invalid-line-no-colon",
			initialIndent:      3,
			wantBreak:          false,
			wantPermissionsLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permissions := make(map[string]string)
			indent := tt.initialIndent

			shouldBreak := processPermissionEntry(tt.line, tt.content, &indent, permissions)

			if shouldBreak != tt.wantBreak {
				t.Errorf("processPermissionEntry() shouldBreak = %v, want %v", shouldBreak, tt.wantBreak)
			}

			if len(permissions) != tt.wantPermissionsLen {
				t.Errorf(
					"processPermissionEntry() permissions length = %d, want %d",
					len(permissions),
					tt.wantPermissionsLen,
				)
			}
		})
	}
}

// TestParsePermissionsFromCommentsEdgeCases tests edge cases in comment parsing.
func TestParsePermissionsFromCommentsEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantPerms   map[string]string
		wantErr     bool
		description string
	}{
		{
			name: "duplicate permissions",
			content: testutil.TestPermissionsHeader +
				testutil.TestContentsRead +
				"#   contents: write\n",
			wantPerms:   map[string]string{testutil.PermissionContents: testutil.PermissionWrite},
			wantErr:     false,
			description: "last value wins",
		},
		{
			name: "mixed valid and invalid lines",
			content: testutil.TestPermissionsHeader +
				testutil.TestContentsRead +
				"#   invalid-line-no-value\n" +
				"#   issues: write\n",
			wantPerms: map[string]string{
				testutil.PermissionContents: testutil.PermissionRead,
				testutil.PermissionIssues:   testutil.PermissionWrite,
			},
			wantErr:     false,
			description: "invalid lines skipped",
		},
		{
			name: "permissions block ends at non-comment",
			content: testutil.TestPermissionsHeader +
				testutil.TestContentsRead +
				testutil.TestActionNameLine +
				"#   issues: write\n",
			wantPerms:   map[string]string{testutil.PermissionContents: testutil.PermissionRead},
			wantErr:     false,
			description: "stops at first non-comment",
		},
		{
			name: "only permissions header",
			content: testutil.TestPermissionsHeader +
				testutil.TestActionNameLine,
			wantPerms:   map[string]string{},
			wantErr:     false,
			description: "empty permissions block",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actionPath := testutil.CreateTempActionFile(t, tt.content)
			hc, err := scanHeaderComments(actionPath)
			perms := hc.Permissions

			if (err != nil) != tt.wantErr {
				t.Errorf("scanHeaderComments() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(perms, tt.wantPerms) {
				t.Errorf("scanHeaderComments().Permissions = %v, want %v (%s)", perms, tt.wantPerms, tt.description)
			}
		})
	}
}

// TestMergePermissionsEdgeCases tests permission merging edge cases.
func TestMergePermissionsEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		yamlPerms    map[string]string
		commentPerms map[string]string
		wantPerms    map[string]string
	}{
		{
			name:         "both nil",
			yamlPerms:    nil,
			commentPerms: nil,
			wantPerms:    nil,
		},
		{
			name:         "yaml nil, comments empty",
			yamlPerms:    nil,
			commentPerms: map[string]string{},
			wantPerms:    nil,
		},
		{
			name:         "yaml empty, comments nil",
			yamlPerms:    map[string]string{},
			commentPerms: nil,
			wantPerms:    map[string]string{},
		},
		{
			name:         "yaml has value, comments override",
			yamlPerms:    map[string]string{testutil.PermissionContents: testutil.PermissionRead},
			commentPerms: map[string]string{testutil.PermissionIssues: testutil.PermissionWrite},
			wantPerms: map[string]string{
				testutil.PermissionContents: testutil.PermissionRead,
				testutil.PermissionIssues:   testutil.PermissionWrite,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &ActionYML{Permissions: PermissionMap(tt.yamlPerms)}
			mergePermissions(action, tt.commentPerms)

			if !reflect.DeepEqual(action.Permissions, PermissionMap(tt.wantPerms)) {
				t.Errorf("mergePermissions() = %v, want %v", action.Permissions, tt.wantPerms)
			}
		})
	}
}

// TestDiscoverActionFilesWalkErrors tests error handling during directory walk.
func TestDiscoverActionFilesWalkErrors(t *testing.T) {
	// Test with a path that doesn't exist
	_, err := DiscoverActionFiles("/nonexistent/path/that/does/not/exist", true, []string{})
	if err == nil {
		t.Error("Expected error for nonexistent directory, got nil")
	}

	// Test that error message mentions the path
	if err != nil && !strings.Contains(err.Error(), "/nonexistent/path/that/does/not/exist") {
		t.Errorf("Expected error to mention path, got: %v", err)
	}
}

// TestWalkFuncErrorHandling tests walkFunc error propagation.
func TestWalkFuncErrorHandling(t *testing.T) {
	walker := &actionFileWalker{
		ignoredDirs: []string{},
		actionFiles: []string{},
	}

	// Create a valid FileInfo for testing
	tmpDir := t.TempDir()
	info, err := os.Stat(tmpDir)
	if err != nil {
		t.Fatalf("Failed to stat temp dir: %v", err)
	}

	// Test with valid directory - should return nil
	err = walker.walkFunc(tmpDir, info, nil)
	if err != nil {
		t.Errorf("walkFunc() with valid directory should return nil, got: %v", err)
	}

	// Test with pre-existing error - should propagate
	testErr := filepath.SkipDir
	err = walker.walkFunc(tmpDir, info, testErr)
	if err != testErr {
		t.Errorf("walkFunc() should propagate error, "+testutil.TestMsgGotWant, err, testErr)
	}
}

// TestDiscoverActionFiles_YAMLExtension kills the CONDITIONALS_NEGATION at parser.go:245
// by creating a directory with only "action.yaml" (not "action.yml") and asserting it is
// discovered. The mutation negates the second operand of the || condition, preventing
// "action.yaml" files from being collected.
func TestDiscoverActionFiles_YAMLExtension(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	yamlPath := filepath.Join(tmpDir, appconstants.ActionFileNameYAML)
	if err := os.WriteFile(yamlPath, []byte("name: Test\n"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	files, err := DiscoverActionFiles(tmpDir, true, nil)
	if err != nil {
		t.Fatalf("DiscoverActionFiles() error: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("DiscoverActionFiles() = %d files, want 1", len(files))

		return
	}

	if files[0] != yamlPath {
		t.Errorf("DiscoverActionFiles() found %q, want %q", files[0], yamlPath)
	}
}

// TestParseActionYMLOnlyComments tests file with only comments.
func TestParseActionYMLOnlyComments(t *testing.T) {
	t.Parallel()

	content := "# This is a comment\n" +
		"# Another comment\n" +
		testutil.TestPermissionsHeader +
		testutil.TestContentsRead

	_, err := parseActionFromContent(t, content)
	// File with only comments should return EOF error from YAML parser
	// (comments are parsed separately, but YAML decoder still needs valid YAML)
	if err == nil {
		t.Error("Expected EOF error for comment-only file, got nil")
	}
}
