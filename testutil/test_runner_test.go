package testutil

import (
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
