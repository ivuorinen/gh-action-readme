package testutil

import "testing"

// StringTestCase represents a test case for string transformation functions.
type StringTestCase struct {
	Name  string
	Input string
	Want  string
}

// RunStringTests runs a suite of string transformation tests.
// The function fn should transform the input string and return the result.
func RunStringTests(t *testing.T, tests []StringTestCase, fn func(string) string) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			got := fn(tt.Input)
			if got != tt.Want {
				t.Errorf("got %q, want %q", got, tt.Want)
			}
		})
	}
}
