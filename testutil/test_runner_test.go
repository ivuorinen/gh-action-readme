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

	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{"empty substring", "hello", "", true},
		{"exact match", "test", "test", true},
		{"substring at start", "hello world", "hello", true},
		{"substring at end", "hello world", "world", true},
		{"substring in middle", "hello world", "lo wo", true},
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
