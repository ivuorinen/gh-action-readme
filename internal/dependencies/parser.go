package dependencies

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// parseCompositeActionFromFile reads and parses a composite action file.
func (a *Analyzer) parseCompositeActionFromFile(actionPath string) (*ActionWithComposite, error) {
	// Read the file
	data, err := os.ReadFile(actionPath)
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
	if action.Runs.Using != compositeUsing {
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

	return action.Runs.Using == compositeUsing, nil
}
