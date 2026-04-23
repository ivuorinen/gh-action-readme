package testutil

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGitHelpers tests the git setup helper functions.
func TestGitHelpers(t *testing.T) {
	t.Parallel()

	t.Run("SetupGitDirectory", func(t *testing.T) {
		t.Parallel()
		tmpDir, cleanup := TempDir(t)
		defer cleanup()

		gitDir := SetupGitDirectory(t, tmpDir)

		expectedGitDir := filepath.Join(tmpDir, ".git")
		if gitDir != expectedGitDir {
			t.Errorf("SetupGitDirectory() = %v, want %v", gitDir, expectedGitDir)
		}

		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			t.Errorf("SetupGitDirectory() did not create .git directory")
		}
	})

	t.Run("CreateGitConfigWithRemote", func(t *testing.T) {
		t.Parallel()
		tmpDir, cleanup := TempDir(t)
		defer cleanup()

		gitDir := SetupGitDirectory(t, tmpDir)
		configPath := CreateGitConfigWithRemote(t, gitDir, "https://github.com/test/repo.git", "main")

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("CreateGitConfigWithRemote() did not create config file")
		}

		content, err := os.ReadFile(configPath) // #nosec G304 -- test file path from helper
		if err != nil {
			t.Fatalf("failed to read config file: %v", err)
		}

		if !strings.Contains(string(content), "https://github.com/test/repo.git") {
			t.Error("config file should contain remote URL")
		}
		if !strings.Contains(string(content), "main") {
			t.Error("config file should contain branch name")
		}
	})

	t.Run("WriteGitConfigFile", func(t *testing.T) {
		t.Parallel()
		tmpDir, cleanup := TempDir(t)
		defer cleanup()

		configContent := `[remote "origin"]
	url = https://github.com/test/repo.git`
		configPath := WriteGitConfigFile(t, tmpDir, configContent)

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("WriteGitConfigFile() did not create config file")
		}

		content, err := os.ReadFile(configPath) // #nosec G304 -- test file path from helper
		if err != nil {
			t.Fatalf("failed to read config: %v", err)
		}
		if string(content) != configContent {
			t.Errorf("unexpected config content: %q", string(content))
		}
	})
}

// TestCapturedOutputMethods tests CapturedOutput methods.
func TestCapturedOutputMethods(t *testing.T) {
	t.Parallel()

	t.Run("AllMessages", func(t *testing.T) {
		t.Parallel()
		co := &CapturedOutput{}
		co.Info("info message")
		co.Success("success message")
		co.Warning("warning message")
		co.Error("error message %s", "here")
		co.Bold("bold message")
		co.Printf("printf message")
		co.Progress("progress message")

		all := co.AllMessages()
		if len(all) != 7 {
			t.Errorf("AllMessages() returned %d messages, want 7", len(all))
		}
	})

	t.Run("ContainsMessage", func(t *testing.T) {
		t.Parallel()
		co := &CapturedOutput{}
		co.Info("hello world")

		if !co.ContainsMessage("hello") {
			t.Error("ContainsMessage() should find 'hello'")
		}
		if co.ContainsMessage("xyz") {
			t.Error("ContainsMessage() should not find 'xyz'")
		}
	})

	t.Run("ContainsError", func(t *testing.T) {
		t.Parallel()
		co := &CapturedOutput{}
		co.Error("something failed")

		if !co.ContainsError("failed") {
			t.Error("ContainsError() should find 'failed'")
		}
		if co.ContainsError("success") {
			t.Error("ContainsError() should not find 'success'")
		}
	})
}

// TestErrorFormatterMock tests the ErrorFormatterMock.
func TestErrorFormatterMock(t *testing.T) {
	t.Parallel()

	t.Run("formats non-nil error", func(t *testing.T) {
		t.Parallel()
		mock := &ErrorFormatterMock{}
		err := errors.New("test error")
		result := mock.FormatContextualError(err)

		if result != "test error" {
			t.Errorf("FormatContextualError() = %q, want %q", result, "test error")
		}
		if len(mock.FormatContextualErrorCalls) != 1 {
			t.Errorf("expected 1 call, got %d", len(mock.FormatContextualErrorCalls))
		}
	})

	t.Run("returns empty string for nil error", func(t *testing.T) {
		t.Parallel()
		mock := &ErrorFormatterMock{}
		result := mock.FormatContextualError(nil)

		if result != "" {
			t.Errorf("FormatContextualError(nil) = %q, want empty", result)
		}
		if len(mock.FormatContextualErrorCalls) != 0 {
			t.Errorf("expected 0 calls, got %d", len(mock.FormatContextualErrorCalls))
		}
	})
}

// TestWriteActionFile tests WriteActionFile helper.
func TestWriteActionFile(t *testing.T) {
	t.Parallel()

	tmpDir, cleanup := TempDir(t)
	defer cleanup()

	content := "name: Test\ndescription: test action\nruns:\n  using: composite\n  steps: []"
	actionPath := WriteActionFile(t, tmpDir, content)

	if _, err := os.Stat(actionPath); os.IsNotExist(err) {
		t.Error("WriteActionFile() did not create file")
	}

	got, err := os.ReadFile(actionPath) // #nosec G304 -- test file path from helper
	if err != nil {
		t.Fatalf("failed to read action file: %v", err)
	}
	if string(got) != content {
		t.Errorf("WriteActionFile() content mismatch")
	}
}

// TestAssertBackupNotExists tests AssertBackupNotExists helper.
func TestAssertBackupNotExists(t *testing.T) {
	t.Parallel()

	tmpDir, cleanup := TempDir(t)
	defer cleanup()

	filePath := filepath.Join(tmpDir, "testfile.yml")
	WriteTestFile(t, filePath, "content")

	// No backup file exists → should not fail
	AssertBackupNotExists(t, filePath)
}

// TestAssertFileContentEquals tests AssertFileContentEquals helper.
func TestAssertFileContentEquals(t *testing.T) {
	t.Parallel()

	tmpDir, cleanup := TempDir(t)
	defer cleanup()

	filePath := filepath.Join(tmpDir, "testfile.yml")
	content := "name: test\ndescription: test"
	WriteTestFile(t, filePath, content)

	AssertFileContentEquals(t, filePath, content)
}
