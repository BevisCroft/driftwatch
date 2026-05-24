// Package cooldown provides a per-service cooldown tracker that suppresses
// repeated drift notifications within a configurable quiet period.
package cooldown

import (
	"sync"
	"time"
)

// entry holds the expiry time of an active cooldown for a service.
type entry struct {
	expiry time.Time
}

// Tracker tracks per-service cooldown windows.
type Tracker struct {
	mu       sync.Mutex
	entries  map[string]entry
	duration time.Duration
	now      func() time.Time
}

// New creates a Tracker with the given cooldown duration.
func New(duration time.Duration) *Tracker {
	return &Tracker{
		entries:  make(map[string]entry),
		duration: duration,
		now:      time.Now,
	}
}

// Allow returns true if the service is not currently in a cooldown window,
// and starts a new cooldown window for it. Returns false if suppressed.
func (t *Tracker) Allow(service string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	if e, ok := t.entries[service]; ok && now.Before(e.expiry) {
		return false
	}
	t.entries[service] = entry{expiry: now.Add(t.duration)}
	return true
}

// Reset clears the cooldown for a specific service, allowing the next
// notification through immediately.
func (t *Tracker) Reset(service string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.entries, service)
}

// Purge removes all expired cooldown entries.
func (t *Tracker) Purge() {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	for svc, e := range t.entries {
		if now.After(e.expiry) {
			delete(t.entries, svc)
		}
	}
}

// Active returns the number of services currently within a cooldown window.
func (t *Tracker) Active() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	count := 0
	for _, e := range t.entries {
		if now.Before(e.expiry) {
			count++
		}
	}
	return count
}
