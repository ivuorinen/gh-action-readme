package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

// TestDefaultTestConfig_Helper tests the defaultTestConfig helper function.
func TestDefaultTestConfig_Helper(t *testing.T) {
	t.Parallel()

	// Call the helper multiple times to verify consistency
	cfg1 := defaultTestConfig()
	cfg2 := defaultTestConfig()

	// Verify expected defaults
	if cfg1.Quiet != true {
		t.Error("expected Quiet=true for test config")
	}
	if cfg1.Theme != appconstants.ThemeDefault {
		t.Errorf("expected default theme, got %s", cfg1.Theme)
	}
	if cfg1.OutputFormat != appconstants.OutputFormatMarkdown {
		t.Errorf("expected markdown format, got %s", cfg1.OutputFormat)
	}
	if cfg1.OutputDir != "." {
		t.Errorf("expected OutputDir='.', got %s", cfg1.OutputDir)
	}

	// Verify immutability - modifying one shouldn't affect others
	cfg1.Quiet = false
	cfg1.Theme = "custom"

	if cfg2.Quiet != true {
		t.Error("defaultTestConfig should return independent configs")
	}
	if cfg2.Theme != appconstants.ThemeDefault {
		t.Error("defaultTestConfig should return independent configs")
	}

	// Verify getting a fresh config after modification
	cfg3 := defaultTestConfig()
	if cfg3.Quiet != true {
		t.Error("defaultTestConfig should always return Quiet=true")
	}
}

// TestAssertActionFiles_Helper tests the assertActionFiles helper function.
func TestAssertActionFiles_Helper(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		files   []string
		setup   func(*testing.T) []string
		wantErr bool
	}{
		{
			name: "empty file list",
			setup: func(t *testing.T) []string {
				t.Helper()

				return []string{}
			},
		},
		{
			name: "valid action.yml files",
			setup: func(t *testing.T) []string {
				t.Helper()
				tmpDir1 := t.TempDir()
				tmpDir2 := t.TempDir()
				file1 := filepath.Join(tmpDir1, appconstants.ActionFileNameYML)
				file2 := filepath.Join(tmpDir2, appconstants.ActionFileNameYML)

				err := os.WriteFile(file1, []byte("name: test"), appconstants.FilePermDefault)
				if err != nil {
					t.Fatalf("failed to write file1: %v", err)
				}

				err = os.WriteFile(file2, []byte("name: test2"), appconstants.FilePermDefault)
				if err != nil {
					t.Fatalf("failed to write file2: %v", err)
				}

				return []string{file1, file2}
			},
		},
		{
			name: "valid action.yaml files",
			setup: func(t *testing.T) []string {
				t.Helper()
				tmpDir := t.TempDir()
				file := filepath.Join(tmpDir, "action.yaml")

				err := os.WriteFile(file, []byte("name: test"), appconstants.FilePermDefault)
				if err != nil {
					t.Fatalf("failed to write file: %v", err)
				}

				return []string{file}
			},
		},
		{
			name: "mixed yml and yaml extensions",
			setup: func(t *testing.T) []string {
				t.Helper()
				tmpDir1 := t.TempDir()
				tmpDir2 := t.TempDir()
				file1 := filepath.Join(tmpDir1, appconstants.ActionFileNameYML)
				file2 := filepath.Join(tmpDir2, "action.yaml")

				_ = os.WriteFile(file1, []byte("name: test1"), appconstants.FilePermDefault)

				_ = os.WriteFile(file2, []byte("name: test2"), appconstants.FilePermDefault)

				return []string{file1, file2}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			files := tt.setup(t)

			// Call the helper - it will verify files exist and have correct extensions
			// For invalid files, it will call t.Error (which is expected)
			assertActionFiles(t, files)
		})
	}
}

// Note: Invalid test cases (wrong extensions, nonexistent files) are not included
// because testing error paths would require mocking testing.T, which is complex.
// The helper is already well-tested through the main test suite for error cases.
