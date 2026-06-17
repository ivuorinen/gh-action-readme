// Package cache provides XDG-compliant caching functionality for gh-action-readme.
package cache

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/adrg/xdg"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

// Entry represents a cached item with TTL support.
type Entry struct {
	Value     any       `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
	Size      int64     `json:"size"`
}

// Cache provides thread-safe caching with TTL and XDG compliance.
type Cache struct {
	path       string           // XDG cache directory
	data       map[string]Entry // In-memory cache
	mutex      sync.RWMutex     // Thread safety
	ticker     *time.Ticker     // Cleanup ticker
	done       chan bool        // Cleanup shutdown
	defaultTTL time.Duration    // Default TTL for entries
	maxSize    int64            // Max total entry size in bytes (0 = unbounded)
	saveWG     sync.WaitGroup   // Wait group for pending save operations
	saveMutex  sync.Mutex       // Serializes disk writes in saveToDisk
	closed     bool             // Set under mutex by Close; gates new async saves
}

// Config represents cache configuration.
type Config struct {
	DefaultTTL      time.Duration // Default TTL for entries
	CleanupInterval time.Duration // How often to clean expired entries
	MaxSize         int64         // Maximum cache size in bytes (0 = unlimited)
	// Path, when non-empty, overrides the XDG cache directory. It exists so
	// tests can supply an isolated temp directory instead of sharing the real
	// per-user cache (which would let cached entries leak between test cases).
	// Production code leaves this empty to use the XDG location.
	Path string
}

// DefaultConfig returns default cache configuration.
func DefaultConfig() *Config {
	return &Config{
		DefaultTTL:      15 * time.Minute,  // 15 minutes for API responses
		CleanupInterval: 5 * time.Minute,   // Clean up every 5 minutes
		MaxSize:         100 * 1024 * 1024, // 100MB max cache size
	}
}

// NewCache creates a new XDG-compliant cache instance.
func NewCache(config *Config) (*Cache, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Resolve the cache directory. An explicit config.Path (used by tests for
	// isolation) takes precedence; otherwise use the XDG cache location.
	cacheDir := config.Path
	if cacheDir == "" {
		// Resolve the full cache file path (<xdg-cache>/gh-action-readme/cache.json).
		// Passing the complete relative path namespaces the cache under its own
		// directory; passing only the app name would resolve c.path to the shared
		// XDG cache root and write cache.json there alongside other apps' files.
		cacheFile, err := xdg.CacheFile(filepath.Join(appconstants.AppName, appconstants.CacheJSON))
		if err != nil {
			return nil, fmt.Errorf("failed to get XDG cache directory: %w", err)
		}
		cacheDir = filepath.Dir(cacheFile)
	}
	// #nosec G301 -- cache directory permissions
	if err := os.MkdirAll(cacheDir, appconstants.FilePermDir); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	cache := &Cache{
		path:       cacheDir,
		data:       make(map[string]Entry),
		defaultTTL: config.DefaultTTL,
		maxSize:    config.MaxSize,
		done:       make(chan bool),
	}

	// Load existing cache from disk
	_ = cache.loadFromDisk() // Log error but don't fail - we can start with empty cache

	// Start cleanup goroutine
	cache.ticker = time.NewTicker(config.CleanupInterval)
	go cache.cleanupLoop()

	return cache, nil
}

// Set stores a value in the cache with default TTL.
func (c *Cache) Set(key string, value any) error {
	return c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores a value in the cache with custom TTL.
func (c *Cache) SetWithTTL(key string, value any, ttl time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Calculate size (rough estimate)
	size := c.estimateSize(value)

	entry := Entry{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
		Size:      size,
	}

	c.data[key] = entry

	// Persist to disk asynchronously
	c.saveToDiskAsync()

	return nil
}

// Get retrieves a value from the cache.
func (c *Cache) Get(key string) (any, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		// Remove expired entry (will be cleaned up by cleanup goroutine)
		return nil, false
	}

	return entry.Value, true
}

// Delete removes a key from the cache.
func (c *Cache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.data, key)
	// Use the tracked async save so Close()'s saveWG.Wait() waits for this
	// persist to finish; a bare goroutine would race past Close and be lost.
	c.saveToDiskAsync()
}

// Clear removes all entries from the cache.
func (c *Cache) Clear() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data = make(map[string]Entry)

	// Remove cache file
	cacheFile := filepath.Join(c.path, appconstants.CacheJSON)
	if err := os.Remove(cacheFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove cache file: %w", err)
	}

	return nil
}

// Stats returns cache statistics.
func (c *Cache) Stats() map[string]any {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var totalSize int64
	expiredCount := 0
	now := time.Now()

	for _, entry := range c.data {
		totalSize += entry.Size
		if now.After(entry.ExpiresAt) {
			expiredCount++
		}
	}

	return map[string]any{
		"total_entries": len(c.data),
		"expired_count": expiredCount,
		"total_size":    totalSize,
		"cache_dir":     c.path,
	}
}

// Close shuts down the cache and stops background processes.
func (c *Cache) Close() error {
	if c.ticker != nil {
		c.ticker.Stop()
	}

	// Signal cleanup goroutine to stop
	select {
	case c.done <- true:
	default:
	}

	// Mark closed under the mutex so no further saveToDiskAsync starts a new
	// WaitGroup task after Wait() returns (which would be a WaitGroup-reuse race
	// against concurrent Set/Delete/cleanup calls).
	c.mutex.Lock()
	c.closed = true
	c.mutex.Unlock()

	// Wait for any pending async save operations to complete
	c.saveWG.Wait()

	// Save final state to disk
	return c.saveToDisk()
}

// GetOrSet retrieves a value from cache or sets it if not found.
func (c *Cache) GetOrSet(key string, getter func() (any, error)) (any, error) {
	// Try to get from cache first
	if value, exists := c.Get(key); exists {
		return value, nil
	}

	// Not in cache, get from source
	value, err := getter()
	if err != nil {
		return nil, err
	}

	// Store in cache
	_ = c.Set(key, value) // Log error but don't fail - we have the value

	return value, nil
}

// cleanupLoop runs periodically to remove expired entries.
func (c *Cache) cleanupLoop() {
	for {
		select {
		case <-c.ticker.C:
			c.cleanup()
		case <-c.done:
			return
		}
	}
}

// cleanup removes expired entries.
func (c *Cache) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for key, entry := range c.data {
		if now.After(entry.ExpiresAt) {
			delete(c.data, key)
		}
	}

	// Enforce the size bound after expiry removal.
	c.evictToMaxSize()

	// Save to disk after cleanup
	c.saveToDiskAsync()
}

// evictToMaxSize removes entries (oldest expiry first) until the total entry
// size is within maxSize. A maxSize of 0 means unbounded. The caller must hold
// c.mutex.
func (c *Cache) evictToMaxSize() {
	if c.maxSize <= 0 {
		return
	}

	var total int64
	for _, entry := range c.data {
		total += entry.Size
	}
	if total <= c.maxSize {
		return
	}

	// Evict the entry with the earliest ExpiresAt repeatedly until under the
	// bound. O(n^2) worst case, but the cache holds few entries and eviction is
	// rare (cleanup interval + tiny entries).
	for total > c.maxSize && len(c.data) > 0 {
		var oldestKey string
		var oldestExp time.Time
		first := true
		for key, entry := range c.data {
			if first || entry.ExpiresAt.Before(oldestExp) {
				oldestKey, oldestExp, first = key, entry.ExpiresAt, false
			}
		}
		total -= c.data[oldestKey].Size
		delete(c.data, oldestKey)
	}
}

// loadFromDisk loads cache data from disk.
func (c *Cache) loadFromDisk() error {
	cacheFile := filepath.Join(c.path, appconstants.CacheJSON)

	data, err := os.ReadFile(cacheFile) // #nosec G304 -- cache file path constructed internally
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No cache file is fine
		}

		return fmt.Errorf("failed to read cache file: %w", err)
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if err := json.Unmarshal(data, &c.data); err != nil {
		return fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return nil
}

// saveToDisk persists cache data to disk.
func (c *Cache) saveToDisk() error {
	c.mutex.RLock()
	data := maps.Clone(c.data)
	c.mutex.RUnlock()

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	cacheFile := filepath.Join(c.path, appconstants.CacheJSON)

	// Serialize concurrent writers (Set/Delete/cleanup each spawn an async save)
	// and stage the write through a temp file + rename so a concurrent reader or
	// a crash mid-write never observes a torn cache.json.
	c.saveMutex.Lock()
	defer c.saveMutex.Unlock()

	tmp, err := os.CreateTemp(c.path, appconstants.CacheJSON+".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp cache file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op once the rename succeeds

	if _, err := tmp.Write(jsonData); err != nil {
		_ = tmp.Close()

		return fmt.Errorf("failed to write cache file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close temp cache file: %w", err)
	}
	// #nosec G302 -- cache file permissions
	if err := os.Chmod(tmpName, appconstants.FilePermDefault); err != nil {
		return fmt.Errorf("failed to set cache file permissions: %w", err)
	}
	if err := os.Rename(tmpName, cacheFile); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// saveToDiskAsync saves the cache to disk asynchronously.
// Cache save failures are non-critical and silently ignored.
//
// Callers must hold c.mutex (Set, Delete, cleanup do). The closed check under
// that lock guarantees no new WaitGroup task is started once Close has begun, so
// Close's saveWG.Wait never races a fresh saveWG.Go.
func (c *Cache) saveToDiskAsync() {
	if c.closed {
		return
	}
	c.saveWG.Go(func() {
		_ = c.saveToDisk() // Ignore errors - cache save failures are non-critical
	})
}

// estimateSize provides a rough estimate of the memory size of a value.
func (c *Cache) estimateSize(value any) int64 {
	// This is a simple estimation - could be improved with reflection
	jsonData, err := json.Marshal(value)
	if err != nil {
		return 100 // Default estimate
	}

	return int64(len(jsonData))
}
