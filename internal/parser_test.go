package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// createTestDirWithAction creates a directory with an action.yml file and returns both paths.
func createTestDirWithAction(t *testing.T, baseDir, dirName, yamlContent string) (string, string) {
	t.Helper()
	dirPath := filepath.Join(baseDir, dirName)
	if err := os.Mkdir(dirPath, appconstants.FilePermDir); err != nil { // nolint:gosec
		t.Fatalf(testutil.ErrCreateDir(dirName), err)
	}
	actionPath := filepath.Join(dirPath, appconstants.ActionFileNameYML)
	if err := os.WriteFile(
		actionPath, []byte(yamlContent), appconstants.FilePermDefault,
	); err != nil { // nolint:gosec
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
	); err != nil { // nolint:gosec
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
	if err := os.MkdirAll(nodeModulesDir, appconstants.FilePermDir); err != nil { // nolint:gosec
		t.Fatalf(testutil.ErrCreateDir("nested"), err)
	}

	nestedAction := filepath.Join(nodeModulesDir, appconstants.ActionFileNameYML)
	if err := os.WriteFile(
		nestedAction, []byte(appconstants.TestYAMLNested), appconstants.FilePermDefault,
	); err != nil { // nolint:gosec
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
	); err != nil { // nolint:gosec
		t.Fatalf(testutil.ErrCreateFile("action.yml"), err)
	}

	// Create subdirectory (should not be searched in non-recursive mode)
	subDir := filepath.Join(tmpDir, "sub")
	if err := os.Mkdir(subDir, appconstants.FilePermDir); err != nil { // nolint:gosec
		t.Fatalf(testutil.ErrCreateDir("sub"), err)
	}
	subAction := filepath.Join(subDir, appconstants.ActionFileNameYML)
	if err := os.WriteFile(
		subAction, []byte(appconstants.TestYAMLSub), appconstants.FilePermDefault,
	); err != nil { // nolint:gosec
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
