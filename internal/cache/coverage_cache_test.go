package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ivuorinen/gh-action-readme/appconstants"
	"github.com/ivuorinen/gh-action-readme/testutil"
)

// covCacheTTL is a reusable non-zero TTL/interval for cache configs built in
// these coverage tests (NewCache panics on a zero CleanupInterval).
const covCacheTTL = time.Minute

// newDirectCache builds a Cache struct directly (no background goroutine, no
// ticker) so saveToDisk/loadFromDisk/Stats/Clear can be exercised in isolation
// without needing Close().
func newDirectCache(path string) *Cache {
	return &Cache{
		path: path,
		data: make(map[string]Entry),
	}
}

// TestCovCacheNewCacheInvalidPath verifies that an explicit Path containing a
// parent-traversal component is rejected.
func TestCovCacheNewCacheInvalidPath(t *testing.T) {
	_, err := NewCache(&Config{
		DefaultTTL:      covCacheTTL,
		CleanupInterval: covCacheTTL,
		MaxSize:         1 << 20,
		Path:            "../escape",
	})
	testutil.AssertError(t, err)
}

// TestCovCacheNewCacheExplicitPath verifies that an explicit Path overrides the
// XDG location and the directory is created.
func TestCovCacheNewCacheExplicitPath(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	sub := filepath.Join(tmpDir, "explicit", "cache")
	cache, err := NewCache(&Config{
		DefaultTTL:      covCacheTTL,
		CleanupInterval: covCacheTTL,
		MaxSize:         1 << 20,
		Path:            sub,
	})
	testutil.AssertNoError(t, err)
	defer testutil.CleanupCache(t, cache)()

	testutil.AssertEqual(t, sub, cache.path)
	testutil.AssertFileExists(t, sub)
}

// TestCovCacheSaveToDiskMarshalError drives the json.MarshalIndent failure
// branch of saveToDisk using a value (a channel) that cannot be marshaled.
func TestCovCacheSaveToDiskMarshalError(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	c := newDirectCache(tmpDir)
	c.data[testutil.CacheTestKey1] = Entry{Value: make(chan int)}

	err := c.saveToDisk()
	testutil.AssertError(t, err)
}

// TestCovCacheSaveToDiskTempCreateError drives the os.CreateTemp failure branch
// by pointing the cache at a directory that does not exist.
func TestCovCacheSaveToDiskTempCreateError(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	c := newDirectCache(filepath.Join(tmpDir, "does-not-exist"))
	c.data[testutil.CacheTestKey1] = Entry{Value: testutil.CacheTestValue1}

	err := c.saveToDisk()
	testutil.AssertError(t, err)
}

// TestCovCacheSaveToDiskSuccess confirms the happy path writes the cache file.
func TestCovCacheSaveToDiskSuccess(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	c := newDirectCache(tmpDir)
	c.data[testutil.CacheTestKey1] = Entry{
		Value:     testutil.CacheTestValue1,
		ExpiresAt: time.Now().Add(covCacheTTL),
	}

	testutil.AssertNoError(t, c.saveToDisk())
	testutil.AssertFileExists(t, filepath.Join(tmpDir, appconstants.CacheJSON))
}

// TestCovCacheEstimateSize covers both the marshal-success and marshal-failure
// (default estimate) branches of estimateSize.
func TestCovCacheEstimateSize(t *testing.T) {
	c := newDirectCache(t.TempDir())

	// Marshalable value: size equals the JSON byte length.
	got := c.estimateSize(testutil.CacheTestValue1)
	marshaled, err := json.Marshal(testutil.CacheTestValue1)
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, int64(len(marshaled)), got)

	// Unmarshalable value (channel) falls back to the default estimate of 100.
	testutil.AssertEqual(t, int64(100), c.estimateSize(make(chan int)))
}

// TestCovCacheLoadFromDiskMissing verifies a missing cache file is not an error.
func TestCovCacheLoadFromDiskMissing(t *testing.T) {
	c := newDirectCache(t.TempDir())
	testutil.AssertNoError(t, c.loadFromDisk())
}

// TestCovCacheLoadFromDiskCorrupt drives the json.Unmarshal failure branch.
func TestCovCacheLoadFromDiskCorrupt(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	testutil.WriteFileInDir(t, tmpDir, appconstants.CacheJSON, "not-valid-json{")

	c := newDirectCache(tmpDir)
	err := c.loadFromDisk()
	testutil.AssertError(t, err)
}

// TestCovCacheLoadFromDiskUnreadable drives the non-ENOENT read-error branch by
// making the cache path a directory instead of a file (os.ReadFile fails with a
// non IsNotExist error regardless of process privileges).
func TestCovCacheLoadFromDiskUnreadable(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	cacheAsDir := filepath.Join(tmpDir, appconstants.CacheJSON)
	// #nosec G301 -- test directory permissions
	if err := os.MkdirAll(cacheAsDir, appconstants.FilePermDir); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	c := newDirectCache(tmpDir)
	err := c.loadFromDisk()
	testutil.AssertError(t, err)
}

// TestCovCacheLoadFromDiskRoundTrip confirms saved data loads back.
func TestCovCacheLoadFromDiskRoundTrip(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	writer := newDirectCache(tmpDir)
	writer.data[testutil.CacheTestKey1] = Entry{
		Value:     testutil.CacheTestValue1,
		ExpiresAt: time.Now().Add(covCacheTTL),
	}
	testutil.AssertNoError(t, writer.saveToDisk())

	reader := newDirectCache(tmpDir)
	testutil.AssertNoError(t, reader.loadFromDisk())
	if _, ok := reader.data[testutil.CacheTestKey1]; !ok {
		t.Fatalf("expected key %q to load from disk", testutil.CacheTestKey1)
	}
}

// TestCovCacheStatsExpiredCount covers the expired-entry counting branch of
// Stats.
func TestCovCacheStatsExpiredCount(t *testing.T) {
	c := newDirectCache(t.TempDir())
	c.data[testutil.CacheTestKey1] = Entry{
		Value:     testutil.CacheTestValue1,
		ExpiresAt: time.Now().Add(-covCacheTTL), // already expired
		Size:      10,
	}
	c.data[testutil.CacheTestKey2] = Entry{
		Value:     testutil.CacheTestValue2,
		ExpiresAt: time.Now().Add(covCacheTTL), // still valid
		Size:      20,
	}

	stats := c.Stats()
	testutil.AssertEqual(t, 2, stats["total_entries"])
	testutil.AssertEqual(t, 1, stats["expired_count"])
	testutil.AssertEqual(t, int64(30), stats["total_size"])
}

// TestCovCacheClearMissingFile verifies Clear succeeds (returns nil) when no
// cache file exists on disk.
func TestCovCacheClearMissingFile(t *testing.T) {
	c := newDirectCache(t.TempDir())
	c.data[testutil.CacheTestKey1] = Entry{Value: testutil.CacheTestValue1}

	testutil.AssertNoError(t, c.Clear())
	testutil.AssertEqual(t, 0, len(c.data))
}
