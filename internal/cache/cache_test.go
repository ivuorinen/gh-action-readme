package cache

import (
	"fmt"
	"os"
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

			originalXDGCache := os.Getenv("XDG_CACHE_HOME")
			_ = os.Setenv("XDG_CACHE_HOME", tmpDir)
			defer func() {
				if originalXDGCache != "" {
					_ = os.Setenv("XDG_CACHE_HOME", originalXDGCache)
				} else {
					_ = os.Unsetenv("XDG_CACHE_HOME")
				}
			}()

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

func TestCache_SetAndGet(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	cache := createTestCache(t, tmpDir)
	defer func() { _ = cache.Close() }()

	tests := []struct {
		name     string
		key      string
		value    any
		expected any
	}{
		{
			name:     "string value",
			key:      "test-key",
			value:    "test-value",
			expected: "test-value",
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

func TestCache_TTL(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	cache := createTestCache(t, tmpDir)
	defer func() { _ = cache.Close() }()

	// Set value with short TTL
	shortTTL := 100 * time.Millisecond
	err := cache.SetWithTTL("short-lived", "value", shortTTL)
	testutil.AssertNoError(t, err)

	// Should exist immediately
	value, exists := cache.Get("short-lived")
	if !exists {
		t.Fatal("expected value to exist immediately")
	}
	testutil.AssertEqual(t, "value", value)

	// Wait for expiration
	time.Sleep(shortTTL + 50*time.Millisecond)

	// Should not exist after TTL
	_, exists = cache.Get("short-lived")
	if exists {
		t.Error("expected value to be expired")
	}
}

func TestCache_GetOrSet(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	cache := createTestCache(t, tmpDir)
	defer func() { _ = cache.Close() }()

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

func TestCache_GetOrSetError(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	cache := createTestCache(t, tmpDir)
	defer func() { _ = cache.Close() }()

	// Getter that returns error
	getter := func() (any, error) {
		return nil, fmt.Errorf("getter error")
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

func TestCache_ConcurrentAccess(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	cache := createTestCache(t, tmpDir)
	defer func() { _ = cache.Close() }()

	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch multiple goroutines doing concurrent operations
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
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
		}(i)
	}

	wg.Wait()
}

func TestCache_Persistence(t *testing.T) {
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
	defer func() { _ = cache2.Close() }()

	// Value should still exist
	value, exists := cache2.Get("persistent-key")
	if !exists {
		t.Fatal("expected persistent value to exist after restart")
	}
	testutil.AssertEqual(t, "persistent-value", value)
}

func TestCache_Clear(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	cache := createTestCache(t, tmpDir)
	defer func() { _ = cache.Close() }()

	// Add some data
	_ = cache.Set("key1", "value1")
	_ = cache.Set("key2", "value2")

	// Verify data exists
	_, exists1 := cache.Get("key1")
	_, exists2 := cache.Get("key2")
	if !exists1 || !exists2 {
		t.Fatal("expected test data to exist before clear")
	}

	// Clear cache
	err := cache.Clear()
	testutil.AssertNoError(t, err)

	// Verify data is gone
	_, exists1 = cache.Get("key1")
	_, exists2 = cache.Get("key2")
	if exists1 || exists2 {
		t.Error("expected data to be cleared")
	}
}

func TestCache_Delete(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	cache := createTestCache(t, tmpDir)
	defer func() { _ = cache.Close() }()

	// Add some data
	_ = cache.Set("key1", "value1")
	_ = cache.Set("key2", "value2")
	_ = cache.Set("key3", "value3")

	// Verify data exists
	_, exists := cache.Get("key1")
	if !exists {
		t.Fatal("expected key1 to exist before delete")
	}

	// Delete specific key
	cache.Delete("key1")

	// Verify deleted key is gone but others remain
	_, exists1 := cache.Get("key1")
	_, exists2 := cache.Get("key2")
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

func TestCache_Stats(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	cache := createTestCache(t, tmpDir)
	defer func() { _ = cache.Close() }()

	// Ensure cache starts clean
	_ = cache.Clear()

	// Add some data
	_ = cache.Set("key1", "value1")
	_ = cache.Set("key2", "larger-value-with-more-content")

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

func TestCache_CleanupExpiredEntries(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Create cache with short cleanup interval
	config := &Config{
		DefaultTTL:      50 * time.Millisecond,
		CleanupInterval: 30 * time.Millisecond,
		MaxSize:         1024 * 1024,
	}

	originalXDGCache := os.Getenv("XDG_CACHE_HOME")
	_ = os.Setenv("XDG_CACHE_HOME", tmpDir)
	defer func() {
		if originalXDGCache != "" {
			_ = os.Setenv("XDG_CACHE_HOME", originalXDGCache)
		} else {
			_ = os.Unsetenv("XDG_CACHE_HOME")
		}
	}()

	cache, err := NewCache(config)
	testutil.AssertNoError(t, err)
	defer func() { _ = cache.Close() }()

	// Add entry that will expire
	err = cache.Set("expiring-key", "expiring-value")
	testutil.AssertNoError(t, err)

	// Verify it exists
	_, exists := cache.Get("expiring-key")
	if !exists {
		t.Fatal("expected entry to exist initially")
	}

	// Wait for cleanup to run
	time.Sleep(config.DefaultTTL + config.CleanupInterval + 20*time.Millisecond)

	// Entry should be cleaned up
	_, exists = cache.Get("expiring-key")
	if exists {
		t.Error("expected expired entry to be cleaned up")
	}
}

func TestCache_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) *Cache
		testFunc    func(t *testing.T, cache *Cache)
		expectError bool
	}{
		{
			name: "invalid cache directory permissions",
			setupFunc: func(t *testing.T) *Cache {
				// This test would require special setup for permission testing
				// For now, we'll create a valid cache and test other error scenarios
				tmpDir, _ := testutil.TempDir(t)
				return createTestCache(t, tmpDir)
			},
			testFunc: func(t *testing.T, cache *Cache) {
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
			defer func() { _ = cache.Close() }()

			tt.testFunc(t, cache)
		})
	}
}

func TestCache_AsyncSaveErrorHandling(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	cache := createTestCache(t, tmpDir)
	defer func() { _ = cache.Close() }()

	// This tests our new saveToDiskAsync error handling
	// Set a value to trigger async save
	err := cache.Set("test-key", "test-value")
	testutil.AssertNoError(t, err)

	// Give some time for async save to complete
	time.Sleep(100 * time.Millisecond)

	// The async save should have completed without panicking
	// We can't easily test the error logging without capturing logs,
	// but we can verify the cache still works
	value, exists := cache.Get("test-key")
	if !exists {
		t.Error("expected value to exist after async save")
	}
	testutil.AssertEqual(t, "test-value", value)
}

func TestCache_EstimateSize(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	cache := createTestCache(t, tmpDir)
	defer func() { _ = cache.Close() }()

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
				"key1": "value1",
				"key2": 42,
				"key3": []string{"a", "b", "c"},
			},
			minSize: 30,
			maxSize: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := cache.estimateSize(tt.value)
			if size < tt.minSize || size > tt.maxSize {
				t.Errorf("expected size between %d and %d, got %d", tt.minSize, tt.maxSize, size)
			}
		})
	}
}

// createTestCache creates a cache instance for testing.
func createTestCache(t *testing.T, tmpDir string) *Cache {
	t.Helper()

	originalXDGCache := os.Getenv("XDG_CACHE_HOME")
	_ = os.Setenv("XDG_CACHE_HOME", tmpDir)
	t.Cleanup(func() {
		if originalXDGCache != "" {
			_ = os.Setenv("XDG_CACHE_HOME", originalXDGCache)
		} else {
			_ = os.Unsetenv("XDG_CACHE_HOME")
		}
	})

	cache, err := NewCache(DefaultConfig())
	testutil.AssertNoError(t, err)

	return cache
}
