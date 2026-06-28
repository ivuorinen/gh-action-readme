package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/internal/cache"
	"github.com/ivuorinen/gh-action-readme/internal/dependencies"
	"github.com/ivuorinen/gh-action-readme/internal/wizard"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// covSeamMockAnalyzer builds an analyzer backed by the mocked GitHub client and an
// isolated on-disk cache, mirroring internal/dependencies/analyzer_test.go. The deps
// handlers defer closeAnalyzer (which closes the cache), so no extra cleanup is
// registered here to avoid a double Close.
func covSeamMockAnalyzer(t *testing.T) *dependencies.Analyzer {
	t.Helper()

	cfg := cache.DefaultConfig()
	cfg.Path = t.TempDir()
	c, err := cache.NewCache(cfg)
	testutil.AssertNoError(t, err)

	return &dependencies.Analyzer{
		GitHubClient: testutil.MockGitHubClient(testutil.MockGitHubResponses()),
		Cache:        dependencies.NewCacheAdapter(c),
	}
}

// covSeamSubCmd locates a subcommand by name on a freshly built parent command so the
// handler is exercised with the same flag set production wires up.
func covSeamSubCmd(t *testing.T, parent *cobra.Command, name string) *cobra.Command {
	t.Helper()

	for _, c := range parent.Commands() {
		if c.Name() == name {
			return c
		}
	}
	t.Fatalf("subcommand %q not found", name)

	return nil
}

// covSeamCapture runs fn with os.Stdout redirected to a pipe and returns everything
// written, so the result-rendering branches can be asserted on.
func covSeamCapture(t *testing.T, fn func() error) (string, error) {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	testutil.AssertNoError(t, err)
	os.Stdout = w

	runErr := fn()

	_ = w.Close()
	os.Stdout = old
	data, _ := io.ReadAll(r)
	_ = r.Close()

	return string(data), runErr
}

// TestCovSeamDepsOutdatedRendersResults overrides the createAnalyzer seam so the
// outdated handler renders the "Found N outdated dependencies" branch using the mock
// GitHub client instead of a live network call.
func TestCovSeamDepsOutdatedRendersResults(t *testing.T) {
	origCfg := globalConfig
	origCreate := createAnalyzer
	defer func() {
		globalConfig = origCfg
		createAnalyzer = origCreate
	}()

	globalConfig = internal.DefaultAppConfig()
	globalConfig.GitHubToken = testutil.TestTokenConfig

	dir := t.TempDir()
	testutil.WriteActionFixture(t, dir, testutil.TestFixtureTestCompositeAction)
	t.Chdir(dir)

	createAnalyzer = func(_ *internal.Generator, _ *internal.ColoredOutput) *dependencies.Analyzer {
		return covSeamMockAnalyzer(t)
	}

	cmd := covSeamSubCmd(t, newDepsCmd(), "outdated")
	out, err := covSeamCapture(t, func() error {
		return depsOutdatedHandler(cmd, nil)
	})

	testutil.AssertNoError(t, err)
	testutil.AssertStringContains(t, out, "outdated dependencies")
}

// TestCovSeamDepsSecurityRendersResults overrides the createAnalyzer seam so the
// security handler renders the pinned/floating summary branch using the mock client.
func TestCovSeamDepsSecurityRendersResults(t *testing.T) {
	origCfg := globalConfig
	origCreate := createAnalyzer
	defer func() {
		globalConfig = origCfg
		createAnalyzer = origCreate
	}()

	globalConfig = internal.DefaultAppConfig()
	globalConfig.GitHubToken = testutil.TestTokenConfig

	dir := t.TempDir()
	testutil.WriteActionFixture(t, dir, testutil.TestFixtureTestCompositeAction)
	t.Chdir(dir)

	createAnalyzer = func(_ *internal.Generator, _ *internal.ColoredOutput) *dependencies.Analyzer {
		return covSeamMockAnalyzer(t)
	}

	cmd := covSeamSubCmd(t, newDepsCmd(), "security")
	out, err := covSeamCapture(t, func() error {
		return depsSecurityHandler(cmd, nil)
	})

	testutil.AssertNoError(t, err)
	testutil.AssertStringContains(t, out, "Pinned versions")
	testutil.AssertStringContains(t, out, "Floating versions")
}

// TestCovSeamConfigWizardWritesConfig overrides the newConfigWizard seam so the wizard
// reads scripted input (all defaults, declining the GitHub-token prompt) and never
// touches a real terminal, exercising configWizardHandler end-to-end.
func TestCovSeamConfigWizardWritesConfig(t *testing.T) {
	origCfg := globalConfig
	origWizard := newConfigWizard
	defer func() {
		globalConfig = origCfg
		newConfigWizard = origWizard
	}()

	globalConfig = internal.DefaultAppConfig()
	globalConfig.Quiet = true

	xdgHome := filepath.Join(t.TempDir(), "xdg")
	// Register the xdg.Reload cleanup before t.Setenv so it runs *after* Setenv's
	// env-restore cleanup (cleanups are LIFO). Otherwise Reload would re-cache paths
	// while XDG_CONFIG_HOME still points at the temp dir, leaking stale paths to
	// later tests in this package.
	t.Cleanup(xdg.Reload)
	t.Setenv("XDG_CONFIG_HOME", xdgHome)
	xdg.Reload()

	// Blank lines accept every default; "Set up GitHub token now?" defaults to false
	// (so promptSensitive is never reached) and "Save this configuration?" defaults to
	// true, completing the wizard without a terminal.
	newConfigWizard = func(output internal.MessageLogger) *wizard.ConfigWizard {
		return wizard.NewConfigWizardWithInput(output, strings.NewReader(strings.Repeat("\n", 12)))
	}

	cmd := covSeamSubCmd(t, newConfigCmd(), appconstants.CommandWizard)
	err := configWizardHandler(cmd, nil)
	testutil.AssertNoError(t, err)

	testutil.AssertFileExists(t, filepath.Join(xdgHome, "gh-action-readme", appconstants.ConfigFileNameFull))
}
