// Package internal provides core logic for parsing, validating, and schema-checking GitHub Actions.
package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ghodssyaml "github.com/ghodss/yaml"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"

	"github.com/ivuorinen/gh-action-readme/schemas"
)

// ValidationResult holds the result of validating an ActionYML struct.
// MissingFields: required fields not present
// Warnings: non-critical issues
// SchemaErrors: errors from JSON schema validation
type ValidationResult struct {
	MissingFields []string
	Warnings      []string
	SchemaErrors  []string
}

// ValidateActionYML checks for missing required fields in the ActionYML struct.
// Returns a ValidationResult with missing fields and warnings.
func ValidateActionYML(action *ActionYML) ValidationResult {
	result := ValidationResult{}
	if action.Name == "" {
		result.MissingFields = append(result.MissingFields, "name")
	}
	if action.Description == "" {
		result.MissingFields = append(result.MissingFields, "description")
	}
	if len(action.Runs) == 0 {
		result.MissingFields = append(result.MissingFields, "runs")
	}

	return result
}

// ValidateActionYMLSchema validates the action.yml file at the given path against the provided
// schema file. Returns a slice of error messages if validation fails, or nil if valid.
// Uses gojsonschema for validation.
func ValidateActionYMLSchema(actionYMLPath, schemaPath string) ([]string, error) {
	var schemaBytes []byte
	var err error
	if schemaPath != "" {
		cleanPath := filepath.Clean(schemaPath)
		schemaBytes, err = os.ReadFile(cleanPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read schema: %w", err)
		}
	} else {
		schemaBytes = schemas.ActionSchema
	}

	schemaLoader := gojsonschema.NewBytesLoader(schemaBytes)

	// Convert YAML to JSON for validation
	cleanActionYMLPath := filepath.Clean(actionYMLPath)
	yamlFile, err := os.ReadFile(cleanActionYMLPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read action.yml: %w", err)
	}
	jsonBytes, err := yamlToJSON(yamlFile)
	if err != nil {
		return nil, fmt.Errorf("failed to convert YAML to JSON: %w", err)
	}
	docLoader := gojsonschema.NewBytesLoader(jsonBytes)

	result, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		return nil, fmt.Errorf("schema validation error: %w", err)
	}
	errs := make([]string, 0, len(result.Errors()))
	for _, e := range result.Errors() {
		errs = append(errs, e.String())
	}

	// Strict mode: check for unknown fields in the YAML that are not allowed by the schema
	allowedFields, afErr := getAllowedTopLevelFieldsFromSchema(schemaBytes)
	if afErr != nil {
		return nil, fmt.Errorf("failed to parse schema for allowed fields: %w", afErr)
	}
	unknownFields, ufErr := getUnknownTopLevelFields(yamlFile, allowedFields)
	if ufErr != nil {
		return nil, fmt.Errorf("failed to check unknown fields: %w", ufErr)
	}
	if len(unknownFields) > 0 {
		unknownFieldsStr := strings.Join(unknownFields, ", ")
		errs = append(
			errs,
			"Unknown top-level fields: "+unknownFieldsStr,
		)
	}

	if len(errs) == 0 {
		return nil, nil
	}

	return errs, nil
}

// yamlToJSON converts YAML bytes to JSON bytes for schema validation using ghodss/yaml.
func yamlToJSON(yamlBytes []byte) ([]byte, error) {
	return ghodssyaml.YAMLToJSON(yamlBytes)
}

// getAllowedTopLevelFieldsFromSchema parses the schema file and returns the allowed
// top-level fields.
func getAllowedTopLevelFieldsFromSchema(schemaBytes []byte) ([]string, error) {
	var schema struct {
		Properties map[string]any `json:"properties"`
	}
	if unmarshalErr := json.Unmarshal(schemaBytes, &schema); unmarshalErr != nil {
		return nil, unmarshalErr
	}
	fields := make([]string, 0, len(schema.Properties))
	for k := range schema.Properties {
		fields = append(fields, k)
	}

	return fields, nil
}

// getUnknownTopLevelFields returns a slice of unknown top-level fields in the YAML
// that are not in allowedFields.
func getUnknownTopLevelFields(yamlBytes []byte, allowedFields []string) ([]string, error) {
	var node yaml.Node
	if err := yaml.Unmarshal(yamlBytes, &node); err != nil {
		return nil, err
	}
	allowed := make(map[string]struct{}, len(allowedFields))
	for _, f := range allowedFields {
		allowed[f] = struct{}{}
	}
	var unknown []string
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		m := node.Content[0]
		if m.Kind == yaml.MappingNode {
			for i := 0; i < len(m.Content); i += 2 {
				key := m.Content[i].Value
				if _, ok := allowed[key]; !ok {
					unknown = append(unknown, key)
				}
			}
		}
	}

	return unknown, nil
}
