// Package retrylog tracks per-service retry attempts and exposes
// aggregate statistics for observability and alerting.
package retrylog

import (
	"fmt"
	"sync"
	"time"
)

// Entry records a single retry event for a service.
type Entry struct {
	Service   string
	Attempt   int
	Reason    string
	Timestamp time.Time
}

// Summary holds aggregate retry statistics for a service.
type Summary struct {
	Service      string
	TotalRetries int
	LastAttempt  time.Time
	LastReason   string
}

// Log maintains an in-memory record of retry events per service.
type Log struct {
	mu      sync.RWMutex
	entries map[string][]Entry
	maxAge  time.Duration
	now     func() time.Time
}

// New creates a Log that prunes entries older than maxAge.
func New(maxAge time.Duration) *Log {
	return &Log{
		entries: make(map[string][]Entry),
		maxAge:  maxAge,
		now:     time.Now,
	}
}

// Record appends a retry event for the given service.
func (l *Log) Record(service, reason string, attempt int) error {
	if service == "" {
		return fmt.Errorf("retrylog: service name must not be empty")
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries[service] = append(l.entries[service], Entry{
		Service:   service,
		Attempt:   attempt,
		Reason:    reason,
		Timestamp: l.now(),
	})
	return nil
}

// Summaries returns aggregate statistics for all tracked services,
// excluding entries older than maxAge.
func (l *Log) Summaries() []Summary {
	l.mu.RLock()
	defer l.mu.RUnlock()
	cutoff := l.now().Add(-l.maxAge)
	var out []Summary
	for svc, entries := range l.entries {
		var recent []Entry
		for _, e := range entries {
			if !e.Timestamp.Before(cutoff) {
				recent = append(recent, e)
			}
		}
		if len(recent) == 0 {
			continue
		}
		last := recent[len(recent)-1]
		out = append(out, Summary{
			Service:      svc,
			TotalRetries: len(recent),
			LastAttempt:  last.Timestamp,
			LastReason:   last.Reason,
		})
	}
	return out
}

// Reset clears all retry entries for the given service.
func (l *Log) Reset(service string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, service)
}
