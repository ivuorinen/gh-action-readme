// Package internal provides core logic for gh-action-readme, including config loading and parsing.
package internal

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Config holds configuration loaded from config.yaml.
type Config struct {
	DefaultValues    DefaultValues `yaml:"defaults"`
	GitHubOrg        string        `yaml:"github_org"`
	Template         string        `yaml:"template"`
	Header           string        `yaml:"header"`
	Footer           string        `yaml:"footer"`
	Schema           string        `yaml:"schema"`
	MainTemplatePath string        `yaml:"main_template_path"`
	HTMLTemplatePath string        `yaml:"html_template_path"`
}

// LoadConfig loads the configuration from the given YAML file path.
func LoadConfig(path string) (*Config, error) {
	cleanPath := filepath.Clean(path)
	f, openErr := os.Open(cleanPath)
	if openErr != nil {
		return nil, openErr
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			// Log the error but do not return it, as we already have the config loaded.
			// This is a best-effort close.
			logrus.Error(err.Error())
		}
	}(f)
	var cfg Config
	dec := yaml.NewDecoder(f)
	if decodeErr := dec.Decode(&cfg); decodeErr != nil {
		return nil, decodeErr
	}

	return &cfg, nil
}
