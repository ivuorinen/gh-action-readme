// Package dependencies provides GitHub Actions dependency analysis functionality.
package dependencies

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v74/github"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/git"
)

// Package-level compiled regexps for performance (compiled once at startup).
var (
	reUsesStatement   = regexp.MustCompile(`^([^/]+)/([^@]+)@(.+)$`)
	reGitSHA          = regexp.MustCompile(appconstants.RegexGitSHA)
	reSemanticVersion = regexp.MustCompile(`^v?\d+(\.\d+)*(\.\d+)?(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$`)
	rePinnedVersion   = regexp.MustCompile(`^v?\d+\.\d+\.\d+`)
	// reSafeTagName accepts only plain git tag characters, blocking newline/YAML
	// injection from a tag name that gets embedded into a pinned uses-statement.
	reSafeTagName = regexp.MustCompile(`^[A-Za-z0-9._/+-]+$`)
)

// validActionRuntimes is the authoritative list of valid GitHub Actions runtime identifiers.
var validActionRuntimes = []string{
	appconstants.NodeRuntimeNode12,
	appconstants.NodeRuntimeNode16,
	appconstants.NodeRuntimeNode20,
	appconstants.NodeRuntimeNode24,
	appconstants.ActionTypeDocker,
	appconstants.ActionTypeComposite,
}

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
		if updateType != appconstants.UpdateTypeNone {
			outdated = append(outdated, OutdatedDependency{
				Current:       dep,
				LatestVersion: latestVersion,
				LatestSHA:     latestSHA,
				UpdateType:    updateType,
				// Don't assume major version bumps are security updates
				// This should only be set if confirmed by security advisory data
				// Future enhancement: integrate with GitHub Security Advisories API
				IsSecurityUpdate: false,
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

	// Defense in depth: latestSHA and latestVersion originate from the GitHub API
	// and are written verbatim into the user's action.yml. Require a full 40-char
	// commit SHA and a plain tag name so a compromised or proxied response cannot
	// inject newlines or extra YAML into the pinned reference.
	if len(latestSHA) != appconstants.FullSHALength || !reGitSHA.MatchString(latestSHA) {
		return nil, fmt.Errorf("refusing to pin %s: invalid commit SHA %q", dep.Uses, latestSHA)
	}
	if !reSafeTagName.MatchString(latestVersion) {
		return nil, fmt.Errorf("refusing to pin %s: invalid version tag %q", dep.Uses, latestVersion)
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
	if action.Runs.Using != appconstants.ActionTypeComposite {
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
	for _, validType := range validActionRuntimes {
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
		Name:          fmt.Sprintf(appconstants.URLPatternGitHubRepo, owner, repo),
		Uses:          step.Uses,
		Version:       version,
		VersionType:   versionType,
		IsPinned:      versionType == CommitSHA || (versionType == SemanticVersion && a.isVersionPinned(version)),
		Author:        owner,
		SourceURL:     fmt.Sprintf("%s/%s/%s", appconstants.GitHubBaseURL, owner, repo),
		IsLocalAction: isLocal,
		IsShellScript: false,
		WithParams:    a.convertWithParams(step.With),
	}

	// Add marketplace URL for public actions
	if !isLocal {
		dep.MarketplaceURL = fmt.Sprintf("%s%s/%s", appconstants.MarketplaceBaseURL, owner, repo)
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
			appconstants.GitHubBaseURL,
			a.RepoInfo.Organization,
			a.RepoInfo.Repository,
			a.RepoInfo.DefaultBranch,
			stepNumber*appconstants.ScriptLineEstimate,
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

	if strings.HasPrefix(uses, appconstants.LocalPathPrefix) ||
		strings.HasPrefix(uses, appconstants.LocalPathUpPrefix) {
		return "", "", uses, LocalPath
	}

	if strings.HasPrefix(uses, appconstants.DockerPrefix) {
		return "", "", uses, LocalPath
	}

	// Standard GitHub action format: owner/repo@version
	matches := reUsesStatement.FindStringSubmatch(uses)
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
	return len(version) >= appconstants.MinSHALength && reGitSHA.MatchString(version)
}

// isSemanticVersion checks if a version string follows semantic versioning.
func (a *Analyzer) isSemanticVersion(version string) bool {
	// Check for vX, vX.Y, vX.Y.Z format
	return reSemanticVersion.MatchString(version)
}

// isVersionPinned checks if a semantic version is pinned to a specific version.
func (a *Analyzer) isVersionPinned(version string) bool {
	// Consider it pinned if it specifies patch version (v1.2.3) or is a commit SHA
	// Also check for full commit SHAs (40 chars)
	if len(version) == appconstants.FullSHALength {
		return true
	}

	return rePinnedVersion.MatchString(version)
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

	ctx, cancel := context.WithTimeout(context.Background(), appconstants.APICallTimeout)
	defer cancel()

	// Check cache first
	cacheKey := appconstants.CacheKeyLatest + fmt.Sprintf(appconstants.URLPatternGitHubRepo, owner, repo)
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
//
// Annotated tags (used by virtually every major action repository) point at a
// tag object, not a commit; GetRef returns that tag-object SHA. GitHub Actions
// only accepts commit SHAs in `uses:` pins, so an annotated tag's tag-object SHA
// would produce a broken pin. When the ref resolves to a tag object, dereference
// it via GetTag to obtain the underlying commit SHA.
func (a *Analyzer) getCommitSHAForTag(ctx context.Context, owner, repo, tagName string) string {
	ref, _, err := a.GitHubClient.Git.GetRef(ctx, owner, repo, "tags/"+tagName)
	if err != nil || ref.GetObject() == nil {
		return ""
	}

	obj := ref.GetObject()
	if obj.GetType() == appconstants.GitObjectTypeTag {
		if tagObj, _, tagErr := a.GitHubClient.Git.GetTag(ctx, owner, repo, obj.GetSHA()); tagErr == nil &&
			tagObj.GetObject() != nil {
			return tagObj.GetObject().GetSHA()
		}
		// Dereferencing the annotated tag failed. The ref's object SHA is the
		// tag-object SHA, not a commit, and GitHub Actions rejects it as a `uses:`
		// pin. Return "" so callers reject the update instead of writing a broken
		// pin that a 40-char-hex validation cannot distinguish from a commit SHA.
		return ""
	}

	return obj.GetSHA()
}

// getLatestTag fetches the most recent tag and its commit SHA.
func (a *Analyzer) getLatestTag(ctx context.Context, owner, repo string) (version, sha string, err error) {
	tags, _, err := a.GitHubClient.Repositories.ListTags(ctx, owner, repo, &github.ListOptions{
		PerPage: 10,
	})
	if err != nil {
		// Surface the real API error (rate limit, auth, network) rather than
		// masking it as "no tags", so callers can report the actual cause.
		return "", "", fmt.Errorf("failed to list tags: %w", err)
	}
	if len(tags) == 0 {
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
	_ = a.Cache.SetWithTTL(cacheKey, versionInfo, appconstants.CacheDefaultTTL)
}

// compareVersions compares two version strings and returns the update type.
func (a *Analyzer) compareVersions(current, latest string) string {
	currentClean := strings.TrimPrefix(current, "v")
	latestClean := strings.TrimPrefix(latest, "v")

	if currentClean == latestClean {
		return appconstants.UpdateTypeNone
	}

	// Special case: floating major version (e.g., "4" -> "4.1.1") should be patch
	if !strings.Contains(currentClean, ".") && strings.HasPrefix(latestClean, currentClean+".") {
		return appconstants.UpdateTypePatch
	}

	currentParts := a.parseVersionParts(currentClean)
	latestParts := a.parseVersionParts(latestClean)

	return a.determineUpdateType(currentParts, latestParts)
}

// parseVersionParts normalizes version string to 3-part semantic version.
func (a *Analyzer) parseVersionParts(version string) []string {
	parts := strings.Split(version, ".")
	// For floating versions like "v4", treat as "v4.0.0" for comparison
	for len(parts) < appconstants.VersionPartsCount {
		parts = append(parts, "0")
	}

	return parts
}

// determineUpdateType compares version parts and returns update type.
func (a *Analyzer) determineUpdateType(currentParts, latestParts []string) string {
	if currentParts[0] != latestParts[0] {
		return appconstants.UpdateTypeMajor
	}
	if currentParts[1] != latestParts[1] {
		return appconstants.UpdateTypeMinor
	}
	if currentParts[2] != latestParts[2] {
		return appconstants.UpdateTypePatch
	}

	return appconstants.UpdateTypeNone
}

// updateActionFile applies updates to a single action file.
func (a *Analyzer) updateActionFile(filePath string, updates []PinnedUpdate) error {
	// filepath.Clean normalises the path (removes redundant separators, ".", "..").
	// It does NOT validate containment within a root directory; the actual security
	// justification for the #nosec annotations below is that filePath originates
	// from the tool's own filesystem discovery (DiscoverActionFilesWithValidation),
	// not from direct, uncontrolled user input.
	cleanPath := filepath.Clean(filePath)

	// Read the file
	content, err := os.ReadFile(cleanPath) // #nosec G304 -- path from tool-internal filesystem scan
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Create backup
	backupPath := cleanPath + appconstants.BackupExtension
	if err := os.WriteFile( // #nosec G306 G703 -- path from tool-internal filesystem scan
		backupPath,
		content,
		appconstants.FilePermDefault,
	); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Apply updates to content
	lines := strings.Split(string(content), "\n")
	applyUpdatesToLines(lines, updates)

	// Write updated content
	updatedContent := strings.Join(lines, "\n")
	if err := os.WriteFile( // #nosec G306 G703 -- path from tool-internal filesystem scan
		cleanPath,
		[]byte(updatedContent),
		appconstants.FilePermDefault,
	); err != nil {
		// Do not remove backupPath here — it is the recovery copy for this failure.
		return fmt.Errorf("failed to write updated file: %w", err)
	}

	// Validate and rollback on failure
	if err := a.validateAndRollbackOnFailure(cleanPath, backupPath); err != nil {
		return err
	}

	// Remove backup on success
	_ = os.Remove(backupPath)

	return nil
}

// applyUpdatesToLines applies all updates to the file lines in place.
// Preserves indentation and YAML list markers.
func applyUpdatesToLines(lines []string, updates []PinnedUpdate) {
	for _, update := range updates {
		target := appconstants.UsesFieldPrefix + update.OldUses

		for i, line := range lines {
			// Skip comment lines
			if strings.HasPrefix(strings.TrimLeft(line, " \t"), "#") {
				continue
			}

			idx := strings.Index(line, target)
			if idx < 0 {
				continue
			}

			// Skip if the match is inside an inline comment (# before the match position)
			if commentIdx := strings.Index(line, "#"); commentIdx >= 0 && commentIdx < idx {
				continue
			}

			// Require OldUses to be a complete token — not a prefix of a longer version string
			afterTarget := strings.TrimLeft(line[idx+len(target):], " \t")
			if afterTarget != "" && !strings.HasPrefix(afterTarget, "#") {
				continue
			}

			// Preserve both indentation AND list markers. Capture the original
			// leading whitespace verbatim rather than re-emitting spaces, so any
			// tabs or mixed indentation in the source line survive the rewrite.
			trimmed := strings.TrimLeft(line, " \t")
			indent := line[:len(line)-len(trimmed)]

			// Check if this is a list item (starts with "- ")
			listMarker := ""
			if strings.HasPrefix(trimmed, "- ") {
				listMarker = "- "
			}

			// Reconstruct: indent + list marker + uses field
			lines[i] = indent + listMarker + appconstants.UsesFieldPrefix + update.NewUses
		}
	}
}

// validateAndRollbackOnFailure validates the action file and rolls back changes on failure.
func (a *Analyzer) validateAndRollbackOnFailure(filePath, backupPath string) error {
	if err := a.validateActionFile(filePath); err != nil {
		// Rollback on validation failure
		if rollbackErr := os.Rename(backupPath, filePath); rollbackErr != nil {
			return fmt.Errorf("validation failed and rollback failed: %w (original error: %w)", rollbackErr, err)
		}

		return fmt.Errorf("validation failed, rolled back changes: %w", err)
	}

	return nil
}

// validateActionFile validates that an action.yml file conforms to GitHub Actions schema.
// Schema reference: https://www.schemastore.org/github-action.json
func (a *Analyzer) validateActionFile(filePath string) error {
	// Parse to check YAML syntax
	action, err := a.parseCompositeAction(filePath)
	if err != nil {
		return err
	}

	// Validate required fields per GitHub Actions schema
	if action.Name == "" {
		return errors.New("validation failed: missing required field 'name'")
	}
	if action.Description == "" {
		return errors.New("validation failed: missing required field 'description'")
	}
	if action.Runs.Using == "" {
		return errors.New("validation failed: missing required field 'runs.using'")
	}

	validUsing := false
	runtime := strings.TrimSpace(strings.ToLower(action.Runs.Using))
	for _, valid := range validActionRuntimes {
		if runtime == valid {
			validUsing = true

			break
		}
	}

	if !validUsing {
		return fmt.Errorf(
			"validation failed: invalid value for 'runs.using': %s (valid: %s)",
			action.Runs.Using,
			strings.Join(validActionRuntimes, ", "),
		)
	}

	return nil
}

// enrichWithGitHubData fetches additional information from GitHub API.
func (a *Analyzer) enrichWithGitHubData(dep *Dependency, owner, repo string) error {
	ctx, cancel := context.WithTimeout(context.Background(), appconstants.APICallTimeout)
	defer cancel()

	// Check cache first
	cacheKey := appconstants.CacheKeyRepo + fmt.Sprintf("%s/%s", owner, repo)
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
		_ = a.Cache.SetWithTTL(cacheKey, repository, appconstants.CacheDefaultTTL) // Ignore cache errors
	}

	// Enrich dependency with API data
	dep.Description = repository.GetDescription()

	return nil
}
