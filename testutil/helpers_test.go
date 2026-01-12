package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// TestContextHelpers tests the context map factory functions.
func TestContextHelpers(t *testing.T) {
	t.Parallel()

	t.Run("ContextWithPath", func(t *testing.T) {
		t.Parallel()
		result := ContextWithPath("/test/path")
		if result["path"] != "/test/path" {
			t.Errorf("ContextWithPath() path = %v, want /test/path", result["path"])
		}
		if len(result) != 1 {
			t.Errorf("ContextWithPath() len = %v, want 1", len(result))
		}
	})

	t.Run("ContextWithError", func(t *testing.T) {
		t.Parallel()
		result := ContextWithError("test error")
		if result["error"] != "test error" {
			t.Errorf("ContextWithError() error = %v, want test error", result["error"])
		}
		if len(result) != 1 {
			t.Errorf("ContextWithError() len = %v, want 1", len(result))
		}
	})

	t.Run("ContextWithStatusCode", func(t *testing.T) {
		t.Parallel()
		result := ContextWithStatusCode("404")
		if result["status_code"] != "404" {
			t.Errorf("ContextWithStatusCode() status_code = %v, want 404", result["status_code"])
		}
		if len(result) != 1 {
			t.Errorf("ContextWithStatusCode() len = %v, want 1", len(result))
		}
	})

	t.Run("EmptyContext", func(t *testing.T) {
		t.Parallel()
		result := EmptyContext()
		if len(result) != 0 {
			t.Errorf("EmptyContext() len = %v, want 0", len(result))
		}
	})
}

// TestGitHelpers tests the git setup helper functions.
func TestGitHelpers(t *testing.T) {
	t.Parallel()

	t.Run("SetupGitDirectory", func(t *testing.T) {
		t.Parallel()
		tmpDir, cleanup := TempDir(t)
		defer cleanup()

		gitDir := SetupGitDirectory(t, tmpDir)

		// Verify git directory exists
		expectedGitDir := filepath.Join(tmpDir, ".git")
		if gitDir != expectedGitDir {
			t.Errorf("SetupGitDirectory() = %v, want %v", gitDir, expectedGitDir)
		}

		// Verify directory was created
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			t.Errorf("SetupGitDirectory() did not create .git directory")
		}
	})

	t.Run("SetupGitConfig", func(t *testing.T) {
		t.Parallel()
		tmpDir, cleanup := TempDir(t)
		defer cleanup()

		gitDir := SetupGitDirectory(t, tmpDir)
		remoteURL := "https://github.com/owner/repo.git"
		SetupGitConfig(t, gitDir, remoteURL)

		// Verify config file exists
		configPath := filepath.Join(gitDir, "config")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Errorf("SetupGitConfig() did not create config file")
		}

		// Verify config content
		content, err := os.ReadFile(configPath) // #nosec G304 -- Test code reading from controlled temp directory
		if err != nil {
			t.Fatalf("Failed to read config file: %v", err)
		}
		contentStr := string(content)
		if !contains(contentStr, remoteURL) {
			t.Errorf("SetupGitConfig() config does not contain remote URL: %s", remoteURL)
		}
		if !contains(contentStr, `[remote "origin"]`) {
			t.Errorf("SetupGitConfig() config does not contain remote origin section")
		}
	})
}
