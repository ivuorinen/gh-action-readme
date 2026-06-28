package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// These tests exercise the under-covered branches of the cache, config, and
// validate command handlers. They mutate the shared globalConfig, so none of
// them use t.Parallel(); each saves and restores globalConfig.

// isolateXDGCache points the XDG cache directory at a fresh temp dir and forces
// the xdg package to re-read the environment, so cache handlers never touch the
// developer's real cache. Returns the cache home so callers can assert on it.
func isolateXDGCache(t *testing.T) string {
	t.Helper()
	cacheHome := t.TempDir()
	t.Cleanup(xdg.Reload)
	t.Setenv("XDG_CACHE_HOME", cacheHome)
	xdg.Reload()

	return cacheHome
}

// TestCovCmdCacheHandlersIsolated runs all three cache handlers against an
// isolated, non-quiet output manager and asserts the cache directory is created
// (the "Directory exists" branch of cachePathHandler).
func TestCovCmdCacheHandlersIsolated(t *testing.T) {
	origConfig := globalConfig
	defer func() { globalConfig = origConfig }()
	globalConfig = internal.DefaultAppConfig()
	globalConfig.Quiet = false

	cacheHome := isolateXDGCache(t)
	cmd := newCacheCmd()

	if err := cacheClearHandler(cmd, []string{}); err != nil {
		t.Errorf("cacheClearHandler() "+testutil.TestErrUnexpected, err)
	}
	if err := cacheStatsHandler(cmd, []string{}); err != nil {
		t.Errorf("cacheStatsHandler() "+testutil.TestErrUnexpected, err)
	}
	if err := cachePathHandler(cmd, []string{}); err != nil {
		t.Errorf("cachePathHandler() "+testutil.TestErrUnexpected, err)
	}

	// NewCache (invoked by every handler) creates <XDG_CACHE_HOME>/<app>; the
	// path handler then takes its "Directory exists" branch.
	wantDir := filepath.Join(cacheHome, appconstants.AppName)
	if _, err := os.Stat(wantDir); err != nil {
		t.Errorf("expected cache directory %q to exist: %v", wantDir, err)
	}
}

// TestCovCmdCacheStatsVerbose exercises cacheStatsHandler under verbose config.
func TestCovCmdCacheStatsVerbose(t *testing.T) {
	origConfig := globalConfig
	defer func() { globalConfig = origConfig }()
	globalConfig = internal.DefaultAppConfig()
	globalConfig.Verbose = true
	globalConfig.Quiet = false

	isolateXDGCache(t)

	if err := cacheStatsHandler(newCacheCmd(), []string{}); err != nil {
		t.Errorf("cacheStatsHandler() "+testutil.TestErrUnexpected, err)
	}
}

// TestCovCmdConfigInitWritesAndDetectsExisting covers BOTH branches of
// configInitHandler: the write-default-config path on first run, and the
// "configuration already exists" warning path on the second run.
func TestCovCmdConfigInitWritesAndDetectsExisting(t *testing.T) {
	origConfig := globalConfig
	defer func() { globalConfig = origConfig }()
	globalConfig = internal.DefaultAppConfig()
	globalConfig.Quiet = true

	// Isolate the XDG config dir so GetConfigPath/WriteDefaultConfig operate on a
	// throwaway location instead of the developer's real config.
	t.Cleanup(xdg.Reload)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "xdg"))
	xdg.Reload()

	cmd := newConfigCmd()

	// First run: config does not exist yet -> writes default config.
	if err := configInitHandler(cmd, []string{}); err != nil {
		t.Fatalf("first configInitHandler() "+testutil.TestErrUnexpected, err)
	}

	configPath, err := internal.GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() "+testutil.TestErrUnexpected, err)
	}
	if _, statErr := os.Stat(configPath); statErr != nil {
		t.Fatalf("expected config file created at %q: %v", configPath, statErr)
	}

	// Second run: config now exists -> warning branch, still returns nil.
	if err := configInitHandler(cmd, []string{}); err != nil {
		t.Errorf("second configInitHandler() "+testutil.TestErrUnexpected, err)
	}
}

// TestCovCmdConfigThemesCurrentHighlight covers the "(current)" highlight branch
// of configThemesHandler, which only runs when a theme matches globalConfig.Theme.
func TestCovCmdConfigThemesCurrentHighlight(t *testing.T) {
	origConfig := globalConfig
	defer func() { globalConfig = origConfig }()
	globalConfig = &internal.AppConfig{Quiet: false, Theme: appconstants.ThemeGitHub}

	if err := configThemesHandler(newConfigCmd(), []string{}); err != nil {
		t.Errorf("configThemesHandler() "+testutil.TestErrUnexpected, err)
	}
}

// TestCovCmdConfigRootVerbose covers the verbose branch of configRootHandler,
// which dumps the redacted config.
func TestCovCmdConfigRootVerbose(t *testing.T) {
	origConfig := globalConfig
	defer func() { globalConfig = origConfig }()
	globalConfig = internal.DefaultAppConfig()
	globalConfig.Verbose = true
	globalConfig.Quiet = false

	if err := configRootHandler(newConfigCmd(), []string{}); err != nil {
		t.Errorf("configRootHandler() "+testutil.TestErrUnexpected, err)
	}
}

// TestCovCmdSchemaHandlerBranches covers both schemaHandler branches: the
// "not configured" path (empty schema) and the configured path.
func TestCovCmdSchemaHandlerBranches(t *testing.T) {
	tests := []struct {
		name   string
		schema string
	}{
		{name: "empty schema not configured", schema: ""},
		{name: "configured schema", schema: "schemas/custom.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origConfig := globalConfig
			defer func() { globalConfig = origConfig }()
			globalConfig = &internal.AppConfig{Quiet: false, Schema: tt.schema}

			if err := schemaHandler(newSchemaCmd(), []string{}); err != nil {
				t.Errorf("schemaHandler() "+testutil.TestErrUnexpected, err)
			}
		})
	}
}

// TestCovCmdValidateHandlerEmptyDir covers the discovery-error path of
// validateHandler when the working directory has no action files.
func TestCovCmdValidateHandlerEmptyDir(t *testing.T) {
	origConfig := globalConfig
	defer func() { globalConfig = origConfig }()

	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	globalConfig = internal.DefaultAppConfig()

	if err := validateHandler(newValidateCmd(), []string{}); err == nil {
		t.Error("validateHandler() expected error for directory with no action files, got nil")
	}
}

// TestCovCmdValidateHandlerSuccessVerbose covers the successful validation path
// of validateHandler under verbose config.
func TestCovCmdValidateHandlerSuccessVerbose(t *testing.T) {
	origConfig := globalConfig
	defer func() { globalConfig = origConfig }()

	tmpDir := t.TempDir()
	// Write the fixture BEFORE changing directory: fixtures are read relative to
	// the project root testdata/ directory.
	testutil.WriteActionFixture(t, tmpDir, testutil.TestFixtureJavaScriptSimple)
	t.Chdir(tmpDir)

	globalConfig = internal.DefaultAppConfig()
	globalConfig.Verbose = true

	if err := validateHandler(newValidateCmd(), []string{}); err != nil {
		t.Errorf("validateHandler() "+testutil.TestErrUnexpected, err)
	}
}
