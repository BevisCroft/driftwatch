// Package window provides a sliding time-window counter for tracking
// event frequencies over a configurable duration.
package window

import (
	"sync"
	"time"
)

// entry records a single timestamped event.
type entry struct {
	at time.Time
}

// Counter is a thread-safe sliding window counter keyed by service name.
type Counter struct {
	mu     sync.Mutex
	window time.Duration
	events map[string][]entry
	now    func() time.Time
}

// New creates a Counter with the given sliding window duration.
func New(window time.Duration) *Counter {
	return &Counter{
		window: window,
		events: make(map[string][]entry),
		now:    time.Now,
	}
}

// Add records one event for the given service and returns the total count
// of events within the current window.
func (c *Counter) Add(service string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()
	c.events[service] = append(c.events[service], entry{at: now})
	c.evict(service, now)
	return len(c.events[service])
}

// Count returns the number of events recorded for service within the window
// without adding a new event.
func (c *Counter) Count(service string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()
	c.evict(service, now)
	return len(c.events[service])
}

// Reset clears all recorded events for the given service.
func (c *Counter) Reset(service string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.events, service)
}

// Services returns the list of service names currently tracked.
func (c *Counter) Services() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	keys := make([]string, 0, len(c.events))
	for k := range c.events {
		keys = append(keys, k)
	}
	return keys
}

// evict removes events older than the window boundary. Must be called with mu held.
func (c *Counter) evict(service string, now time.Time) {
	cutoff := now.Add(-c.window)
	bucket := c.events[service]
	i := 0
	for i < len(bucket) && bucket[i].at.Before(cutoff) {
		i++
	}
	if i > 0 {
		c.events[service] = bucket[i:]
	}
}
