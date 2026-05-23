// Package lineage tracks the change history of a service's drift results,
// allowing operators to observe how configuration drift evolves over time.
package lineage

import (
	"fmt"
	"sync"
	"time"

	"github.com/driftwatch/internal/drift"
)

// Entry records a single point-in-time snapshot of drift results for a service.
type Entry struct {
	Service   string
	Timestamp time.Time
	Results   []drift.Result
}

// Tracker maintains an in-memory ring buffer of drift history per service.
type Tracker struct {
	mu      sync.RWMutex
	history map[string][]Entry
	maxAge  time.Duration
	now     func() time.Time
}

// New creates a Tracker that retains entries no older than maxAge.
func New(maxAge time.Duration) *Tracker {
	return &Tracker{
		history: make(map[string][]Entry),
		maxAge:  maxAge,
		now:     time.Now,
	}
}

// Record appends a new drift observation for the given service and prunes
// entries that have exceeded the retention window.
func (t *Tracker) Record(service string, results []drift.Result) {
	t.mu.Lock()
	defer t.mu.Unlock()

	entry := Entry{
		Service:   service,
		Timestamp: t.now(),
		Results:   results,
	}
	t.history[service] = append(t.history[service], entry)
	t.prune(service)
}

// Get returns all retained entries for the given service, oldest first.
// Returns an empty slice if no history exists.
func (t *Tracker) Get(service string) []Entry {
	t.mu.RLock()
	defer t.mu.RUnlock()

	entries := t.history[service]
	out := make([]Entry, len(entries))
	copy(out, entries)
	return out
}

// Services returns the list of services currently tracked.
func (t *Tracker) Services() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	svcs := make([]string, 0, len(t.history))
	for svc := range t.history {
		svcs = append(svcs, svc)
	}
	return svcs
}

// Summary returns a human-readable summary of drift change counts for a service.
func (t *Tracker) Summary(service string) string {
	entries := t.Get(service)
	if len(entries) == 0 {
		return fmt.Sprintf("%s: no history", service)
	}
	total := 0
	for _, e := range entries {
		total += len(e.Results)
	}
	return fmt.Sprintf("%s: %d entries over %d observations", service, total, len(entries))
}

// prune removes entries older than maxAge. Caller must hold t.mu.
func (t *Tracker) prune(service string) {
	cutoff := t.now().Add(-t.maxAge)
	entries := t.history[service]
	i := 0
	for i < len(entries) && entries[i].Timestamp.Before(cutoff) {
		i++
	}
	t.history[service] = entries[i:]
}
