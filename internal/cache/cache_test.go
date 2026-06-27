package cache

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

func TestNewCache(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name:        "default config",
			config:      nil,
			expectError: false,
		},
		{
			name: "custom config",
			config: &Config{
				DefaultTTL:      30 * time.Minute,
				CleanupInterval: 10 * time.Minute,
				MaxSize:         50 * 1024 * 1024,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set XDG_CACHE_HOME to temp directory
			tmpDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			t.Setenv("XDG_CACHE_HOME", tmpDir)

			cache, err := NewCache(tt.config)

			if tt.expectError {
				testutil.AssertError(t, err)

				return
			}

			testutil.AssertNoError(t, err)

			// Verify cache was created
			if cache == nil {
				t.Fatal("expected cache to be created")
			}

			// Verify default TTL
			expectedTTL := 15 * time.Minute
			if tt.config != nil && tt.config.DefaultTTL != 0 {
				expectedTTL = tt.config.DefaultTTL
			}
			testutil.AssertEqual(t, expectedTTL, cache.defaultTTL)

			// Clean up
			_ = cache.Close()
		})
	}
}

func TestCacheSetAndGet(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	cache := createTestCache(t, tmpDir)
	defer testutil.CleanupCache(t, cache)()

	tests := []struct {
		name     string
		key      string
		value    any
		expected any
	}{
		{
			name:     "string value",
			key:      testutil.CacheTestKey,
			value:    testutil.CacheTestValue,
			expected: testutil.CacheTestValue,
		},
		{
			name:     "struct value",
			key:      "struct-key",
			value:    map[string]string{"foo": "bar"},
			expected: map[string]string{"foo": "bar"},
		},
		{
			name:     "nil value",
			key:      "nil-key",
			value:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Set value
			err := cache.Set(tt.key, tt.value)
			testutil.AssertNoError(t, err)

			// Get value
			value, exists := cache.Get(tt.key)
			if !exists {
				t.Fatal("expected value to exist in cache")
			}

			testutil.AssertEqual(t, tt.expected, value)
		})
	}
}

func TestCacheTTL(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	cache := createTestCache(t, tmpDir)
	defer testutil.CleanupCache(t, cache)()

	// Set value with short TTL
	shortTTL := 100 * time.Millisecond
	err := cache.SetWithTTL(testutil.CacheShortLivedKey, "value", shortTTL)
	testutil.AssertNoError(t, err)

	// Should exist immediately
	value, exists := cache.Get(testutil.CacheShortLivedKey)
	if !exists {
		t.Fatal("expected value to exist immediately")
	}
	testutil.AssertEqual(t, "value", value)

	// Wait for expiration
	time.Sleep(shortTTL + 50*time.Millisecond)

	// Should not exist after TTL
	_, exists = cache.Get(testutil.CacheShortLivedKey)
	if exists {
		t.Error("expected value to be expired")
	}
}

func TestCacheGetOrSet(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	cache := createTestCache(t, tmpDir)
	defer testutil.CleanupCache(t, cache)()

	// Use unique key to avoid interference from other tests
	testKey := fmt.Sprintf("test-key-%d", time.Now().UnixNano())

	callCount := 0
	getter := func() (any, error) {
		callCount++

		return fmt.Sprintf("generated-value-%d", callCount), nil
	}

	// First call should invoke getter
	value1, err := cache.GetOrSet(testKey, getter)
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, "generated-value-1", value1)
	testutil.AssertEqual(t, 1, callCount)

	// Second call should use cached value
	value2, err := cache.GetOrSet(testKey, getter)
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, "generated-value-1", value2) // Same value
	testutil.AssertEqual(t, 1, callCount)                // Getter not called again
}

func TestCacheGetOrSetError(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	cache := createTestCache(t, tmpDir)
	defer testutil.CleanupCache(t, cache)()

	// Getter that returns error
	getter := func() (any, error) {
		return nil, errors.New("getter error")
	}

	value, err := cache.GetOrSet("error-key", getter)
	testutil.AssertError(t, err)
	testutil.AssertStringContains(t, err.Error(), "getter error")

	if value != nil {
		t.Errorf("expected nil value on error, got: %v", value)
	}

	// Verify nothing was cached
	_, exists := cache.Get("error-key")
	if exists {
		t.Error("expected no value to be cached on error")
	}
}

func TestCacheConcurrentAccess(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	cache := createTestCache(t, tmpDir)
	defer testutil.CleanupCache(t, cache)()

	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch multiple goroutines doing concurrent operations
	for i := 0; i < numGoroutines; i++ {
		go performConcurrentCacheOperations(t, cache, i, numOperations, &wg)
	}

	wg.Wait()
}

func performConcurrentCacheOperations(t *testing.T, cache *Cache, goroutineID, numOperations int, wg *sync.WaitGroup) {
	t.Helper()
	defer wg.Done()

	for j := 0; j < numOperations; j++ {
		key := fmt.Sprintf("key-%d-%d", goroutineID, j)
		value := fmt.Sprintf("value-%d-%d", goroutineID, j)

		// Set value
		err := cache.Set(key, value)
		if err != nil {
			t.Errorf("error setting value: %v", err)

			return
		}

		// Get value
		retrieved, exists := cache.Get(key)
		if !exists {
			t.Errorf("expected key %s to exist", key)

			return
		}

		if retrieved != value {
			t.Errorf("expected %s, got %s", value, retrieved)

			return
		}
	}
}

func TestCachePersistence(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Create cache and add some data
	cache1 := createTestCache(t, tmpDir)
	err := cache1.Set("persistent-key", "persistent-value")
	testutil.AssertNoError(t, err)

	// Close cache to trigger save
	err = cache1.Close()
	testutil.AssertNoError(t, err)

	// Create new cache instance (should load from disk)
	cache2 := createTestCache(t, tmpDir)
	defer testutil.CleanupCache(t, cache2)()

	// Value should still exist
	value, exists := cache2.Get("persistent-key")
	if !exists {
		t.Fatal("expected persistent value to exist after restart")
	}
	testutil.AssertEqual(t, "persistent-value", value)
}

func TestCacheClear(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	cache := createTestCache(t, tmpDir)
	defer testutil.CleanupCache(t, cache)()

	// Add some data
	_ = cache.Set(testutil.CacheTestKey1, testutil.CacheTestValue1)
	_ = cache.Set(testutil.CacheTestKey2, "value2")

	// Verify data exists
	_, exists1 := cache.Get(testutil.CacheTestKey1)
	_, exists2 := cache.Get(testutil.CacheTestKey2)
	if !exists1 || !exists2 {
		t.Fatal("expected test data to exist before clear")
	}

	// Clear cache
	err := cache.Clear()
	testutil.AssertNoError(t, err)

	// Verify data is gone
	_, exists1 = cache.Get(testutil.CacheTestKey1)
	_, exists2 = cache.Get(testutil.CacheTestKey2)
	if exists1 || exists2 {
		t.Error("expected data to be cleared")
	}
}

func TestCacheDelete(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	cache := createTestCache(t, tmpDir)
	defer testutil.CleanupCache(t, cache)()

	// Add some data
	_ = cache.Set(testutil.CacheTestKey1, testutil.CacheTestValue1)
	_ = cache.Set(testutil.CacheTestKey2, "value2")
	_ = cache.Set("key3", "value3")

	// Verify data exists
	_, exists := cache.Get(testutil.CacheTestKey1)
	if !exists {
		t.Fatal("expected key1 to exist before delete")
	}

	// Delete specific key
	cache.Delete(testutil.CacheTestKey1)

	// Verify deleted key is gone but others remain
	_, exists1 := cache.Get(testutil.CacheTestKey1)
	_, exists2 := cache.Get(testutil.CacheTestKey2)
	_, exists3 := cache.Get("key3")

	if exists1 {
		t.Error("expected key1 to be deleted")
	}
	if !exists2 || !exists3 {
		t.Error("expected key2 and key3 to still exist")
	}

	// Test deleting non-existent key (should not panic)
	cache.Delete("nonexistent")
}

func TestCacheStats(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	cache := createTestCache(t, tmpDir)
	defer testutil.CleanupCache(t, cache)()

	// Ensure cache starts clean
	_ = cache.Clear()

	// Add some data
	_ = cache.Set(testutil.CacheTestKey1, testutil.CacheTestValue1)
	_ = cache.Set(testutil.CacheTestKey2, "larger-value-with-more-content")

	stats := cache.Stats()

	// Check stats structure
	if _, ok := stats["cache_dir"]; !ok {
		t.Error("expected cache_dir in stats")
	}

	if _, ok := stats["total_entries"]; !ok {
		t.Error("expected total_entries in stats")
	}

	if _, ok := stats["total_size"]; !ok {
		t.Error("expected total_size in stats")
	}

	// Verify entry count
	totalEntries, ok := stats["total_entries"].(int)
	if !ok {
		t.Error("expected total_entries to be int")
	}
	if totalEntries != 2 {
		t.Errorf("expected 2 entries, got %d", totalEntries)
	}

	// Verify size is reasonable
	totalSize, ok := stats["total_size"].(int64)
	if !ok {
		t.Error("expected total_size to be int64")
	}
	if totalSize <= 0 {
		t.Errorf("expected positive total size, got %d", totalSize)
	}
}

func TestCacheCleanupExpiredEntries(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Create cache with short cleanup interval
	config := &Config{
		DefaultTTL:      50 * time.Millisecond,
		CleanupInterval: 30 * time.Millisecond,
		MaxSize:         1024 * 1024,
	}

	t.Setenv("XDG_CACHE_HOME", tmpDir)

	cache, err := NewCache(config)
	testutil.AssertNoError(t, err)
	defer testutil.CleanupCache(t, cache)()

	// Add entry that will expire
	err = cache.Set(testutil.CacheExpiringKey, "expiring-value")
	testutil.AssertNoError(t, err)

	// Verify it exists
	_, exists := cache.Get(testutil.CacheExpiringKey)
	if !exists {
		t.Fatal("expected entry to exist initially")
	}

	// Wait for cleanup to run
	time.Sleep(config.DefaultTTL + config.CleanupInterval + 20*time.Millisecond)

	// Entry should be cleaned up
	_, exists = cache.Get(testutil.CacheExpiringKey)
	if exists {
		t.Error("expected expired entry to be cleaned up")
	}
}

func TestCacheErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) *Cache
		testFunc    func(t *testing.T, cache *Cache)
		expectError bool
	}{
		{
			name: "invalid cache directory permissions",
			setupFunc: func(t *testing.T) *Cache {
				t.Helper()
				// This test would require special setup for permission testing
				// For now, we'll create a valid cache and test other error scenarios
				tmpDir, _ := testutil.TempDir(t)

				return createTestCache(t, tmpDir)
			},
			testFunc: func(t *testing.T, cache *Cache) {
				t.Helper()
				// Test setting a value that might cause issues during marshaling
				// Circular reference would cause JSON marshal to fail, but
				// Go's JSON package handles most cases gracefully
				err := cache.Set("test", "normal-value")
				testutil.AssertNoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := tt.setupFunc(t)
			defer testutil.CleanupCache(t, cache)()

			tt.testFunc(t, cache)
		})
	}
}

func TestCacheAsyncSaveErrorHandling(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	cache := createTestCache(t, tmpDir)
	defer testutil.CleanupCache(t, cache)()

	// This tests our new saveToDiskAsync error handling
	// Set a value to trigger async save
	err := cache.Set(testutil.CacheTestKey, testutil.CacheTestValue)
	testutil.AssertNoError(t, err)

	// Give some time for async save to complete
	time.Sleep(100 * time.Millisecond)

	// The async save should have completed without panicking
	// We can't easily test the error logging without capturing logs,
	// but we can verify the cache still works
	value, exists := cache.Get(testutil.CacheTestKey)
	if !exists {
		t.Error("expected value to exist after async save")
	}
	testutil.AssertEqual(t, testutil.CacheTestValue, value)
}

func TestCacheEstimateSize(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	cache := createTestCache(t, tmpDir)
	defer testutil.CleanupCache(t, cache)()

	tests := []struct {
		name    string
		value   any
		minSize int64
		maxSize int64
	}{
		{
			name:    "small string",
			value:   "test",
			minSize: 4,
			maxSize: 50,
		},
		{
			name:    "large string",
			value:   strings.Repeat("a", 1000),
			minSize: 1000,
			maxSize: 1100,
		},
		{
			name: "struct",
			value: map[string]any{
				testutil.CacheTestKey1: testutil.CacheTestValue1,
				testutil.CacheTestKey2: 42,
				"key3":                 []string{"a", "b", "c"},
			},
			minSize: 30,
			maxSize: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			size := cache.estimateSize(tt.value)
			if size < tt.minSize || size > tt.maxSize {
				t.Errorf("expected size between %d and %d, got %d", tt.minSize, tt.maxSize, size)
			}
		})
	}
}

// TestCacheClose_NilTicker kills the CONDITIONALS_NEGATION at cache.go:187
// (c.ticker != nil → c.ticker == nil). A negated mutation calls c.ticker.Stop()
// on a nil ticker, causing a nil pointer panic. This test verifies that Close()
// completes without panic when ticker is nil.
func TestCacheClose_NilTicker(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	c := &Cache{
		path: tmpDir,
		data: make(map[string]Entry),
		done: make(chan bool, 1),
		// ticker is nil (zero value)
	}

	err := c.Close()
	if err != nil {
		t.Errorf("Close() with nil ticker unexpected error: %v", err)
	}
}

// createTestCache creates a cache instance for testing.
func createTestCache(t *testing.T, tmpDir string) *Cache {
	t.Helper()

	t.Setenv("XDG_CACHE_HOME", tmpDir)

	cache, err := NewCache(DefaultConfig())
	testutil.AssertNoError(t, err)

	return cache
}

// TestCacheEvictsToMaxSize verifies that cleanup enforces the MaxSize bound by
// evicting entries (previously MaxSize was dead config and the cache was
// unbounded).
func TestCacheEvictsToMaxSize(t *testing.T) {
	t.Parallel()

	const maxSize = int64(60)
	cache, err := NewCache(&Config{
		DefaultTTL:      time.Hour, // long TTL so entries are not expiry-evicted
		CleanupInterval: time.Hour,
		MaxSize:         maxSize,
		Path:            t.TempDir(),
	})
	testutil.AssertNoError(t, err)
	defer testutil.CleanupCache(t, cache)()

	// Each ~12-byte value; 10 entries far exceed the 60-byte bound.
	for i := range 10 {
		testutil.AssertNoError(t, cache.Set(fmt.Sprintf("k%d", i), strings.Repeat("x", 10)))
	}

	cache.cleanup() // runs evictToMaxSize under the lock

	var total int64
	cache.mutex.RLock()
	remaining := len(cache.data)
	for _, e := range cache.data {
		total += e.Size
	}
	cache.mutex.RUnlock()

	if total > maxSize {
		t.Errorf("after eviction total size = %d, want <= %d", total, maxSize)
	}
	if remaining == 0 {
		t.Error("eviction removed all entries; expected some to remain under the bound")
	}
	if remaining == 10 {
		t.Error("no entries evicted; MaxSize bound not enforced")
	}
}

// TestNewCacheRejectsTraversalPath verifies an explicit config.Path containing a
// parent-traversal component is rejected before any directory is created.
func TestNewCacheRejectsTraversalPath(t *testing.T) {
	t.Parallel()

	// A literal relative "../" survives filepath.Clean (unlike filepath.Join,
	// which would resolve it away), so it reaches the guard.
	_, err := NewCache(&Config{Path: "../evil-cache"})
	if err == nil {
		t.Fatal("expected NewCache to reject a path containing '..'")
	}
	if !strings.Contains(err.Error(), "parent traversal") {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestCacheCloseProunesExpiredAndEnforcesMaxSize verifies the final save on Close
// prunes expired entries and enforces MaxSize. The cleanupLoop ticker never fires
// in a short-lived process, so without pruning on Close the persisted cache.json
// would grow without bound. CleanupInterval is set far in the future so the loop
// cannot run during the test — only Close's prune can produce the asserted state.
func TestCacheCloseProunesExpiredAndEnforcesMaxSize(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	const maxSize = int64(60)
	cfg := &Config{
		DefaultTTL:      time.Millisecond,
		CleanupInterval: time.Hour,
		MaxSize:         maxSize,
		Path:            dir,
	}

	cache, err := NewCache(cfg)
	testutil.AssertNoError(t, err)

	// Seed state directly under the lock (no Set, to avoid the async-save race):
	// one already-expired entry plus live entries that overflow MaxSize.
	freshCfg := *cfg
	freshCfg.DefaultTTL = time.Hour
	cache.mutex.Lock()
	cache.data["expired"] = Entry{Value: "x", ExpiresAt: time.Now().Add(-time.Hour), Size: 1}
	for i := range 10 {
		cache.data[fmt.Sprintf("k%d", i)] = Entry{
			Value:     strings.Repeat("x", 10),
			ExpiresAt: time.Now().Add(time.Hour),
			Size:      12,
		}
	}
	cache.mutex.Unlock()

	testutil.AssertNoError(t, cache.Close())

	// Reopen: loadFromDisk does not filter, so persisted state is observed verbatim.
	reopened, err := NewCache(&freshCfg)
	testutil.AssertNoError(t, err)
	defer testutil.CleanupCache(t, reopened)()

	reopened.mutex.RLock()
	_, expiredKept := reopened.data["expired"]
	var total int64
	for _, e := range reopened.data {
		total += e.Size
	}
	count := len(reopened.data)
	reopened.mutex.RUnlock()

	if expiredKept {
		t.Error("Close persisted an expired entry; expiry prune did not run on shutdown")
	}
	if total > maxSize {
		t.Errorf("Close persisted total size %d > MaxSize %d; eviction did not run on shutdown", total, maxSize)
	}
	if count == 0 {
		t.Error("Close evicted everything; expected some live entries to remain")
	}
}
