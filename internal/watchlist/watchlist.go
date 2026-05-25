// Package watchlist manages a named set of services that driftwatch
// actively monitors. Entries can be added, removed, and queried with
// optional metadata attached to each service.
package watchlist

import (
	"errors"
	"sync"
	"time"
)

// Entry represents a single watched service.
type Entry struct {
	Service   string            `json:"service"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels,omitempty"`
	AddedAt   time.Time         `json:"added_at"`
}

// Watchlist holds the set of actively monitored services.
type Watchlist struct {
	mu      sync.RWMutex
	entries map[string]Entry
}

// New returns an empty Watchlist.
func New() *Watchlist {
	return &Watchlist{
		entries: make(map[string]Entry),
	}
}

// Add registers a service for monitoring. Returns an error if the
// service name is empty or already present.
func (w *Watchlist) Add(e Entry) error {
	if e.Service == "" {
		return errors.New("watchlist: service name must not be empty")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, exists := w.entries[e.Service]; exists {
		return errors.New("watchlist: service already registered: " + e.Service)
	}
	if e.AddedAt.IsZero() {
		e.AddedAt = time.Now()
	}
	w.entries[e.Service] = e
	return nil
}

// Remove deletes a service from the watchlist. Returns false if not found.
func (w *Watchlist) Remove(service string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	_, exists := w.entries[service]
	if exists {
		delete(w.entries, service)
	}
	return exists
}

// Get returns the entry for a service, or false if not found.
func (w *Watchlist) Get(service string) (Entry, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	e, ok := w.entries[service]
	return e, ok
}

// All returns a snapshot of all registered entries.
func (w *Watchlist) All() []Entry {
	w.mu.RLock()
	defer w.mu.RUnlock()
	out := make([]Entry, 0, len(w.entries))
	for _, e := range w.entries {
		out = append(out, e)
	}
	return out
}

// Contains reports whether the named service is being watched.
func (w *Watchlist) Contains(service string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	_, ok := w.entries[service]
	return ok
}
