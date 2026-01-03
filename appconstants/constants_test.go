package appconstants

import (
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
		found := false
		for _, theme := range themes {
			if theme == expected {
				found = true

				break
			}
		}
		if !found {
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
		found := false
		for _, dir := range dirs {
			if dir == expected {
				found = true

				break
			}
		}
		if !found {
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

// TestConfigurationSource_String tests the String method for ConfigurationSource.
func TestConfigurationSource_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source ConfigurationSource
		want   string
	}{
		{
			name:   "defaults source",
			source: SourceDefaults,
			want:   ConfigKeyDefaults,
		},
		{
			name:   "global source",
			source: SourceGlobal,
			want:   ScopeGlobal,
		},
		{
			name:   "repo override source",
			source: SourceRepoOverride,
			want:   "repo-override",
		},
		{
			name:   "repo config source",
			source: SourceRepoConfig,
			want:   "repo-config",
		},
		{
			name:   "action config source",
			source: SourceActionConfig,
			want:   "action-config",
		},
		{
			name:   "environment source",
			source: SourceEnvironment,
			want:   "environment",
		},
		{
			name:   "CLI flags source",
			source: SourceCLIFlags,
			want:   "cli-flags",
		},
		{
			name:   "unknown source",
			source: ConfigurationSource(999),
			want:   "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.source.String()
			if got != tt.want {
				t.Errorf("ConfigurationSource.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
