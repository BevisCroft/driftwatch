// Package scorecard computes a drift health score for each tracked service
// based on accumulated drift results over a rolling window.
package scorecard

import (
	"fmt"
	"sync"
	"time"

	"github.com/example/driftwatch/internal/drift"
)

// Score represents the health score for a single service.
type Score struct {
	Service   string
	Score     float64 // 0.0 (fully drifted) – 1.0 (clean)
	DriftRuns int
	TotalRuns int
	UpdatedAt time.Time
}

// entry tracks per-run outcomes for a service.
type entry struct {
	at    time.Time
	drift bool
}

// Scorecard accumulates run history and computes rolling health scores.
type Scorecard struct {
	mu      sync.Mutex
	window  time.Duration
	history map[string][]entry
	now     func() time.Time
}

// New returns a Scorecard that evaluates scores over the given rolling window.
func New(window time.Duration) *Scorecard {
	return &Scorecard{
		window:  window,
		history: make(map[string][]entry),
		now:     time.Now,
	}
}

// Record adds a drift result to the service's history.
func (s *Scorecard) Record(result drift.Result) error {
	if result.Service == "" {
		return fmt.Errorf("scorecard: service name must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	s.history[result.Service] = append(s.history[result.Service], entry{
		at:    now,
		drift: result.Drifted,
	})
	s.evict(result.Service, now)
	return nil
}

// Get returns the current Score for a service, or false if unknown.
func (s *Scorecard) Get(service string) (Score, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, ok := s.history[service]
	if !ok || len(entries) == 0 {
		return Score{}, false
	}
	now := s.now()
	s.evict(service, now)
	entries = s.history[service]

	total := len(entries)
	driftCount := 0
	for _, e := range entries {
		if e.drift {
			driftCount++
		}
	}
	var sc float64
	if total > 0 {
		sc = 1.0 - float64(driftCount)/float64(total)
	}
	return Score{
		Service:   service,
		Score:     sc,
		DriftRuns: driftCount,
		TotalRuns: total,
		UpdatedAt: entries[len(entries)-1].at,
	}, true
}

// Services returns the names of all tracked services.
func (s *Scorecard) Services() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, 0, len(s.history))
	for svc := range s.history {
		out = append(out, svc)
	}
	return out
}

// evict removes entries older than the rolling window. Must be called with mu held.
func (s *Scorecard) evict(service string, now time.Time) {
	cutoff := now.Add(-s.window)
	entries := s.history[service]
	i := 0
	for i < len(entries) && entries[i].at.Before(cutoff) {
		i++
	}
	s.history[service] = entries[i:]
}
