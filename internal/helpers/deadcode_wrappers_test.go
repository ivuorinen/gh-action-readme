package helpers

// Test-only wrapper. GetGitRepoRootAndInfo has no production caller (production
// resolves the repo root and info separately), but tests use it to exercise the
// git.FindRepositoryRoot + git.DetectRepository pair. Kept here as a test helper
// so that coverage stays without shipping dead code. See docs/audit finding N-161.

import "github.com/ivuorinen/gh-action-readme/internal/git"

// GetGitRepoRootAndInfo gets git repository root and info with error handling.
func GetGitRepoRootAndInfo(startPath string) (string, *git.RepoInfo, error) {
	repoRoot, err := git.FindRepositoryRoot(startPath)
	if err != nil {
		return "", nil, err
	}

	gitInfo, err := git.DetectRepository(repoRoot)
	if err != nil {
		return repoRoot, nil, err
	}

	return repoRoot, gitInfo, nil
}
