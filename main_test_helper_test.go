package main

import (
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// testSimpleHandler is a helper for testing simple command handlers that:
// - Don't need specific setup beyond globalConfig
// - Return an error
// - Should complete without error
//
// This reduces duplication in tests like TestCacheClearHandler, TestCacheStatsHandler, etc.
func testSimpleHandler(
	t *testing.T,
	handlerFunc func(cmd *cobra.Command, args []string) error,
	handlerName string,
) {
	t.Helper()

	// Save and restore globalConfig
	originalConfig := globalConfig
	defer func() { globalConfig = originalConfig }()

	globalConfig = &internal.AppConfig{Quiet: true}

	// Execute handler
	cmd := &cobra.Command{}
	err := handlerFunc(cmd, []string{})
	if err != nil {
		t.Errorf("%s() unexpected error: %v", handlerName, err)
	}
}

// setupFixtureReturningPath is a helper for test setup functions that:
// - Write a single action fixture to tmpDir
// - Return []string{actionPath} pointing to the created action file
//
// This reduces duplication in tests that need the action file path for processing.
func setupFixtureReturningPath(fixturePath string) func(*testing.T, string) []string {
	return func(t *testing.T, tmpDir string) []string {
		t.Helper()
		actionPath := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
		testutil.WriteActionFixture(t, tmpDir, fixturePath)

		return []string{actionPath}
	}
}

// setupFixtureInDir is a helper for E2E test setup functions that:
// - Write a single action fixture to tmpDir
// - Don't return anything (void setupFunc)
//
// This reduces duplication in E2E integration tests where many cases write a single fixture.
func setupFixtureInDir(fixturePath string) func(*testing.T, string) {
	return func(t *testing.T, tmpDir string) {
		t.Helper()
		testutil.WriteActionFixture(t, tmpDir, fixturePath)
	}
}

// setupWithSingleFixture is a helper for test setup functions that:
// - Write a single action fixture to tmpDir
// - Return []string{tmpDir} for test processing
//
// This reduces duplication in genHandler tests where many cases follow the same pattern.
func setupWithSingleFixture(fixturePath string) func(*testing.T, string) []string {
	return func(t *testing.T, tmpDir string) []string {
		t.Helper()
		testutil.WriteActionFixture(t, tmpDir, fixturePath)

		return []string{tmpDir}
	}
}

// setupWithActionContent is a helper for test setup functions that:
// - Write action content to tmpDir/action.yml
// - Return []string{actionPath} pointing to the created action file
//
// This reduces duplication in tests that need to create action files from string content.
func setupWithActionContent(content string) func(*testing.T, string) []string {
	return func(t *testing.T, tmpDir string) []string {
		t.Helper()
		actionPath := filepath.Join(tmpDir, appconstants.ActionFileNameYML)
		testutil.WriteTestFile(t, actionPath, content)

		return []string{actionPath}
	}
}
