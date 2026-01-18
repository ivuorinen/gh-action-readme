package templatesembed

import (
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

// TestGetEmbeddedTemplate tests reading templates from embedded filesystem.
func TestGetEmbeddedTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		templatePath string
		expectError  bool
		description  string
	}{
		{
			name:         "valid default template",
			templatePath: testutil.TestTemplateReadme,
			expectError:  false,
			description:  "Should read default template successfully",
		},
		{
			name:         "valid template with templates/ prefix",
			templatePath: testutil.TestTemplateWithPrefix,
			expectError:  false,
			description:  "Should handle templates/ prefix correctly",
		},
		{
			name:         "valid GitHub theme",
			templatePath: testutil.TestTemplateGitHub,
			expectError:  false,
			description:  "Should read theme template successfully",
		},
		{
			name:         "valid template with leading slash",
			templatePath: "/readme.tmpl",
			expectError:  false,
			description:  "Should strip leading slash and read template",
		},
		{
			name:         "non-existent template",
			templatePath: "nonexistent.tmpl",
			expectError:  true,
			description:  "Should return error for missing template",
		},
		{
			name:         testutil.TestCaseNameEmptyPath,
			templatePath: "",
			expectError:  true,
			description:  "Should return error for empty path",
		},
		{
			name:         "path traversal attempt",
			templatePath: "../../../etc/passwd",
			expectError:  true,
			description:  "Should reject path traversal",
		},
		{
			name:         "Windows-style path",
			templatePath: "themes\\github\\readme.tmpl",
			expectError:  true,
			description:  "Windows paths won't work directly in embedded FS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			content, err := GetEmbeddedTemplate(tt.templatePath)

			assertTemplateLoaded(t, content, err, tt.expectError, 1)
		})
	}
}

// TestGetEmbeddedTemplateFS verifies the filesystem is accessible.
func TestGetEmbeddedTemplateFS(t *testing.T) {
	t.Parallel()

	fs := GetEmbeddedTemplateFS()
	if fs == nil {
		t.Fatal("GetEmbeddedTemplateFS() returned nil")
	}

	// Verify we can read from the filesystem
	file, err := fs.Open(testutil.TestTemplateWithPrefix)
	if err != nil {
		t.Errorf("failed to open default template: %v", err)
	}
	if file != nil {
		_ = file.Close()
	}
}

// TestIsEmbeddedTemplateAvailable tests template existence checking.
func TestIsEmbeddedTemplateAvailable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		templatePath string
		expectExists bool
	}{
		{
			name:         "default template exists",
			templatePath: testutil.TestTemplateReadme,
			expectExists: true,
		},
		{
			name:         "GitHub theme exists",
			templatePath: testutil.TestTemplateGitHub,
			expectExists: true,
		},
		{
			name:         "non-existent template",
			templatePath: "nonexistent.tmpl",
			expectExists: false,
		},
		{
			name:         testutil.TestCaseNameEmptyPath,
			templatePath: "",
			expectExists: false,
		},
		{
			name:         "path with leading slash",
			templatePath: "/readme.tmpl",
			expectExists: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			exists := IsEmbeddedTemplateAvailable(tt.templatePath)
			if exists != tt.expectExists {
				t.Errorf("IsEmbeddedTemplateAvailable(%q) = %v, want %v",
					tt.templatePath, exists, tt.expectExists)
			}
		})
	}
}

// TestReadTemplate tests the fallback logic (embedded → filesystem).
func TestReadTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		templatePath string
		expectError  bool
		description  string
	}{
		{
			name:         "read embedded template",
			templatePath: testutil.TestTemplateReadme,
			expectError:  false,
			description:  "Should read from embedded filesystem",
		},
		{
			name:         "absolute path - valid",
			templatePath: "/tmp/test-template.tmpl",
			expectError:  true, // Will fail unless file exists
			description:  "Should attempt filesystem read for absolute path",
		},
		{
			name:         "path traversal protection - relative",
			templatePath: "../../../etc/passwd",
			expectError:  true,
			description:  "Should reject path traversal in relative paths",
		},
		{
			name:         "path traversal protection - with dots",
			templatePath: "templates/../../../etc/passwd",
			expectError:  true,
			description:  "Should detect unclean paths",
		},
		{
			name:         "non-existent embedded template",
			templatePath: "missing.tmpl",
			expectError:  true,
			description:  "Should fail when template doesn't exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			content, err := ReadTemplate(tt.templatePath)

			assertTemplateLoaded(t, content, err, tt.expectError, 1)
		})
	}
}

// TestReadTemplate_PathValidation tests security aspects of path handling.
func TestReadTemplatePathValidation(t *testing.T) {
	t.Parallel()

	securityTests := []struct {
		name        string
		path        string
		description string
	}{
		{
			name:        "double dot traversal",
			path:        "../templates/readme.tmpl",
			description: "Should reject paths with ..",
		},
		{
			name:        "null byte injection",
			path:        "readme.tmpl\x00.evil",
			description: "Should reject null bytes",
		},
		{
			name:        "absolute traversal",
			path:        "/nonexistent/absolute/path/file.txt",
			description: "Should validate absolute paths and fail for non-existent",
		},
	}

	for _, tt := range securityTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ReadTemplate(tt.path)
			// All these should fail (either not found or security rejection)
			if err == nil {
				t.Errorf("security test failed: %s should have been rejected", tt.description)
			}
		})
	}
}
