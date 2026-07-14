package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/internal/dependencies"
	"github.com/ivuorinen/gh-action-readme/internal/git"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// These tests squeeze the remaining reachable error/edge branches of the command
// handlers. Several mutate the shared globalConfig (or the package-global
// configFile), so those do not use t.Parallel() and always save/restore.

const (
	covMopErrWantErr    = "expected error, got nil"
	covMopPinnedFixture = "dependencies/already-pinned.yml"
)

// covMopXDGCacheAsFile points XDG_CACHE_HOME at a regular file so xdg.CacheFile —
// and therefore cache.NewCache — fails to create the cache directory, exercising
// the NewCache error branches of the cache handlers.
func covMopXDGCacheAsFile(t *testing.T) {
	t.Helper()
	cacheFile := filepath.Join(t.TempDir(), "cache-home-is-a-file")
	if err := os.WriteFile(cacheFile, []byte("x"), 0o600); err != nil {
		t.Fatalf("failed to seed cache-home file: %v", err)
	}
	t.Cleanup(xdg.Reload)
	t.Setenv("XDG_CACHE_HOME", cacheFile)
	xdg.Reload()
}

// covMopXDGConfigAsFile neutralizes every config candidate xdg.ConfigFile searches
// so that GetConfigPath fails, exercising the config-path error branches. Unlike
// xdg.CacheFile (single candidate), xdg.ConfigFile falls back through a chain —
// XDG_CONFIG_HOME, then the platform home defaults (on darwin: ~/Library/Application
// Support, ~/Library/Preferences, ~/.config). Pointing only XDG_CONFIG_HOME at a file
// lets xdg fall back to a writable home default and succeed, so HOME and
// XDG_CONFIG_DIRS must be neutralized too, or the test passes on Linux CI but fails
// on darwin.
func covMopXDGConfigAsFile(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	seed := func(name string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
			t.Fatalf("failed to seed %s: %v", name, err)
		}

		return p
	}
	t.Cleanup(xdg.Reload)
	t.Setenv("HOME", seed("home-is-a-file"))
	t.Setenv("XDG_CONFIG_HOME", seed("config-home-is-a-file"))
	t.Setenv("XDG_CONFIG_DIRS", "")
	xdg.Reload()
}

// TestCovMopCacheHandlersNewCacheError covers the cache.NewCache error branches of
// cacheClearHandler and cacheStatsHandler (NewCache fails when the XDG cache home
// is a file, so the cache directory cannot be created).
func TestCovMopCacheHandlersNewCacheError(t *testing.T) {
	origConfig := globalConfig
	defer func() { globalConfig = origConfig }()
	globalConfig = internal.DefaultAppConfig()
	globalConfig.Quiet = true

	covMopXDGCacheAsFile(t)
	cmd := newCacheCmd()

	if err := cacheClearHandler(cmd, []string{}); err == nil {
		t.Error("cacheClearHandler() " + covMopErrWantErr)
	}
	if err := cacheStatsHandler(cmd, []string{}); err == nil {
		t.Error("cacheStatsHandler() " + covMopErrWantErr)
	}
}

// TestCovMopConfigHandlersGetConfigPathError covers the GetConfigPath error
// branches of configRootHandler and configInitHandler.
func TestCovMopConfigHandlersGetConfigPathError(t *testing.T) {
	origConfig := globalConfig
	defer func() { globalConfig = origConfig }()
	globalConfig = internal.DefaultAppConfig()
	globalConfig.Quiet = true

	covMopXDGConfigAsFile(t)

	if err := configRootHandler(newConfigCmd(), []string{}); err == nil {
		t.Error("configRootHandler() " + covMopErrWantErr)
	}
	if err := configInitHandler(newConfigCmd(), []string{}); err == nil {
		t.Error("configInitHandler() " + covMopErrWantErr)
	}
}

// TestCovMopDepsListPinnedDependency drives depsListHandler with a GitHub token and
// a composite action whose only dependency is pinned to a commit SHA. The token
// makes createAnalyzer return a real analyzer, so analyzeActionFileDeps takes its
// "pinned" success branch and depsListHandler prints the non-zero total. Listing
// only parses the file locally, so no network access occurs.
func TestCovMopDepsListPinnedDependency(t *testing.T) {
	origConfig := globalConfig
	defer func() { globalConfig = origConfig }()

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()
	testutil.InitGitRepo(t, tmpDir)
	testutil.WriteActionFixture(t, tmpDir, covMopPinnedFixture)
	t.Chdir(tmpDir)

	globalConfig = internal.DefaultAppConfig()
	globalConfig.Quiet = false
	globalConfig.GitHubToken = testutil.TestToken123

	if err := depsListHandler(newDepsCmd(), []string{}); err != nil {
		t.Errorf("depsListHandler() "+testutil.TestErrUnexpected, err)
	}
}

// TestCovMopSetupDepsUpgradeNilGlobalConfig covers the branch where globalConfig is
// nil and setupDepsUpgrade falls back to DefaultAppConfig(), which carries no token,
// so the function returns the GitHub-auth error before discovering files.
func TestCovMopSetupDepsUpgradeNilGlobalConfig(t *testing.T) {
	origConfig := globalConfig
	defer func() { globalConfig = origConfig }()
	globalConfig = nil

	// DefaultAppConfig() must observe an empty token, so clear the token env vars.
	t.Setenv(appconstants.EnvGitHubToken, "")
	t.Setenv(appconstants.EnvGitHubTokenStandard, "")

	analyzer, files, err := setupDepsUpgrade(t.TempDir(), nil)
	if err == nil {
		t.Error("setupDepsUpgrade() " + covMopErrWantErr)
	}
	if analyzer != nil || files != nil {
		t.Errorf("expected nil analyzer and files on error, got %v / %v", analyzer, files)
	}
}

// TestCovMopApplyUpdatesReadError covers the non-EOF read-error branch of
// applyUpdates: an interactive prompt whose reader returns a non-EOF error must
// surface as a wrapped error (distinct from the silent EOF-cancel path).
func TestCovMopApplyUpdatesReadError(t *testing.T) {
	t.Parallel()

	output := createOutputManager(true)
	analyzer := dependencies.NewAnalyzer(nil, git.RepoInfo{}, nil)
	updates := []dependencies.PinnedUpdate{
		{OldUses: testutil.TestActionCheckoutV3, NewUses: testutil.TestActionCheckoutV4},
	}

	// An exhausted TestInputReader returns a non-EOF error from ReadLine.
	reader := &TestInputReader{responses: nil}

	if err := applyUpdates(output, analyzer, updates, false, reader); err == nil {
		t.Error("applyUpdates() " + covMopErrWantErr)
	}
}

// TestCovMopLoadGenConfigError covers the LoadConfiguration error branch of
// loadGenConfig by pointing the package-global configFile at a malformed YAML file
// so viper's ReadInConfig fails during the global-config load step.
func TestCovMopLoadGenConfigError(t *testing.T) {
	origConfigFile := configFile
	defer func() { configFile = origConfigFile }()

	tmpDir := t.TempDir()
	badConfig := filepath.Join(tmpDir, "bad-config.yaml")
	testutil.WriteTestFile(t, badConfig, testutil.TestInvalidYAMLPrefix)
	configFile = badConfig

	cfg, err := loadGenConfig(tmpDir, tmpDir)
	if err == nil {
		t.Error("loadGenConfig() " + covMopErrWantErr)
	}
	if cfg != nil {
		t.Errorf("expected nil config on error, got %+v", cfg)
	}
}
