package testutil

import (
	"errors"
	"strings"
	"testing"
)

func TestRunStringTests(t *testing.T) {
	t.Parallel()

	tests := []StringTestCase{
		{Name: "uppercase", Input: "hello", Want: "HELLO"},
		{Name: "lowercase", Input: "WORLD", Want: "world"},
	}

	RunStringTests(t, tests, func(s string) string {
		if s == "hello" {
			return strings.ToUpper(s)
		}

		return strings.ToLower(s)
	})
}

func TestRunBoolTests(t *testing.T) {
	t.Parallel()

	tests := []BoolTestCase{
		{Name: "empty string", Input: "", Want: false},
		{Name: "non-empty string", Input: "test", Want: true},
	}

	RunBoolTests(t, tests, func(s string) bool {
		return len(s) > 0
	})
}

func TestRunErrorTests(t *testing.T) {
	t.Parallel()

	tests := []ErrorTestCase{
		{Name: "valid input", Input: "valid", WantErr: false},
		{Name: "invalid input", Input: "invalid", WantErr: true, ErrContains: "invalid"},
		{Name: "error without check", Input: "bad", WantErr: true},
	}

	RunErrorTests(t, tests, func(s string) error {
		if s == "valid" {
			return nil
		}
		if s == "invalid" {
			return errors.New("invalid input")
		}

		return errors.New("something went wrong")
	})
}

func TestContains(t *testing.T) {
	t.Parallel()

	const testString = "hello world"

	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{"empty substring", "hello", "", true},
		{"exact match", "test", "test", true},
		{"substring at start", testString, "hello", true},
		{"substring at end", testString, "world", true},
		{"substring in middle", testString, "lo wo", true},
		{"not found", "hello", "goodbye", false},
		{"longer substring", "hi", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := contains(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

func TestRunMapValidationTests(t *testing.T) {
	t.Parallel()

	tests := []MapValidationTestCase{
		{
			Name:  "valid map",
			Input: map[string]string{"key": "value"},
			Validate: func(m map[string]string) error {
				if m["key"] != "value" {
					return errors.New("unexpected value")
				}

				return nil
			},
		},
		{
			Name:  "empty map",
			Input: map[string]string{},
			Validate: func(m map[string]string) error {
				if len(m) != 0 {
					return errors.New("expected empty map")
				}

				return nil
			},
		},
		{
			Name:  "map with multiple keys",
			Input: map[string]string{"key1": "value1", "key2": "value2"},
			Validate: func(m map[string]string) error {
				if len(m) != 2 {
					return errors.New("expected 2 keys")
				}

				return nil
			},
		},
	}

	RunMapValidationTests(t, tests)
}

func TestRunStringSliceTests(t *testing.T) {
	t.Parallel()

	tests := []StringSliceTestCase{
		{
			Name:  "reverse slice",
			Input: []string{"a", "b", "c"},
			Want:  []string{"c", "b", "a"},
			Fn: func(s []string) []string {
				result := make([]string, len(s))
				for i, v := range s {
					result[len(s)-1-i] = v
				}

				return result
			},
		},
		{
			Name:  "uppercase slice",
			Input: []string{"hello", "world"},
			Want:  []string{"HELLO", "WORLD"},
			Fn: func(s []string) []string {
				result := make([]string, len(s))
				for i, v := range s {
					result[i] = strings.ToUpper(v)
				}

				return result
			},
		},
		{
			Name:  "empty slice",
			Input: []string{},
			Want:  []string{},
			Fn: func(s []string) []string {
				return s
			},
		},
	}

	RunStringSliceTests(t, tests)
}
