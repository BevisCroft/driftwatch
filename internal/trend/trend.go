// Package trend tracks drift frequency over time and surfaces services
// that are drifting repeatedly within a configurable observation window.
package trend

import (
	"sync"
	"time"
)

// Entry records a single drift observation for a service.
type Entry struct {
	Service   string
	FieldPath string
	ObservedAt time.Time
}

// Summary describes the drift trend for a single service.
type Summary struct {
	Service string
	Count   int
	Fields  []string
	First   time.Time
	Last    time.Time
}

// Tracker accumulates drift observations and reports trending services.
type Tracker struct {
	mu      sync.Mutex
	window  time.Duration
	entries []Entry
	now     func() time.Time
}

// New returns a Tracker that keeps observations within the given window.
func New(window time.Duration) *Tracker {
	return &Tracker{
		window: window,
		now:    time.Now,
	}
}

// Record adds a drift observation for the given service and field.
func (t *Tracker) Record(service, fieldPath string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.entries = append(t.entries, Entry{
		Service:    service,
		FieldPath:  fieldPath,
		ObservedAt: t.now(),
	})
}

// Summaries returns per-service drift summaries for observations within
// the observation window, ordered by descending drift count.
func (t *Tracker) Summaries() []Summary {
	t.mu.Lock()
	defer t.mu.Unlock()

	cutoff := t.now().Add(-t.window)
	type bucket struct {
		count  int
		fields map[string]struct{}
		first  time.Time
		last   time.Time
	}
	buckets := make(map[string]*bucket)

	for _, e := range t.entries {
		if e.ObservedAt.Before(cutoff) {
			continue
		}
		b, ok := buckets[e.Service]
		if !ok {
			b = &bucket{fields: make(map[string]struct{}), first: e.ObservedAt}
			buckets[e.Service] = b
		}
		b.count++
		b.fields[e.FieldPath] = struct{}{}
		if e.ObservedAt.Before(b.first) {
			b.first = e.ObservedAt
		}
		if e.ObservedAt.After(b.last) {
			b.last = e.ObservedAt
		}
	}

	summaries := make([]Summary, 0, len(buckets))
	for svc, b := range buckets {
		fields := make([]string, 0, len(b.fields))
		for f := range b.fields {
			fields = append(fields, f)
		}
		summaries = append(summaries, Summary{
			Service: svc,
			Count:   b.count,
			Fields:  fields,
			First:   b.first,
			Last:    b.last,
		})
	}
	// sort descending by count
	for i := 1; i < len(summaries); i++ {
		for j := i; j > 0 && summaries[j].Count > summaries[j-1].Count; j-- {
			summaries[j], summaries[j-1] = summaries[j-1], summaries[j]
		}
	}
	return summaries
}

// Purge removes all observations older than the observation window.
func (t *Tracker) Purge() {
	t.mu.Lock()
	defer t.mu.Unlock()
	cutoff := t.now().Add(-t.window)
	filtered := t.entries[:0]
	for _, e := range t.entries {
		if !e.ObservedAt.Before(cutoff) {
			filtered = append(filtered, e)
		}
	}
	t.entries = filtered
}
