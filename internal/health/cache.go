package health

import (
	"sync"
	"time"
)

// CachedResult holds a health check result with a timestamp.
type CachedResult struct {
	Healthy   bool
	CheckedAt time.Time
}

// Cache stores recent health check results to reduce upstream gRPC calls.
type Cache struct {
	mu      sync.RWMutex
	results map[string]CachedResult
	ttl     time.Duration
}

// NewCache creates a Cache with the given TTL duration.
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		results: make(map[string]CachedResult),
		ttl:     ttl,
	}
}

// Get returns the cached result for a service, and whether it is still valid.
func (c *Cache) Get(service string) (CachedResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result, ok := c.results[service]
	if !ok {
		return CachedResult{}, false
	}
	if time.Since(result.CheckedAt) > c.ttl {
		return CachedResult{}, false
	}
	return result, true
}

// Set stores a health check result for a service.
func (c *Cache) Set(service string, healthy bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.results[service] = CachedResult{
		Healthy:   healthy,
		CheckedAt: time.Now(),
	}
}

// Invalidate removes the cached entry for a service.
func (c *Cache) Invalidate(service string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.results, service)
}
