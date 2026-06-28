package dependencies

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v74/github"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/cache"
	"github.com/ivuorinen/gh-action-readme/internal/git"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

const (
	testDepOwnerActions = "actions"
	testDepRepoCheckout = "checkout"
	testDepVersionV3    = "v3"
	testDepVersionV4    = "v4"
)

// analyzeActionFileTestCase describes a single test case for AnalyzeActionFile.
type analyzeActionFileTestCase struct {
	name         string
	actionYML    string
	expectError  bool
	expectDeps   bool
	expectedLen  int
	expectedDeps []string
}

// runAnalyzeActionFileTest executes a single test case with setup, analysis, and validation.
func runAnalyzeActionFileTest(t *testing.T, tt analyzeActionFileTestCase) {
	t.Helper()

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	actionPath := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
	testutil.WriteTestFile(t, actionPath, tt.actionYML)

	mockResponses := testutil.MockGitHubResponses()
	githubClient := testutil.MockGitHubClient(mockResponses)
	cacheInstance := newIsolatedCache(t)

	analyzer := &Analyzer{
		GitHubClient: githubClient,
		Cache:        NewCacheAdapter(cacheInstance),
	}

	deps, err := analyzer.AnalyzeActionFile(actionPath)

	if tt.expectError {
		testutil.AssertError(t, err)

		return
	}
	testutil.AssertNoError(t, err)

	validateAnalyzedDependencies(t, tt, deps)
}

// validateAnalyzedDependencies checks that analyzed dependencies match expectations.
func validateAnalyzedDependencies(t *testing.T, tt analyzeActionFileTestCase, deps []Dependency) {
	t.Helper()

	if tt.expectDeps {
		validateExpectedDeps(t, tt, deps)
	} else if len(deps) != 0 {
		t.Errorf("expected no dependencies, got %d", len(deps))
	}
}

// validateExpectedDeps validates dependencies when deps are expected.
func validateExpectedDeps(t *testing.T, tt analyzeActionFileTestCase, deps []Dependency) {
	t.Helper()

	if len(deps) != tt.expectedLen {
		t.Errorf("expected %d dependencies, got %d", tt.expectedLen, len(deps))
	}

	if tt.expectedDeps == nil {
		return
	}

	for i, expectedDep := range tt.expectedDeps {
		if i >= len(deps) {
			t.Errorf("expected dependency %s but got fewer dependencies", expectedDep)

			continue
		}
		if !strings.Contains(deps[i].Name+"@"+deps[i].Version, expectedDep) {
			t.Errorf("expected dependency %s, got %s@%s", expectedDep, deps[i].Name, deps[i].Version)
		}
	}
}

func TestAnalyzerAnalyzeActionFile(t *testing.T) {
	t.Parallel()

	tests := []analyzeActionFileTestCase{
		{
			name:        "simple action - no dependencies",
			actionYML:   testutil.MustReadFixture(testutil.TestFixtureJavaScriptSimple),
			expectError: false,
			expectDeps:  false,
			expectedLen: 0,
		},
		{
			name:         "composite action with dependencies",
			actionYML:    testutil.MustReadFixture(testutil.TestFixtureCompositeWithDeps),
			expectError:  false,
			expectDeps:   true,
			expectedLen:  5,
			expectedDeps: []string{testutil.TestActionCheckoutV4, "actions/setup-node@v4", "actions/setup-python@v4"},
		},
		{
			name:        "docker action - no step dependencies",
			actionYML:   testutil.MustReadFixture(testutil.TestFixtureDockerBasic),
			expectError: false,
			expectDeps:  false,
			expectedLen: 0,
		},
		{
			name:        testutil.TestCaseNameInvalidActionFile,
			actionYML:   testutil.MustReadFixture(testutil.TestFixtureInvalidInvalidUsing),
			expectError: true,
		},
		{
			name:        "minimal action - no dependencies",
			actionYML:   testutil.MustReadFixture(testutil.TestFixtureMinimalAction),
			expectError: false,
			expectDeps:  false,
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			runAnalyzeActionFileTest(t, tt)
		})
	}
}

// TestAnalyzeActionFileLocalAndDockerSteps verifies N141: composite steps that
// reference local (./) and docker:// actions are emitted as dependencies instead
// of being silently dropped.
func TestAnalyzeActionFileLocalAndDockerSteps(t *testing.T) {
	t.Parallel()

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	actionPath := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
	testutil.WriteTestFile(t, actionPath, testutil.MustReadFixture(testutil.TestFixtureLocalDockerSteps))

	analyzer := &Analyzer{} // no GitHub client: local/docker/remote all parse without enrichment
	deps, err := analyzer.AnalyzeActionFile(actionPath)
	testutil.AssertNoError(t, err)

	var sawLocal, sawDocker, sawRemote bool
	for _, d := range deps {
		switch d.Uses {
		case "./setup":
			sawLocal = d.IsLocalAction && d.VersionType == LocalPath
		case "docker://alpine:3.14":
			sawDocker = d.VersionType == LocalPath && !d.IsLocalAction
		case testutil.TestActionCheckoutV4:
			sawRemote = true
		}
	}

	if !sawLocal {
		t.Error("local ./setup step was dropped or mislabeled")
	}
	if !sawDocker {
		t.Error("docker:// step was dropped or mislabeled")
	}
	if !sawRemote {
		t.Error("remote checkout step missing")
	}
}

func TestAnalyzerParseUsesStatement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		uses            string
		expectedOwner   string
		expectedRepo    string
		expectedVersion string
		expectedType    VersionType
	}{
		{
			name:            testutil.TestCaseNameSemanticVersion,
			uses:            testutil.TestActionCheckoutV4,
			expectedOwner:   testDepOwnerActions,
			expectedRepo:    testDepRepoCheckout,
			expectedVersion: testDepVersionV4,
			expectedType:    SemanticVersion,
		},
		{
			name:            "semantic version with patch",
			uses:            "actions/setup-node@v3.8.1",
			expectedOwner:   testDepOwnerActions,
			expectedRepo:    "setup-node",
			expectedVersion: "v3.8.1",
			expectedType:    SemanticVersion,
		},
		{
			name:            testutil.TestCaseNameCommitSHA,
			uses:            "actions/checkout@8f4b7f84bd579b95d7f0b90f8d8b6e5d9b8a7f6e",
			expectedOwner:   testDepOwnerActions,
			expectedRepo:    testDepRepoCheckout,
			expectedVersion: testutil.TestSHAForTesting,
			expectedType:    CommitSHA,
		},
		{
			name:            "branch reference",
			uses:            "octocat/hello-world@main",
			expectedOwner:   "octocat",
			expectedRepo:    "hello-world",
			expectedVersion: "main",
			expectedType:    BranchName,
		},
		{
			// N140: monorepo subpath action — repo must be the bare repository so
			// API/URL lookups resolve; the subpath stays in Uses.
			name:            "monorepo subpath action",
			uses:            "actions/cache/restore@v4",
			expectedOwner:   testDepOwnerActions,
			expectedRepo:    "cache",
			expectedVersion: testDepVersionV4,
			expectedType:    SemanticVersion,
		},
		{
			name:            "monorepo subpath action with branch",
			uses:            "github/codeql-action/analyze@" + testutil.TestBranchMain,
			expectedOwner:   "github",
			expectedRepo:    "codeql-action",
			expectedVersion: testutil.TestBranchMain,
			expectedType:    BranchName,
		},
	}

	analyzer := &Analyzer{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			owner, repo, version, versionType := analyzer.parseUsesStatement(tt.uses)

			testutil.AssertEqual(t, tt.expectedOwner, owner)
			testutil.AssertEqual(t, tt.expectedRepo, repo)
			testutil.AssertEqual(t, tt.expectedVersion, version)
			testutil.AssertEqual(t, tt.expectedType, versionType)
		})
	}
}

func TestAnalyzerVersionChecking(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		version     string
		isPinned    bool
		isCommitSHA bool
		isSemantic  bool
	}{
		{
			name:        "semantic version major",
			version:     testDepVersionV4,
			isPinned:    false,
			isCommitSHA: false,
			isSemantic:  true,
		},
		{
			name:        "semantic version full",
			version:     "v3.8.1",
			isPinned:    true,
			isCommitSHA: false,
			isSemantic:  true,
		},
		{
			name:        "commit SHA full",
			version:     testutil.TestSHAForTesting,
			isPinned:    true,
			isCommitSHA: true,
			isSemantic:  false,
		},
		{
			name:        "commit SHA short",
			version:     "8f4b7f8",
			isPinned:    false,
			isCommitSHA: true,
			isSemantic:  false,
		},
		{
			name:        "branch reference",
			version:     "main",
			isPinned:    false,
			isCommitSHA: false,
			isSemantic:  false,
		},
		{
			name:        "numeric version",
			version:     "1.2.3",
			isPinned:    true,
			isCommitSHA: false,
			isSemantic:  true,
		},
	}

	analyzer := &Analyzer{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			isPinned := analyzer.isVersionPinned(tt.version)
			isCommitSHA := analyzer.isCommitSHA(tt.version)
			isSemantic := analyzer.isSemanticVersion(tt.version)

			testutil.AssertEqual(t, tt.isPinned, isPinned)
			testutil.AssertEqual(t, tt.isCommitSHA, isCommitSHA)
			testutil.AssertEqual(t, tt.isSemantic, isSemantic)
		})
	}
}

func TestAnalyzerGetLatestVersion(t *testing.T) {
	t.Parallel()

	// Create mock GitHub client with test responses
	mockResponses := testutil.MockGitHubResponses()
	githubClient := testutil.MockGitHubClient(mockResponses)
	cacheInstance := newIsolatedCache(t)

	analyzer := &Analyzer{
		GitHubClient: githubClient,
		Cache:        cacheInstance,
	}

	tests := []struct {
		name            string
		owner           string
		repo            string
		expectedVersion string
		expectedSHA     string
		expectError     bool
	}{
		{
			name:            "valid repository",
			owner:           testDepOwnerActions,
			repo:            testDepRepoCheckout,
			expectedVersion: testutil.TestVersionV4_1_1,
			expectedSHA:     testutil.TestSHAForTesting,
			expectError:     false,
		},
		{
			name:            "another valid repository",
			owner:           testDepOwnerActions,
			repo:            "setup-node",
			expectedVersion: testutil.TestVersionV4_0_0,
			expectedSHA:     "1a4e6d7c9f8e5b2a3c4d5e6f7a8b9c0d1e2f3a4b",
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			version, sha, err := analyzer.getLatestVersion(tt.owner, tt.repo)

			if tt.expectError {
				testutil.AssertError(t, err)

				return
			}

			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, tt.expectedVersion, version)
			testutil.AssertEqual(t, tt.expectedSHA, sha)
		})
	}
}

func TestAnalyzerCheckOutdated(t *testing.T) {
	t.Parallel()

	// Create mock GitHub client
	mockResponses := testutil.MockGitHubResponses()
	githubClient := testutil.MockGitHubClient(mockResponses)
	cacheInstance := newIsolatedCache(t)

	analyzer := &Analyzer{
		GitHubClient: githubClient,
		Cache:        cacheInstance,
	}

	// Create test dependencies
	dependencies := []Dependency{
		{
			Name:        testutil.TestActionCheckout,
			Uses:        testutil.TestActionCheckoutV3,
			Version:     testDepVersionV3,
			IsPinned:    false,
			VersionType: SemanticVersion,
			Description: "Action for checking out a repo",
		},
		{
			Name:        "actions/setup-node",
			Uses:        "actions/setup-node@v4.0.0",
			Version:     testutil.TestVersionV4_0_0,
			IsPinned:    true,
			VersionType: SemanticVersion,
			Description: "Setup Node.js",
		},
	}

	outdated, err := analyzer.CheckOutdated(dependencies)
	testutil.AssertNoError(t, err)

	// Should detect that actions/checkout v3 is outdated (latest is v4.1.1)
	if len(outdated) == 0 {
		t.Error("expected to find outdated dependencies")
	}

	found := false
	for _, dep := range outdated {
		if dep.Current.Name == testutil.TestActionCheckout && dep.Current.Version == testDepVersionV3 {
			found = true
			if dep.LatestVersion != testutil.TestVersionV4_1_1 {
				t.Errorf("expected latest version v4.1.1, got %s", dep.LatestVersion)
			}
			if dep.UpdateType != appconstants.UpdateTypeMajor {
				t.Errorf("expected major update, got %s", dep.UpdateType)
			}
		}
	}

	if !found {
		t.Error("expected to find actions/checkout v3 as outdated")
	}
}

func TestAnalyzerCompareVersions(t *testing.T) {
	t.Parallel()

	analyzer := &Analyzer{}

	tests := []struct {
		name         string
		current      string
		latest       string
		expectedType string
	}{
		{
			name:         "major version difference",
			current:      "v3.0.0",
			latest:       testutil.TestVersionV4_0_0,
			expectedType: appconstants.UpdateTypeMajor,
		},
		{
			name:         "minor version difference",
			current:      testutil.TestVersionV4_0_0,
			latest:       "v4.1.0",
			expectedType: appconstants.UpdateTypeMinor,
		},
		{
			name:         "patch version difference",
			current:      "v4.1.0",
			latest:       testutil.TestVersionV4_1_1,
			expectedType: appconstants.UpdateTypePatch,
		},
		{
			name:         "no difference",
			current:      testutil.TestVersionV4_1_1,
			latest:       testutil.TestVersionV4_1_1,
			expectedType: appconstants.UpdateTypeNone,
		},
		{
			name:         "floating to specific",
			current:      testDepVersionV4,
			latest:       testutil.TestVersionV4_1_1,
			expectedType: appconstants.UpdateTypePatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			updateType := analyzer.compareVersions(tt.current, tt.latest)
			testutil.AssertEqual(t, tt.expectedType, updateType)
		})
	}
}

// TestAnalyzerGeneratePinnedUpdatePreservesSubpath verifies N148: pinning a
// monorepo sub-action keeps the subpath instead of collapsing the reference to
// the repo-root action.
func TestAnalyzerGeneratePinnedUpdatePreservesSubpath(t *testing.T) {
	t.Parallel()

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()
	actionPath := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
	testutil.WriteTestFile(t, actionPath, testutil.MustReadFixture(testutil.TestFixtureTestCompositeAction))

	analyzer := &Analyzer{} // GeneratePinnedUpdate formats from its args; no client/cache needed

	dep := Dependency{
		Name:        "github.com/github/codeql-action",
		Uses:        "github/codeql-action/analyze@v3",
		Version:     "v3",
		VersionType: SemanticVersion,
	}

	update, err := analyzer.GeneratePinnedUpdate(
		actionPath,
		dep,
		testutil.TestVersionV4_1_1,
		testutil.TestSHAForTesting,
	)
	testutil.AssertNoError(t, err)

	testutil.AssertStringContains(t, update.NewUses, "github/codeql-action/analyze@")
	testutil.AssertEqual(t, "github/codeql-action/analyze@v3", update.OldUses)
}

// TestAnalyzerOutdatedUpdateType verifies that a commit-SHA pin is classified by
// SHA equality rather than semver-compared (N136). A SHA pin equal to the latest
// release SHA is up to date; a differing one is a digest update — never a bogus
// "major". Non-SHA versions keep semver classification.
func TestAnalyzerOutdatedUpdateType(t *testing.T) {
	t.Parallel()

	analyzer := &Analyzer{}
	const sha = "8f4b7f84bd579b95d7f0b90f8d8b6e5d9b8a7f6e"
	const otherSHA = "1111111111111111111111111111111111111111"

	tests := []struct {
		name           string
		versionType    VersionType
		currentVersion string
		latestVersion  string
		latestSHA      string
		want           string
	}{
		{
			name:           "sha pin equal to latest sha is current",
			versionType:    CommitSHA,
			currentVersion: sha,
			latestVersion:  testutil.TestVersionV4_1_1,
			latestSHA:      sha,
			want:           appconstants.UpdateTypeNone,
		},
		{
			name:           "sha pin differing from latest sha is a digest update",
			versionType:    CommitSHA,
			currentVersion: otherSHA,
			latestVersion:  testutil.TestVersionV4_1_1,
			latestSHA:      sha,
			want:           appconstants.UpdateTypeDigest,
		},
		{
			name:           "sha pin with unknown latest sha cannot be compared",
			versionType:    CommitSHA,
			currentVersion: otherSHA,
			latestVersion:  testutil.TestVersionV4_1_1,
			latestSHA:      "",
			want:           appconstants.UpdateTypeNone,
		},
		{
			// N149: an abbreviated SHA pin that is a prefix of the latest SHA is
			// current, not perpetually "digest".
			name:           "short sha pin matching latest prefix is current",
			versionType:    CommitSHA,
			currentVersion: "8f4b7f8",
			latestVersion:  testutil.TestVersionV4_1_1,
			latestSHA:      sha,
			want:           appconstants.UpdateTypeNone,
		},
		{
			name:           "short sha pin not matching latest is a digest update",
			versionType:    CommitSHA,
			currentVersion: "deadbee",
			latestVersion:  testutil.TestVersionV4_1_1,
			latestSHA:      sha,
			want:           appconstants.UpdateTypeDigest,
		},
		{
			name:           "semver still classified by version",
			versionType:    SemanticVersion,
			currentVersion: "v3.0.0",
			latestVersion:  testutil.TestVersionV4_0_0,
			latestSHA:      sha,
			want:           appconstants.UpdateTypeMajor,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := analyzer.outdatedUpdateType(tt.versionType, tt.currentVersion, tt.latestVersion, tt.latestSHA)
			testutil.AssertEqual(t, tt.want, got)
		})
	}
}

func TestAnalyzerGeneratePinnedUpdate(t *testing.T) {
	t.Parallel()

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Create a test action file with composite steps
	actionContent := testutil.MustReadFixture(testutil.TestFixtureTestCompositeAction)

	actionPath := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
	testutil.WriteTestFile(t, actionPath, actionContent)

	// Create analyzer
	mockResponses := testutil.MockGitHubResponses()
	githubClient := testutil.MockGitHubClient(mockResponses)
	cacheInstance := newIsolatedCache(t)

	analyzer := &Analyzer{
		GitHubClient: githubClient,
		Cache:        cacheInstance,
	}

	// Create test dependency
	dep := Dependency{
		Name:        testutil.TestActionCheckout,
		Uses:        testutil.TestActionCheckoutV3,
		Version:     testDepVersionV3,
		IsPinned:    false,
		VersionType: SemanticVersion,
		Description: "Action for checking out a repo",
	}

	// Generate pinned update
	update, err := analyzer.GeneratePinnedUpdate(
		actionPath,
		dep,
		testutil.TestVersionV4_1_1,
		testutil.TestSHAForTesting,
	)

	testutil.AssertNoError(t, err)

	// Verify update details
	testutil.AssertEqual(t, actionPath, update.FilePath)
	testutil.AssertEqual(t, testutil.TestActionCheckoutV3, update.OldUses)
	testutil.AssertStringContains(t, update.NewUses, "actions/checkout@8f4b7f84bd579b95d7f0b90f8d8b6e5d9b8a7f6e")
	testutil.AssertStringContains(t, update.NewUses, "# v4.1.1")
	testutil.AssertEqual(t, appconstants.UpdateTypeMajor, update.UpdateType)
}

func TestAnalyzerWithCache(t *testing.T) {
	t.Parallel()

	// Test that caching works properly
	mockResponses := testutil.MockGitHubResponses()
	githubClient := testutil.MockGitHubClient(mockResponses)
	cacheInstance := newIsolatedCache(t)

	analyzer := &Analyzer{
		GitHubClient: githubClient,
		Cache:        cacheInstance,
	}

	// First call should hit the API
	version1, sha1, err1 := analyzer.getLatestVersion(testDepOwnerActions, testDepRepoCheckout)
	testutil.AssertNoError(t, err1)

	// Second call should hit the cache
	version2, sha2, err2 := analyzer.getLatestVersion(testDepOwnerActions, testDepRepoCheckout)
	testutil.AssertNoError(t, err2)

	// Results should be identical
	testutil.AssertEqual(t, version1, version2)
	testutil.AssertEqual(t, sha1, sha2)
}

func TestAnalyzerRateLimitHandling(t *testing.T) {
	t.Parallel()

	rlHeader := http.Header{
		"X-RateLimit-Remaining": []string{"0"},
		"X-RateLimit-Reset":     []string{strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10)},
	}
	// Separate responses: an http.Response Body can only be read once, so the
	// release and tag lookups need their own.
	rateLimitRelease := &http.Response{
		StatusCode: http.StatusForbidden, Header: rlHeader,
		Body: testutil.NewStringReader(`{"message": "API rate limit exceeded"}`),
	}
	rateLimitTags := &http.Response{
		StatusCode: http.StatusForbidden, Header: rlHeader,
		Body: testutil.NewStringReader(`{"message": "API rate limit exceeded"}`),
	}

	// Rate-limit BOTH the release and the tag-list endpoints (tags is the
	// fallback path) so the surfaced error genuinely reflects the rate limit.
	// ListTags adds ?per_page=10, so the mock key must include it.
	mockClient := &testutil.MockHTTPClient{
		Responses: map[string]*http.Response{
			"GET https://api.github.com/repos/actions/checkout/releases/latest":  rateLimitRelease,
			"GET https://api.github.com/repos/actions/checkout/tags?per_page=10": rateLimitTags,
		},
	}

	client := github.NewClient(&http.Client{Transport: &testutil.MockTransport{Client: mockClient}})
	cacheInstance := newIsolatedCache(t)

	analyzer := &Analyzer{
		GitHubClient: client,
		Cache:        cacheInstance,
	}

	// This should handle the rate limit gracefully
	_, _, err := analyzer.getLatestVersion(testDepOwnerActions, testDepRepoCheckout)
	if err == nil {
		t.Error("expected rate limit error to be returned")
	}

	// With both endpoints rate-limited, the surfaced error must reflect the rate
	// limit rather than being masked as a generic "no tags" message.
	if !strings.Contains(err.Error(), "rate limit") {
		t.Errorf("expected error to surface rate limit info, got: %s", err.Error())
	}
}

func TestAnalyzerWithoutGitHubClient(t *testing.T) {
	t.Parallel()

	// Test graceful degradation when GitHub client is not available
	analyzer := &Analyzer{
		GitHubClient: nil,
		Cache:        nil,
	}

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	actionPath := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
	testutil.WriteTestFile(t, actionPath, testutil.MustReadFixture(testutil.TestFixtureCompositeBasic))

	deps, err := analyzer.AnalyzeActionFile(actionPath)

	// Should still parse dependencies but without GitHub API data
	testutil.AssertNoError(t, err)
	if len(deps) > 0 {
		// Dependencies should have basic info but no GitHub API data
		for _, dep := range deps {
			// Only check action dependencies (not shell scripts which have hardcoded descriptions)
			if !dep.IsShellScript && dep.Description != "" {
				t.Error("expected empty description when GitHub client is not available")
			}
		}
	}
}

// TestNewAnalyzer tests the analyzer constructor.
func TestNewAnalyzer(t *testing.T) {
	t.Parallel()

	// Create test dependencies
	mockResponses := testutil.MockGitHubResponses()
	githubClient := testutil.MockGitHubClient(mockResponses)
	cacheInstance, err := cache.NewCache(isolatedCacheConfig(t))
	testutil.AssertNoError(t, err)
	defer testutil.CleanupCache(t, cacheInstance)()

	repoInfo := git.RepoInfo{
		Organization: "test-owner",
		Repository:   "test-repo",
	}

	tests := []struct {
		name         string
		client       *github.Client
		repoInfo     git.RepoInfo
		cache        DependencyCache
		expectNotNil bool
	}{
		{
			name:         "creates analyzer with all dependencies",
			client:       githubClient,
			repoInfo:     repoInfo,
			cache:        NewCacheAdapter(cacheInstance),
			expectNotNil: true,
		},
		{
			name:         "creates analyzer with nil client",
			client:       nil,
			repoInfo:     repoInfo,
			cache:        NewCacheAdapter(cacheInstance),
			expectNotNil: true,
		},
		{
			name:         "creates analyzer with nil cache",
			client:       githubClient,
			repoInfo:     repoInfo,
			cache:        nil,
			expectNotNil: true,
		},
		{
			name:         "creates analyzer with empty repo info",
			client:       githubClient,
			repoInfo:     git.RepoInfo{},
			cache:        NewCacheAdapter(cacheInstance),
			expectNotNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			analyzer := NewAnalyzer(tt.client, tt.repoInfo, tt.cache)

			if tt.expectNotNil && analyzer == nil {
				t.Fatal("expected non-nil analyzer")
			}

			// Verify fields are set correctly
			if analyzer.GitHubClient != tt.client {
				t.Error("GitHub client not set correctly")
			}
			if analyzer.Cache != tt.cache {
				t.Error("cache not set correctly")
			}
			if analyzer.RepoInfo != tt.repoInfo {
				t.Error("repo info not set correctly")
			}
		})
	}
}

// TestNoOpCache tests the no-op cache implementation.
func TestNoOpCache(t *testing.T) {
	t.Parallel()

	noc := NewNoOpCache()
	if noc == nil {
		t.Fatal("NewNoOpCache() returned nil")
	}

	// Test Get - should always return false
	val, ok := noc.Get(testutil.CacheTestKey)
	if ok {
		t.Error("NoOpCache.Get() should return false")
	}
	if val != nil {
		t.Error("NoOpCache.Get() should return nil value")
	}

	// Test Set - should not error
	err := noc.Set(testutil.CacheTestKey, testutil.CacheTestValue)
	if err != nil {
		t.Errorf("NoOpCache.Set() returned error: %v", err)
	}

	// Test SetWithTTL - should not error
	err = noc.SetWithTTL(testutil.CacheTestKey, testutil.CacheTestValue, time.Hour)
	if err != nil {
		t.Errorf("NoOpCache.SetWithTTL() returned error: %v", err)
	}
}

// TestCacheAdapterSet tests the cache adapter Set method.
func TestCacheAdapterSet(t *testing.T) {
	t.Parallel()

	c, err := cache.NewCache(isolatedCacheConfig(t))
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}
	defer testutil.CleanupCache(t, c)()

	adapter := NewCacheAdapter(c)

	// Test Set
	err = adapter.Set(testutil.CacheTestKey, testutil.CacheTestValue)
	if err != nil {
		t.Errorf("CacheAdapter.Set() returned error: %v", err)
	}

	// Verify value was set
	val, ok := adapter.Get(testutil.CacheTestKey)
	if !ok {
		t.Error("CacheAdapter.Get() should return true after Set")
	}
	if val != testutil.CacheTestValue {
		t.Errorf("CacheAdapter.Get() = %v, want %q", val, testutil.CacheTestValue)
	}
}

// TestIsCompositeAction tests composite action detection.
func TestValidateActionType(t *testing.T) {
	t.Parallel()

	analyzer, cleanup := newTestAnalyzer(t)
	defer cleanup()

	validTypes := []string{
		appconstants.NodeRuntimeNode24,
		appconstants.NodeRuntimeNode20,
		appconstants.NodeRuntimeNode16,
		appconstants.NodeRuntimeNode12,
		appconstants.ActionTypeDocker,
		appconstants.ActionTypeComposite,
	}
	for _, typ := range validTypes {
		if err := analyzer.validateActionType(typ); err != nil {
			t.Errorf("validateActionType(%q) unexpected error: %v", typ, err)
		}
	}

	if err := analyzer.validateActionType("node18"); err == nil {
		t.Error("validateActionType(\"node18\") expected error, got nil")
	}
}

func TestIsCompositeAction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		fixture string
		want    bool
		wantErr bool
	}{
		{
			name:    testutil.TestCaseNameCompositeAction,
			fixture: testutil.TestFixtureCompositeActionAnalyzer,
			want:    true,
			wantErr: false,
		},
		{
			name:    "docker action",
			fixture: "docker-action.yml",
			want:    false,
			wantErr: false,
		},
		{
			name:    testutil.TestCaseNameJavaScriptAction,
			fixture: "javascript-action.yml",
			want:    false,
			wantErr: false,
		},
		{
			name:    testutil.TestCaseNameInvalidYAML,
			fixture: "invalid.yml",
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Read fixture content using safe helper
			yamlContent := testutil.MustReadAnalyzerFixture(tt.fixture)

			// Create temp file with action YAML
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			actionPath := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
			testutil.WriteTestFile(t, actionPath, yamlContent)

			got, err := IsCompositeAction(actionPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsCompositeAction() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if got != tt.want {
				t.Errorf("IsCompositeAction() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetCommitSHAForTag_AnnotatedTag verifies that an annotated tag (ref object
// type "tag") is dereferenced via GetTag to the underlying commit SHA, instead of
// returning the unusable tag-object SHA that would produce a broken `uses:` pin.
func TestGetCommitSHAForTag_AnnotatedTag(t *testing.T) {
	t.Parallel()

	const (
		annTag       = "v9.9.9"
		annTagObjSHA = "1111111111111111111111111111111111111111"
		annCommitSHA = "2222222222222222222222222222222222222222"
	)

	base := "GET https://api.github.com/repos/annorg/annrepo"
	refJSON := `{"ref":"refs/tags/` + annTag + `","object":{"type":"tag","sha":"` + annTagObjSHA + `"}}`
	tagJSON := `{"sha":"` + annTagObjSHA + `","object":{"type":"commit","sha":"` + annCommitSHA + `"}}`
	mockResponses := map[string]string{
		base + "/releases/latest":          `{"tag_name": "` + annTag + `"}`,
		base + "/git/ref/tags/" + annTag:   refJSON,
		base + "/git/tags/" + annTagObjSHA: tagJSON,
	}

	cacheInstance := newIsolatedCache(t)
	analyzer := &Analyzer{
		GitHubClient: testutil.MockGitHubClient(mockResponses),
		Cache:        cacheInstance,
	}

	version, sha, err := analyzer.getLatestVersion("annorg", "annrepo")
	if err != nil {
		t.Fatalf("getLatestVersion returned error: %v", err)
	}
	if version != annTag {
		t.Errorf("version = %q, want %q", version, annTag)
	}
	if sha != annCommitSHA {
		t.Errorf("sha = %q, want dereferenced commit SHA %q (annotated tag not peeled)", sha, annCommitSHA)
	}
}

// TestGetCommitSHAForTag_AnnotatedTagDerefFailure verifies that when an annotated
// tag's GetTag dereference fails (e.g. rate limit/404 between the GetRef and
// GetTag calls), getCommitSHAForTag returns "" rather than the unusable
// tag-object SHA. Returning the tag-object SHA would produce a broken `uses:` pin
// that the 40-char-hex validation cannot catch.
func TestGetCommitSHAForTag_AnnotatedTagDerefFailure(t *testing.T) {
	t.Parallel()

	const (
		annTag       = "v9.9.9"
		annTagObjSHA = "1111111111111111111111111111111111111111"
	)

	base := "GET https://api.github.com/repos/annorg/annrepo"
	refJSON := `{"ref":"refs/tags/` + annTag + `","object":{"type":"tag","sha":"` + annTagObjSHA + `"}}`
	// Note: the "/git/tags/<sha>" endpoint is intentionally NOT mocked, so the
	// mock client returns 404 for GetTag and the dereference fails.
	mockResponses := map[string]string{
		base + "/releases/latest":        `{"tag_name": "` + annTag + `"}`,
		base + "/git/ref/tags/" + annTag: refJSON,
	}

	cacheInstance := newIsolatedCache(t)
	analyzer := &Analyzer{
		GitHubClient: testutil.MockGitHubClient(mockResponses),
		Cache:        cacheInstance,
	}

	version, sha, err := analyzer.getLatestVersion("annorg", "annrepo")
	if err != nil {
		t.Fatalf("getLatestVersion returned error: %v", err)
	}
	if version != annTag {
		t.Errorf("version = %q, want %q", version, annTag)
	}
	if sha != "" {
		t.Errorf("sha = %q, want \"\" (deref failed; must not fall back to tag-object SHA %q)", sha, annTagObjSHA)
	}
}

// TestGeneratePinnedUpdate_RejectsInjection verifies that a malformed/injection-
// bearing SHA or version tag (e.g. from a compromised/proxied API) is refused
// rather than written verbatim into the user's action.yml.
func TestGeneratePinnedUpdate_RejectsInjection(t *testing.T) {
	t.Parallel()

	cacheInstance := newIsolatedCache(t)
	analyzer := &Analyzer{Cache: cacheInstance}
	dep := Dependency{Uses: testutil.TestActionCheckoutV3}

	if _, err := analyzer.GeneratePinnedUpdate(
		"a.yml", dep, "v4", "abc123 # evil\nrun: bad",
	); err == nil {
		t.Error("expected GeneratePinnedUpdate to reject a non-SHA / injection-bearing commit value")
	}

	if _, err := analyzer.GeneratePinnedUpdate(
		"a.yml", dep, "v4\nrun: bad", testutil.TestSHAForTesting,
	); err == nil {
		t.Error("expected GeneratePinnedUpdate to reject an injection-bearing version tag")
	}
}

// TestGetCachedVersionTreatsPartialDataAsMiss ensures incomplete cached entries
// (missing, empty, or non-string version/sha) are reported as a cache miss so the
// analyzer refetches rather than locking onto unusable data.
func TestGetCachedVersionTreatsPartialDataAsMiss(t *testing.T) {
	t.Parallel()

	analyzer := &Analyzer{Cache: NewCacheAdapter(newIsolatedCache(t))}

	const wantVer, wantSHA = "v1", "abc"
	partial := map[string]map[string]any{
		"missing-sha":     {cacheFieldVersion: wantVer},
		"missing-version": {cacheFieldSHA: wantSHA},
		"empty-values":    {cacheFieldVersion: "", cacheFieldSHA: ""},
		"non-string":      {cacheFieldVersion: 1, cacheFieldSHA: wantSHA},
	}
	for key, val := range partial {
		testutil.AssertNoError(t, analyzer.Cache.Set(key, val))
		if _, _, found := analyzer.getCachedVersion(key); found {
			t.Errorf("%s: expected cache miss for partial data, got found=true", key)
		}
	}

	testutil.AssertNoError(t, analyzer.Cache.Set("complete",
		map[string]any{cacheFieldVersion: wantVer, cacheFieldSHA: wantSHA}))
	version, sha, found := analyzer.getCachedVersion("complete")
	if !found || version != wantVer || sha != wantSHA {
		t.Errorf("complete data: got version=%q sha=%q found=%v, want %s/%s/true",
			version, sha, found, wantVer, wantSHA)
	}
}
