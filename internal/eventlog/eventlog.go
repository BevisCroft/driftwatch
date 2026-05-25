// Package eventlog provides a bounded, in-memory log of drift events
// that can be queried by service name or time range.
package eventlog

import (
	"errors"
	"sync"
	"time"

	"github.com/org/driftwatch/internal/drift"
)

// Entry is a single recorded drift event.
type Entry struct {
	Service   string
	Timestamp time.Time
	Results   []drift.Result
}

// Log is a bounded, thread-safe event log.
type Log struct {
	mu      sync.RWMutex
	entries []Entry
	maxSize int
	now     func() time.Time
}

// New returns a Log that retains at most maxSize entries.
// Older entries are evicted when the limit is exceeded.
func New(maxSize int) (*Log, error) {
	if maxSize <= 0 {
		return nil, errors.New("eventlog: maxSize must be greater than zero")
	}
	return &Log{
		maxSize: maxSize,
		now:     time.Now,
	}, nil
}

// Record appends a drift event for the given service.
// If the log is full, the oldest entry is dropped.
func (l *Log) Record(service string, results []drift.Result) error {
	if service == "" {
		return errors.New("eventlog: service name must not be empty")
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := Entry{
		Service:   service,
		Timestamp: l.now(),
		Results:   results,
	}
	l.entries = append(l.entries, entry)
	if len(l.entries) > l.maxSize {
		l.entries = l.entries[len(l.entries)-l.maxSize:]
	}
	return nil
}

// Query returns all entries recorded for the given service.
// If service is empty, all entries are returned.
func (l *Log) Query(service string) []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var out []Entry
	for _, e := range l.entries {
		if service == "" || e.Service == service {
			out = append(out, e)
		}
	}
	return out
}

// Len returns the current number of entries in the log.
func (l *Log) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.entries)
}

// Clear removes all entries from the log.
func (l *Log) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = nil
}
