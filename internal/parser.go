package internal

import (
	"bufio"
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
	Permissions map[string]string       `yaml:"permissions,omitempty"`
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
	// Parse permissions from header comments FIRST
	commentPermissions, err := parsePermissionsFromComments(path)
	if err != nil {
		// Don't fail if comment parsing fails, just log and continue
		commentPermissions = nil
	}

	// Standard YAML parsing
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

	// Merge permissions: YAML permissions override comment permissions
	mergePermissions(&a, commentPermissions)

	return &a, nil
}

// mergePermissions combines comment and YAML permissions.
// YAML permissions take precedence when both exist.
func mergePermissions(action *ActionYML, commentPerms map[string]string) {
	if len(commentPerms) == 0 {
		return
	}
	if action.Permissions == nil {
		action.Permissions = commentPerms

		return
	}
	for key, value := range commentPerms {
		if _, exists := action.Permissions[key]; !exists {
			action.Permissions[key] = value
		}
	}
}

// parsePermissionsFromComments extracts permissions from header comments.
// Looks for lines like:
//
//	# permissions:
//	#   - contents: read  # Required for checking out repository
//	#   contents: read    # Alternative format without dash
func parsePermissionsFromComments(path string) (map[string]string, error) {
	file, err := os.Open(path) // #nosec G304 -- path from function parameter
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close() // Ignore close error in defer
	}()

	permissions := make(map[string]string)
	scanner := bufio.NewScanner(file)

	inPermissionsBlock := false
	var expectedItemIndent int

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Stop parsing at first non-comment line
		if !strings.HasPrefix(trimmed, "#") {
			break
		}

		// Remove leading # and spaces
		content := strings.TrimPrefix(trimmed, "#")
		content = strings.TrimSpace(content)

		// Check for permissions block start
		if content == "permissions:" {
			inPermissionsBlock = true
			// Calculate expected indent for permission items (after the # and any spaces)
			// We expect items to be indented relative to the content
			expectedItemIndent = -1 // Will be set on first item

			continue
		}

		// Parse permission entries
		if inPermissionsBlock && content != "" {
			shouldBreak := processPermissionEntry(line, content, &expectedItemIndent, permissions)
			if shouldBreak {
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

// parsePermissionLine extracts key-value from a permission line.
// Supports formats:
//   - "- contents: read  # comment"
//   - "contents: read  # comment"
func parsePermissionLine(content string) (key, value string, ok bool) {
	// Remove leading dash if present
	content = strings.TrimPrefix(content, "-")
	content = strings.TrimSpace(content)

	// Remove inline comments
	if idx := strings.Index(content, "#"); idx > 0 {
		content = strings.TrimSpace(content[:idx])
	}

	// Parse key: value
	parts := strings.SplitN(content, ":", 2)
	if len(parts) == 2 {
		key = strings.TrimSpace(parts[0])
		value = strings.TrimSpace(parts[1])
		if key != "" && value != "" {
			return key, value, true
		}
	}

	return "", "", false
}

// processPermissionEntry processes a single line in the permissions block.
// Returns true if parsing should break (dedented out of block), false to continue.
func processPermissionEntry(line, content string, expectedItemIndent *int, permissions map[string]string) bool {
	// Get the indent of the content (after removing #). Strip both spaces and
	// tabs so a tab-indented comment item is recognized as indentation rather
	// than content (which would otherwise truncate the permissions block early).
	lineAfterHash := strings.TrimPrefix(line, "#")
	contentIndent := len(lineAfterHash) - len(strings.TrimLeft(lineAfterHash, " \t"))

	// Set expected indent on first item
	if *expectedItemIndent == -1 {
		*expectedItemIndent = contentIndent
	}

	// If dedented relative to expected item indent, we've left the permissions block
	if contentIndent < *expectedItemIndent {
		return true
	}

	// Parse permission line and add to map if valid
	if key, value, ok := parsePermissionLine(content); ok {
		permissions[key] = value
	}

	return false
}

// shouldIgnoreDirectory checks if a directory name exactly matches the ignore list.
func shouldIgnoreDirectory(dirName string, ignoredDirs []string) bool {
	for _, ignored := range ignoredDirs {
		if dirName == ignored {
			return true
		}
	}

	return false
}

// actionFileWalker encapsulates the logic for walking directories and finding action files.
type actionFileWalker struct {
	ignoredDirs []string
	actionFiles []string
}

// walkFunc is the callback function for filepath.Walk.
func (w *actionFileWalker) walkFunc(path string, info os.FileInfo, err error) error {
	if err != nil {
		if os.IsPermission(err) {
			if info != nil && info.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		return err
	}

	if info.IsDir() {
		if shouldIgnoreDirectory(info.Name(), w.ignoredDirs) {
			return filepath.SkipDir
		}

		return nil
	}

	// Check for action.yml or action.yaml files
	filename := strings.ToLower(info.Name())
	if filename == appconstants.ActionFileNameYML || filename == appconstants.ActionFileNameYAML {
		w.actionFiles = append(w.actionFiles, path)
	}

	return nil
}

// DiscoverActionFiles finds action.yml and action.yaml files in the given directory.
// This consolidates the file discovery logic from both generator.go and dependencies/parser.go.
func DiscoverActionFiles(dir string, recursive bool, ignoredDirs []string) ([]string, error) {
	// Check if dir exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", dir)
	}

	if recursive {
		walker := &actionFileWalker{ignoredDirs: ignoredDirs}
		if err := filepath.Walk(dir, walker.walkFunc); err != nil {
			return nil, fmt.Errorf("failed to walk directory %s: %w", dir, err)
		}

		return walker.actionFiles, nil
	}

	// Check only the specified directory (non-recursive)
	return DiscoverActionFilesNonRecursive(dir), nil
}

// DiscoverActionFilesNonRecursive finds action files (action.yml or action.yaml) in a single directory.
// This is exported for use by other packages that need to discover action files.
func DiscoverActionFilesNonRecursive(dir string) []string {
	var actionFiles []string
	for _, filename := range []string{appconstants.ActionFileNameYML, appconstants.ActionFileNameYAML} {
		path := filepath.Join(dir, filename)
		if _, err := os.Stat(path); err == nil {
			actionFiles = append(actionFiles, path)
		}
	}

	return actionFiles
}
