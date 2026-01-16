package main

import (
	"testing"

	"github.com/spf13/cobra"

	"github.com/ivuorinen/gh-action-readme/internal"
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
