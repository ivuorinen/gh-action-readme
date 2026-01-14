package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-github/v74/github"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// TestAssertBooleanConfigFields_Helper tests the assertBooleanConfigFields helper.
func TestAssertBooleanConfigFields_Helper(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  *AppConfig
		want *AppConfig
	}{
		{
			name: "all fields match",
			got: &AppConfig{
				AnalyzeDependencies: true,
				ShowSecurityInfo:    false,
				Verbose:             true,
				Quiet:               false,
				UseDefaultBranch:    true,
			},
			want: &AppConfig{
				AnalyzeDependencies: true,
				ShowSecurityInfo:    false,
				Verbose:             true,
				Quiet:               false,
				UseDefaultBranch:    true,
			},
		},
		{
			name: "all fields false",
			got: &AppConfig{
				AnalyzeDependencies: false,
				ShowSecurityInfo:    false,
				Verbose:             false,
				Quiet:               false,
				UseDefaultBranch:    false,
			},
			want: &AppConfig{
				AnalyzeDependencies: false,
				ShowSecurityInfo:    false,
				Verbose:             false,
				Quiet:               false,
				UseDefaultBranch:    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call the helper - it will call t.Error if fields don't match
			// For matching cases, it should not error
			assertBooleanConfigFields(t, tt.got, tt.want)
		})
	}
}

// TestAssertGitHubClientValid_Helper tests the assertGitHubClientValid helper.
func TestAssertGitHubClientValid_Helper(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		client        *GitHubClient
		expectedToken string
	}{
		{
			name: "valid client with token",
			client: &GitHubClient{
				Client: github.NewClient(nil),
				Token:  "test-token-123",
			},
			expectedToken: "test-token-123",
		},
		{
			name: "valid client with empty token",
			client: &GitHubClient{
				Client: github.NewClient(nil),
				Token:  "",
			},
			expectedToken: "",
		},
		{
			name: "valid client with github PAT",
			client: &GitHubClient{
				Client: github.NewClient(nil),
				Token:  "ghp_1234567890abcdefghijklmnopqrstuvwxyzABCD",
			},
			expectedToken: "ghp_1234567890abcdefghijklmnopqrstuvwxyzABCD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call the helper - it will verify the client is valid
			// For valid clients, it should not error
			assertGitHubClientValid(t, tt.client, tt.expectedToken)
		})
	}
}

// TestRunTemplatePathTest_Helper tests the runTemplatePathTest helper.
func TestRunTemplatePathTest_Helper(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setupFunc    func(*testing.T) (string, func())
		checkFunc    func(*testing.T, string)
		expectResult string
	}{
		{
			name: "absolute path setup",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()
				tmpDir := t.TempDir()
				templatePath := filepath.Join(tmpDir, "test.tmpl")

				err := os.WriteFile(templatePath, []byte("test template"), appconstants.FilePermDefault)
				if err != nil {
					t.Fatalf("failed to write template: %v", err)
				}

				return templatePath, func() { /* Cleanup handled by t.TempDir() */ }
			},
			checkFunc: func(t *testing.T, result string) {
				t.Helper()
				if result == "" {
					t.Error(testutil.TestMsgExpectedNonEmpty)
				}
			},
		},
		{
			name: "relative path setup",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()

				return "templates/readme.tmpl", func() { /* No cleanup needed for relative path test */ }
			},
			checkFunc: func(t *testing.T, result string) {
				t.Helper()
				if result == "" {
					t.Error(testutil.TestMsgExpectedNonEmpty)
				}
			},
		},
		{
			name: "nil checkFunc (just runs setup)",
			setupFunc: func(t *testing.T) (string, func()) {
				t.Helper()

				return "test/path.tmpl", func() { /* No cleanup needed for nil checkFunc test */ }
			},
			checkFunc: nil, // No validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call the helper - it runs setup, calls resolveTemplatePath, and validates
			runTemplatePathTest(t, tt.setupFunc, tt.checkFunc)
		})
	}
}
