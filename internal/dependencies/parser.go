package dependencies

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

// validateFilePath ensures a file path is safe to read.
// Returns an error if the path contains traversal attempts.
func validateFilePath(path string) error {
	// Check the ORIGINAL path for ".." components rather than the cleaned path:
	// filepath.Clean collapses an absolute traversal such as
	// "/a/../../etc/passwd" into "/etc/passwd" (no ".." remaining), which would
	// otherwise slip past a cleaned-path check. Action file paths derived from
	// directory walks never legitimately contain "..".
	for _, component := range strings.Split(filepath.ToSlash(path), "/") {
		if component == ".." {
			return fmt.Errorf("invalid file path: traversal detected in %q", path)
		}
	}

	return nil
}

// parseCompositeActionFromFile reads and parses a composite action file.
func (a *Analyzer) parseCompositeActionFromFile(actionPath string) (*ActionWithComposite, error) {
	// Validate path before reading
	if err := validateFilePath(actionPath); err != nil {
		return nil, err
	}

	// Read the file
	data, err := os.ReadFile(actionPath) // #nosec G304 -- path validated above
	if err != nil {
		return nil, fmt.Errorf("failed to read action file %s: %w", actionPath, err)
	}

	// Parse YAML
	var action ActionWithComposite
	if err := yaml.Unmarshal(data, &action); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &action, nil
}

// parseCompositeAction parses an action.yml file with composite action support.
func (a *Analyzer) parseCompositeAction(actionPath string) (*ActionWithComposite, error) {
	// Use the real file parser
	action, err := a.parseCompositeActionFromFile(actionPath)
	if err != nil {
		return nil, err
	}

	// If this is not a composite action, return empty steps
	if action.Runs.Using != appconstants.ActionTypeComposite {
		action.Runs.Steps = []CompositeStep{}
	}

	return action, nil
}

// IsCompositeAction checks if an action file defines a composite action.
func IsCompositeAction(actionPath string) (bool, error) {
	action, err := (&Analyzer{}).parseCompositeActionFromFile(actionPath)
	if err != nil {
		return false, err
	}

	return action.Runs.Using == appconstants.ActionTypeComposite, nil
}
