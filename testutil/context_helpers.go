package testutil

// ContextWithPath creates a context map with a path entry.
// Used in error handler and suggestions tests to reduce duplication.
func ContextWithPath(path string) map[string]string {
	return map[string]string{TestKeyPath: path}
}

// ContextWithError creates a context map with an error entry.
// Used in error handler and suggestions tests to reduce duplication.
func ContextWithError(err string) map[string]string {
	return map[string]string{TestKeyError: err}
}

// ContextWithStatusCode creates a context map with a status code entry.
// Used in error handler and suggestions tests to reduce duplication.
func ContextWithStatusCode(code string) map[string]string {
	return map[string]string{TestKeyStatusCode: code}
}

// EmptyContext creates an empty context map.
// Used in error handler and suggestions tests to reduce duplication.
func EmptyContext() map[string]string {
	return map[string]string{}
}

// ContextWithLine creates a context with a line number.
// Useful for YAML parsing error suggestions.
func ContextWithLine(line string) map[string]string {
	return map[string]string{"line": line}
}

// ContextWithMissingFields creates a context with missing field names.
// Useful for validation error suggestions.
func ContextWithMissingFields(fields string) map[string]string {
	return map[string]string{"missing_fields": fields}
}

// ContextWithDirectory creates a context with a directory path.
// Useful for file discovery error suggestions.
func ContextWithDirectory(dir string) map[string]string {
	return map[string]string{"directory": dir}
}

// ContextWithConfigPath creates a context with a config file path.
// Useful for configuration error suggestions.
func ContextWithConfigPath(path string) map[string]string {
	return map[string]string{"config_path": path}
}

// ContextWithField creates a context with a single field value.
// Generic helper for any single-field context.
func ContextWithField(key, value string) map[string]string {
	return map[string]string{key: value}
}

// MergeContexts merges multiple context maps into one.
// Later maps override earlier maps for duplicate keys.
func MergeContexts(contexts ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, ctx := range contexts {
		for k, v := range ctx {
			result[k] = v
		}
	}

	return result
}
