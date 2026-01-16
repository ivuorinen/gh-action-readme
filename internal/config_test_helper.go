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

// createTokenMergeTest creates a test table entry for testing token merging behavior.
// This helper reduces duplication for the 4 token merge test cases.
func createTokenMergeTest(
	name, dstToken, srcToken, wantToken string,
	allowTokens bool,
) struct {
	name        string
	dst         *AppConfig
	src         *AppConfig
	allowTokens bool
	want        *AppConfig
} {
	return struct {
		name        string
		dst         *AppConfig
		src         *AppConfig
		allowTokens bool
		want        *AppConfig
	}{
		name:        name,
		dst:         &AppConfig{GitHubToken: dstToken},
		src:         &AppConfig{GitHubToken: srcToken},
		allowTokens: allowTokens,
		want:        &AppConfig{GitHubToken: wantToken},
	}
}

// createMapMergeTest creates a test table entry for testing map field merging (permissions/variables).
// This helper reduces duplication for tests that merge map[string]string fields.
func createMapMergeTest(
	name string,
	dstMap, srcMap, expectedMap map[string]string,
	isPermissions bool,
) struct {
	name     string
	dst      *AppConfig
	src      *AppConfig
	expected *AppConfig
} {
	dst := &AppConfig{}
	src := &AppConfig{}
	expected := &AppConfig{}

	if isPermissions {
		dst.Permissions = dstMap
		src.Permissions = srcMap
		expected.Permissions = expectedMap
	} else {
		dst.Variables = dstMap
		src.Variables = srcMap
		expected.Variables = expectedMap
	}

	return struct {
		name     string
		dst      *AppConfig
		src      *AppConfig
		expected *AppConfig
	}{
		name:     name,
		dst:      dst,
		src:      src,
		expected: expected,
	}
}
