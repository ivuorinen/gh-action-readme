package appconstants

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const testModifiedValue = "modified"

// TestGetSupportedThemes tests the GetSupportedThemes function.
func TestGetSupportedThemes(t *testing.T) {
	t.Parallel()

	themes := GetSupportedThemes()

	// Check that we get a non-empty slice
	if len(themes) == 0 {
		t.Error("GetSupportedThemes() returned empty slice")
	}

	// Check that known themes are included
	expectedThemes := []string{ThemeDefault, ThemeGitHub, ThemeMinimal, ThemeProfessional}
	for _, expected := range expectedThemes {
		if !slices.Contains(themes, expected) {
			t.Errorf("GetSupportedThemes() missing expected theme: %s", expected)
		}
	}

	// Verify it returns a copy (modifying returned slice shouldn't affect original)
	themes1 := GetSupportedThemes()
	themes2 := GetSupportedThemes()
	if len(themes1) != len(themes2) {
		t.Error("GetSupportedThemes() not returning consistent results")
	}

	// Modify the returned slice
	if len(themes1) > 0 {
		themes1[0] = testModifiedValue
		// Get a fresh copy
		themes3 := GetSupportedThemes()
		// Should not be modified
		if themes3[0] == testModifiedValue {
			t.Error("GetSupportedThemes() not returning a copy - original was modified")
		}
	}
}

// TestGetConfigSearchPaths tests the GetConfigSearchPaths function.
func TestGetConfigSearchPaths(t *testing.T) {
	t.Parallel()

	paths := GetConfigSearchPaths()

	// Check that we get a non-empty slice
	if len(paths) == 0 {
		t.Error("GetConfigSearchPaths() returned empty slice")
	}

	// Check that it contains path-like strings
	for _, path := range paths {
		if path == "" {
			t.Error("GetConfigSearchPaths() contains empty string")
		}

		// Validate path doesn't contain traversal components
		if strings.Contains(path, "..") {
			t.Errorf("GetConfigSearchPaths() path %q contains unsafe .. component", path)
		}

		// Validate path is already cleaned
		cleanPath := filepath.Clean(path)
		if path != cleanPath {
			t.Errorf("GetConfigSearchPaths() path %q is not cleaned (should be %q)", path, cleanPath)
		}
	}

	// Verify it returns a copy (modifying returned slice shouldn't affect original)
	paths1 := GetConfigSearchPaths()
	paths2 := GetConfigSearchPaths()
	if len(paths1) != len(paths2) {
		t.Error("GetConfigSearchPaths() not returning consistent results")
	}

	// Modify the returned slice
	if len(paths1) > 0 {
		paths1[0] = testModifiedValue
		// Get a fresh copy
		paths3 := GetConfigSearchPaths()
		// Should not be modified
		if paths3[0] == testModifiedValue {
			t.Error("GetConfigSearchPaths() not returning a copy - original was modified")
		}
	}
}

// TestGetDefaultIgnoredDirectories tests the GetDefaultIgnoredDirectories function.
func TestGetDefaultIgnoredDirectories(t *testing.T) {
	t.Parallel()

	dirs := GetDefaultIgnoredDirectories()

	// Check that we get a non-empty slice
	if len(dirs) == 0 {
		t.Error("GetDefaultIgnoredDirectories() returned empty slice")
	}

	// Check that known ignored directories are included
	expectedDirs := []string{DirGit, DirNodeModules, DirVendor, DirDist}
	for _, expected := range expectedDirs {
		if !slices.Contains(dirs, expected) {
			t.Errorf("GetDefaultIgnoredDirectories() missing expected directory: %s", expected)
		}
	}

	// Verify it returns a copy (modifying returned slice shouldn't affect original)
	dirs1 := GetDefaultIgnoredDirectories()
	dirs2 := GetDefaultIgnoredDirectories()
	if len(dirs1) != len(dirs2) {
		t.Error("GetDefaultIgnoredDirectories() not returning consistent results")
	}

	// Modify the returned slice
	if len(dirs1) > 0 {
		dirs1[0] = testModifiedValue
		// Get a fresh copy
		dirs3 := GetDefaultIgnoredDirectories()
		// Should not be modified
		if dirs3[0] == testModifiedValue {
			t.Error("GetDefaultIgnoredDirectories() not returning a copy - original was modified")
		}
	}
}

// TestGetSupportedOutputFormats tests the GetSupportedOutputFormats function.
func TestGetSupportedOutputFormats(t *testing.T) {
	t.Parallel()

	formats := GetSupportedOutputFormats()

	if len(formats) == 0 {
		t.Error("GetSupportedOutputFormats() returned empty slice")
	}

	expectedFormats := []string{
		OutputFormatMarkdown, OutputFormatHTML, OutputFormatJSON, OutputFormatASCIIDoc,
	}
	for _, expected := range expectedFormats {
		if !slices.Contains(formats, expected) {
			t.Errorf("GetSupportedOutputFormats() missing expected format: %s", expected)
		}
	}
	// yaml/toml are config-export formats with no document generator; they must
	// NOT be advertised as gen output formats (see generateByFormat).
	for _, unsupported := range []string{OutputFormatYAML, OutputFormatTOML} {
		if slices.Contains(formats, unsupported) {
			t.Errorf("GetSupportedOutputFormats() must not list ungeneratable format: %s", unsupported)
		}
	}

	formats1 := GetSupportedOutputFormats()
	formats2 := GetSupportedOutputFormats()
	if len(formats1) != len(formats2) {
		t.Error("GetSupportedOutputFormats() not returning consistent results")
	}

	if len(formats1) > 0 {
		formats1[0] = testModifiedValue
		formats3 := GetSupportedOutputFormats()
		if formats3[0] == testModifiedValue {
			t.Error("GetSupportedOutputFormats() not returning a copy - original was modified")
		}
	}
}
