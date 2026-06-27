package validation

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/git"
)

var (
	reCommitSHA       = regexp.MustCompile(appconstants.RegexGitSHA)
	reSemanticVersion = regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$`)
)

// IsCommitSHA checks if a version string is a commit SHA.
func IsCommitSHA(version string) bool {
	return len(version) >= 7 && reCommitSHA.MatchString(version)
}

// IsSemanticVersion checks if a version string follows semantic versioning.
func IsSemanticVersion(version string) bool {
	return reSemanticVersion.MatchString(version)
}

// IsVersionPinned checks if a semantic version is pinned to a specific version.
func IsVersionPinned(version string) bool {
	// Consider it pinned if it specifies patch version (v1.2.3) or is a full commit SHA
	return IsSemanticVersion(version) || (IsCommitSHA(version) && len(version) == 40)
}

// ValidateGitBranch checks if a branch exists in the given repository.
func ValidateGitBranch(repoRoot, branch string) bool {
	cmd := exec.Command(
		appconstants.GitCommand,
		appconstants.GitShowRef,
		appconstants.GitVerify,
		appconstants.GitQuiet,
		"refs/heads/"+branch,
	) // #nosec G204 -- branch name validated by git
	cmd.Dir = repoRoot

	return cmd.Run() == nil
}

// ValidateActionYMLPath validates that a path points to a valid action.yml file.
func ValidateActionYMLPath(path string) error {
	// Check the file is accessible. Surface every stat failure (not just
	// not-exist) wrapped with %w, so an unreadable/permission-denied path is
	// reported rather than silently treated as valid.
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot access action file %q: %w", path, err)
	}
	// A directory named action.yml/action.yaml would otherwise pass this check
	// and fail later with a less precise error.
	if info.IsDir() {
		return fmt.Errorf("action file path %q is a directory, not a file", path)
	}

	// Check if it's an action.yml or action.yaml file
	filename := filepath.Base(path)
	if filename != appconstants.ActionFileNameYML && filename != appconstants.ActionFileNameYAML {
		return os.ErrInvalid
	}

	return nil
}

// IsGitRepository checks if the given path is within a git repository.
func IsGitRepository(path string) bool {
	_, err := git.FindRepositoryRoot(path)

	return err == nil
}
