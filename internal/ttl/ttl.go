// Package ttl provides a generic time-to-live cache for drift-related state.
// Entries expire automatically after a configurable duration and can be
// invalidated explicitly. A background goroutine sweeps stale entries on a
// configurable interval.
package ttl

import (
	"sync"
	"time"
)

// entry holds a cached value and its expiry timestamp.
type entry struct {
	value     interface{}
	expiresAt time.Time
}

// Cache is a thread-safe TTL key-value store.
type Cache struct {
	mu      sync.RWMutex
	items   map[string]entry
	ttl     time.Duration
	now     func() time.Time
	stopCh  chan struct{}
}

// New creates a Cache with the given TTL and sweep interval.
// The sweep goroutine runs until Stop is called.
func New(ttl, sweepInterval time.Duration) *Cache {
	c := &Cache{
		items:  make(map[string]entry),
		ttl:    ttl,
		now:    time.Now,
		stopCh: make(chan struct{}),
	}
	go c.sweep(sweepInterval)
	return c
}

// Set stores a value under key, resetting its TTL.
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = entry{value: value, expiresAt: c.now().Add(c.ttl)}
}

// Get returns the value for key and whether it was found and not expired.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.items[key]
	if !ok || c.now().After(e.expiresAt) {
		return nil, false
	}
	return e.value, true
}

// Delete removes a key immediately regardless of TTL.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Len returns the number of non-expired entries currently held.
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	now := c.now()
	count := 0
	for _, e := range c.items {
		if !now.After(e.expiresAt) {
			count++
		}
	}
	return count
}

// Stop halts the background sweep goroutine.
func (c *Cache) Stop() {
	close(c.stopCh)
}

func (c *Cache) sweep(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.purge()
		case <-c.stopCh:
			return
		}
	}
}

func (c *Cache) purge() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.now()
	for k, e := range c.items {
		if now.After(e.expiresAt) {
			delete(c.items, k)
		}
	}
}
