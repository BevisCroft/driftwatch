// Package ownership maps services to their owning teams and provides
// lookup and HTTP handler support for drift attribution.
package ownership

import (
	"errors"
	"sync"
)

// Entry describes a single ownership record.
type Entry struct {
	Service string   `json:"service"`
	Team    string   `json:"team"`
	Contacts []string `json:"contacts,omitempty"`
}

// Registry holds ownership mappings for services.
type Registry struct {
	mu      sync.RWMutex
	entries map[string]Entry
}

// New returns an empty Registry.
func New() *Registry {
	return &Registry{
		entries: make(map[string]Entry),
	}
}

// Set registers or replaces the ownership entry for a service.
func (r *Registry) Set(e Entry) error {
	if e.Service == "" {
		return errors.New("ownership: service name must not be empty")
	}
	if e.Team == "" {
		return errors.New("ownership: team must not be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[e.Service] = e
	return nil
}

// Get returns the ownership entry for a service, or false if not found.
func (r *Registry) Get(service string) (Entry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.entries[service]
	return e, ok
}

// Remove deletes the ownership record for a service.
func (r *Registry) Remove(service string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.entries[service]
	if ok {
		delete(r.entries, service)
	}
	return ok
}

// All returns a snapshot of all registered entries.
func (r *Registry) All() []Entry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Entry, 0, len(r.entries))
	for _, e := range r.entries {
		out = append(out, e)
	}
	return out
}
