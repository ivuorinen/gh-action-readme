//go:build unix

package dependencies

import (
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// limitFileSizeToZero drops RLIMIT_FSIZE to 0 so that writes to regular files fail
// with EFBIG while os.CreateTemp still succeeds, and returns a function restoring
// the previous limit. SIGXFSZ is ignored for the duration: the kernel raises it on
// an over-limit write and its default disposition would kill the test binary.
//
// It probes the limit before handing control back, and skips the test when the
// environment does not enforce it (some sandboxes silently ignore the change), so
// an unsupported platform reports "skipped" rather than a spurious failure.
func limitFileSizeToZero(t *testing.T) func() {
	t.Helper()

	var saved syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_FSIZE, &saved); err != nil {
		t.Skipf("RLIMIT_FSIZE unavailable on this platform: %v", err)
	}

	signal.Ignore(syscall.SIGXFSZ)

	restore := func() {
		_ = syscall.Setrlimit(syscall.RLIMIT_FSIZE, &saved)
		signal.Reset(syscall.SIGXFSZ)
	}

	if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &syscall.Rlimit{Cur: 0, Max: saved.Max}); err != nil {
		restore()
		t.Skipf("cannot lower RLIMIT_FSIZE in this environment: %v", err)
	}

	// Probe: confirm the limit is actually enforced before relying on it.
	probe, err := os.CreateTemp(t.TempDir(), "rlimit-probe-*")
	if err != nil {
		restore()
		t.Skipf("probe file could not be created: %v", err)
	}

	_, werr := probe.Write([]byte("probe"))
	_ = probe.Close()

	if werr == nil {
		restore()
		t.Skip("RLIMIT_FSIZE is not enforced in this environment; cannot induce a write failure")
	}

	return restore
}

// TestCreateBackupWriteFailure covers createBackup's write-failure branch. The
// branch matters because it is the one place a backup can be created but left
// unusable: updateActionFile rolls the original action.yml back from this file, so
// a truncated backup that was still reported as a success would replace a valid
// action.yml with a partial one. createBackup must instead delete the half-written
// file and report the error.
//
// This test is deliberately NOT parallel. RLIMIT_FSIZE is process-global, and Go
// runs top-level serial tests one at a time while parallel ones stay paused until
// the sequential pass finishes — so no other test writes files while the limit is
// in force.
func TestCreateBackupWriteFailure(t *testing.T) {
	// Not parallel: see the doc comment above.
	dir := t.TempDir()
	target := filepath.Join(dir, backupTestFile)

	restore := limitFileSizeToZero(t)
	defer restore()

	backupPath, err := createBackup(target, []byte("content that cannot be written"))
	if err == nil {
		t.Fatalf("createBackup() with writes failing returned %q, want an error", backupPath)
	}
	if !strings.Contains(err.Error(), "failed to create backup") {
		t.Errorf("error = %v, want it to mention 'failed to create backup'", err)
	}
	if backupPath != "" {
		t.Errorf("createBackup() returned path %q alongside an error, want empty", backupPath)
	}

	// The half-written backup must be cleaned up, not left behind as a rollback
	// point that would restore truncated content.
	entries, rerr := os.ReadDir(dir)
	if rerr != nil {
		t.Fatalf("read backup directory: %v", rerr)
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), ".bak-") {
			t.Errorf("createBackup() left a failed backup behind: %s", e.Name())
		}
	}
}
