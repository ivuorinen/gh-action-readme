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

// BoolTestCase represents a test case for boolean validation functions.
type BoolTestCase struct {
	Name  string
	Input string
	Want  bool
}

// RunBoolTests runs a suite of validation tests.
// The function fn should validate the input string and return true/false.
func RunBoolTests(t *testing.T, tests []BoolTestCase, fn func(string) bool) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			got := fn(tt.Input)
			if got != tt.Want {
				t.Errorf(TestMsgGotWant, got, tt.Want)
			}
		})
	}
}

// ErrorTestCase represents a test case for functions that return errors.
type ErrorTestCase struct {
	Name        string
	Input       string
	WantErr     bool
	ErrContains string // Optional: check if error message contains this string
}

// RunErrorTests runs a suite of error-returning function tests.
// The function fn should process the input and return an error if validation fails.
func RunErrorTests(t *testing.T, tests []ErrorTestCase, fn func(string) error) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			err := fn(tt.Input)
			if tt.WantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if tt.ErrContains != "" && !contains(err.Error(), tt.ErrContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.ErrContains)
				}
			} else {
				if err != nil {
					t.Errorf(TestErrUnexpected, err)
				}
			}
		})
	}
}

// contains checks if s contains substr (case-sensitive).
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}

// MapValidationTestCase represents a test case that validates maps.
type MapValidationTestCase struct {
	Name     string
	Input    map[string]string
	Validate func(map[string]string) error
}

// RunMapValidationTests runs validation tests on maps.
func RunMapValidationTests(t *testing.T, tests []MapValidationTestCase) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			if err := tt.Validate(tt.Input); err != nil {
				t.Errorf("validation failed: %v", err)
			}
		})
	}
}

// StringSliceTestCase represents a test case for string slice operations.
type StringSliceTestCase struct {
	Name  string
	Input []string
	Want  []string
	Fn    func([]string) []string
}

// RunStringSliceTests runs tests on string slice functions.
func RunStringSliceTests(t *testing.T, tests []StringSliceTestCase) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			got := tt.Fn(tt.Input)
			if !slicesEqual(got, tt.Want) {
				t.Errorf(TestMsgGotWant, got, tt.Want)
			}
		})
	}
}

// slicesEqual compares two string slices for equality.
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
