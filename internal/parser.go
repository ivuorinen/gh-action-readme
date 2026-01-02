package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

// ActionYML models the action.yml metadata (fields are updateable as schema evolves).
type ActionYML struct {
	Name        string                  `yaml:"name"`
	Description string                  `yaml:"description"`
	Inputs      map[string]ActionInput  `yaml:"inputs"`
	Outputs     map[string]ActionOutput `yaml:"outputs"`
	Runs        map[string]any          `yaml:"runs"`
	Branding    *Branding               `yaml:"branding,omitempty"`
	// Add more fields as the schema evolves
}

// ActionInput represents an input parameter for a GitHub Action.
type ActionInput struct {
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
	Default     any    `yaml:"default"`
}

// ActionOutput represents an output parameter for a GitHub Action.
type ActionOutput struct {
	Description string `yaml:"description"`
}

// Branding represents the branding configuration for a GitHub Action.
type Branding struct {
	Icon  string `yaml:"icon"`
	Color string `yaml:"color"`
}

// ParseActionYML reads and parses action.yml from given path.
func ParseActionYML(path string) (*ActionYML, error) {
	f, err := os.Open(path) // #nosec G304 -- path from function parameter
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close() // Ignore close error in defer
	}()
	var a ActionYML
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&a); err != nil {
		return nil, err
	}

	return &a, nil
}

// shouldIgnoreDirectory checks if a directory name matches the ignore list.
func shouldIgnoreDirectory(dirName string, ignoredDirs []string) bool {
	for _, ignored := range ignoredDirs {
		if strings.HasPrefix(ignored, ".") {
			// Pattern match: ".git" matches ".git", ".github", etc.
			if strings.HasPrefix(dirName, ignored) {
				return true
			}
		} else {
			// Exact match for non-hidden dirs
			if dirName == ignored {
				return true
			}
		}
	}

	return false
}

// DiscoverActionFiles finds action.yml and action.yaml files in the given directory.
// This consolidates the file discovery logic from both generator.go and dependencies/parser.go.
//
//nolint:gocyclo // File discovery requires multiple checks
func DiscoverActionFiles(dir string, recursive bool, ignoredDirs []string) ([]string, error) {
	var actionFiles []string

	// Check if dir exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", dir)
	}

	if recursive {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				// Check if directory should be ignored
				if shouldIgnoreDirectory(info.Name(), ignoredDirs) {
					return filepath.SkipDir
				}

				return nil
			}

			// Check for action.yml or action.yaml files
			filename := strings.ToLower(info.Name())
			if filename == appconstants.ActionFileNameYML || filename == appconstants.ActionFileNameYAML {
				actionFiles = append(actionFiles, path)
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to walk directory %s: %w", dir, err)
		}
	} else {
		// Check only the specified directory
		for _, filename := range []string{appconstants.ActionFileNameYML, appconstants.ActionFileNameYAML} {
			path := filepath.Join(dir, filename)
			if _, err := os.Stat(path); err == nil {
				actionFiles = append(actionFiles, path)
			}
		}
	}

	return actionFiles, nil
}
