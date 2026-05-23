// Package debounce provides a per-service event debouncer that suppresses
// rapid successive drift notifications within a configurable quiet window.
package debounce

import (
	"sync"
	"time"
)

// Debouncer tracks the last notification time per service and suppresses
// events that arrive within the configured quiet window duration.
type Debouncer struct {
	mu     sync.Mutex
	window time.Duration
	last   map[string]time.Time
	nowFn  func() time.Time
}

// New creates a Debouncer with the given quiet window. Events for a service
// are suppressed if a previous event was recorded within that window.
func New(window time.Duration) *Debouncer {
	return &Debouncer{
		window: window,
		last:   make(map[string]time.Time),
		nowFn:  time.Now,
	}
}

// Allow returns true if the event for the given service should be forwarded,
// i.e. no event has been recorded within the quiet window. It records the
// current time as the last-seen timestamp when returning true.
func (d *Debouncer) Allow(service string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.nowFn()
	if t, ok := d.last[service]; ok && now.Sub(t) < d.window {
		return false
	}
	d.last[service] = now
	return true
}

// Reset clears the recorded timestamp for a service, allowing the next
// event to pass through immediately regardless of the window.
func (d *Debouncer) Reset(service string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.last, service)
}

// Len returns the number of services currently tracked by the debouncer.
func (d *Debouncer) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.last)
}

// Purge removes all entries whose last-seen time is older than the window,
// freeing memory for services that are no longer active.
func (d *Debouncer) Purge() {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.nowFn()
	for svc, t := range d.last {
		if now.Sub(t) >= d.window {
			delete(d.last, svc)
		}
	}
}

// LastSeen returns the time of the most recent allowed event for the given
// service and true, or the zero Time and false if the service has no recorded
// timestamp (i.e. it was never seen or has been reset/purged).
func (d *Debouncer) LastSeen(service string) (time.Time, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	t, ok := d.last[service]
	return t, ok
}
