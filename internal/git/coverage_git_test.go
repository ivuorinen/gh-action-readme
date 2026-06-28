package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

// Constants local to the real-git coverage tests. testGitBranchMain ("main") is
// already declared in detector_test.go (same package) and is reused here.
const (
	covGitOrg       = "cov-org"
	covGitRepo      = "cov-repo"
	covGitRemoteURL = "https://github.com/cov-org/cov-repo.git"
	covGitOriginRef = "refs/remotes/origin/HEAD"
)

// initRealGitRepo runs `git init` in dir with global/system config neutralized
// and author identity supplied via the environment, so the repository is fully
// isolated from the host's git configuration. It skips the test if git is not
// installed.
func initRealGitRepo(t *testing.T, dir string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}
	runRealGit(t, dir, "-c", "init.defaultBranch="+testGitBranchMain, "init")
}

// runRealGit executes a git command in dir and fails the test on error.
func runRealGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...) // #nosec G204 -- test-controlled args
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL="+os.DevNull,
		"GIT_CONFIG_SYSTEM="+os.DevNull,
		"GIT_AUTHOR_NAME=cov",
		"GIT_AUTHOR_EMAIL=cov@example.com",
		"GIT_COMMITTER_NAME=cov",
		"GIT_COMMITTER_EMAIL=cov@example.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}

// TestCovGitGetRemoteURLFromGitSuccess drives the success path of
// getRemoteURLFromGit / getRemoteURL using a real `git remote add origin`.
func TestCovGitGetRemoteURLFromGitSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	initRealGitRepo(t, tmpDir)
	runRealGit(t, tmpDir, "remote", "add", "origin", covGitRemoteURL)

	urlFromGit, err := getRemoteURLFromGit(tmpDir)
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, covGitRemoteURL, urlFromGit)

	urlFromWrapper, err := getRemoteURL(tmpDir)
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, covGitRemoteURL, urlFromWrapper)
}

// TestCovGitDetectRepositoryRealGit exercises DetectRepository through the live
// git command path: the remote URL is parsed into organization/repository.
func TestCovGitDetectRepositoryRealGit(t *testing.T) {
	tmpDir := t.TempDir()
	initRealGitRepo(t, tmpDir)
	runRealGit(t, tmpDir, "remote", "add", "origin", covGitRemoteURL)

	info, err := DetectRepository(tmpDir)
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, true, info.IsGitRepo)
	testutil.AssertEqual(t, covGitRemoteURL, info.RemoteURL)
	testutil.AssertEqual(t, covGitOrg, info.Organization)
	testutil.AssertEqual(t, covGitRepo, info.Repository)
}

// TestCovGitGetDefaultBranchSymbolicRef drives the symbolic-ref success branch
// of getDefaultBranch by setting refs/remotes/origin/HEAD to point at main.
func TestCovGitGetDefaultBranchSymbolicRef(t *testing.T) {
	tmpDir := t.TempDir()
	initRealGitRepo(t, tmpDir)
	runRealGit(t, tmpDir, "symbolic-ref", covGitOriginRef, "refs/remotes/origin/"+testGitBranchMain)

	branch := getDefaultBranch(tmpDir)
	testutil.AssertEqual(t, testGitBranchMain, branch)
}

// TestCovGitGetDefaultBranchBranchExists drives the fallback branch of
// getDefaultBranch: with no origin/HEAD symbolic ref, the symbolic-ref command
// fails and the function locates the existing local "main" branch via
// branchExists (which requires a real commit so refs/heads/main exists).
func TestCovGitGetDefaultBranchBranchExists(t *testing.T) {
	tmpDir := t.TempDir()
	initRealGitRepo(t, tmpDir)
	runRealGit(t, tmpDir, "commit", "--allow-empty", "-m", "init")

	// Sanity: branchExists should report the freshly created default branch.
	if !branchExists(tmpDir, testGitBranchMain) {
		t.Fatalf("expected branch %q to exist after a commit", testGitBranchMain)
	}

	branch := getDefaultBranch(tmpDir)
	testutil.AssertEqual(t, testGitBranchMain, branch)
}

// TestCovGitBranchExistsMissing covers the negative result of branchExists for a
// branch that was never created.
func TestCovGitBranchExistsMissing(t *testing.T) {
	tmpDir := t.TempDir()
	initRealGitRepo(t, tmpDir)

	if branchExists(tmpDir, "definitely-missing-branch") {
		t.Error("expected branchExists to be false for a non-existent branch")
	}
}

// TestCovGitGetRemoteURLFromGitError covers the error path of getRemoteURLFromGit
// when the directory is not a git repository (the git command exits non-zero).
func TestCovGitGetRemoteURLFromGitError(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}
	// A bare temp dir with no .git: `git remote get-url origin` fails.
	nonRepo := filepath.Join(t.TempDir(), "plain")
	testutil.CreateTestDir(t, nonRepo)

	_, err := getRemoteURLFromGit(nonRepo)
	testutil.AssertError(t, err)
}
