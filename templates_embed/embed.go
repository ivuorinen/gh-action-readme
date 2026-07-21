// Package templatesembed provides embedded template filesystem functionality for gh-action-readme.
// This package contains all template files embedded in the binary using Go's embed directive,
// making templates available regardless of working directory or filesystem location.
package templatesembed

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ivuorinen/gh-action-readme/appconstants"
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
	if !strings.HasPrefix(cleanPath, appconstants.DirTemplates) {
		cleanPath = appconstants.DirTemplates + cleanPath
	}

	if !fs.ValidPath(cleanPath) {
		return nil, filepath.ErrBadPattern
	}

	return embeddedTemplates.ReadFile(cleanPath)
}

// IsEmbeddedTemplateAvailable checks if a template exists in the embedded filesystem.
func IsEmbeddedTemplateAvailable(templatePath string) bool {
	cleanPath := strings.TrimPrefix(filepath.ToSlash(templatePath), "/")
	if !strings.HasPrefix(cleanPath, appconstants.DirTemplates) {
		cleanPath = appconstants.DirTemplates + cleanPath
	}

	if !fs.ValidPath(cleanPath) {
		return false
	}

	_, err := embeddedTemplates.ReadFile(cleanPath)

	return err == nil
}

// ReadTemplate reads a template from embedded filesystem first, then falls back to filesystem.
func ReadTemplate(templatePath string) ([]byte, error) {
	// If it's an absolute path, read from filesystem with path validation.
	if filepath.IsAbs(templatePath) {
		// This branch is reachable only from trusted input: an absolute template
		// path comes from the user's global config or the --theme flag, never from
		// untrusted repo/action config (those merge with allowTokens off, which
		// gates the template-path fields). We normalize and reject a non-clean
		// path; we deliberately do NOT constrain the location, because a legitimate
		// global theme lives outside the working directory
		// (e.g. ~/.config/gh-action-readme/...).
		cleanPath := filepath.Clean(templatePath)
		if cleanPath != templatePath {
			return nil, filepath.ErrBadPattern
		}

		return os.ReadFile(
			cleanPath,
		) // #nosec G304 -- trusted path (global config / --theme), normalized and clean-checked
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

	// Resolve symlinks and reject paths that escape the current working directory,
	// so a symlink inside the repo cannot point ReadTemplate at an arbitrary file.
	if err := ensureWithinCWD(cleanPath); err != nil {
		return nil, err
	}

	return os.ReadFile(cleanPath) // #nosec G304 -- validated relative path within CWD
}

// ensureWithinCWD verifies that path, after symlink resolution, stays inside the
// current working directory. Returns filepath.ErrBadPattern if it escapes.
func ensureWithinCWD(path string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return err
	}
	absResolved, err := filepath.Abs(resolved)
	if err != nil {
		return err
	}

	rel, err := filepath.Rel(cwd, absResolved)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return filepath.ErrBadPattern
	}

	return nil
}
