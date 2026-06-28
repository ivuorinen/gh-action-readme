package dependencies

import (
	"fmt"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/git"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// Local string constants to avoid duplicate literals across this file.
const (
	covTagOwner   = "tagorg"
	covTagRepo    = "tagrepo"
	covTagSHA     = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	covTagVersion = "v2.0.0"
	covShellRun   = "echo hello"
	covRepoDesc   = "cached repo description"
	covShortSHA   = "8f4b7f8"
	covCore123    = "1.2.3"
)

// TestCovAnalyzerCoreVersion covers the prerelease/build-metadata stripping
// branches of coreVersion (the "-"/"+" suffix path that was under-covered).
func TestCovAnalyzerCoreVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"prerelease suffix", "1.0.0-beta", "1.0.0"},
		{"build metadata suffix", "1.0.0+build", "1.0.0"},
		{"prerelease with dots", "2.3.4-rc.1", "2.3.4"},
		{"no suffix unchanged", covCore123, covCore123},
		{"empty unchanged", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.AssertEqual(t, tt.want, coreVersion(tt.input))
		})
	}
}

// TestCovAnalyzerDetermineUpdateType exercises every comparison branch of
// determineUpdateType directly.
func TestCovAnalyzerDetermineUpdateType(t *testing.T) {
	t.Parallel()

	analyzer := &Analyzer{}

	tests := []struct {
		name    string
		current []string
		latest  []string
		want    string
	}{
		{"major differs", []string{"1", "0", "0"}, []string{"2", "0", "0"}, appconstants.UpdateTypeMajor},
		{"minor differs", []string{"1", "0", "0"}, []string{"1", "1", "0"}, appconstants.UpdateTypeMinor},
		{"patch differs", []string{"1", "0", "0"}, []string{"1", "0", "1"}, appconstants.UpdateTypePatch},
		{"all equal", []string{"1", "2", "3"}, []string{"1", "2", "3"}, appconstants.UpdateTypeNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.AssertEqual(t, tt.want, analyzer.determineUpdateType(tt.current, tt.latest))
		})
	}
}

// TestCovAnalyzerShaMatches covers the empty, prefix, equal, and mismatch
// branches of shaMatches.
func TestCovAnalyzerShaMatches(t *testing.T) {
	t.Parallel()

	const full = "8f4b7f84bd579b95d7f0b90f8d8b6e5d9b8a7f6e"

	tests := []struct {
		name    string
		current string
		latest  string
		want    bool
	}{
		{"empty current", "", full, false},
		{"short prefix match", covShortSHA, full, true},
		{"short prefix mismatch", "deadbee", full, false},
		{"full equal", full, full, true},
		{"full mismatch same length", "1111111111111111111111111111111111111111", full, false},
		{"case insensitive equal", "8F4B7F84BD579B95D7F0B90F8D8B6E5D9B8A7F6E", full, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.AssertEqual(t, tt.want, shaMatches(tt.current, tt.latest))
		})
	}
}

// TestCovAnalyzerUnquoteYAMLScalar covers the quote-stripping and pass-through
// branches of unquoteYAMLScalar.
func TestCovAnalyzerUnquoteYAMLScalar(t *testing.T) {
	t.Parallel()

	v4 := testutil.TestActionCheckoutV4

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"double quoted", `"` + v4 + `"`, v4},
		{"single quoted", `'` + v4 + `'`, v4},
		{"unquoted unchanged", v4, v4},
		{"too short unchanged", "x", "x"},
		{"mismatched quotes unchanged", `"abc'`, `"abc'`},
		{"empty unchanged", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.AssertEqual(t, tt.want, unquoteYAMLScalar(tt.input))
		})
	}
}

// TestCovAnalyzerAnalyzeShellScript covers both the named/with-RepoInfo branch
// (script URL generated) and the unnamed/no-RepoInfo branch (default name, no
// URL).
func TestCovAnalyzerAnalyzeShellScript(t *testing.T) {
	t.Parallel()

	t.Run("with repo info and name", func(t *testing.T) {
		t.Parallel()

		analyzer := &Analyzer{RepoInfo: git.RepoInfo{
			Organization:  testDepOwnerActions,
			Repository:    testDepRepoCheckout,
			DefaultBranch: testutil.TestBranchMain,
		}}

		dep := analyzer.analyzeShellScript(CompositeStep{Name: "Build", Run: covShellRun}, 2)
		if dep == nil {
			t.Fatal("expected non-nil shell script dependency")
		}
		testutil.AssertEqual(t, "Build", dep.Name)
		if !dep.IsShellScript || !dep.IsLocalAction {
			t.Error("shell script dependency should be marked shell + local")
		}
		if dep.ScriptURL == "" {
			t.Error("expected a script URL when RepoInfo is populated")
		}
	})

	t.Run("without repo info and no name", func(t *testing.T) {
		t.Parallel()

		analyzer := &Analyzer{}
		dep := analyzer.analyzeShellScript(CompositeStep{Run: covShellRun}, 3)
		if dep == nil {
			t.Fatal("expected non-nil shell script dependency")
		}
		testutil.AssertEqual(t, "Shell Script #3", dep.Name)
		if dep.ScriptURL != "" {
			t.Errorf("expected empty script URL without RepoInfo, got %q", dep.ScriptURL)
		}
	})
}

// TestCovAnalyzerLocalActionDependency covers the docker:// description branch
// and the local (./) empty-name fallback branch.
func TestCovAnalyzerLocalActionDependency(t *testing.T) {
	t.Parallel()

	analyzer := &Analyzer{}

	t.Run("docker reference", func(t *testing.T) {
		t.Parallel()

		dep := analyzer.localActionDependency(CompositeStep{Name: "Run image", Uses: "docker://alpine:3.14"})
		testutil.AssertEqual(t, "Docker container action", dep.Description)
		if dep.IsLocalAction {
			t.Error("docker reference must not be flagged as a local action")
		}
	})

	t.Run("local reference with empty name", func(t *testing.T) {
		t.Parallel()

		dep := analyzer.localActionDependency(CompositeStep{Uses: "./setup"})
		testutil.AssertEqual(t, "./setup", dep.Name) // falls back to Uses
		testutil.AssertEqual(t, "Local action (same repository)", dep.Description)
		if !dep.IsLocalAction {
			t.Error("local reference should be flagged as a local action")
		}
	})
}

// TestCovAnalyzerProcessStep covers the uses-step, run-step, and empty-step
// branches of processStep. The "uses returns error -> nil" branch is effectively
// unreachable: an unparseable uses statement resolves to LocalPath and yields a
// local dependency rather than an error, so it is intentionally not forced here.
func TestCovAnalyzerProcessStep(t *testing.T) {
	t.Parallel()

	analyzer := &Analyzer{}

	t.Run("uses step yields dependency", func(t *testing.T) {
		t.Parallel()
		dep := analyzer.processStep(CompositeStep{Uses: "./local"}, 1)
		if dep == nil {
			t.Fatal("expected dependency for uses step")
		}
		testutil.AssertEqual(t, "./local", dep.Uses)
	})

	t.Run("run step yields shell dependency", func(t *testing.T) {
		t.Parallel()
		dep := analyzer.processStep(CompositeStep{Run: covShellRun}, 1)
		if dep == nil {
			t.Fatal("expected dependency for run step")
		}
		if !dep.IsShellScript {
			t.Error("run step should produce a shell script dependency")
		}
	})

	t.Run("empty step yields nil", func(t *testing.T) {
		t.Parallel()
		if dep := analyzer.processStep(CompositeStep{}, 1); dep != nil {
			t.Errorf("expected nil for empty step, got %+v", dep)
		}
	})
}

// TestCovAnalyzerGetLatestTagSuccess drives the success path of getLatestTag:
// the latest-release lookup 404s (unmocked) so the analyzer falls back to the
// tag list, which returns a valid tag.
func TestCovAnalyzerGetLatestTagSuccess(t *testing.T) {
	t.Parallel()

	base := "GET https://api.github.com/repos/" + covTagOwner + "/" + covTagRepo
	tagsJSON := `[{"name":"` + covTagVersion + `","commit":{"sha":"` + covTagSHA + `"}}]`
	mockResponses := map[string]string{
		base + "/tags?per_page=10": tagsJSON,
	}

	analyzer := &Analyzer{
		GitHubClient: testutil.MockGitHubClient(mockResponses),
		Cache:        newIsolatedCache(t),
	}

	version, sha, err := analyzer.getLatestVersion(covTagOwner, covTagRepo)
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, covTagVersion, version)
	testutil.AssertEqual(t, covTagSHA, sha)
}

// TestCovAnalyzerCheckOutdatedSkips covers the early-skip branches of
// CheckOutdated (shell scripts, local actions, unparseable uses) plus the
// up-to-date branch that produces no OutdatedDependency.
func TestCovAnalyzerCheckOutdatedSkips(t *testing.T) {
	t.Parallel()

	analyzer := &Analyzer{
		GitHubClient: testutil.MockGitHubClient(testutil.MockGitHubResponses()),
		Cache:        newIsolatedCache(t),
	}

	deps := []Dependency{
		{Name: "shell", IsShellScript: true},
		{Name: "local", IsLocalAction: true},
		{Name: "bad-uses", Uses: "./local-only", VersionType: LocalPath},
		{
			Name:        testutil.TestActionCheckout,
			Uses:        "actions/checkout@v4.1.1", // already latest
			Version:     testutil.TestVersionV4_1_1,
			VersionType: SemanticVersion,
		},
		{
			Name:        testutil.TestActionCheckout,
			Uses:        testutil.TestActionCheckoutV3, // outdated
			Version:     testDepVersionV3,
			VersionType: SemanticVersion,
		},
	}

	outdated, err := analyzer.CheckOutdated(deps)
	testutil.AssertNoError(t, err)

	// Only the v3 checkout is outdated; everything else is skipped or current.
	if len(outdated) != 1 {
		t.Fatalf("expected exactly 1 outdated dependency, got %d", len(outdated))
	}
	testutil.AssertEqual(t, testDepVersionV3, outdated[0].Current.Version)
	testutil.AssertEqual(t, appconstants.UpdateTypeMajor, outdated[0].UpdateType)
}

// TestCovAnalyzerCheckOutdatedGetLatestError covers the branch where
// getLatestVersion errors (repo not present in the mock) and the dependency is
// skipped instead of failing the whole operation.
func TestCovAnalyzerCheckOutdatedGetLatestError(t *testing.T) {
	t.Parallel()

	analyzer := &Analyzer{
		GitHubClient: testutil.MockGitHubClient(testutil.MockGitHubResponses()),
		Cache:        newIsolatedCache(t),
	}

	deps := []Dependency{
		{
			Name:        "unknown/repo",
			Uses:        "unknown/repo@v1",
			Version:     "v1",
			VersionType: SemanticVersion,
		},
	}

	outdated, err := analyzer.CheckOutdated(deps)
	testutil.AssertNoError(t, err)
	if len(outdated) != 0 {
		t.Errorf("expected no outdated deps when latest version lookup fails, got %d", len(outdated))
	}
}

// TestCovAnalyzerEnrichWithGitHubData covers both the cache-hit branch (a stored
// description string short-circuits the API call) and the API-fetch branch.
func TestCovAnalyzerEnrichWithGitHubData(t *testing.T) {
	t.Parallel()

	t.Run("cache hit returns stored description", func(t *testing.T) {
		t.Parallel()

		analyzer := &Analyzer{Cache: NewCacheAdapter(newIsolatedCache(t))}
		cacheKey := appconstants.CacheKeyRepo + fmt.Sprintf("%s/%s", testDepOwnerActions, testDepRepoCheckout)
		testutil.AssertNoError(t, analyzer.Cache.Set(cacheKey, covRepoDesc))

		dep := &Dependency{}
		err := analyzer.enrichWithGitHubData(dep, testDepOwnerActions, testDepRepoCheckout)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, covRepoDesc, dep.Description)
	})

	t.Run("api fetch populates description", func(t *testing.T) {
		t.Parallel()

		analyzer := &Analyzer{
			GitHubClient: testutil.MockGitHubClient(testutil.MockGitHubResponses()),
			Cache:        NewCacheAdapter(newIsolatedCache(t)),
		}

		dep := &Dependency{}
		err := analyzer.enrichWithGitHubData(dep, testDepOwnerActions, testDepRepoCheckout)
		testutil.AssertNoError(t, err)
		if dep.Description == "" {
			t.Error("expected description to be populated from the GitHub API mock")
		}
	})
}

// TestCovAnalyzerAnalyzeActionFileWithProgress drives the progress-callback path
// of AnalyzeActionFileWithProgress (and, through it, processCompositeSteps and
// analyzeShellScript) using a composite fixture with both action and run steps.
func TestCovAnalyzerAnalyzeActionFileWithProgress(t *testing.T) {
	t.Parallel()

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	actionPath := testutil.WriteActionFile(t, tmpDir, testutil.MustReadFixture(testutil.TestFixtureCompositeWithDeps))

	analyzer := &Analyzer{
		GitHubClient: testutil.MockGitHubClient(testutil.MockGitHubResponses()),
		Cache:        newIsolatedCache(t),
	}

	var calls int
	deps, err := analyzer.AnalyzeActionFileWithProgress(actionPath, func(_, _ int, _ string) {
		calls++
	})
	testutil.AssertNoError(t, err)

	if calls == 0 {
		t.Error("expected the progress callback to be invoked at least once")
	}
	if len(deps) == 0 {
		t.Error("expected dependencies from the composite fixture")
	}

	var sawShell bool
	for _, d := range deps {
		if d.IsShellScript {
			sawShell = true
		}
	}
	if !sawShell {
		t.Error("expected at least one shell script dependency from the run steps")
	}
}
