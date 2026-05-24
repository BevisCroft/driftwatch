// Package tagindex provides an in-memory index for looking up manifests
// by arbitrary key-value tags attached during load or annotation.
package tagindex

import (
	"fmt"
	"sync"
)

// Entry holds a service name and its associated tags.
type Entry struct {
	Service string
	Tags    map[string]string
}

// Index maps tag key-value pairs to sets of service names.
type Index struct {
	mu      sync.RWMutex
	entries map[string]Entry           // service -> Entry
	inverted map[string]map[string]bool // "key=value" -> set of services
}

// New creates an empty Index.
func New() *Index {
	return &Index{
		entries:  make(map[string]Entry),
		inverted: make(map[string]map[string]bool),
	}
}

// Add registers a service with the given tags, replacing any prior entry.
func (idx *Index) Add(service string, tags map[string]string) error {
	if service == "" {
		return fmt.Errorf("tagindex: service name must not be empty")
	}
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Remove stale inverted entries for this service.
	if old, ok := idx.entries[service]; ok {
		for k, v := range old.Tags {
			key := k + "=" + v
			delete(idx.inverted[key], service)
		}
	}

	copy := make(map[string]string, len(tags))
	for k, v := range tags {
		copy[k] = v
		key := k + "=" + v
		if idx.inverted[key] == nil {
			idx.inverted[key] = make(map[string]bool)
		}
		idx.inverted[key][service] = true
	}
	idx.entries[service] = Entry{Service: service, Tags: copy}
	return nil
}

// Remove deletes a service from the index.
func (idx *Index) Remove(service string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	if old, ok := idx.entries[service]; ok {
		for k, v := range old.Tags {
			delete(idx.inverted[k+"="+v], service)
		}
		delete(idx.entries, service)
	}
}

// Lookup returns all service names that carry ALL of the supplied tags.
func (idx *Index) Lookup(tags map[string]string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var result []string
outer:
	for service, entry := range idx.entries {
		for k, v := range tags {
			if entry.Tags[k] != v {
				continue outer
			}
		}
		result = append(result, service)
	}
	return result
}

// Get returns the Entry for a service, and whether it was found.
func (idx *Index) Get(service string) (Entry, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	e, ok := idx.entries[service]
	return e, ok
}
