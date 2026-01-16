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

// testSimpleVoidHandler is a helper for testing void command handlers that:
// - Don't need specific setup beyond globalConfig
// - Don't return an error
// - Should complete without panicking
//
// This reduces duplication in tests like TestConfigThemesHandler, TestConfigShowHandler, etc.
func testSimpleVoidHandler(
	t *testing.T,
	handlerFunc func(cmd *cobra.Command, args []string),
) {
	t.Helper()

	// Save and restore globalConfig
	originalConfig := globalConfig
	defer func() { globalConfig = originalConfig }()

	globalConfig = &internal.AppConfig{Quiet: true}

	// Execute handler (should not panic)
	cmd := &cobra.Command{}
	handlerFunc(cmd, []string{})
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
