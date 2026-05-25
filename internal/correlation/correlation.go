// Package correlation tracks relationships between drift events across services,
// allowing operators to identify cascading or co-occurring drift patterns.
package correlation

import (
	"sync"
	"time"

	"github.com/example/driftwatch/internal/drift"
)

// Entry represents a single recorded drift event used for correlation.
type Entry struct {
	Service   string
	Field     string
	RecordedAt time.Time
}

// Match describes a correlated pair of drift events.
type Match struct {
	ServiceA string
	ServiceB string
	Field    string
	Delta    time.Duration
}

// Tracker records drift events and surfaces correlations within a time window.
type Tracker struct {
	mu      sync.Mutex
	entries []Entry
	window  time.Duration
	now     func() time.Time
}

// New returns a Tracker that correlates events occurring within window.
func New(window time.Duration) *Tracker {
	return &Tracker{
		window: window,
		now:    time.Now,
	}
}

// Record ingests drift results for a service, storing each drifted field.
func (t *Tracker) Record(service string, results []drift.Result) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	for _, r := range results {
		if !r.Drifted {
			continue
		}
		for _, f := range r.Fields {
			t.entries = append(t.entries, Entry{
				Service:    service,
				Field:      f,
				RecordedAt: now,
			})
		}
	}
	t.evict(now)
}

// Correlate returns pairs of services that drifted on the same field within
// the configured window.
func (t *Tracker) Correlate() []Match {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.evict(t.now())

	// Group entries by field.
	byField := make(map[string][]Entry)
	for _, e := range t.entries {
		byField[e.Field] = append(byField[e.Field], e)
	}

	var matches []Match
	for field, group := range byField {
		for i := 0; i < len(group); i++ {
			for j := i + 1; j < len(group); j++ {
				a, b := group[i], group[j]
				if a.Service == b.Service {
					continue
				}
				delta := b.RecordedAt.Sub(a.RecordedAt)
				if delta < 0 {
					delta = -delta
				}
				matches = append(matches, Match{
					ServiceA: a.Service,
					ServiceB: b.Service,
					Field:    field,
					Delta:    delta,
				})
			}
		}
	}
	return matches
}

// evict removes entries older than the configured window. Caller must hold mu.
func (t *Tracker) evict(now time.Time) {
	cutoff := now.Add(-t.window)
	kept := t.entries[:0]
	for _, e := range t.entries {
		if e.RecordedAt.After(cutoff) {
			kept = append(kept, e)
		}
	}
	t.entries = kept
}
