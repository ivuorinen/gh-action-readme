// Package internal provides core logic for parsing, validating, and rendering
// documentation for GitHub Actions.
package internal

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// ActionYML represents a parsed GitHub Action's action.yml file.
// Fields correspond to the official action.yml schema.
type ActionYML struct {
	Name        string                  `yaml:"name"`
	Description string                  `yaml:"description"`
	Inputs      map[string]ActionInput  `yaml:"inputs"`
	Outputs     map[string]ActionOutput `yaml:"outputs"`
	Runs        map[string]any          `yaml:"runs"`
	Branding    *Branding               `yaml:"branding,omitempty"`
}

// ActionInput represents a single input parameter for a GitHub Action.
type ActionInput struct {
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
	Default     any    `yaml:"default"`
}

// ActionOutput represents a single output parameter for a GitHub Action.
type ActionOutput struct {
	Description string `yaml:"description"`
}

// Branding represents the branding section of a GitHub Action.
type Branding struct {
	Icon  string `yaml:"icon"`
	Color string `yaml:"color"`
}

// ParseActionYML parses the action.yml file at the given path and returns a pointer to ActionYML.
func ParseActionYML(path string) (*ActionYML, error) {
	cleanPath := filepath.Clean(path)
	f, err := os.Open(cleanPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		cerr := f.Close()
		if cerr != nil {
			logrus.Error("Failed to close file:", cleanPath)
		}
	}()
	var a ActionYML
	dec := yaml.NewDecoder(f)
	if decErr := dec.Decode(&a); decErr != nil {
		return nil, decErr
	}

	return &a, nil
}
