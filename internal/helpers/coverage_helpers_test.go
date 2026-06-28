package helpers

import (
	"os"
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

// TestCovHelpersGetCurrentDirError exercises the error branch of GetCurrentDir.
// On Linux, removing the process working directory makes os.Getwd fail, which is
// the only reachable way to drive the error path. The original working directory
// is restored on cleanup. This test mutates global process state (the cwd), so it
// must not run in parallel.
func TestCovHelpersGetCurrentDirError(t *testing.T) {
	// t.Chdir switches into the throwaway dir and restores the original cwd at
	// cleanup (the original still exists, so cleanup is safe).
	removed := t.TempDir()
	t.Chdir(removed)

	// Some platforms keep the cwd resolvable via the PWD env var even after the
	// directory is unlinked; clearing it forces os.Getwd to hit the syscall.
	t.Setenv("PWD", "")

	if err := os.RemoveAll(removed); err != nil {
		t.Skipf("cannot remove cwd to drive the error path: %v", err)
	}

	dir, err := GetCurrentDir()
	if err == nil {
		// Platform kept the cwd valid; the error branch is unreachable here.
		t.Skip("os.Getwd still succeeds after removing the working directory")
	}
	testutil.AssertEqual(t, "", dir)
}

// TestCovHelpersFindGitRepoRootNonRepo confirms FindGitRepoRoot swallows the
// "not a git repository" error and returns an empty string (the error-handling
// branch of the helper).
func TestCovHelpersFindGitRepoRootNonRepo(t *testing.T) {
	t.Parallel()

	nonRepo := t.TempDir()
	if root := FindGitRepoRoot(nonRepo); root != "" {
		t.Errorf("expected empty repo root for non-git dir, got %q", root)
	}
}

// TestCovHelpersGetGitRepoRootAndInfoSuccess covers the full success path of
// GetGitRepoRootAndInfo: a directory containing a .git directory resolves to a
// repo root and a non-nil RepoInfo with no error.
func TestCovHelpersGetGitRepoRootAndInfoSuccess(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testutil.SetupGitDirectory(t, tmpDir)

	root, info, err := GetGitRepoRootAndInfo(tmpDir)
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, tmpDir, root)
	if info == nil {
		t.Fatal("expected non-nil git info")
	}
	testutil.AssertEqual(t, true, info.IsGitRepo)
}

// Note: the DetectRepository-error branch in GetGitRepoRootAndInfo (the
// `if err != nil` after git.DetectRepository) is unreachable: DetectRepository
// never returns a non-nil error for a path that FindRepositoryRoot accepted.
