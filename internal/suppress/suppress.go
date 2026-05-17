// Package suppress provides a suppression list for drift results,
// allowing operators to silence known or accepted drift for a given
// service and field combination.
package suppress

import (
	"sync"
	"time"
)

// Entry represents a single suppression rule.
type Entry struct {
	Service   string
	Field     string
	Reason    string
	ExpiresAt time.Time
}

// IsExpired reports whether the suppression entry has passed its expiry time.
func (e Entry) IsExpired(now time.Time) bool {
	return !e.ExpiresAt.IsZero() && now.After(e.ExpiresAt)
}

// List manages a set of suppression entries.
type List struct {
	mu      sync.RWMutex
	entries []Entry
	now     func() time.Time
}

// New returns a new, empty suppression List.
func New() *List {
	return &List{now: time.Now}
}

// Add appends a new suppression entry to the list.
func (l *List) Add(e Entry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, e)
}

// IsSuppressed reports whether a given service+field pair is currently suppressed.
func (l *List) IsSuppressed(service, field string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	now := l.now()
	for _, e := range l.entries {
		if e.IsExpired(now) {
			continue
		}
		if e.Service == service && (e.Field == "*" || e.Field == field) {
			return true
		}
	}
	return false
}

// Purge removes all expired entries from the list.
func (l *List) Purge() {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	active := l.entries[:0]
	for _, e := range l.entries {
		if !e.IsExpired(now) {
			active = append(active, e)
		}
	}
	l.entries = active
}

// Snapshot returns a copy of all current (non-expired) entries.
func (l *List) Snapshot() []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	now := l.now()
	out := make([]Entry, 0, len(l.entries))
	for _, e := range l.entries {
		if !e.IsExpired(now) {
			out = append(out, e)
		}
	}
	return out
}
