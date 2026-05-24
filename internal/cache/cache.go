// Package cache provides an in-memory key/value store with optional TTL
// and size-capped eviction for manifest and drift result caching.
package cache

import (
	"sync"
	"time"
)

// Entry holds a cached value along with its expiry.
type Entry struct {
	Value     interface{}
	ExpiresAt time.Time
}

// Cache is a thread-safe, TTL-aware in-memory store.
type Cache struct {
	mu      sync.RWMutex
	items   map[string]Entry
	ttl     time.Duration
	maxSize int
	now     func() time.Time
}

// New creates a Cache with the given TTL and maximum number of entries.
// A maxSize of 0 disables the size cap.
func New(ttl time.Duration, maxSize int) *Cache {
	return &Cache{
		items:   make(map[string]Entry),
		ttl:     ttl,
		maxSize: maxSize,
		now:     time.Now,
	}
}

// Set stores a value under key. If the cache is at capacity the oldest
// entry is evicted before inserting the new one.
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.maxSize > 0 && len(c.items) >= c.maxSize {
		c.evictOldest()
	}

	c.items[key] = Entry{
		Value:     value,
		ExpiresAt: c.now().Add(c.ttl),
	}
}

// Get retrieves a value by key. Returns (nil, false) if the key is absent
// or the entry has expired.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.items[key]
	if !ok || c.now().After(e.ExpiresAt) {
		return nil, false
	}
	return e.Value, true
}

// Delete removes a single entry from the cache.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Flush removes all entries from the cache.
func (c *Cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]Entry)
}

// Len returns the number of entries currently in the cache (including
// expired ones that have not yet been evicted).
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// evictOldest removes the entry with the earliest expiry. Must be called
// with the write lock held.
func (c *Cache) evictOldest() {
	var oldest string
	var oldestTime time.Time

	for k, e := range c.items {
		if oldest == "" || e.ExpiresAt.Before(oldestTime) {
			oldest = k
			oldestTime = e.ExpiresAt
		}
	}
	if oldest != "" {
		delete(c.items, oldest)
	}
}
