package validation

import (
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

// TestTrimAndNormalize tests the TrimAndNormalize function.
func TestTrimAndNormalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no whitespace",
			input: "test",
			want:  "test",
		},
		{
			name:  "leading and trailing whitespace",
			input: "  test  ",
			want:  "test",
		},
		{
			name:  "multiple internal spaces",
			input: "hello    world",
			want:  testutil.HelloWorldStr,
		},
		{
			name:  "mixed whitespace",
			input: "  hello   world  ",
			want:  testutil.HelloWorldStr,
		},
		{
			name:  "newlines and tabs",
			input: "hello\n\t\tworld",
			want:  testutil.HelloWorldStr,
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "whitespace only",
			input: "   \n\t  ",
			want:  "",
		},
		{
			name:  "multiple lines",
			input: "line one\n  line two\n    line three",
			want:  "line one line two line three",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := TrimAndNormalize(tt.input)
			if got != tt.want {
				t.Errorf("TrimAndNormalize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestFormatUsesStatement tests the FormatUsesStatement function.
func TestFormatUsesStatement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		org     string
		repo    string
		version string
		want    string
	}{
		{
			name:    "full statement with version",
			org:     "actions",
			repo:    "checkout",
			version: "v3",
			want:    "actions/checkout@v3",
		},
		{
			name:    "without version defaults to v1",
			org:     "actions",
			repo:    "setup-node",
			version: "",
			want:    "actions/setup-node@v1",
		},
		{
			name:    "version with @ prefix",
			org:     "actions",
			repo:    "cache",
			version: "@v2",
			want:    "actions/cache@v2",
		},
		{
			name:    "version without @ prefix",
			org:     "actions",
			repo:    "upload-artifact",
			version: "v4",
			want:    "actions/upload-artifact@v4",
		},
		{
			name:    "empty org returns empty",
			org:     "",
			repo:    "checkout",
			version: "v3",
			want:    "",
		},
		{
			name:    "empty repo returns empty",
			org:     "actions",
			repo:    "",
			version: "v3",
			want:    "",
		},
		{
			name:    "both org and repo empty",
			org:     "",
			repo:    "",
			version: "v3",
			want:    "",
		},
		{
			name:    "sha as version",
			org:     "actions",
			repo:    "checkout",
			version: "abc123def456",
			want:    "actions/checkout@abc123def456",
		},
		{
			name:    "main branch as version",
			org:     "actions",
			repo:    "checkout",
			version: "main",
			want:    "actions/checkout@main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := FormatUsesStatement(tt.org, tt.repo, tt.version)
			if got != tt.want {
				t.Errorf("FormatUsesStatement(%q, %q, %q) = %q, want %q",
					tt.org, tt.repo, tt.version, got, tt.want)
			}
		})
	}
}
