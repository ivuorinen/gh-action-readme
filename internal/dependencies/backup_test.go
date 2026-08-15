package dependencies

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

const backupTestFile = "action.yml"

// TestCreateBackupWritesContent pins the data-safety contract: updateActionFile
// rewrites an action.yml in place and rolls back from this backup when the rewrite
// produces invalid YAML. A backup whose bytes do not match the original silently
// turns that rollback into corruption.
func TestCreateBackupWritesContent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	target := filepath.Join(dir, backupTestFile)
	content := []byte(testutil.MustReadFixture(testutil.TestFixtureCompositeWithDeps))

	backupPath, err := createBackup(target, content)
	if err != nil {
		t.Fatalf("createBackup() unexpected error: %v", err)
	}

	got, err := os.ReadFile(backupPath) // #nosec G304 -- path returned by the function under test
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("backup content = %q, want %q", got, content)
	}
}

// TestCreateBackupPreservesExistingBackupFile is the reason createBackup uses
// os.CreateTemp rather than a fixed "<file>.backup" name: a backup the user
// maintains themselves must survive. A fixed name would overwrite it.
func TestCreateBackupPreservesExistingBackupFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	target := filepath.Join(dir, backupTestFile)

	userBackup := filepath.Join(dir, backupTestFile+".backup")
	userContent := []byte("the user's own backup\n")
	if err := os.WriteFile(userBackup, userContent, 0o600); err != nil {
		t.Fatalf("seed user backup: %v", err)
	}

	backupPath, err := createBackup(target, []byte("fresh content\n"))
	if err != nil {
		t.Fatalf("createBackup() unexpected error: %v", err)
	}

	if backupPath == userBackup {
		t.Fatal("createBackup() reused the user's .backup path; it must generate a unique name")
	}

	got, err := os.ReadFile(userBackup) // #nosec G304 -- test-controlled path
	if err != nil {
		t.Fatalf("read user backup: %v", err)
	}
	if string(got) != string(userContent) {
		t.Errorf("user's backup was modified: got %q, want %q", got, userContent)
	}
}

// TestCreateBackupIsUniquePerCall guards the same property across repeated calls:
// two backups of the same file must not collide, or the second rollback point
// destroys the first.
func TestCreateBackupIsUniquePerCall(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	target := filepath.Join(dir, backupTestFile)

	first, err := createBackup(target, []byte("one\n"))
	if err != nil {
		t.Fatalf("first createBackup() error: %v", err)
	}

	second, err := createBackup(target, []byte("two\n"))
	if err != nil {
		t.Fatalf("second createBackup() error: %v", err)
	}

	if first == second {
		t.Fatalf("createBackup() returned the same path twice: %q", first)
	}

	// Both must still be readable with their own content.
	for path, want := range map[string]string{first: "one\n", second: "two\n"} {
		got, rerr := os.ReadFile(path) // #nosec G304 -- paths returned by the function under test
		if rerr != nil {
			t.Fatalf("read %s: %v", path, rerr)
		}
		if string(got) != want {
			t.Errorf("backup %s content = %q, want %q", path, got, want)
		}
	}
}

// TestCreateBackupErrorsOnMissingDirectory covers the failure branch: when the
// containing directory does not exist there is nowhere to put a rollback point,
// and the caller must learn that before it starts rewriting the original.
func TestCreateBackupErrorsOnMissingDirectory(t *testing.T) {
	t.Parallel()

	missing := filepath.Join(t.TempDir(), "no-such-dir", backupTestFile)

	backupPath, err := createBackup(missing, []byte("content\n"))
	if err == nil {
		t.Fatalf("createBackup() in a missing directory returned %q, want an error", backupPath)
	}
	if !strings.Contains(err.Error(), "failed to create backup") {
		t.Errorf("error = %v, want it to mention 'failed to create backup'", err)
	}
}
