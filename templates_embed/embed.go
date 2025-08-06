// Package templates_embed provides embedded template filesystem functionality for gh-action-readme.
// This package contains all template files embedded in the binary using Go's embed directive,
// making templates available regardless of working directory or filesystem location.
//
//nolint:revive // Package name with underscore is intentional for clarity
package templates_embed

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// embeddedTemplates contains all template files embedded in the binary
//
//go:embed templates
var embeddedTemplates embed.FS

// GetEmbeddedTemplate reads a template from the embedded filesystem.
func GetEmbeddedTemplate(templatePath string) ([]byte, error) {
	// Normalize path separators and remove leading slash if present
	cleanPath := strings.TrimPrefix(filepath.ToSlash(templatePath), "/")

	// If path doesn't start with templates/, prepend it
	if !strings.HasPrefix(cleanPath, "templates/") {
		cleanPath = "templates/" + cleanPath
	}

	return embeddedTemplates.ReadFile(cleanPath)
}

// GetEmbeddedTemplateFS returns the embedded filesystem for templates.
func GetEmbeddedTemplateFS() fs.FS {
	return embeddedTemplates
}

// IsEmbeddedTemplateAvailable checks if a template exists in the embedded filesystem.
func IsEmbeddedTemplateAvailable(templatePath string) bool {
	cleanPath := strings.TrimPrefix(filepath.ToSlash(templatePath), "/")
	if !strings.HasPrefix(cleanPath, "templates/") {
		cleanPath = "templates/" + cleanPath
	}

	_, err := embeddedTemplates.ReadFile(cleanPath)

	return err == nil
}

// ReadTemplate reads a template from embedded filesystem first, then falls back to filesystem.
func ReadTemplate(templatePath string) ([]byte, error) {
	// If it's an absolute path, read from filesystem with path validation
	if filepath.IsAbs(templatePath) {
		// Validate the path is clean to prevent path traversal attacks
		cleanPath := filepath.Clean(templatePath)
		if cleanPath != templatePath {
			return nil, filepath.ErrBadPattern
		}

		return os.ReadFile(cleanPath) // #nosec G304 -- validated absolute path
	}

	// Try embedded template first
	if IsEmbeddedTemplateAvailable(templatePath) {
		return GetEmbeddedTemplate(templatePath)
	}

	// Fallback to filesystem with path validation
	// Validate the path is clean to prevent path traversal attacks
	cleanPath := filepath.Clean(templatePath)
	if cleanPath != templatePath || strings.Contains(cleanPath, "..") {
		return nil, filepath.ErrBadPattern
	}

	return os.ReadFile(cleanPath) // #nosec G304 -- validated relative path
}
