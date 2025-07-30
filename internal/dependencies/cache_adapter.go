package dependencies

import (
	"time"

	"github.com/ivuorinen/gh-action-readme/internal/cache"
)

// CacheAdapter adapts the cache.Cache to implement DependencyCache interface.
type CacheAdapter struct {
	cache *cache.Cache
}

// NewCacheAdapter creates a new cache adapter.
func NewCacheAdapter(c *cache.Cache) *CacheAdapter {
	return &CacheAdapter{cache: c}
}

// Get retrieves a value from the cache.
func (ca *CacheAdapter) Get(key string) (any, bool) {
	return ca.cache.Get(key)
}

// Set stores a value in the cache with default TTL.
func (ca *CacheAdapter) Set(key string, value any) error {
	return ca.cache.Set(key, value)
}

// SetWithTTL stores a value in the cache with custom TTL.
func (ca *CacheAdapter) SetWithTTL(key string, value any, ttl time.Duration) error {
	return ca.cache.SetWithTTL(key, value, ttl)
}

// NoOpCache implements DependencyCache with no-op operations for when caching is disabled.
type NoOpCache struct{}

// NewNoOpCache creates a new no-op cache.
func NewNoOpCache() *NoOpCache {
	return &NoOpCache{}
}

// Get always returns false (cache miss).
func (noc *NoOpCache) Get(_ string) (any, bool) {
	return nil, false
}

// Set does nothing.
func (noc *NoOpCache) Set(_ string, _ any) error {
	return nil
}

// SetWithTTL does nothing.
func (noc *NoOpCache) SetWithTTL(_ string, _ any, _ time.Duration) error {
	return nil
}
