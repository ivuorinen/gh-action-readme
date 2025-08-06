// Package dependencies provides GitHub Actions dependency analysis functionality.
package dependencies

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v57/github"

	"github.com/ivuorinen/gh-action-readme/internal/git"
)

// VersionType represents the type of version specification used.
type VersionType string

const (
	// SemanticVersion represents semantic versioning format (v1.2.3).
	SemanticVersion VersionType = "semantic"
	// CommitSHA represents a git commit SHA.
	CommitSHA VersionType = "commit"
	// BranchName represents a git branch reference.
	BranchName VersionType = "branch"
	// LocalPath represents a local file path reference.
	LocalPath VersionType = "local"

	// Common string constants.
	compositeUsing  = "composite"
	updateTypeNone  = "none"
	updateTypeMajor = "major"
	updateTypePatch = "patch"
	updateTypeMinor = "minor"
	defaultBranch   = "main"

	// Timeout constants.
	apiCallTimeout  = 10 * time.Second
	cacheDefaultTTL = 1 * time.Hour

	// File permission constants.
	backupFilePerms  = 0600
	updatedFilePerms = 0600

	// GitHub URL patterns.
	githubBaseURL      = "https://github.com"
	marketplaceBaseURL = "https://github.com/marketplace/actions/"

	// Version parsing constants.
	fullSHALength     = 40
	minSHALength      = 7
	versionPartsCount = 3

	// File path patterns.
	dockerPrefix      = "docker://"
	localPathPrefix   = "./"
	localPathUpPrefix = "../"

	// File extensions.
	backupExtension = ".backup"

	// Cache key prefixes.
	cacheKeyLatest = "latest:"
	cacheKeyRepo   = "repo:"

	// YAML structure constants.
	usesFieldPrefix = "uses: "

	// Special line estimation for script URLs.
	scriptLineEstimate = 10
)

// Dependency represents a GitHub Action dependency with detailed information.
type Dependency struct {
	Name           string            `json:"name"`
	Uses           string            `json:"uses"`         // Full uses statement
	Version        string            `json:"version"`      // Readable version
	VersionType    VersionType       `json:"version_type"` // semantic, commit, branch
	IsPinned       bool              `json:"is_pinned"`    // Whether locked to specific version
	Description    string            `json:"description"`  // From GitHub API
	Author         string            `json:"author"`       // Action owner
	MarketplaceURL string            `json:"marketplace_url,omitempty"`
	SourceURL      string            `json:"source_url"`
	WithParams     map[string]string `json:"with_params,omitempty"`
	IsLocalAction  bool              `json:"is_local_action"` // Same repo dependency
	IsShellScript  bool              `json:"is_shell_script"`
	ScriptURL      string            `json:"script_url,omitempty"` // Link to script line
}

// OutdatedDependency represents a dependency that has newer versions available.
type OutdatedDependency struct {
	Current          Dependency `json:"current"`
	LatestVersion    string     `json:"latest_version"`
	LatestSHA        string     `json:"latest_sha"`
	UpdateType       string     `json:"update_type"` // "major", "minor", "patch"
	Changelog        string     `json:"changelog,omitempty"`
	IsSecurityUpdate bool       `json:"is_security_update"`
}

// PinnedUpdate represents an update that pins to a specific commit SHA.
type PinnedUpdate struct {
	FilePath   string `json:"file_path"`
	OldUses    string `json:"old_uses"` // "actions/checkout@v4"
	NewUses    string `json:"new_uses"` // "actions/checkout@8f4b7f84...# v4.1.1"
	CommitSHA  string `json:"commit_sha"`
	Version    string `json:"version"`
	UpdateType string `json:"update_type"` // "major", "minor", "patch"
	LineNumber int    `json:"line_number"`
}

// Analyzer analyzes GitHub Action dependencies.
type Analyzer struct {
	GitHubClient *github.Client
	Cache        DependencyCache // High-performance cache interface
	RepoInfo     git.RepoInfo
}

// DependencyCache defines the caching interface for dependency data.
type DependencyCache interface {
	Get(key string) (any, bool)
	Set(key string, value any) error
	SetWithTTL(key string, value any, ttl time.Duration) error
}

// Note: Using git.RepoInfo instead of local GitInfo to avoid duplication

// NewAnalyzer creates a new dependency analyzer.
func NewAnalyzer(client *github.Client, repoInfo git.RepoInfo, cache DependencyCache) *Analyzer {
	return &Analyzer{
		GitHubClient: client,
		Cache:        cache,
		RepoInfo:     repoInfo,
	}
}

// AnalyzeActionFile analyzes dependencies from an action.yml file.
func (a *Analyzer) AnalyzeActionFile(actionPath string) ([]Dependency, error) {
	return a.AnalyzeActionFileWithProgress(actionPath, nil)
}

// AnalyzeActionFileWithProgress analyzes dependencies with optional progress tracking.
func (a *Analyzer) AnalyzeActionFileWithProgress(
	actionPath string,
	progressCallback func(current, total int, message string),
) ([]Dependency, error) {
	if progressCallback != nil {
		progressCallback(0, 1, "Parsing "+actionPath)
	}

	// Read and parse the action.yml file
	action, err := a.parseCompositeAction(actionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse action file: %w", err)
	}

	// Validate and check if it's a composite action
	deps, isComposite, err := a.validateAndCheckComposite(action, progressCallback)
	if err != nil {
		return nil, err
	}
	if !isComposite {
		return deps, nil
	}

	// Process composite action steps
	return a.processCompositeSteps(action.Runs.Steps, progressCallback)
}

// CheckOutdated analyzes dependencies and finds those with newer versions available.
func (a *Analyzer) CheckOutdated(deps []Dependency) ([]OutdatedDependency, error) {
	var outdated []OutdatedDependency

	for _, dep := range deps {
		if dep.IsShellScript || dep.IsLocalAction {
			continue // Skip shell scripts and local actions
		}

		owner, repo, currentVersion, _ := a.parseUsesStatement(dep.Uses)
		if owner == "" || repo == "" {
			continue
		}

		latestVersion, latestSHA, err := a.getLatestVersion(owner, repo)
		if err != nil {
			continue // Skip on error, don't fail the whole operation
		}

		updateType := a.compareVersions(currentVersion, latestVersion)
		if updateType != updateTypeNone {
			outdated = append(outdated, OutdatedDependency{
				Current:          dep,
				LatestVersion:    latestVersion,
				LatestSHA:        latestSHA,
				UpdateType:       updateType,
				IsSecurityUpdate: updateType == updateTypeMajor, // Assume major updates might be security
			})
		}
	}

	return outdated, nil
}

// GeneratePinnedUpdate creates a pinned update for a dependency.
func (a *Analyzer) GeneratePinnedUpdate(
	actionPath string,
	dep Dependency,
	latestVersion, latestSHA string,
) (*PinnedUpdate, error) {
	if latestSHA == "" {
		return nil, fmt.Errorf("no commit SHA available for %s", dep.Uses)
	}

	// Create the new pinned uses string: "owner/repo@sha # version"
	owner, repo, currentVersion, _ := a.parseUsesStatement(dep.Uses)
	newUses := fmt.Sprintf("%s/%s@%s # %s", owner, repo, latestSHA, latestVersion)

	updateType := a.compareVersions(currentVersion, latestVersion)

	return &PinnedUpdate{
		FilePath:   actionPath,
		OldUses:    dep.Uses,
		NewUses:    newUses,
		CommitSHA:  latestSHA,
		Version:    latestVersion,
		UpdateType: updateType,
		LineNumber: 0, // Will be determined during file update
	}, nil
}

// ApplyPinnedUpdates applies pinned updates to action files.
func (a *Analyzer) ApplyPinnedUpdates(updates []PinnedUpdate) error {
	// Group updates by file path
	updatesByFile := make(map[string][]PinnedUpdate)
	for _, update := range updates {
		updatesByFile[update.FilePath] = append(updatesByFile[update.FilePath], update)
	}

	// Apply updates to each file
	for filePath, fileUpdates := range updatesByFile {
		if err := a.updateActionFile(filePath, fileUpdates); err != nil {
			return fmt.Errorf("failed to update %s: %w", filePath, err)
		}
	}

	return nil
}

// validateAndCheckComposite validates action type and checks if it's composite.
func (a *Analyzer) validateAndCheckComposite(
	action *ActionWithComposite,
	progressCallback func(current, total int, message string),
) ([]Dependency, bool, error) {
	if action.Runs.Using != compositeUsing {
		if err := a.validateActionType(action.Runs.Using); err != nil {
			return nil, false, err
		}
		if progressCallback != nil {
			progressCallback(1, 1, "No dependencies (non-composite action)")
		}

		return []Dependency{}, false, nil
	}

	return nil, true, nil
}

// validateActionType checks if the action type is valid.
func (a *Analyzer) validateActionType(usingType string) error {
	validTypes := []string{"node20", "node16", "node12", "docker", "composite"}
	for _, validType := range validTypes {
		if usingType == validType {
			return nil
		}
	}

	return fmt.Errorf("invalid action runtime: %s", usingType)
}

// processCompositeSteps processes steps in a composite action.
func (a *Analyzer) processCompositeSteps(
	steps []CompositeStep,
	progressCallback func(current, total int, message string),
) ([]Dependency, error) {
	var dependencies []Dependency
	totalSteps := len(steps)

	for i, step := range steps {
		if progressCallback != nil {
			progressCallback(i, totalSteps, fmt.Sprintf("Analyzing step %d/%d", i+1, totalSteps))
		}

		dep := a.processStep(step, i+1)
		if dep != nil {
			dependencies = append(dependencies, *dep)
		}
	}

	if progressCallback != nil {
		progressCallback(totalSteps, totalSteps, fmt.Sprintf("Found %d dependencies", len(dependencies)))
	}

	return dependencies, nil
}

// processStep processes a single step and returns dependency if found.
func (a *Analyzer) processStep(step CompositeStep, stepNumber int) *Dependency {
	if step.Uses != "" {
		// This is an action dependency
		dep, err := a.analyzeActionDependency(step, stepNumber)
		if err != nil {
			// Log error but continue processing
			return nil
		}

		return dep
	} else if step.Run != "" {
		// This is a shell script step
		return a.analyzeShellScript(step, stepNumber)
	}

	return nil
}

// analyzeActionDependency analyzes a single action dependency.
func (a *Analyzer) analyzeActionDependency(step CompositeStep, _ int) (*Dependency, error) {
	// Parse the uses statement
	owner, repo, version, versionType := a.parseUsesStatement(step.Uses)
	if owner == "" || repo == "" {
		return nil, fmt.Errorf("invalid uses statement: %s", step.Uses)
	}

	// Check if it's a local action (same repository)
	isLocal := (owner == a.RepoInfo.Organization && repo == a.RepoInfo.Repository)

	// Build dependency
	dep := &Dependency{
		Name:          fmt.Sprintf("%s/%s", owner, repo),
		Uses:          step.Uses,
		Version:       version,
		VersionType:   versionType,
		IsPinned:      versionType == CommitSHA || (versionType == SemanticVersion && a.isVersionPinned(version)),
		Author:        owner,
		SourceURL:     fmt.Sprintf("%s/%s/%s", githubBaseURL, owner, repo),
		IsLocalAction: isLocal,
		IsShellScript: false,
		WithParams:    a.convertWithParams(step.With),
	}

	// Add marketplace URL for public actions
	if !isLocal {
		dep.MarketplaceURL = marketplaceBaseURL + repo
	}

	// Fetch additional metadata from GitHub API if available
	if a.GitHubClient != nil && !isLocal {
		_ = a.enrichWithGitHubData(dep, owner, repo) // Ignore error - we have basic info
	}

	return dep, nil
}

// analyzeShellScript analyzes a shell script step.
func (a *Analyzer) analyzeShellScript(step CompositeStep, stepNumber int) *Dependency {
	// Create a shell script dependency
	name := step.Name
	if name == "" {
		name = fmt.Sprintf("Shell Script #%d", stepNumber)
	}

	// Try to create a link to the script in the repository
	scriptURL := ""
	if a.RepoInfo.Organization != "" && a.RepoInfo.Repository != "" {
		// This would ideally link to the specific line in the action.yml file
		scriptURL = fmt.Sprintf(
			"%s/%s/%s/blob/%s/action.yml#L%d",
			githubBaseURL,
			a.RepoInfo.Organization,
			a.RepoInfo.Repository,
			a.RepoInfo.DefaultBranch,
			stepNumber*scriptLineEstimate,
		) // Rough estimate
	}

	return &Dependency{
		Name:          name,
		Uses:          "", // No uses for shell scripts
		Version:       "",
		VersionType:   LocalPath,
		IsPinned:      true, // Shell scripts are always "pinned"
		Description:   "Shell script execution",
		Author:        a.RepoInfo.Organization,
		SourceURL:     scriptURL,
		WithParams:    map[string]string{},
		IsLocalAction: true,
		IsShellScript: true,
		ScriptURL:     scriptURL,
	}
}

// parseUsesStatement parses a GitHub Action uses statement.
func (a *Analyzer) parseUsesStatement(uses string) (owner, repo, version string, versionType VersionType) {
	// Handle different uses statement formats:
	// - actions/checkout@v4
	// - actions/checkout@main
	// - actions/checkout@8f4b7f84bd579b95d7f0b90f8d8b6e5d9b8a7f6e
	// - ./local-action
	// - docker://alpine:3.14

	if strings.HasPrefix(uses, localPathPrefix) || strings.HasPrefix(uses, localPathUpPrefix) {
		return "", "", uses, LocalPath
	}

	if strings.HasPrefix(uses, dockerPrefix) {
		return "", "", uses, LocalPath
	}

	// Standard GitHub action format: owner/repo@version
	re := regexp.MustCompile(`^([^/]+)/([^@]+)@(.+)$`)
	matches := re.FindStringSubmatch(uses)
	if len(matches) != 4 {
		return "", "", "", LocalPath
	}

	owner = matches[1]
	repo = matches[2]
	version = matches[3]

	// Determine version type
	switch {
	case a.isCommitSHA(version):
		versionType = CommitSHA
	case a.isSemanticVersion(version):
		versionType = SemanticVersion
	default:
		versionType = BranchName
	}

	return owner, repo, version, versionType
}

// isCommitSHA checks if a version string is a commit SHA.
func (a *Analyzer) isCommitSHA(version string) bool {
	// Check if it's a 40-character hex string (full SHA) or 7+ character hex (short SHA)
	re := regexp.MustCompile(`^[a-f0-9]{7,40}$`)

	return len(version) >= minSHALength && re.MatchString(version)
}

// isSemanticVersion checks if a version string follows semantic versioning.
func (a *Analyzer) isSemanticVersion(version string) bool {
	// Check for vX, vX.Y, vX.Y.Z format
	re := regexp.MustCompile(`^v?\d+(\.\d+)*(\.\d+)?(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$`)

	return re.MatchString(version)
}

// isVersionPinned checks if a semantic version is pinned to a specific version.
func (a *Analyzer) isVersionPinned(version string) bool {
	// Consider it pinned if it specifies patch version (v1.2.3) or is a commit SHA
	// Also check for full commit SHAs (40 chars)
	if len(version) == fullSHALength {
		return true
	}
	re := regexp.MustCompile(`^v?\d+\.\d+\.\d+`)

	return re.MatchString(version)
}

// convertWithParams converts with parameters to string map.
func (a *Analyzer) convertWithParams(with map[string]any) map[string]string {
	params := make(map[string]string)
	for k, v := range with {
		if str, ok := v.(string); ok {
			params[k] = str
		} else {
			params[k] = fmt.Sprintf("%v", v)
		}
	}

	return params
}

// getLatestVersion fetches the latest release/tag for a repository.
func (a *Analyzer) getLatestVersion(owner, repo string) (version, sha string, err error) {
	if a.GitHubClient == nil {
		return "", "", errors.New("GitHub client not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), apiCallTimeout)
	defer cancel()

	// Check cache first
	cacheKey := cacheKeyLatest + fmt.Sprintf("%s/%s", owner, repo)
	if version, sha, found := a.getCachedVersion(cacheKey); found {
		return version, sha, nil
	}

	// Try to get latest release first
	if version, sha, err := a.getLatestRelease(ctx, owner, repo); err == nil {
		a.cacheVersion(cacheKey, version, sha)

		return version, sha, nil
	}

	// Fallback to latest tag
	version, sha, err = a.getLatestTag(ctx, owner, repo)
	if err != nil {
		return "", "", err
	}

	a.cacheVersion(cacheKey, version, sha)

	return version, sha, nil
}

// getCachedVersion retrieves version info from cache if available.
func (a *Analyzer) getCachedVersion(cacheKey string) (version, sha string, found bool) {
	if a.Cache == nil {
		return "", "", false
	}

	cached, exists := a.Cache.Get(cacheKey)
	if !exists {
		return "", "", false
	}

	versionInfo, ok := cached.(map[string]string)
	if !ok {
		return "", "", false
	}

	return versionInfo["version"], versionInfo["sha"], true
}

// getLatestRelease fetches the latest release and its commit SHA.
func (a *Analyzer) getLatestRelease(ctx context.Context, owner, repo string) (version, sha string, err error) {
	release, _, err := a.GitHubClient.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil || release.GetTagName() == "" {
		return "", "", errors.New("no release found")
	}

	version = release.GetTagName()
	sha = a.getCommitSHAForTag(ctx, owner, repo, version)

	return version, sha, nil
}

// getCommitSHAForTag retrieves the commit SHA for a given tag.
func (a *Analyzer) getCommitSHAForTag(ctx context.Context, owner, repo, tagName string) string {
	tag, _, err := a.GitHubClient.Git.GetRef(ctx, owner, repo, "tags/"+tagName)
	if err != nil || tag.GetObject() == nil {
		return ""
	}

	return tag.GetObject().GetSHA()
}

// getLatestTag fetches the most recent tag and its commit SHA.
func (a *Analyzer) getLatestTag(ctx context.Context, owner, repo string) (version, sha string, err error) {
	tags, _, err := a.GitHubClient.Repositories.ListTags(ctx, owner, repo, &github.ListOptions{
		PerPage: 10,
	})
	if err != nil || len(tags) == 0 {
		return "", "", errors.New("no releases or tags found")
	}

	latestTag := tags[0]

	return latestTag.GetName(), latestTag.GetCommit().GetSHA(), nil
}

// cacheVersion stores version information in cache with TTL.
func (a *Analyzer) cacheVersion(cacheKey, version, sha string) {
	if a.Cache == nil {
		return
	}

	versionInfo := map[string]string{"version": version, "sha": sha}
	_ = a.Cache.SetWithTTL(cacheKey, versionInfo, cacheDefaultTTL)
}

// compareVersions compares two version strings and returns the update type.
func (a *Analyzer) compareVersions(current, latest string) string {
	currentClean := strings.TrimPrefix(current, "v")
	latestClean := strings.TrimPrefix(latest, "v")

	if currentClean == latestClean {
		return updateTypeNone
	}

	// Special case: floating major version (e.g., "4" -> "4.1.1") should be patch
	if !strings.Contains(currentClean, ".") && strings.HasPrefix(latestClean, currentClean+".") {
		return updateTypePatch
	}

	currentParts := a.parseVersionParts(currentClean)
	latestParts := a.parseVersionParts(latestClean)

	return a.determineUpdateType(currentParts, latestParts)
}

// parseVersionParts normalizes version string to 3-part semantic version.
func (a *Analyzer) parseVersionParts(version string) []string {
	parts := strings.Split(version, ".")
	// For floating versions like "v4", treat as "v4.0.0" for comparison
	for len(parts) < versionPartsCount {
		parts = append(parts, "0")
	}

	return parts
}

// determineUpdateType compares version parts and returns update type.
func (a *Analyzer) determineUpdateType(currentParts, latestParts []string) string {
	if currentParts[0] != latestParts[0] {
		return updateTypeMajor
	}
	if currentParts[1] != latestParts[1] {
		return updateTypeMinor
	}
	if currentParts[2] != latestParts[2] {
		return updateTypePatch
	}

	return updateTypeNone
}

// updateActionFile applies updates to a single action file.
func (a *Analyzer) updateActionFile(filePath string, updates []PinnedUpdate) error {
	// Read the file
	content, err := os.ReadFile(filePath) // #nosec G304 -- file path from function parameter
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Create backup
	backupPath := filePath + backupExtension
	if err := os.WriteFile(backupPath, content, backupFilePerms); err != nil { // #nosec G306 -- backup file permissions
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Apply updates to content
	lines := strings.Split(string(content), "\n")
	for _, update := range updates {
		// Find and replace the uses line
		for i, line := range lines {
			if strings.Contains(line, update.OldUses) {
				// Replace the uses statement while preserving indentation
				indent := strings.Repeat(" ", len(line)-len(strings.TrimLeft(line, " ")))
				lines[i] = indent + usesFieldPrefix + update.NewUses
				update.LineNumber = i + 1 // Store line number for reference

				break
			}
		}
	}

	// Write updated content
	updatedContent := strings.Join(lines, "\n")
	if err := os.WriteFile(filePath, []byte(updatedContent), updatedFilePerms); err != nil {
		// #nosec G306 -- updated file permissions
		return fmt.Errorf("failed to write updated file: %w", err)
	}

	// Validate the updated file by trying to parse it
	if err := a.validateActionFile(filePath); err != nil {
		// Rollback on validation failure
		if rollbackErr := os.Rename(backupPath, filePath); rollbackErr != nil {
			return fmt.Errorf("validation failed and rollback failed: %w (original error: %w)", rollbackErr, err)
		}

		return fmt.Errorf("validation failed, rolled back changes: %w", err)
	}

	// Remove backup on success
	_ = os.Remove(backupPath)

	return nil
}

// validateActionFile validates that an action.yml file is still valid after updates.
func (a *Analyzer) validateActionFile(filePath string) error {
	_, err := a.parseCompositeAction(filePath)

	return err
}

// enrichWithGitHubData fetches additional information from GitHub API.
func (a *Analyzer) enrichWithGitHubData(dep *Dependency, owner, repo string) error {
	ctx, cancel := context.WithTimeout(context.Background(), apiCallTimeout)
	defer cancel()

	// Check cache first
	cacheKey := cacheKeyRepo + fmt.Sprintf("%s/%s", owner, repo)
	if a.Cache != nil {
		if cached, exists := a.Cache.Get(cacheKey); exists {
			if repository, ok := cached.(*github.Repository); ok {
				dep.Description = repository.GetDescription()

				return nil
			}
		}
	}

	// Fetch from API
	repository, _, err := a.GitHubClient.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return fmt.Errorf("failed to fetch repository info: %w", err)
	}

	// Cache the result with 1 hour TTL
	if a.Cache != nil {
		_ = a.Cache.SetWithTTL(cacheKey, repository, cacheDefaultTTL) // Ignore cache errors
	}

	// Enrich dependency with API data
	dep.Description = repository.GetDescription()

	return nil
}
