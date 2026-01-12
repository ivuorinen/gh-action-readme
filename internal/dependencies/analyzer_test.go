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

func TestAnalyzerAnalyzeActionFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		actionYML    string
		expectError  bool
		expectDeps   bool
		expectedLen  int
		expectedDeps []string
	}{
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
			expectedLen:  5, // 3 action dependencies + 2 shell script dependencies
			expectedDeps: []string{"actions/checkout@v4", "actions/setup-node@v4", "actions/setup-python@v4"},
		},
		{
			name:        "docker action - no step dependencies",
			actionYML:   testutil.MustReadFixture(testutil.TestFixtureDockerBasic),
			expectError: false,
			expectDeps:  false,
			expectedLen: 0,
		},
		{
			name:        "invalid action file",
			actionYML:   testutil.MustReadFixture(testutil.TestFixtureInvalidInvalidUsing),
			expectError: true,
		},
		{
			name:        "minimal action - no dependencies",
			actionYML:   testutil.MustReadFixture("minimal-action.yml"),
			expectError: false,
			expectDeps:  false,
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temporary action file
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			actionPath := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
			testutil.WriteTestFile(t, actionPath, tt.actionYML)

			// Create analyzer with mock GitHub client
			mockResponses := testutil.MockGitHubResponses()
			githubClient := testutil.MockGitHubClient(mockResponses)
			cacheInstance, _ := cache.NewCache(cache.DefaultConfig())

			analyzer := &Analyzer{
				GitHubClient: githubClient,
				Cache:        NewCacheAdapter(cacheInstance),
			}

			// Analyze the action file
			deps, err := analyzer.AnalyzeActionFile(actionPath)

			// Check error expectation
			if tt.expectError {
				testutil.AssertError(t, err)

				return
			}
			testutil.AssertNoError(t, err)

			// Check dependencies
			if tt.expectDeps {
				if len(deps) != tt.expectedLen {
					t.Errorf("expected %d dependencies, got %d", tt.expectedLen, len(deps))
				}

				// Check specific dependencies if provided
				if tt.expectedDeps != nil {
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
			} else if len(deps) != 0 {
				t.Errorf("expected no dependencies, got %d", len(deps))
			}
		})
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
			name:            "semantic version",
			uses:            "actions/checkout@v4",
			expectedOwner:   "actions",
			expectedRepo:    "checkout",
			expectedVersion: "v4",
			expectedType:    SemanticVersion,
		},
		{
			name:            "semantic version with patch",
			uses:            "actions/setup-node@v3.8.1",
			expectedOwner:   "actions",
			expectedRepo:    "setup-node",
			expectedVersion: "v3.8.1",
			expectedType:    SemanticVersion,
		},
		{
			name:            "commit SHA",
			uses:            "actions/checkout@8f4b7f84bd579b95d7f0b90f8d8b6e5d9b8a7f6e",
			expectedOwner:   "actions",
			expectedRepo:    "checkout",
			expectedVersion: appconstants.TestSHAForTesting,
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
			version:     "v4",
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
			version:     appconstants.TestSHAForTesting,
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
	cacheInstance, _ := cache.NewCache(cache.DefaultConfig())

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
			owner:           "actions",
			repo:            "checkout",
			expectedVersion: appconstants.TestVersionV4_1_1,
			expectedSHA:     appconstants.TestSHAForTesting,
			expectError:     false,
		},
		{
			name:            "another valid repository",
			owner:           "actions",
			repo:            "setup-node",
			expectedVersion: appconstants.TestVersionV4_0_0,
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
	cacheInstance, _ := cache.NewCache(cache.DefaultConfig())

	analyzer := &Analyzer{
		GitHubClient: githubClient,
		Cache:        cacheInstance,
	}

	// Create test dependencies
	dependencies := []Dependency{
		{
			Name:        testutil.TestActionCheckout,
			Uses:        appconstants.TestActionCheckoutV3,
			Version:     "v3",
			IsPinned:    false,
			VersionType: SemanticVersion,
			Description: "Action for checking out a repo",
		},
		{
			Name:        "actions/setup-node",
			Uses:        "actions/setup-node@v4.0.0",
			Version:     appconstants.TestVersionV4_0_0,
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
		if dep.Current.Name == testutil.TestActionCheckout && dep.Current.Version == "v3" {
			found = true
			if dep.LatestVersion != appconstants.TestVersionV4_1_1 {
				t.Errorf("expected latest version v4.1.1, got %s", dep.LatestVersion)
			}
			if dep.UpdateType != "major" {
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
			latest:       appconstants.TestVersionV4_0_0,
			expectedType: "major",
		},
		{
			name:         "minor version difference",
			current:      appconstants.TestVersionV4_0_0,
			latest:       "v4.1.0",
			expectedType: "minor",
		},
		{
			name:         "patch version difference",
			current:      "v4.1.0",
			latest:       appconstants.TestVersionV4_1_1,
			expectedType: "patch",
		},
		{
			name:         "no difference",
			current:      appconstants.TestVersionV4_1_1,
			latest:       appconstants.TestVersionV4_1_1,
			expectedType: "none",
		},
		{
			name:         "floating to specific",
			current:      "v4",
			latest:       appconstants.TestVersionV4_1_1,
			expectedType: "patch",
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
	cacheInstance, _ := cache.NewCache(cache.DefaultConfig())

	analyzer := &Analyzer{
		GitHubClient: githubClient,
		Cache:        cacheInstance,
	}

	// Create test dependency
	dep := Dependency{
		Name:        testutil.TestActionCheckout,
		Uses:        appconstants.TestActionCheckoutV3,
		Version:     "v3",
		IsPinned:    false,
		VersionType: SemanticVersion,
		Description: "Action for checking out a repo",
	}

	// Generate pinned update
	update, err := analyzer.GeneratePinnedUpdate(
		actionPath,
		dep,
		appconstants.TestVersionV4_1_1,
		appconstants.TestSHAForTesting,
	)

	testutil.AssertNoError(t, err)

	// Verify update details
	testutil.AssertEqual(t, actionPath, update.FilePath)
	testutil.AssertEqual(t, appconstants.TestActionCheckoutV3, update.OldUses)
	testutil.AssertStringContains(t, update.NewUses, "actions/checkout@8f4b7f84bd579b95d7f0b90f8d8b6e5d9b8a7f6e")
	testutil.AssertStringContains(t, update.NewUses, "# v4.1.1")
	testutil.AssertEqual(t, "major", update.UpdateType)
}

func TestAnalyzerWithCache(t *testing.T) {
	t.Parallel()

	// Test that caching works properly
	mockResponses := testutil.MockGitHubResponses()
	githubClient := testutil.MockGitHubClient(mockResponses)
	cacheInstance, _ := cache.NewCache(cache.DefaultConfig())

	analyzer := &Analyzer{
		GitHubClient: githubClient,
		Cache:        cacheInstance,
	}

	// First call should hit the API
	version1, sha1, err1 := analyzer.getLatestVersion("actions", "checkout")
	testutil.AssertNoError(t, err1)

	// Second call should hit the cache
	version2, sha2, err2 := analyzer.getLatestVersion("actions", "checkout")
	testutil.AssertNoError(t, err2)

	// Results should be identical
	testutil.AssertEqual(t, version1, version2)
	testutil.AssertEqual(t, sha1, sha2)
}

func TestAnalyzerRateLimitHandling(t *testing.T) {
	t.Parallel()

	// Create mock client that returns rate limit error
	rateLimitResponse := &http.Response{
		StatusCode: http.StatusForbidden,
		Header: http.Header{
			"X-RateLimit-Remaining": []string{"0"},
			"X-RateLimit-Reset":     []string{strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10)},
		},
		Body: testutil.NewStringReader(`{"message": "API rate limit exceeded"}`),
	}

	mockClient := &testutil.MockHTTPClient{
		Responses: map[string]*http.Response{
			"GET https://api.github.com/repos/actions/checkout/releases/latest": rateLimitResponse,
		},
	}

	client := github.NewClient(&http.Client{Transport: &testutil.MockTransport{Client: mockClient}})
	cacheInstance, _ := cache.NewCache(cache.DefaultConfig())

	analyzer := &Analyzer{
		GitHubClient: client,
		Cache:        cacheInstance,
	}

	// This should handle the rate limit gracefully
	_, _, err := analyzer.getLatestVersion("actions", "checkout")
	if err == nil {
		t.Error("expected rate limit error to be returned")
	}

	// The error message depends on GitHub client implementation
	// It should fail with either rate limit or API error
	if !strings.Contains(err.Error(), "rate limit") && !strings.Contains(err.Error(), "no releases or tags found") {
		t.Errorf("expected error to contain rate limit info or no releases message, got: %s", err.Error())
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
	cacheInstance, err := cache.NewCache(cache.DefaultConfig())
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

	c, err := cache.NewCache(cache.DefaultConfig())
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
func TestIsCompositeAction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		fixture string
		want    bool
		wantErr bool
	}{
		{
			name:    "composite action",
			fixture: "composite-action.yml",
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
			name:    "javascript action",
			fixture: "javascript-action.yml",
			want:    false,
			wantErr: false,
		},
		{
			name:    "invalid yaml",
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
