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
	// owner/repo@version, with an optional subdirectory for monorepo actions
	// (actions/cache/restore@v4, github/codeql-action/analyze@v3). The subpath is
	// non-capturing: repo must be the bare repository name so GitHub API lookups
	// and source/marketplace URLs resolve (a "cache/restore" repo 404s). The full
	// reference is preserved in Dependency.Uses.
	reUsesStatement   = regexp.MustCompile(`^([^/]+)/([^/@]+)(?:/[^@]+)?@(.+)$`)
	reGitSHA          = regexp.MustCompile(appconstants.RegexGitSHA)
	reSemanticVersion = regexp.MustCompile(`^v?\d+(\.\d+)*(\.\d+)?(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$`)
	rePinnedVersion   = regexp.MustCompile(`^v?\d+\.\d+\.\d+`)
	// reSafeTagName accepts only plain git tag characters, blocking newline/YAML
	// injection from a tag name that gets embedded into a pinned uses-statement.
	reSafeTagName = regexp.MustCompile(`^[A-Za-z0-9._/+-]+$`)
)

// Cache map field keys for the stored version/sha entry (see getCachedVersion/cacheVersion).
const (
	cacheFieldVersion = "version"
	cacheFieldSHA     = "sha"
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
	// Close stops any background cleanup goroutine and flushes pending writes to
	// disk. Callers that construct a cache must Close it (e.g. via Analyzer.Close)
	// so the final async saves are not lost on process exit.
	Close() error
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

// Close releases the analyzer's cache (stopping its background goroutine and
// flushing pending disk writes). It is safe to call on an analyzer with a nil
// cache. Callers that construct an analyzer should defer Close.
func (a *Analyzer) Close() error {
	if a.Cache == nil {
		return nil
	}

	return a.Cache.Close()
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

		owner, repo, currentVersion, versionType := a.parseUsesStatement(dep.Uses)
		if owner == "" || repo == "" {
			continue
		}

		latestVersion, latestSHA, err := a.getLatestVersion(owner, repo)
		if err != nil {
			continue // Skip on error, don't fail the whole operation
		}

		updateType := a.outdatedUpdateType(versionType, currentVersion, latestVersion, latestSHA)
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

	// Create the new pinned uses string, preserving the full reference path
	// (including any monorepo subpath such as codeql-action/analyze). Rebuilding
	// from the bare owner/repo would drop the subpath and silently repoint the pin
	// at the repo-root action (parseUsesStatement returns a bare repo by design).
	_, _, currentVersion, _ := a.parseUsesStatement(dep.Uses)
	refPath := dep.Uses
	if at := strings.LastIndex(refPath, "@"); at >= 0 {
		refPath = refPath[:at]
	}
	newUses := fmt.Sprintf("%s@%s # %s", refPath, latestSHA, latestVersion)

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

// localActionDependency builds a Dependency for a local (./, ../) or docker://
// reference. These have no owner/repo, so no GitHub enrichment or marketplace
// URL is attempted; the full reference is kept in Uses.
func (a *Analyzer) localActionDependency(step CompositeStep) *Dependency {
	name := step.Name
	if name == "" {
		name = step.Uses
	}

	isDocker := strings.HasPrefix(step.Uses, appconstants.DockerPrefix)
	description := "Local action (same repository)"
	if isDocker {
		description = "Docker container action"
	}

	return &Dependency{
		Name:          name,
		Uses:          step.Uses,
		Version:       step.Uses,
		VersionType:   LocalPath,
		IsPinned:      true, // local/docker refs are fixed to the given path/tag
		Description:   description,
		WithParams:    a.convertWithParams(step.With),
		IsLocalAction: !isDocker,
	}
}

// analyzeActionDependency analyzes a single action dependency.
func (a *Analyzer) analyzeActionDependency(step CompositeStep, _ int) (*Dependency, error) {
	// Parse the uses statement
	owner, repo, version, versionType := a.parseUsesStatement(step.Uses)

	// Local (./, ../) and docker:// references resolve to no owner/repo, but they
	// are still real dependencies — emit them instead of dropping the step, which
	// is what returning an error here used to do (processStep discards nil).
	if versionType == LocalPath {
		return a.localActionDependency(step), nil
	}

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

	// Read as map[string]any: an in-memory hit returns the stored map[string]any,
	// and a hit loaded from cache.json is decoded by encoding/json as
	// map[string]interface{} (== map[string]any). Asserting map[string]string
	// here previously failed for every disk-loaded entry, silently defeating the
	// on-disk cache across CLI invocations.
	versionInfo, ok := cached.(map[string]any)
	if !ok {
		return "", "", false
	}

	// Treat incomplete cached data as a miss: a partial entry (missing/non-string/
	// empty version or sha) would otherwise lock the analyzer into unusable data
	// and skip the refetch path.
	version, vOK := versionInfo[cacheFieldVersion].(string)
	sha, sOK := versionInfo[cacheFieldSHA].(string)
	if !vOK || !sOK || version == "" || sha == "" {
		return "", "", false
	}

	return version, sha, true
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

	// Store as map[string]any so the value type is identical whether served from
	// memory or reloaded from cache.json (see getCachedVersion).
	versionInfo := map[string]any{cacheFieldVersion: version, cacheFieldSHA: sha}
	_ = a.Cache.SetWithTTL(cacheKey, versionInfo, appconstants.CacheDefaultTTL)
}

// compareVersions compares two version strings and returns the update type.
// outdatedUpdateType classifies how far a dependency is behind. For commit-SHA
// pins the version string is a 40-char hex SHA, which cannot be semver-compared
// (parseVersionParts would treat the SHA as a major version and always report
// "major"). A SHA pin is current iff it already equals the latest release's SHA;
// otherwise it is behind by an unknowable amount, reported as a digest update.
// When the latest SHA is unknown we cannot compare, so it is treated as current.
func (a *Analyzer) outdatedUpdateType(versionType VersionType, currentVersion, latestVersion, latestSHA string) string {
	if versionType == CommitSHA {
		if latestSHA == "" || shaMatches(currentVersion, latestSHA) {
			return appconstants.UpdateTypeNone
		}

		return appconstants.UpdateTypeDigest
	}

	return a.compareVersions(currentVersion, latestVersion)
}

// shaMatches reports whether a pinned commit SHA refers to the latest SHA. A pin
// may be abbreviated (isCommitSHA accepts 7+ chars), so a shorter current value
// is compared as a case-insensitive prefix of the full 40-char latest SHA rather
// than required to equal it (which would flag every short pin as outdated).
func shaMatches(current, latest string) bool {
	current = strings.ToLower(current)
	latest = strings.ToLower(latest)
	if current == "" {
		return false
	}
	if len(current) < len(latest) {
		return strings.HasPrefix(latest, current)
	}

	return current == latest
}

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
// Prerelease (-suffix) and build metadata (+suffix) are stripped first so that,
// per semver, "1.0.0-beta" and "1.0.0+build" both reduce to the core "1.0.0";
// otherwise the trailing segment ("0-beta") would be string-compared against
// "0" and misclassified as a patch update.
func (a *Analyzer) parseVersionParts(version string) []string {
	parts := strings.Split(coreVersion(version), ".")
	// For floating versions like "v4", treat as "v4.0.0" for comparison
	for len(parts) < appconstants.VersionPartsCount {
		parts = append(parts, "0")
	}

	return parts
}

// coreVersion strips any semver prerelease (-...) or build-metadata (+...)
// suffix, returning just the numeric "major.minor.patch" core.
func coreVersion(version string) string {
	if i := strings.IndexAny(version, "-+"); i >= 0 {
		return version[:i]
	}

	return version
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
		// os.WriteFile opened the original with O_TRUNC, so a mid-write failure
		// (e.g. ENOSPC) has already truncated the user's action.yml. Restore it
		// from the backup instead of leaving it corrupt — mirroring the rollback
		// the validation-failure path below performs.
		if rollbackErr := os.Rename(backupPath, cleanPath); rollbackErr != nil {
			return fmt.Errorf("failed to write updated file: %w (restore from %s also failed: %w)",
				err, backupPath, rollbackErr)
		}

		return fmt.Errorf("failed to write updated file, restored original: %w", err)
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
		for i, line := range lines {
			if rewritten, ok := applyUpdateToLine(line, update); ok {
				lines[i] = rewritten
			}
		}
	}
}

// applyUpdateToLine rewrites a single line if it is a `uses:` field whose value
// equals update.OldUses. The action reference is matched whether or not it is
// wrapped in single or double quotes (a quoted `uses: "actions/checkout@v4"` was
// previously skipped, silently leaving the dependency unpinned). The original
// leading whitespace and YAML list marker are preserved. Returns the rewritten
// line and true when a rewrite occurred, otherwise the original line and false.
func applyUpdateToLine(line string, update PinnedUpdate) (string, bool) {
	trimmed := strings.TrimLeft(line, " \t")
	if strings.HasPrefix(trimmed, "#") {
		return line, false // whole-line comment
	}

	// Capture the original leading whitespace verbatim so tabs/mixed indentation
	// survive the rewrite.
	indent := line[:len(line)-len(trimmed)]

	listMarker := ""
	body := trimmed
	if strings.HasPrefix(body, "- ") {
		listMarker = "- "
		body = strings.TrimLeft(body[len(listMarker):], " \t")
	}

	if !strings.HasPrefix(body, appconstants.UsesFieldPrefix) {
		return line, false
	}

	value := strings.TrimSpace(body[len(appconstants.UsesFieldPrefix):])
	// Drop any trailing inline comment before comparing the reference value.
	if ci := strings.Index(value, "#"); ci >= 0 {
		value = strings.TrimSpace(value[:ci])
	}
	value = unquoteYAMLScalar(value)

	if value != update.OldUses {
		return line, false
	}

	return indent + listMarker + appconstants.UsesFieldPrefix + update.NewUses, true
}

// unquoteYAMLScalar strips a single matching pair of surrounding single or double
// quotes from a YAML scalar; otherwise returns the input unchanged.
func unquoteYAMLScalar(s string) string {
	if len(s) >= 2 {
		first, last := s[0], s[len(s)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			return s[1 : len(s)-1]
		}
	}

	return s
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
			// Cache only the description string (the sole field consumed here).
			// A string survives the cache.json JSON round-trip as a string, unlike
			// the *github.Repository SDK struct, whose type assertion would fail
			// for every disk-loaded entry.
			if description, ok := cached.(string); ok {
				dep.Description = description

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
		// Ignore cache errors.
		_ = a.Cache.SetWithTTL(cacheKey, repository.GetDescription(), appconstants.CacheDefaultTTL)
	}

	// Enrich dependency with API data
	dep.Description = repository.GetDescription()

	return nil
}
