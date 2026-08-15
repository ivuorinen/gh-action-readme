package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

// initConfig is the PersistentPreRunE for every command, so a defect in it
// misconfigures the whole CLI. It reads and writes package-level state, which is
// why none of these tests are parallel.

// isolateInitConfigState saves the package-level state initConfig touches and
// restores it when the test ends, so these tests cannot leak into the rest of
// package main. It also redirects XDG_CONFIG_HOME and HOME at a temporary
// directory: with those left alone, initConfig would read the developer's real
// ~/.config/gh-action-readme/config.yaml and the assertions below would depend
// on the machine running them.
func isolateInitConfigState(t *testing.T) {
	t.Helper()

	origConfig, origFile, origVerbose, origQuiet := globalConfig, configFile, verbose, quiet
	t.Cleanup(func() {
		globalConfig, configFile, verbose, quiet = origConfig, origFile, origVerbose, origQuiet
	})

	empty := t.TempDir()
	testutil.SetupXDGEnv(t, empty, empty)

	configFile, verbose, quiet = "", false, false
}

// writeInitConfigFixture writes a config fixture to its own temp directory and
// returns the path, for use as the --config value.
func writeInitConfigFixture(t *testing.T, fixture string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	testutil.WriteTestFile(t, path, testutil.MustReadFixture(fixture))

	return path
}

// TestInitConfigLoadsGlobalConfig covers the success path: the named config file
// is loaded into globalConfig, which every command handler then reads.
func TestInitConfigLoadsGlobalConfig(t *testing.T) {
	isolateInitConfigState(t)
	configFile = writeInitConfigFixture(t, testutil.TestConfigGlobalDefault)

	if err := initConfig(nil, nil); err != nil {
		t.Fatalf("initConfig() unexpected error: %v", err)
	}

	if globalConfig == nil {
		t.Fatal("initConfig() left globalConfig nil")
	}
	if globalConfig.Theme != "default" {
		t.Errorf("Theme = %q, want %q", globalConfig.Theme, "default")
	}
	if globalConfig.OutputFormat != "md" {
		t.Errorf("OutputFormat = %q, want %q", globalConfig.OutputFormat, "md")
	}
	if globalConfig.Verbose {
		t.Error("Verbose should stay false when the flag is not set and the file says false")
	}
}

// TestInitConfigVerboseFlagOverridesFile covers the --verbose branch: the flag
// must win over a config file that disables it, which is the whole point of
// applying the flags after the load.
func TestInitConfigVerboseFlagOverridesFile(t *testing.T) {
	isolateInitConfigState(t)
	configFile = writeInitConfigFixture(t, testutil.TestConfigGlobalDefault)
	verbose = true

	if err := initConfig(nil, nil); err != nil {
		t.Fatalf("initConfig() unexpected error: %v", err)
	}

	if !globalConfig.Verbose {
		t.Error("--verbose did not override the config file's verbose: false")
	}
}

// TestInitConfigQuietOverridesVerbose pins the documented precedence: quiet wins
// over verbose even when both flags are given. Without the explicit reset, a
// `-qv` invocation would produce a config that is both quiet and verbose.
func TestInitConfigQuietOverridesVerbose(t *testing.T) {
	isolateInitConfigState(t)
	configFile = writeInitConfigFixture(t, testutil.TestConfigGlobalDefault)
	verbose, quiet = true, true

	if err := initConfig(nil, nil); err != nil {
		t.Fatalf("initConfig() unexpected error: %v", err)
	}

	if !globalConfig.Quiet {
		t.Error("Quiet = false, want true when --quiet is set")
	}
	if globalConfig.Verbose {
		t.Error("Verbose = true, want false: quiet must override verbose")
	}
}

// TestInitConfigWrapsLoadError covers the error branch. initConfig runs as
// PersistentPreRunE, so returning the error is what stops a command from running
// against a half-loaded config; the wrapping is what tells the user which stage
// failed.
func TestInitConfigWrapsLoadError(t *testing.T) {
	isolateInitConfigState(t)
	configFile = writeInitConfigFixture(t, testutil.TestConfigInvalidMalformed)

	err := initConfig(nil, nil)
	if err == nil {
		t.Fatal("initConfig() with a malformed config file returned nil, want an error")
	}
	if !strings.Contains(err.Error(), "failed to initialize configuration") {
		t.Errorf("error = %v, want it to name the initialization stage", err)
	}
}
