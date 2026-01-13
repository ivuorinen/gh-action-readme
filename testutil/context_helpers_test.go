package testutil

import "testing"

const testErrorMessage = "test error"

func TestContextWithPath(t *testing.T) {
	t.Parallel()

	result := ContextWithPath("/test/path")
	if result["path"] != "/test/path" {
		t.Errorf("expected path='/test/path', got '%s'", result["path"])
	}
}

func TestContextWithError(t *testing.T) {
	t.Parallel()

	result := ContextWithError(testErrorMessage)
	if result["error"] != testErrorMessage {
		t.Errorf("expected error='%s', got '%s'", testErrorMessage, result["error"])
	}
}

func TestContextWithStatusCode(t *testing.T) {
	t.Parallel()

	result := ContextWithStatusCode("404")
	if result["status_code"] != "404" {
		t.Errorf("expected status_code='404', got '%s'", result["status_code"])
	}
}

func TestEmptyContext(t *testing.T) {
	t.Parallel()

	result := EmptyContext()
	if len(result) != 0 {
		t.Errorf("expected empty context, got %d entries", len(result))
	}
}

func TestContextWithLine(t *testing.T) {
	t.Parallel()

	result := ContextWithLine("25")
	if result["line"] != "25" {
		t.Errorf("expected line='25', got '%s'", result["line"])
	}
}

func TestContextWithMissingFields(t *testing.T) {
	t.Parallel()

	result := ContextWithMissingFields("name, description")
	if result["missing_fields"] != "name, description" {
		t.Errorf("expected missing_fields='name, description', got '%s'", result["missing_fields"])
	}
}

func TestContextWithDirectory(t *testing.T) {
	t.Parallel()

	result := ContextWithDirectory("/project")
	if result["directory"] != "/project" {
		t.Errorf("expected directory='/project', got '%s'", result["directory"])
	}
}

func TestContextWithConfigPath(t *testing.T) {
	t.Parallel()

	result := ContextWithConfigPath("~/.config/app/config.yaml")
	if result["config_path"] != "~/.config/app/config.yaml" {
		t.Errorf("expected config_path='~/.config/app/config.yaml', got '%s'", result["config_path"])
	}
}

func TestContextWithCommand(t *testing.T) {
	t.Parallel()

	result := ContextWithCommand("gh-action-readme")
	if result["command"] != "gh-action-readme" {
		t.Errorf("expected command='gh-action-readme', got '%s'", result["command"])
	}
}

func TestContextWithField(t *testing.T) {
	t.Parallel()

	result := ContextWithField("theme", "custom")
	if result["theme"] != "custom" {
		t.Errorf("expected theme='custom', got '%s'", result["theme"])
	}
}

func TestMergeContexts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		contexts []map[string]string
		expected map[string]string
	}{
		{
			name:     "empty contexts",
			contexts: []map[string]string{},
			expected: map[string]string{},
		},
		{
			name: "single context",
			contexts: []map[string]string{
				{"key": "value"},
			},
			expected: map[string]string{"key": "value"},
		},
		{
			name: "multiple contexts without overlap",
			contexts: []map[string]string{
				{"key1": "value1"},
				{"key2": "value2"},
			},
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "multiple contexts with overlap - later wins",
			contexts: []map[string]string{
				{"key": "first"},
				{"key": "second"},
			},
			expected: map[string]string{"key": "second"},
		},
		{
			name: "complex merge",
			contexts: []map[string]string{
				{"path": "/test", "error": "not found"},
				{"status_code": "404"},
				{"error": "file not found"},
			},
			expected: map[string]string{
				"path":        "/test",
				"error":       "file not found",
				"status_code": "404",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := MergeContexts(tt.contexts...)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d entries, got %d", len(tt.expected), len(result))
			}

			for key, expectedValue := range tt.expected {
				if result[key] != expectedValue {
					t.Errorf("expected %s='%s', got '%s'", key, expectedValue, result[key])
				}
			}
		})
	}
}
