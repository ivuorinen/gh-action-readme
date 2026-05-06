package testutil

import "testing"

const testErrorMessage = "test error"

func TestContextHelpers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		key         string
		value       string
		contextFunc func(string) map[string]string
	}{
		{"ContextWithPath", TestKeyPath, "/test/path", ContextWithPath},
		{"ContextWithError", TestKeyError, testErrorMessage, ContextWithError},
		{"ContextWithStatusCode", TestKeyStatusCode, "404", ContextWithStatusCode},
		{"ContextWithLine", "line", "42", ContextWithLine},
		{"ContextWithMissingFields", "missing_fields", "field1,field2", ContextWithMissingFields},
		{"ContextWithDirectory", "directory", "/test/dir", ContextWithDirectory},
		{"ContextWithConfigPath", "config_path", "/config.yaml", ContextWithConfigPath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.contextFunc(tt.value)
			if result[tt.key] != tt.value {
				t.Errorf("expected %s='%s', got '%s'", tt.key, tt.value, result[tt.key])
			}
		})
	}
}

func TestEmptyContext(t *testing.T) {
	t.Parallel()

	result := EmptyContext()
	if len(result) != 0 {
		t.Errorf("expected empty context, got %d entries", len(result))
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
				{TestKeySingle: "value"},
			},
			expected: map[string]string{TestKeySingle: "value"},
		},
		{
			name: "multiple contexts without overlap",
			contexts: []map[string]string{
				{"key1": "value1"},
				{"key2": CacheTestValue2},
			},
			expected: map[string]string{
				"key1": "value1",
				"key2": CacheTestValue2,
			},
		},
		{
			name: "multiple contexts with overlap - later wins",
			contexts: []map[string]string{
				{TestKeySingle: "first"},
				{TestKeySingle: "second"},
			},
			expected: map[string]string{TestKeySingle: "second"},
		},
		{
			name: "complex merge",
			contexts: []map[string]string{
				{TestKeyPath: "/test", TestKeyError: "not found"},
				{TestKeyStatusCode: "404"},
				{TestKeyError: "file not found"},
			},
			expected: map[string]string{
				TestKeyPath:       "/test",
				TestKeyError:      "file not found",
				TestKeyStatusCode: "404",
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
