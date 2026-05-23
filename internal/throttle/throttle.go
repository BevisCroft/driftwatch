// Package throttle provides per-service notification throttling to prevent
// alert storms when many fields drift simultaneously within a short window.
package throttle

import (
	"sync"
	"time"
)

// Throttler limits how frequently notifications are emitted for a given service.
type Throttler struct {
	mu       sync.Mutex
	window   time.Duration
	maxBurst int
	records  map[string]*record
	now      func() time.Time
}

type record struct {
	count     int
	windowEnd time.Time
}

// New creates a Throttler that allows at most maxBurst notifications per
// service within the given window duration.
func New(window time.Duration, maxBurst int, now func() time.Time) *Throttler {
	if now == nil {
		now = time.Now
	}
	return &Throttler{
		window:   window,
		maxBurst: maxBurst,
		records:  make(map[string]*record),
		now:      now,
	}
}

// Allow reports whether a notification for the given service should be
// forwarded. It returns false once the burst limit is reached within the
// current window.
func (t *Throttler) Allow(service string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	r, ok := t.records[service]
	if !ok || now.After(r.windowEnd) {
		t.records[service] = &record{count: 1, windowEnd: now.Add(t.window)}
		return true
	}
	if r.count >= t.maxBurst {
		return false
	}
	r.count++
	return true
}

// Reset clears throttle state for the given service, allowing the next
// notification through regardless of the current window.
func (t *Throttler) Reset(service string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.records, service)
}

// Purge removes all expired window records to reclaim memory.
func (t *Throttler) Purge() {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := t.now()
	for svc, r := range t.records {
		if now.After(r.windowEnd) {
			delete(t.records, svc)
		}
	}
}
