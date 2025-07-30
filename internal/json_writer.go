package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// getVersion returns the current version - can be overridden at build time.
var getVersion = func() string {
	return "0.1.0" // Default version, should be overridden at build time
}

// JSONOutput represents the structured JSON documentation output.
type JSONOutput struct {
	Meta          MetaInfo          `json:"meta"`
	Action        ActionYMLForJSON  `json:"action"`
	Documentation DocumentationInfo `json:"documentation"`
	Examples      []ExampleInfo     `json:"examples"`
	Generated     GeneratedInfo     `json:"generated"`
}

// MetaInfo contains metadata about the documentation generation.
type MetaInfo struct {
	Version   string `json:"version"`
	Format    string `json:"format"`
	Schema    string `json:"schema"`
	Generator string `json:"generator"`
}

// ActionYMLForJSON represents the action.yml data in JSON format.
type ActionYMLForJSON struct {
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	Inputs      map[string]ActionInputForJSON  `json:"inputs,omitempty"`
	Outputs     map[string]ActionOutputForJSON `json:"outputs,omitempty"`
	Runs        map[string]any                 `json:"runs"`
	Branding    *BrandingForJSON               `json:"branding,omitempty"`
}

// ActionInputForJSON represents an input parameter in JSON format.
type ActionInputForJSON struct {
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Default     any    `json:"default,omitempty"`
}

// ActionOutputForJSON represents an output parameter in JSON format.
type ActionOutputForJSON struct {
	Description string `json:"description"`
}

// BrandingForJSON represents branding information in JSON format.
type BrandingForJSON struct {
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

// DocumentationInfo contains information about the generated documentation.
type DocumentationInfo struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Usage       string            `json:"usage"`
	Badges      []BadgeInfo       `json:"badges,omitempty"`
	Sections    []SectionInfo     `json:"sections"`
	Links       map[string]string `json:"links"`
}

// BadgeInfo represents a documentation badge.
type BadgeInfo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Alt  string `json:"alt"`
}

// SectionInfo represents a documentation section.
type SectionInfo struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Type    string `json:"type"` // "inputs", "outputs", "examples", "text"
}

// ExampleInfo represents a usage example.
type ExampleInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Code        string `json:"code"`
	Language    string `json:"language"`
}

// GeneratedInfo contains metadata about when and how the documentation was generated.
type GeneratedInfo struct {
	Timestamp string `json:"timestamp"`
	Tool      string `json:"tool"`
	Version   string `json:"version"`
	Theme     string `json:"theme,omitempty"`
}

// JSONWriter handles JSON output generation.
type JSONWriter struct {
	Config *AppConfig
}

// NewJSONWriter creates a new JSON writer.
func NewJSONWriter(config *AppConfig) *JSONWriter {
	return &JSONWriter{Config: config}
}

// Write generates JSON documentation from the action data.
func (jw *JSONWriter) Write(action *ActionYML, outputPath string) error {
	jsonOutput := jw.convertToJSONOutput(action)

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(jsonOutput, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(outputPath, data, 0644)
}

// convertToJSONOutput converts ActionYML to structured JSON output.
func (jw *JSONWriter) convertToJSONOutput(action *ActionYML) *JSONOutput {
	// Convert inputs
	inputs := make(map[string]ActionInputForJSON)
	for key, input := range action.Inputs {
		inputs[key] = ActionInputForJSON(input)
	}

	// Convert outputs
	outputs := make(map[string]ActionOutputForJSON)
	for key, output := range action.Outputs {
		outputs[key] = ActionOutputForJSON(output)
	}

	// Convert branding
	var branding *BrandingForJSON
	if action.Branding != nil {
		branding = &BrandingForJSON{
			Icon:  action.Branding.Icon,
			Color: action.Branding.Color,
		}
	}

	// Generate badges
	var badges []BadgeInfo
	if branding != nil {
		badges = append(badges, BadgeInfo{
			Name: "Icon",
			URL:  "https://img.shields.io/badge/icon-" + branding.Icon + "-" + branding.Color,
			Alt:  branding.Icon,
		})
	}
	badges = append(badges,
		BadgeInfo{
			Name: "GitHub Action",
			URL:  "https://img.shields.io/badge/GitHub%20Action-" + action.Name + "-blue",
			Alt:  "GitHub Action",
		},
		BadgeInfo{
			Name: "License",
			URL:  "https://img.shields.io/badge/license-MIT-green",
			Alt:  "MIT License",
		},
	)

	// Generate examples
	examples := []ExampleInfo{
		{
			Title:       "Basic Usage",
			Description: "Basic example of using " + action.Name,
			Code:        jw.generateBasicExample(action),
			Language:    "yaml",
		},
	}

	// Build sections
	sections := []SectionInfo{
		{
			Title:   "Overview",
			Content: action.Description,
			Type:    "text",
		},
	}

	if len(action.Inputs) > 0 {
		sections = append(sections, SectionInfo{
			Title:   "Inputs",
			Content: "Input parameters for this action",
			Type:    "inputs",
		})
	}

	if len(action.Outputs) > 0 {
		sections = append(sections, SectionInfo{
			Title:   "Outputs",
			Content: "Output parameters from this action",
			Type:    "outputs",
		})
	}

	return &JSONOutput{
		Meta: MetaInfo{
			Version:   "1.0.0",
			Format:    "gh-action-readme-json",
			Schema:    "https://github.com/ivuorinen/gh-action-readme/schema/v1",
			Generator: "gh-action-readme",
		},
		Action: ActionYMLForJSON{
			Name:        action.Name,
			Description: action.Description,
			Inputs:      inputs,
			Outputs:     outputs,
			Runs:        action.Runs,
			Branding:    branding,
		},
		Documentation: DocumentationInfo{
			Title:       action.Name,
			Description: action.Description,
			Usage:       jw.generateBasicExample(action),
			Badges:      badges,
			Sections:    sections,
			Links: map[string]string{
				"action.yml": "./action.yml",
				"repository": "https://github.com/your-org/" + action.Name,
			},
		},
		Examples: examples,
		Generated: GeneratedInfo{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Tool:      "gh-action-readme",
			Version:   getVersion(),
			Theme:     jw.Config.Theme,
		},
	}
}

// generateBasicExample creates a basic usage example.
func (jw *JSONWriter) generateBasicExample(action *ActionYML) string {
	example := "- name: " + action.Name + "\n"
	example += "  uses: your-org/" + action.Name + "@v1"

	if len(action.Inputs) > 0 {
		example += "\n  with:"
		for key, input := range action.Inputs {
			value := "value"
			if input.Default != nil {
				if str, ok := input.Default.(string); ok {
					value = str
				} else {
					value = fmt.Sprintf("%v", input.Default)
				}
			}
			example += "\n    " + key + ": \"" + value + "\""
		}
	}

	return example
}
