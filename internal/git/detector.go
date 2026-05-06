// Package git provides Git repository detection and information extraction.
package git

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

var (
	reGitHubURLNoSuffix   = regexp.MustCompile(`github\.com[:/]([^/]+)/([^/\.]+)`)
	reGitHubURLWithSuffix = regexp.MustCompile(`github\.com[:/]([^/]+)/([^/]+)\.git$`)
)

// RepoInfo contains information about a Git repository.
type RepoInfo struct {
	Organization  string `json:"organization"`
	Repository    string `json:"repository"`
	RemoteURL     string `json:"remote_url"`
	DefaultBranch string `json:"default_branch"`
	IsGitRepo     bool   `json:"is_git_repo"`
}

// GetRepositoryName returns the full repository name in org/repo format.
func (r *RepoInfo) GetRepositoryName() string {
	if r.Organization != "" && r.Repository != "" {
		return fmt.Sprintf(appconstants.URLPatternGitHubRepo, r.Organization, r.Repository)
	}

	return ""
}

// FindRepositoryRoot finds the root directory of a Git repository.
func FindRepositoryRoot(startPath string) (string, error) {
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", err
	}

	// Walk up the directory tree looking for .git
	for {
		gitPath := filepath.Join(absPath, appconstants.DirGit)
		if _, err := os.Stat(gitPath); err == nil {
			return absPath, nil
		}

		parent := filepath.Dir(absPath)
		if parent == absPath {
			// Reached root without finding .git
			return "", errors.New("not a git repository")
		}
		absPath = parent
	}
}

// DetectRepository detects Git repository information from the current directory.
func DetectRepository(repoRoot string) (*RepoInfo, error) {
	if repoRoot == "" {
		return &RepoInfo{IsGitRepo: false}, nil
	}

	// Check if this is actually a git repository
	gitPath := filepath.Join(repoRoot, appconstants.DirGit)
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		return &RepoInfo{IsGitRepo: false}, nil
	}

	info := &RepoInfo{IsGitRepo: true}

	// Try to get remote URL
	remoteURL, err := getRemoteURL(repoRoot)
	if err == nil {
		info.RemoteURL = remoteURL
		org, repo := parseGitHubURL(remoteURL)
		info.Organization = org
		info.Repository = repo
	}

	// Try to get default branch
	info.DefaultBranch = getDefaultBranch(repoRoot)

	return info, nil
}

// getRemoteURL gets the remote URL for the origin remote.
func getRemoteURL(repoRoot string) (string, error) {
	// First try using git command
	if url, err := getRemoteURLFromGit(repoRoot); err == nil {
		return url, nil
	}

	// Fallback to parsing .git/config directly
	return getRemoteURLFromConfig(repoRoot)
}

// getRemoteURLFromGit uses git command to get remote URL.
func getRemoteURLFromGit(repoRoot string) (string, error) {
	cmd := exec.Command(
		appconstants.GitCommand,
		"remote",
		"get-url",
		"origin",
	) // #nosec G204 -- git command is a constant
	cmd.Dir = repoRoot

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL from git: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// getRemoteURLFromConfig parses .git/config to extract remote URL.
func getRemoteURLFromConfig(repoRoot string) (string, error) {
	configPath := filepath.Join(repoRoot, appconstants.DirGit, appconstants.ConfigFileName)
	file, err := os.Open(configPath) // #nosec G304 -- git config path constructed from repo root
	if err != nil {
		return "", fmt.Errorf("failed to open git config: %w", err)
	}
	defer func() {
		_ = file.Close() // File will be closed, error not actionable in defer
	}()

	scanner := bufio.NewScanner(file)
	inOriginSection := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Check for [remote "origin"] section
		if strings.Contains(line, `[remote "origin"]`) {
			inOriginSection = true

			continue
		}

		// Check for new section
		if strings.HasPrefix(line, "[") && inOriginSection {
			inOriginSection = false

			continue
		}

		// Look for url = in origin section
		if inOriginSection && strings.HasPrefix(line, appconstants.GitConfigURL) {
			return strings.TrimPrefix(line, appconstants.GitConfigURL), nil
		}
	}

	return "", errors.New("no origin remote URL found in git config")
}

// getDefaultBranch gets the default branch name.
func getDefaultBranch(repoRoot string) string {
	cmd := exec.Command(
		appconstants.GitCommand,
		"symbolic-ref",
		"refs/remotes/origin/HEAD",
	) // #nosec G204 -- controlled git command
	cmd.Dir = repoRoot

	output, err := cmd.Output()
	if err != nil {
		// Fallback to common default branches
		for _, branch := range []string{appconstants.GitDefaultBranch, "master"} {
			if branchExists(repoRoot, branch) {
				return branch
			}
		}

		return appconstants.GitDefaultBranch // Default fallback
	}

	// Extract branch name from refs/remotes/origin/HEAD -> refs/remotes/origin/main
	parts := strings.Split(strings.TrimSpace(string(output)), "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return appconstants.GitDefaultBranch
}

// branchExists checks if a branch exists in the repository.
func branchExists(repoRoot, branch string) bool {
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

// parseGitHubURL extracts organization and repository name from various GitHub URL formats.
func parseGitHubURL(url string) (organization, repository string) {
	for _, re := range []*regexp.Regexp{reGitHubURLWithSuffix, reGitHubURLNoSuffix} {
		matches := re.FindStringSubmatch(url)
		if len(matches) >= 3 {
			repo := strings.TrimSuffix(matches[2], appconstants.DirGit)

			return matches[1], repo
		}
	}

	return "", ""
}

// GenerateUsesStatement generates a proper uses statement for GitHub Actions.
func (r *RepoInfo) GenerateUsesStatement(actionName, version string) string {
	if r.Organization != "" && r.Repository != "" {
		// For same repository actions, use relative path
		if actionName != "" && actionName != r.Repository {
			return fmt.Sprintf("%s/%s/%s@%s", r.Organization, r.Repository, actionName, version)
		}
		// For repository-level actions
		return fmt.Sprintf("%s/%s@%s", r.Organization, r.Repository, version)
	}

	// Fallback to generic format
	if actionName != "" {
		return fmt.Sprintf("your-org/%s@%s", actionName, version)
	}

	return "your-org/your-action@v1"
}
