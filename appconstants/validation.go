// Package appconstants provides common constants used throughout the application.
package appconstants

// Field names for validation.
const (
	// FieldName is the name field.
	FieldName = "name"
	// FieldDescription is the description field.
	FieldDescription = "description"
	// FieldRuns is the runs field.
	FieldRuns = "runs"
	// FieldRunsUsing is the runs.using field.
	FieldRunsUsing = "runs.using"
)

// YAML format string constants for test fixtures and action generation.
const (
	// YAMLFieldName is the YAML name field format.
	YAMLFieldName = "name: %s\n"
	// YAMLFieldDescription is the YAML description field format.
	YAMLFieldDescription = "description: %s\n"
	// YAMLFieldRuns is the YAML runs field.
	YAMLFieldRuns = "runs:\n"
	// JSONCloseBrace is the JSON closing brace with newline.
	JSONCloseBrace = "  },\n"
)
