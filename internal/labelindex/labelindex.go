// Package labelindex provides an in-memory index for looking up services
// by one or more key=value label pairs, supporting fast intersection queries.
package labelindex

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

// Index maps label key=value pairs to sets of service names.
type Index struct {
	mu    sync.RWMutex
	// index["env=prod"] -> set of service names
	index map[string]map[string]struct{}
	// labels[service] -> map of label key -> value
	labels map[string]map[string]string
}

// New returns an empty Index.
func New() *Index {
	return &Index{
		index:  make(map[string]map[string]struct{}),
		labels: make(map[string]map[string]string),
	}
}

// Set registers a service with the given labels, replacing any previous labels.
func (idx *Index) Set(service string, lbls map[string]string) error {
	if service == "" {
		return errors.New("labelindex: service name must not be empty")
	}
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Remove old label entries for this service.
	if old, ok := idx.labels[service]; ok {
		for k, v := range old {
			key := fmt.Sprintf("%s=%s", k, v)
			delete(idx.index[key], service)
			if len(idx.index[key]) == 0 {
				delete(idx.index, key)
			}
		}
	}

	// Store new labels.
	copy := make(map[string]string, len(lbls))
	for k, v := range lbls {
		copy[k] = v
		key := fmt.Sprintf("%s=%s", k, v)
		if idx.index[key] == nil {
			idx.index[key] = make(map[string]struct{})
		}
		idx.index[key][service] = struct{}{}
	}
	idx.labels[service] = copy
	return nil
}

// Remove deletes a service from the index entirely.
func (idx *Index) Remove(service string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	if old, ok := idx.labels[service]; ok {
		for k, v := range old {
			key := fmt.Sprintf("%s=%s", k, v)
			delete(idx.index[key], service)
			if len(idx.index[key]) == 0 {
				delete(idx.index, key)
			}
		}
		delete(idx.labels, service)
	}
}

// Lookup returns all services that match ALL of the provided key=value labels
// (intersection). An empty selector returns all known services.
func (idx *Index) Lookup(selector map[string]string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if len(selector) == 0 {
		out := make([]string, 0, len(idx.labels))
		for svc := range idx.labels {
			out = append(out, svc)
		}
		sort.Strings(out)
		return out
	}

	var result map[string]struct{}
	for k, v := range selector {
		key := fmt.Sprintf("%s=%s", k, v)
		set, ok := idx.index[key]
		if !ok {
			return nil
		}
		if result == nil {
			result = make(map[string]struct{}, len(set))
			for svc := range set {
				result[svc] = struct{}{}
			}
		} else {
			for svc := range result {
				if _, found := set[svc]; !found {
					delete(result, svc)
				}
			}
		}
	}

	out := make([]string, 0, len(result))
	for svc := range result {
		out = append(out, svc)
	}
	sort.Strings(out)
	return out
}
