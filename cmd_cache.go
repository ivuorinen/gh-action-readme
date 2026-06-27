// Package main provides the cache management commands.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/internal/cache"
)

func newCacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   appconstants.CommandCache,
		Short: "Cache management commands",
		Long:  "Manage the XDG-compliant dependency cache",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "clear",
		Short: "Clear the dependency cache",
		Run:   wrapHandlerWithErrorHandling(cacheClearHandler),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   appconstants.CommandStats,
		Short: "Show cache statistics",
		Run:   wrapHandlerWithErrorHandling(cacheStatsHandler),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Show cache directory path",
		Run:   wrapHandlerWithErrorHandling(cachePathHandler),
	})

	return cmd
}

func cacheClearHandler(_ *cobra.Command, _ []string) error {
	output := createOutputManager(globalConfig.Quiet)
	output.Info("Clearing dependency cache...")

	// Create a cache instance
	cacheInstance, err := cache.NewCache(cache.DefaultConfig())
	if err != nil {
		return wrapError(appconstants.ErrFailedToAccessCache, err)
	}

	if err := cacheInstance.Clear(); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	output.Success("Cache cleared successfully")

	return nil
}

func cacheStatsHandler(_ *cobra.Command, _ []string) error {
	output := createOutputManager(globalConfig.Quiet)

	// Create a cache instance
	cacheInstance, err := cache.NewCache(cache.DefaultConfig())
	if err != nil {
		return wrapError(appconstants.ErrFailedToAccessCache, err)
	}

	stats := cacheInstance.Stats()

	cacheDir, ok := stats[appconstants.CacheStatsKeyDir].(string)
	if !ok {
		cacheDir = appconstants.CachePathUnknown
	}

	output.Bold("Cache Statistics:")
	output.Printf("Cache location: %s\n", cacheDir)
	output.Printf("Total entries: %d\n", stats["total_entries"])
	output.Printf("Expired entries: %d\n", stats["expired_count"])

	// Format size nicely
	totalSize, ok := stats["total_size"].(int64)
	if !ok {
		totalSize = 0
	}
	sizeStr := formatSize(totalSize)
	output.Printf("Total size: %s\n", sizeStr)

	return nil
}

func cachePathHandler(_ *cobra.Command, _ []string) error {
	output := createOutputManager(globalConfig.Quiet)

	// Create a cache instance
	cacheInstance, err := cache.NewCache(cache.DefaultConfig())
	if err != nil {
		return wrapError(appconstants.ErrFailedToAccessCache, err)
	}

	stats := cacheInstance.Stats()
	cachePath, ok := stats[appconstants.CacheStatsKeyDir].(string)
	if !ok {
		cachePath = appconstants.CachePathUnknown
	}

	output.Bold("Cache Directory:")
	output.Printf("%s\n", cachePath)

	// Only stat a real path. The sentinel is not a filesystem path, so statting
	// it would resolve a bogus relative path in the current directory.
	if !ok {
		output.Warning("Cache directory could not be determined")

		return nil
	}

	// Check if directory exists
	if _, err := os.Stat(cachePath); err == nil {
		output.Success("Directory exists")
	} else if os.IsNotExist(err) {
		output.Warning("Directory does not exist (will be created on first use)")
	}

	return nil
}
