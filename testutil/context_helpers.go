package testutil

// ContextWithPath creates a context map with a path entry.
// Used in error handler and suggestions tests to reduce duplication.
func ContextWithPath(path string) map[string]string {
	return map[string]string{"path": path}
}

// ContextWithError creates a context map with an error entry.
// Used in error handler and suggestions tests to reduce duplication.
func ContextWithError(err string) map[string]string {
	return map[string]string{"error": err}
}

// ContextWithStatusCode creates a context map with a status code entry.
// Used in error handler and suggestions tests to reduce duplication.
func ContextWithStatusCode(code string) map[string]string {
	return map[string]string{"status_code": code}
}

// EmptyContext creates an empty context map.
// Used in error handler and suggestions tests to reduce duplication.
func EmptyContext() map[string]string {
	return map[string]string{}
}
