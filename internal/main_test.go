package internal

import (
	"os"
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

// TestMain clears the GitHub token environment variables for the whole package
// run. They outrank config-file values in the token resolution hierarchy, so a
// developer with GITHUB_TOKEN exported would otherwise see config-precedence
// tests fail on a clean checkout — TestConfigMerging and
// TestCreateDependencyAnalyzer_TokenGuard both did, with nothing in the failure
// pointing at their shell.
//
// Tests that exercise the environment layer (TestConfigTokenHierarchy,
// TestGetGitHubToken) set what they need with t.Setenv, which restores
// afterwards and is unaffected by this.
func TestMain(m *testing.M) {
	testutil.ClearGitHubTokenEnv()

	os.Exit(m.Run())
}
