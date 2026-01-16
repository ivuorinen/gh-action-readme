package internal

import (
	"path/filepath"
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

// boolFields represents the boolean configuration fields used in merge tests.
type boolFields struct {
	AnalyzeDependencies bool
	ShowSecurityInfo    bool
	Verbose             bool
	Quiet               bool
	UseDefaultBranch    bool
}

// createBoolFieldMergeTest creates a test table entry for testing boolean field merging.
// This helper reduces duplication by standardizing the creation of AppConfig test structures
// with boolean fields.
func createBoolFieldMergeTest(name string, dst, src, want boolFields) struct {
	name string
	dst  *AppConfig
	src  *AppConfig
	want *AppConfig
} {
	return struct {
		name string
		dst  *AppConfig
		src  *AppConfig
		want *AppConfig
	}{
		name: name,
		dst: &AppConfig{
			AnalyzeDependencies: dst.AnalyzeDependencies,
			ShowSecurityInfo:    dst.ShowSecurityInfo,
			Verbose:             dst.Verbose,
			Quiet:               dst.Quiet,
			UseDefaultBranch:    dst.UseDefaultBranch,
		},
		src: &AppConfig{
			AnalyzeDependencies: src.AnalyzeDependencies,
			ShowSecurityInfo:    src.ShowSecurityInfo,
			Verbose:             src.Verbose,
			Quiet:               src.Quiet,
			UseDefaultBranch:    src.UseDefaultBranch,
		},
		want: &AppConfig{
			AnalyzeDependencies: want.AnalyzeDependencies,
			ShowSecurityInfo:    want.ShowSecurityInfo,
			Verbose:             want.Verbose,
			Quiet:               want.Quiet,
			UseDefaultBranch:    want.UseDefaultBranch,
		},
	}
}

// createGitRemoteTestCase creates a test table entry for git remote detection tests.
// This helper reduces duplication for tests that set up a git repo with a remote config.
func createGitRemoteTestCase(
	name, configContent, expectedResult, description string,
) struct {
	name           string
	setupFunc      func(t *testing.T) string
	expectedResult string
	description    string
} {
	return struct {
		name           string
		setupFunc      func(t *testing.T) string
		expectedResult string
		description    string
	}{
		name: name,
		setupFunc: func(t *testing.T) string {
			t.Helper()
			tmpDir, _ := testutil.TempDir(t)
			testutil.InitGitRepo(t, tmpDir)

			if configContent != "" {
				configPath := filepath.Join(tmpDir, ".git", "config")
				testutil.WriteTestFile(t, configPath, configContent)
			}

			return tmpDir
		},
		expectedResult: expectedResult,
		description:    description,
	}
}
