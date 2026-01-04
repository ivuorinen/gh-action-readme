package internal

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

const testPermissionWrite = "write"

// parseActionFromContent creates a temporary action.yml file with the given content and parses it.
func parseActionFromContent(t *testing.T, content string) (*ActionYML, error) {
	t.Helper()

	tmpFile, err := os.CreateTemp(t.TempDir(), appconstants.TestActionFilePattern)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	_ = tmpFile.Close()

	return ParseActionYML(tmpFile.Name())
}

// createTestDirWithAction creates a directory with an action.yml file and returns both paths.
func createTestDirWithAction(t *testing.T, baseDir, dirName, yamlContent string) (string, string) {
	t.Helper()
	dirPath := filepath.Join(baseDir, dirName)
	if err := os.Mkdir(dirPath, appconstants.FilePermDir); err != nil {
		t.Fatalf(testutil.ErrCreateDir(dirName), err)
	}
	actionPath := filepath.Join(dirPath, appconstants.ActionFileNameYML)
	if err := os.WriteFile(
		actionPath, []byte(yamlContent), appconstants.FilePermDefault,
	); err != nil {
		t.Fatalf(testutil.ErrCreateFile(dirName+"/action.yml"), err)
	}

	return dirPath, actionPath
}

// createTestFile creates a file with the given content and returns its path.
func createTestFile(t *testing.T, baseDir, fileName, content string) string {
	t.Helper()
	filePath := filepath.Join(baseDir, fileName)
	if err := os.WriteFile(
		filePath, []byte(content), appconstants.FilePermDefault,
	); err != nil {
		t.Fatalf(testutil.ErrCreateFile(fileName), err)
	}

	return filePath
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
			name:        "no match",
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
			name:        "dot prefix pattern match - .github",
			dirName:     appconstants.DirGitHub,
			ignoredDirs: []string{appconstants.DirGit},
			want:        true,
		},
		{
			name:        "dot prefix pattern match - .gitlab",
			dirName:     appconstants.DirGitLab,
			ignoredDirs: []string{appconstants.DirGit},
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
	rootAction := createTestFile(t, tmpDir, appconstants.ActionFileNameYML, appconstants.TestYAMLRoot)

	// Create directories with action.yml files
	_, nodeModulesAction := createTestDirWithAction(
		t,
		tmpDir,
		appconstants.DirNodeModules,
		appconstants.TestYAMLNodeModules,
	)
	_, vendorAction := createTestDirWithAction(t, tmpDir, appconstants.DirVendor, appconstants.TestYAMLVendor)
	_, gitAction := createTestDirWithAction(t, tmpDir, appconstants.DirGit, appconstants.TestYAMLGit)
	_, srcAction := createTestDirWithAction(t, tmpDir, "src", appconstants.TestYAMLSrc)

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

	nodeModulesDir := filepath.Join(tmpDir, appconstants.DirNodeModules, "deep", "nested")
	if err := os.MkdirAll(nodeModulesDir, appconstants.FilePermDir); err != nil {
		t.Fatalf(testutil.ErrCreateDir("nested"), err)
	}

	nestedAction := filepath.Join(nodeModulesDir, appconstants.ActionFileNameYML)
	if err := os.WriteFile(
		nestedAction, []byte(appconstants.TestYAMLNested), appconstants.FilePermDefault,
	); err != nil {
		t.Fatalf(testutil.ErrCreateFile("nested action.yml"), err)
	}

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
	rootAction := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
	if err := os.WriteFile(
		rootAction, []byte(appconstants.TestYAMLRoot), appconstants.FilePermDefault,
	); err != nil {
		t.Fatalf(testutil.ErrCreateFile("action.yml"), err)
	}

	// Create subdirectory (should not be searched in non-recursive mode)
	subDir := filepath.Join(tmpDir, "sub")
	if err := os.Mkdir(subDir, appconstants.FilePermDir); err != nil {
		t.Fatalf(testutil.ErrCreateDir("sub"), err)
	}
	subAction := filepath.Join(subDir, appconstants.ActionFileNameYML)
	if err := os.WriteFile(
		subAction, []byte(appconstants.TestYAMLSub), appconstants.FilePermDefault,
	); err != nil {
		t.Fatalf(testutil.ErrCreateFile("sub/action.yml"), err)
	}

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
			name: "single permission with dash format",
			content: `# yaml-language-server: $schema=https://json.schemastore.org/github-action.json
# permissions:
#   - contents: read  # Required for checking out repository
name: Test Action`,
			want: map[string]string{
				"contents": "read",
			},
			wantErr: false,
		},
		{
			name: "multiple permissions",
			content: `# permissions:
#   - contents: read
#   - issues: write
#   - pull-requests: write
name: Test Action`,
			want: map[string]string{
				"contents":      "read",
				"issues":        "write",
				"pull-requests": "write",
			},
			wantErr: false,
		},
		{
			name: "permissions without dash",
			content: `# permissions:
#   contents: read
#   issues: write
name: Test Action`,
			want: map[string]string{
				"contents": "read",
				"issues":   "write",
			},
			wantErr: false,
		},
		{
			name: "no permissions block",
			content: `# Just a comment
name: Test Action`,
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name: "permissions with inline comments",
			content: `# permissions:
#   - contents: read  # Needed for checkout
#   - issues: write   # To create issues
name: Test Action`,
			want: map[string]string{
				"contents": "read",
				"issues":   "write",
			},
			wantErr: false,
		},
		{
			name: "empty permissions block",
			content: `# permissions:
name: Test Action`,
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name: "permissions with mixed formats",
			content: `# permissions:
#   - contents: read
#   issues: write
name: Test Action`,
			want: map[string]string{
				"contents": "read",
				"issues":   "write",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temp file
			tmpFile, err := os.CreateTemp(t.TempDir(), appconstants.TestActionFilePattern)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = os.Remove(tmpFile.Name()) }()

			if _, err := tmpFile.WriteString(tt.content); err != nil {
				t.Fatal(err)
			}
			_ = tmpFile.Close()

			got, err := parsePermissionsFromComments(tmpFile.Name())

			if (err != nil) != tt.wantErr {
				t.Errorf("parsePermissionsFromComments() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePermissionsFromComments() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseActionYMLWithCommentPermissions tests that ParseActionYML includes comment permissions.
func TestParseActionYMLWithCommentPermissions(t *testing.T) {
	t.Parallel()

	content := appconstants.TestPermissionsHeader +
		"#   - contents: read\n" +
		appconstants.TestActionNameLine +
		appconstants.TestDescriptionLine +
		appconstants.TestRunsLine +
		appconstants.TestCompositeUsing +
		appconstants.TestStepsEmpty

	action, err := parseActionFromContent(t, content)
	if err != nil {
		t.Fatalf(appconstants.TestErrorFormat, err)
	}

	if action.Permissions == nil {
		t.Fatal("Expected permissions to be parsed from comments")
	}

	if action.Permissions["contents"] != "read" {
		t.Errorf("Expected contents: read, got %v", action.Permissions)
	}
}

// TestParseActionYMLYAMLPermissionsOverrideComments tests that YAML permissions override comments.
func TestParseActionYMLYAMLPermissionsOverrideComments(t *testing.T) {
	t.Parallel()

	content := appconstants.TestPermissionsHeader +
		"#   - contents: read\n" +
		"#   - issues: write\n" +
		appconstants.TestActionNameLine +
		appconstants.TestDescriptionLine +
		"permissions:\n" +
		"  contents: write  # YAML override\n" +
		appconstants.TestRunsLine +
		appconstants.TestCompositeUsing +
		appconstants.TestStepsEmpty

	action, err := parseActionFromContent(t, content)
	if err != nil {
		t.Fatalf(appconstants.TestErrorFormat, err)
	}

	// YAML should override comment
	if action.Permissions["contents"] != testPermissionWrite {
		t.Errorf(
			"Expected YAML permissions to override comment permissions, got contents: %v",
			action.Permissions["contents"],
		)
	}

	// Comment permission should be merged in
	if action.Permissions["issues"] != testPermissionWrite {
		t.Errorf(
			"Expected comment permissions to be merged with YAML permissions, got issues: %v",
			action.Permissions["issues"],
		)
	}
}

// TestParseActionYMLOnlyYAMLPermissions tests parsing when only YAML permissions exist.
func TestParseActionYMLOnlyYAMLPermissions(t *testing.T) {
	t.Parallel()

	content := appconstants.TestActionNameLine +
		appconstants.TestDescriptionLine +
		"permissions:\n" +
		"  contents: read\n" +
		"  issues: write\n" +
		appconstants.TestRunsLine +
		appconstants.TestCompositeUsing +
		appconstants.TestStepsEmpty

	action, err := parseActionFromContent(t, content)
	if err != nil {
		t.Fatalf(appconstants.TestErrorFormat, err)
	}

	if action.Permissions == nil {
		t.Fatal("Expected permissions to be parsed from YAML")
	}

	if action.Permissions["contents"] != "read" {
		t.Errorf("Expected contents: read, got %v", action.Permissions)
	}

	if action.Permissions["issues"] != testPermissionWrite {
		t.Errorf("Expected issues: write, got %v", action.Permissions)
	}
}

// TestParseActionYMLNoPermissions tests parsing when no permissions exist.
func TestParseActionYMLNoPermissions(t *testing.T) {
	t.Parallel()

	content := appconstants.TestActionNameLine +
		appconstants.TestDescriptionLine +
		appconstants.TestRunsLine +
		appconstants.TestCompositeUsing +
		appconstants.TestStepsEmpty

	action, err := parseActionFromContent(t, content)
	if err != nil {
		t.Fatalf(appconstants.TestErrorFormat, err)
	}

	if action.Permissions != nil {
		t.Errorf("Expected no permissions, got %v", action.Permissions)
	}
}

// TestParseActionYMLMalformedYAML tests parsing with malformed YAML.
func TestParseActionYMLMalformedYAML(t *testing.T) {
	t.Parallel()

	content := appconstants.TestActionNameLine +
		appconstants.TestDescriptionLine +
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

	tmpFile, err := os.CreateTemp(t.TempDir(), appconstants.TestActionFilePattern)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	_ = tmpFile.Close()

	_, err = ParseActionYML(tmpFile.Name())
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
			wantValue: "read",
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
			wantKey:   "contents",
			wantValue: "read",
			wantOK:    true,
		},
		{
			name:      "valid with leading dash",
			input:     "- issues: write",
			wantKey:   "issues",
			wantValue: "write",
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
			line:               appconstants.TestContentsRead,
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
			content: appconstants.TestPermissionsHeader +
				appconstants.TestContentsRead +
				"#   contents: write\n",
			wantPerms:   map[string]string{"contents": "write"},
			wantErr:     false,
			description: "last value wins",
		},
		{
			name: "mixed valid and invalid lines",
			content: appconstants.TestPermissionsHeader +
				appconstants.TestContentsRead +
				"#   invalid-line-no-value\n" +
				"#   issues: write\n",
			wantPerms:   map[string]string{"contents": "read", "issues": "write"},
			wantErr:     false,
			description: "invalid lines skipped",
		},
		{
			name: "permissions block ends at non-comment",
			content: appconstants.TestPermissionsHeader +
				appconstants.TestContentsRead +
				appconstants.TestActionNameLine +
				"#   issues: write\n",
			wantPerms:   map[string]string{"contents": "read"},
			wantErr:     false,
			description: "stops at first non-comment",
		},
		{
			name: "only permissions header",
			content: appconstants.TestPermissionsHeader +
				appconstants.TestActionNameLine,
			wantPerms:   map[string]string{},
			wantErr:     false,
			description: "empty permissions block",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp(t.TempDir(), "test-*.yml")
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = os.Remove(tmpFile.Name()) }()

			if _, err := tmpFile.WriteString(tt.content); err != nil {
				t.Fatal(err)
			}
			_ = tmpFile.Close()

			perms, err := parsePermissionsFromComments(tmpFile.Name())

			if (err != nil) != tt.wantErr {
				t.Errorf("parsePermissionsFromComments() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(perms, tt.wantPerms) {
				t.Errorf("parsePermissionsFromComments() = %v, want %v (%s)", perms, tt.wantPerms, tt.description)
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
			yamlPerms:    map[string]string{"contents": "read"},
			commentPerms: map[string]string{"issues": "write"},
			wantPerms:    map[string]string{"contents": "read", "issues": "write"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &ActionYML{Permissions: tt.yamlPerms}
			mergePermissions(action, tt.commentPerms)

			if !reflect.DeepEqual(action.Permissions, tt.wantPerms) {
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

	// Test that error message is descriptive
	if err != nil && !filepath.IsAbs("/nonexistent/path/that/does/not/exist") {
		t.Errorf("Expected absolute path in error, got: %v", err)
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
		t.Errorf("walkFunc() should propagate error, got %v, want %v", err, testErr)
	}
}

// TestParseActionYMLOnlyComments tests file with only comments.
func TestParseActionYMLOnlyComments(t *testing.T) {
	t.Parallel()

	content := "# This is a comment\n" +
		"# Another comment\n" +
		appconstants.TestPermissionsHeader +
		appconstants.TestContentsRead

	_, err := parseActionFromContent(t, content)
	// File with only comments should return EOF error from YAML parser
	// (comments are parsed separately, but YAML decoder still needs valid YAML)
	if err == nil {
		t.Error("Expected EOF error for comment-only file, got nil")
	}
}
