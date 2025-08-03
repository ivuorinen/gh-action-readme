package dependencies

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v57/github"

	"github.com/ivuorinen/gh-action-readme/internal/cache"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

func TestAnalyzer_AnalyzeActionFile(t *testing.T) {
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
			actionYML:   testutil.SimpleActionYML,
			expectError: false,
			expectDeps:  false,
			expectedLen: 0,
		},
		{
			name:         "composite action with dependencies",
			actionYML:    testutil.CompositeActionYML,
			expectError:  false,
			expectDeps:   true,
			expectedLen:  3,
			expectedDeps: []string{"actions/checkout@v4", "actions/setup-node@v3"},
		},
		{
			name:        "docker action - no step dependencies",
			actionYML:   testutil.DockerActionYML,
			expectError: false,
			expectDeps:  false,
			expectedLen: 0,
		},
		{
			name:        "invalid action file",
			actionYML:   testutil.InvalidActionYML,
			expectError: true,
		},
		{
			name:        "minimal action - no dependencies",
			actionYML:   testutil.MinimalActionYML,
			expectError: false,
			expectDeps:  false,
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary action file
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			actionPath := filepath.Join(tmpDir, "action.yml")
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

func TestAnalyzer_ParseUsesStatement(t *testing.T) {
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
			expectedVersion: "8f4b7f84bd579b95d7f0b90f8d8b6e5d9b8a7f6e",
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
			owner, repo, version, versionType := analyzer.parseUsesStatement(tt.uses)

			testutil.AssertEqual(t, tt.expectedOwner, owner)
			testutil.AssertEqual(t, tt.expectedRepo, repo)
			testutil.AssertEqual(t, tt.expectedVersion, version)
			testutil.AssertEqual(t, tt.expectedType, versionType)
		})
	}
}

func TestAnalyzer_VersionChecking(t *testing.T) {
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
			version:     "8f4b7f84bd579b95d7f0b90f8d8b6e5d9b8a7f6e",
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
			isPinned := analyzer.isVersionPinned(tt.version)
			isCommitSHA := analyzer.isCommitSHA(tt.version)
			isSemantic := analyzer.isSemanticVersion(tt.version)

			testutil.AssertEqual(t, tt.isPinned, isPinned)
			testutil.AssertEqual(t, tt.isCommitSHA, isCommitSHA)
			testutil.AssertEqual(t, tt.isSemantic, isSemantic)
		})
	}
}

func TestAnalyzer_GetLatestVersion(t *testing.T) {
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
			expectedVersion: "v4.1.1",
			expectedSHA:     "8f4b7f84bd579b95d7f0b90f8d8b6e5d9b8a7f6e",
			expectError:     false,
		},
		{
			name:            "another valid repository",
			owner:           "actions",
			repo:            "setup-node",
			expectedVersion: "v4.0.0",
			expectedSHA:     "1a4e6d7c9f8e5b2a3c4d5e6f7a8b9c0d1e2f3a4b",
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

func TestAnalyzer_CheckOutdated(t *testing.T) {
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
			Name:        "actions/checkout",
			Uses:        "actions/checkout@v3",
			Version:     "v3",
			IsPinned:    false,
			VersionType: SemanticVersion,
			Description: "Action for checking out a repo",
		},
		{
			Name:        "actions/setup-node",
			Uses:        "actions/setup-node@v4.0.0",
			Version:     "v4.0.0",
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
		if dep.Current.Name == "actions/checkout" && dep.Current.Version == "v3" {
			found = true
			if dep.LatestVersion != "v4.1.1" {
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

func TestAnalyzer_CompareVersions(t *testing.T) {
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
			latest:       "v4.0.0",
			expectedType: "major",
		},
		{
			name:         "minor version difference",
			current:      "v4.0.0",
			latest:       "v4.1.0",
			expectedType: "minor",
		},
		{
			name:         "patch version difference",
			current:      "v4.1.0",
			latest:       "v4.1.1",
			expectedType: "patch",
		},
		{
			name:         "no difference",
			current:      "v4.1.1",
			latest:       "v4.1.1",
			expectedType: "none",
		},
		{
			name:         "floating to specific",
			current:      "v4",
			latest:       "v4.1.1",
			expectedType: "patch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateType := analyzer.compareVersions(tt.current, tt.latest)
			testutil.AssertEqual(t, tt.expectedType, updateType)
		})
	}
}

func TestAnalyzer_GeneratePinnedUpdate(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Create a test action file with composite steps
	actionContent := `name: 'Test Composite Action'
description: 'Test action for update testing'
runs:
	using: 'composite'
	steps:
		- name: Checkout code
			uses: actions/checkout@v3
		- name: Setup Node
			uses: actions/setup-node@v3.8.0
			with:
				node-version: '18'
`

	actionPath := filepath.Join(tmpDir, "action.yml")
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
		Name:        "actions/checkout",
		Uses:        "actions/checkout@v3",
		Version:     "v3",
		IsPinned:    false,
		VersionType: SemanticVersion,
		Description: "Action for checking out a repo",
	}

	// Generate pinned update
	update, err := analyzer.GeneratePinnedUpdate(
		actionPath,
		dep,
		"v4.1.1",
		"8f4b7f84bd579b95d7f0b90f8d8b6e5d9b8a7f6e",
	)

	testutil.AssertNoError(t, err)

	// Verify update details
	testutil.AssertEqual(t, actionPath, update.FilePath)
	testutil.AssertEqual(t, "actions/checkout@v3", update.OldUses)
	testutil.AssertStringContains(t, update.NewUses, "actions/checkout@8f4b7f84bd579b95d7f0b90f8d8b6e5d9b8a7f6e")
	testutil.AssertStringContains(t, update.NewUses, "# v4.1.1")
	testutil.AssertEqual(t, "major", update.UpdateType)
}

func TestAnalyzer_WithCache(t *testing.T) {
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

func TestAnalyzer_RateLimitHandling(t *testing.T) {
	// Create mock client that returns rate limit error
	rateLimitResponse := &http.Response{
		StatusCode: 403,
		Header: http.Header{
			"X-RateLimit-Remaining": []string{"0"},
			"X-RateLimit-Reset":     []string{fmt.Sprintf("%d", time.Now().Add(time.Hour).Unix())},
		},
		Body: testutil.NewStringReader(`{"message": "API rate limit exceeded"}`),
	}

	mockClient := &testutil.MockHTTPClient{
		Responses: map[string]*http.Response{
			"GET https://api.github.com/repos/actions/checkout/releases/latest": rateLimitResponse,
		},
	}

	client := github.NewClient(&http.Client{Transport: &mockTransport{client: mockClient}})
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

func TestAnalyzer_WithoutGitHubClient(t *testing.T) {
	// Test graceful degradation when GitHub client is not available
	analyzer := &Analyzer{
		GitHubClient: nil,
		Cache:        nil,
	}

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	actionPath := filepath.Join(tmpDir, "action.yml")
	testutil.WriteTestFile(t, actionPath, testutil.CompositeActionYML)

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

// mockTransport wraps our mock HTTP client for GitHub client.
type mockTransport struct {
	client *testutil.MockHTTPClient
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.client.Do(req)
}
