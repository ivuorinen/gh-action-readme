package internal

import (
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// covCfgToken is a fake GitHub token reused across these coverage tests.
const covCfgToken = "cov-github-token" // #nosec G101 -- test token, not a real credential

// TestCovIntCfgNewGitHubClientWithToken covers the token branch of
// NewGitHubClient.
func TestCovIntCfgNewGitHubClientWithToken(t *testing.T) {
	t.Parallel()

	client, err := NewGitHubClient(covCfgToken)
	testutil.AssertNoError(t, err)
	if client == nil || client.Client == nil {
		t.Fatal("expected a non-nil client")
	}
	testutil.AssertEqual(t, covCfgToken, client.Token)
}

// TestCovIntCfgNewGitHubClientNoToken covers the no-token branch of
// NewGitHubClient.
func TestCovIntCfgNewGitHubClientNoToken(t *testing.T) {
	t.Parallel()

	client, err := NewGitHubClient("")
	testutil.AssertNoError(t, err)
	if client == nil || client.Client == nil {
		t.Fatal("expected a non-nil client")
	}
	testutil.AssertEqual(t, "", client.Token)
}

// TestCovIntCfgResolveTemplatePath exercises the absolute, embedded-marker, and
// filesystem-fallback branches of resolveTemplatePath.
func TestCovIntCfgResolveTemplatePath(t *testing.T) {
	t.Parallel()

	abs := filepath.Join(string(filepath.Separator), "abs", "template.tmpl")
	testutil.AssertEqual(t, abs, resolveTemplatePath(abs))

	// Default embedded template returns the path unchanged (the embedded marker).
	embedded := "templates/readme.tmpl"
	testutil.AssertEqual(t, embedded, resolveTemplatePath(embedded))

	// A non-existent, non-embedded relative path falls back to the input path.
	missing := "nonexistent/custom-template.tmpl"
	testutil.AssertEqual(t, missing, resolveTemplatePath(missing))
}

// TestCovIntCfgMergeDefaultsFieldsRunsBranding covers the Runs and Branding
// merge branches of mergeDefaultsFields.
func TestCovIntCfgMergeDefaultsFieldsRunsBranding(t *testing.T) {
	t.Parallel()

	dst := &AppConfig{}
	const wantColor = "green"
	src := &AppConfig{Defaults: DefaultValues{
		Runs:     map[string]any{"using": "composite"},
		Branding: Branding{Icon: appconstants.ActivityWorkflowType, Color: wantColor},
	}}

	MergeConfigs(dst, src, false)

	if len(dst.Defaults.Runs) != 1 {
		t.Fatalf("expected Runs merged, got %v", dst.Defaults.Runs)
	}
	testutil.AssertEqual(t, appconstants.ActivityWorkflowType, dst.Defaults.Branding.Icon)
	testutil.AssertEqual(t, wantColor, dst.Defaults.Branding.Color)
}

// TestCovIntCfgDetectRepositoryName covers the empty-root and non-git branches.
func TestCovIntCfgDetectRepositoryName(t *testing.T) {
	t.Parallel()

	testutil.AssertEqual(t, "", DetectRepositoryName(""))
	testutil.AssertEqual(t, "", DetectRepositoryName(t.TempDir()))
}

// TestCovIntCfgInitConfigFromFile loads an explicit config file via InitConfig.
func TestCovIntCfgInitConfigFromFile(t *testing.T) {
	tmpDir, cleanup := testutil.SetupTestEnvironment(t)
	defer cleanup()

	configPath := WriteConfigFixture(t, tmpDir, testutil.TestConfigGitHubVerbose)

	config, err := InitConfig(configPath)
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, testutil.TestThemeGitHub, config.Theme)
}

// TestCovIntCfgInitConfigMissingFile drives the read-error branch of InitConfig
// when an explicit, non-existent config file is requested.
func TestCovIntCfgInitConfigMissingFile(t *testing.T) {
	tmpDir, cleanup := testutil.SetupTestEnvironment(t)
	defer cleanup()

	_, err := InitConfig(filepath.Join(tmpDir, "does-not-exist.yaml"))
	testutil.AssertError(t, err)
}

// TestCovIntCfgValidateTheme covers the empty, built-in, custom-path, and
// unsupported branches of validateTheme.
func TestCovIntCfgValidateTheme(t *testing.T) {
	t.Parallel()

	cl := NewConfigurationLoader()

	testutil.AssertError(t, cl.validateTheme(""))
	testutil.AssertNoError(t, cl.validateTheme(appconstants.ThemeDefault))
	testutil.AssertNoError(t, cl.validateTheme("custom/path/readme.tmpl"))
	testutil.AssertError(t, cl.validateTheme("nonexistent-theme-name"))
}

// TestCovIntCfgApplyRepoOverridesNoRepo covers the no-repository-detected branch
// of applyRepoOverrides (a non-git temp dir resolves to an empty repo name).
func TestCovIntCfgApplyRepoOverridesNoRepo(t *testing.T) {
	t.Parallel()

	cl := NewConfigurationLoader()
	config := DefaultAppConfig()
	original := config.Theme

	cl.applyRepoOverrides(config, t.TempDir())

	// No repo detected: config is left untouched.
	testutil.AssertEqual(t, original, config.Theme)
}

// TestCovIntCfgLoadGlobalConfig covers loadGlobalConfig via the public
// LoadGlobalConfig entry point with an explicit config file.
func TestCovIntCfgLoadGlobalConfig(t *testing.T) {
	tmpDir, cleanup := testutil.SetupTestEnvironment(t)
	defer cleanup()

	configPath := WriteConfigFixture(t, tmpDir, testutil.TestConfigGitHubVerbose)

	config, err := NewConfigurationLoader().LoadGlobalConfig(configPath)
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, testutil.TestThemeGitHub, config.Theme)
}

// TestCovIntCfgLoadConfigurationDefaultsOnly exercises LoadConfiguration with
// only the defaults source enabled, driving the disabled-source short-circuit in
// loadGlobalStep and loadConfigStep.
func TestCovIntCfgLoadConfigurationDefaultsOnly(t *testing.T) {
	tmpDir, cleanup := testutil.SetupTestEnvironment(t)
	defer cleanup()

	loader := NewConfigurationLoaderWithOptions(ConfigurationOptions{
		EnabledSources: []appconstants.ConfigurationSource{appconstants.SourceDefaults},
	})

	config, err := loader.LoadConfiguration("", tmpDir, tmpDir)
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, testutil.TestThemeDefault, config.Theme)
}
