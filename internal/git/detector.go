// Package git provides Git repository detection and information extraction.
package git

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// DefaultBranch is the default branch name used as fallback.
	DefaultBranch = "main"
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
		return fmt.Sprintf("%s/%s", r.Organization, r.Repository)
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
		gitPath := filepath.Join(absPath, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return absPath, nil
		}

		parent := filepath.Dir(absPath)
		if parent == absPath {
			// Reached root without finding .git
			return "", fmt.Errorf("not a git repository")
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
	gitPath := filepath.Join(repoRoot, ".git")
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
	if defaultBranch, err := getDefaultBranch(repoRoot); err == nil {
		info.DefaultBranch = defaultBranch
	}

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
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = repoRoot

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL from git: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// getRemoteURLFromConfig parses .git/config to extract remote URL.
func getRemoteURLFromConfig(repoRoot string) (string, error) {
	configPath := filepath.Join(repoRoot, ".git", "config")
	file, err := os.Open(configPath)
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
		if inOriginSection && strings.HasPrefix(line, "url = ") {
			return strings.TrimPrefix(line, "url = "), nil
		}
	}

	return "", fmt.Errorf("no origin remote URL found in git config")
}

// getDefaultBranch gets the default branch name.
func getDefaultBranch(repoRoot string) (string, error) {
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	cmd.Dir = repoRoot

	output, err := cmd.Output()
	if err != nil {
		// Fallback to common default branches
		for _, branch := range []string{DefaultBranch, "master"} {
			if branchExists(repoRoot, branch) {
				return branch, nil
			}
		}
		return DefaultBranch, nil // Default fallback
	}

	// Extract branch name from refs/remotes/origin/HEAD -> refs/remotes/origin/main
	parts := strings.Split(strings.TrimSpace(string(output)), "/")
	if len(parts) > 0 {
		return parts[len(parts)-1], nil
	}

	return DefaultBranch, nil
}

// branchExists checks if a branch exists in the repository.
func branchExists(repoRoot, branch string) bool {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	cmd.Dir = repoRoot
	return cmd.Run() == nil
}

// parseGitHubURL extracts organization and repository name from various GitHub URL formats.
func parseGitHubURL(url string) (organization, repository string) {
	// Common GitHub URL patterns
	patterns := []string{
		`github\.com[:/]([^/]+)/([^/\.]+)`,     // github.com:org/repo or github.com/org/repo
		`github\.com[:/]([^/]+)/([^/]+)\.git$`, // github.com:org/repo.git or github.com/org/repo.git
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(url)
		if len(matches) >= 3 {
			org := matches[1]
			repo := matches[2]

			// Remove .git suffix if present
			repo = strings.TrimSuffix(repo, ".git")

			return org, repo
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
