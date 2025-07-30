package validation

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/ivuorinen/gh-action-readme/internal/git"
)

// IsCommitSHA checks if a version string is a commit SHA.
func IsCommitSHA(version string) bool {
	// Check if it's a 40-character hex string (full SHA) or 7+ character hex (short SHA)
	re := regexp.MustCompile(`^[a-f0-9]{7,40}$`)
	return len(version) >= 7 && re.MatchString(version)
}

// IsSemanticVersion checks if a version string follows semantic versioning.
func IsSemanticVersion(version string) bool {
	// Check for vX.Y.Z format (requires major.minor.patch)
	re := regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$`)
	return re.MatchString(version)
}

// IsVersionPinned checks if a semantic version is pinned to a specific version.
func IsVersionPinned(version string) bool {
	// Consider it pinned if it specifies patch version (v1.2.3) or is a commit SHA
	if IsSemanticVersion(version) {
		return true
	}
	return IsCommitSHA(version) && len(version) == 40 // Only full SHAs are considered pinned
}

// ValidateGitBranch checks if a branch exists in the given repository.
func ValidateGitBranch(repoRoot, branch string) bool {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	cmd.Dir = repoRoot
	return cmd.Run() == nil
}

// ValidateActionYMLPath validates that a path points to a valid action.yml file.
func ValidateActionYMLPath(path string) error {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}

	// Check if it's an action.yml or action.yaml file
	filename := filepath.Base(path)
	if filename != "action.yml" && filename != "action.yaml" {
		return os.ErrInvalid
	}

	return nil
}

// IsGitRepository checks if the given path is within a git repository.
func IsGitRepository(path string) bool {
	_, err := git.FindRepositoryRoot(path)
	return err == nil
}
