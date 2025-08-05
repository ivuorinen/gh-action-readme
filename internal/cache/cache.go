// Package cache provides XDG-compliant caching functionality for gh-action-readme.
package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/adrg/xdg"
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
	saveWG     sync.WaitGroup   // Wait group for pending save operations
}

// Config represents cache configuration.
type Config struct {
	DefaultTTL      time.Duration // Default TTL for entries
	CleanupInterval time.Duration // How often to clean expired entries
	MaxSize         int64         // Maximum cache size in bytes (0 = unlimited)
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

	// Get XDG cache directory
	cacheDir, err := xdg.CacheFile("gh-action-readme")
	if err != nil {
		return nil, fmt.Errorf("failed to get XDG cache directory: %w", err)
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(filepath.Dir(cacheDir), 0750); err != nil { // #nosec G301 -- cache directory permissions
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	cache := &Cache{
		path:       filepath.Dir(cacheDir),
		data:       make(map[string]Entry),
		defaultTTL: config.DefaultTTL,
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
	go func() {
		_ = c.saveToDisk() // Async operation, error logged internally
	}()
}

// Clear removes all entries from the cache.
func (c *Cache) Clear() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data = make(map[string]Entry)

	// Remove cache file
	cacheFile := filepath.Join(c.path, "cache.json")
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

	// Wait for any pending async save operations to complete
	c.saveWG.Wait()

	// Save final state to disk
	return c.saveToDisk()
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

	// Save to disk after cleanup
	c.saveToDiskAsync()
}

// loadFromDisk loads cache data from disk.
func (c *Cache) loadFromDisk() error {
	cacheFile := filepath.Join(c.path, "cache.json")

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
	data := make(map[string]Entry)
	for k, v := range c.data {
		data[k] = v
	}
	c.mutex.RUnlock()

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	cacheFile := filepath.Join(c.path, "cache.json")
	if err := os.WriteFile(cacheFile, jsonData, 0600); err != nil { // #nosec G306 -- cache file permissions
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// saveToDiskAsync saves the cache to disk asynchronously.
// Cache save failures are non-critical and silently ignored.
func (c *Cache) saveToDiskAsync() {
	c.saveWG.Add(1)
	go func() {
		defer c.saveWG.Done()
		_ = c.saveToDisk() // Ignore errors - cache save failures are non-critical
	}()
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
