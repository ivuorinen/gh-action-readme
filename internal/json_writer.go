package internal

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/validation"
)

// buildVersion is the CLI version reported in generated JSON metadata. main sets
// it from its ldflags-injected `version` via SetVersion at startup; a func value
// like the old getVersion could not be set by `-X`, so JSON output was stuck at a
// hardcoded placeholder regardless of the release version.
var buildVersion = "dev"

// SetVersion records the CLI build version for inclusion in generated JSON
// metadata. Called once from main with the ldflags-injected version.
func SetVersion(v string) {
	if v != "" {
		buildVersion = v
	}
}

// getVersion returns the recorded build version.
var getVersion = func() string {
	return buildVersion
}

// shieldsBadgeEncode escapes a value for a shields.io badge URL segment. First it
// applies shields.io's separator escaping ("_" → "__", "-" → "--", " " → "_") so
// an action name like "Setup-Node" does not split the badge into the wrong
// label/message/color segments. Then it percent-encodes any remaining
// URL-reserved characters (e.g. "/", "?", "#", "%", "(", ")") so a legitimate
// name or color such as "C/C++ build" or a branding value with a slash does not
// corrupt the URL. url.PathEscape leaves the "-"/"_" produced above untouched
// (both are RFC 3986 unreserved).
func shieldsBadgeEncode(s string) string {
	s = strings.ReplaceAll(s, "_", "__")
	s = strings.ReplaceAll(s, "-", "--")
	s = strings.ReplaceAll(s, " ", "_")

	return url.PathEscape(s)
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
	Runs        map[string]any                 `json:"runs,omitempty"`
	Branding    *BrandingForJSON               `json:"branding,omitempty"`
	Permissions map[string]string              `json:"permissions,omitempty"`
}

// ActionInputForJSON represents an input parameter in JSON format.
type ActionInputForJSON struct {
	Description string   `json:"description"`
	Required    FlexBool `json:"required"`
	Default     any      `json:"default,omitempty"`
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

	// Write to file, tightening mode on rewrites of an existing file.
	return writeFileTightMode(outputPath, data, appconstants.FilePermDefault)
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
			URL: "https://img.shields.io/badge/icon-" + shieldsBadgeEncode(branding.Icon) +
				"-" + shieldsBadgeEncode(branding.Color),
			Alt: branding.Icon,
		})
	}
	badges = append(badges,
		BadgeInfo{
			Name: appconstants.LabelGitHubAction,
			URL:  "https://img.shields.io/badge/GitHub%20Action-" + shieldsBadgeEncode(action.Name) + "-blue",
			Alt:  appconstants.LabelGitHubAction,
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
			Permissions: action.Permissions,
		},
		Documentation: DocumentationInfo{
			Title:       action.Name,
			Description: action.Description,
			Usage:       jw.generateBasicExample(action),
			Badges:      badges,
			Sections:    sections,
			Links: map[string]string{
				appconstants.ActionFileNameYML: "./" + appconstants.ActionFileNameYML,
				"repository": "https://github.com/your-org/" + validation.SanitizeActionName(
					action.Name,
				),
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

// escapeYAMLDoubleQuoted escapes a value for embedding inside a double-quoted
// YAML scalar: backslash first, then double-quote, so an input default that
// contains `"` or `\` produces valid YAML the user can copy rather than a broken
// scalar like `key: "a"b"`.
func escapeYAMLDoubleQuoted(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)

	return s
}

// generateBasicExample creates a basic usage example.
func (jw *JSONWriter) generateBasicExample(action *ActionYML) string {
	example := "- name: " + action.Name + "\n"
	example += "  uses: your-org/" + validation.SanitizeActionName(action.Name) + "@v1"

	if len(action.Inputs) > 0 {
		example += "\n  with:"
		var exampleSb247 strings.Builder
		// Iterate inputs in a stable, sorted order so the generated example is
		// reproducible across runs (Go map iteration order is randomized).
		keys := make([]string, 0, len(action.Inputs))
		for key := range action.Inputs {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			input := action.Inputs[key]
			value := "value"
			if input.Default != nil {
				if str, ok := input.Default.(string); ok {
					value = str
				} else {
					value = fmt.Sprintf("%v", input.Default)
				}
			}
			exampleSb247.WriteString("\n    ")
			exampleSb247.WriteString(key)
			exampleSb247.WriteString(`: "`)
			exampleSb247.WriteString(escapeYAMLDoubleQuoted(value))
			exampleSb247.WriteString(`"`)
		}
		example += exampleSb247.String()
	}

	return example
}
