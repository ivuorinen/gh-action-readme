package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

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

// TestDiscoverActionFiles_WithIgnoredDirectories tests file discovery with directory filtering.
// nolint:gocyclo // Test functions can be complex
func TestDiscoverActionFiles_WithIgnoredDirectories(t *testing.T) {
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
	rootAction := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
	if err := os.WriteFile(
		rootAction, []byte(appconstants.TestYAMLRoot), appconstants.FilePermDefault,
	); err != nil { // nolint:gosec
		t.Fatalf(testutil.ErrCreateFile("root action.yml"), err)
	}

	// Create node_modules with action.yml
	nodeModulesDir := filepath.Join(tmpDir, appconstants.DirNodeModules)
	if err := os.Mkdir(nodeModulesDir, appconstants.FilePermDir); err != nil { // nolint:gosec
		t.Fatalf(testutil.ErrCreateDir("node_modules"), err)
	}
	nodeModulesAction := filepath.Join(nodeModulesDir, appconstants.ActionFileNameYML)
	if err := os.WriteFile(
		nodeModulesAction, []byte(appconstants.TestYAMLNodeModules), appconstants.FilePermDefault,
	); err != nil { // nolint:gosec
		t.Fatalf(testutil.ErrCreateFile("node_modules/action.yml"), err)
	}

	// Create vendor with action.yml
	vendorDir := filepath.Join(tmpDir, appconstants.DirVendor)
	if err := os.Mkdir(vendorDir, appconstants.FilePermDir); err != nil { // nolint:gosec
		t.Fatalf(testutil.ErrCreateDir("vendor"), err)
	}
	vendorAction := filepath.Join(vendorDir, appconstants.ActionFileNameYML)
	if err := os.WriteFile(
		vendorAction, []byte(appconstants.TestYAMLVendor), appconstants.FilePermDefault,
	); err != nil { // nolint:gosec
		t.Fatalf(testutil.ErrCreateFile("vendor/action.yml"), err)
	}

	// Create .git with action.yml
	gitDir := filepath.Join(tmpDir, appconstants.DirGit)
	if err := os.Mkdir(gitDir, appconstants.FilePermDir); err != nil { // nolint:gosec
		t.Fatalf(testutil.ErrCreateDir(".git"), err)
	}
	gitAction := filepath.Join(gitDir, appconstants.ActionFileNameYML)
	if err := os.WriteFile(
		gitAction, []byte(appconstants.TestYAMLGit), appconstants.FilePermDefault,
	); err != nil { // nolint:gosec
		t.Fatalf(testutil.ErrCreateFile(".git/action.yml"), err)
	}

	// Create src with action.yml
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.Mkdir(srcDir, appconstants.FilePermDir); err != nil { // nolint:gosec
		t.Fatalf(testutil.ErrCreateDir("src"), err)
	}
	srcAction := filepath.Join(srcDir, appconstants.ActionFileNameYML)
	if err := os.WriteFile(
		srcAction, []byte(appconstants.TestYAMLSrc), appconstants.FilePermDefault,
	); err != nil { // nolint:gosec
		t.Fatalf(testutil.ErrCreateFile("src/action.yml"), err)
	}

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
				t.Fatalf("DiscoverActionFiles() error = %v", err)
			}

			if len(files) != tt.wantCount {
				t.Errorf("DiscoverActionFiles() returned %d files, want %d", len(files), tt.wantCount)
				t.Logf("Got files: %v", files)
				t.Logf("Want files: %v", tt.wantPaths)
			}

			// Check that all expected files are present
			fileMap := make(map[string]bool)
			for _, f := range files {
				fileMap[f] = true
			}

			for _, wantPath := range tt.wantPaths {
				if !fileMap[wantPath] {
					t.Errorf("Expected file %s not found in results", wantPath)
				}
			}
		})
	}
}

// TestDiscoverActionFiles_NestedIgnoredDirs tests that subdirectories of ignored dirs are skipped.
func TestDiscoverActionFiles_NestedIgnoredDirs(t *testing.T) {
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
		t.Fatalf("DiscoverActionFiles() error = %v", err)
	}

	if len(files) != 0 {
		t.Errorf("DiscoverActionFiles() returned %d files, want 0 (nested dirs should be skipped)", len(files))
		t.Logf("Got files: %v", files)
	}
}

// TestDiscoverActionFiles_NonRecursive tests that non-recursive mode ignores the filter.
func TestDiscoverActionFiles_NonRecursive(t *testing.T) {
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
		t.Fatalf("DiscoverActionFiles() error = %v", err)
	}

	if len(files) != 1 {
		t.Errorf("DiscoverActionFiles() non-recursive returned %d files, want 1", len(files))
	}

	if len(files) > 0 && files[0] != rootAction {
		t.Errorf("DiscoverActionFiles() = %v, want %v", files[0], rootAction)
	}
}
