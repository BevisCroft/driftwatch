// Package routing provides weighted round-robin routing across multiple
// manifest source endpoints, enabling driftwatch to fan out checks across
// clusters or remote manifest stores.
package routing

import (
	"errors"
	"sync"
)

// Endpoint represents a named manifest source with an associated weight.
type Endpoint struct {
	Name   string
	URL    string
	Weight int
}

// Router selects endpoints using weighted round-robin.
type Router struct {
	mu        sync.Mutex
	endpoints []Endpoint
	current   int
	counts    []int
}

// New creates a Router from the given endpoints. Returns an error if the
// slice is empty or any weight is non-positive.
func New(endpoints []Endpoint) (*Router, error) {
	if len(endpoints) == 0 {
		return nil, errors.New("routing: at least one endpoint required")
	}
	for _, e := range endpoints {
		if e.Weight <= 0 {
			return nil, errors.New("routing: endpoint weight must be positive")
		}
	}
	counts := make([]int, len(endpoints))
	return &Router{endpoints: endpoints, counts: counts}, nil
}

// Next returns the next endpoint according to weighted round-robin.
func (r *Router) Next() Endpoint {
	r.mu.Lock()
	defer r.mu.Unlock()

	for {
		ep := r.endpoints[r.current]
		if r.counts[r.current] < ep.Weight {
			r.counts[r.current]++
			return ep
		}
		r.counts[r.current] = 0
		r.current = (r.current + 1) % len(r.endpoints)
	}
}

// All returns a snapshot of all registered endpoints.
func (r *Router) All() []Endpoint {
	r.mu.Lock()
	defer r.mu.Unlock()
	snap := make([]Endpoint, len(r.endpoints))
	copy(snap, r.endpoints)
	return snap
}

// Reset resets the internal counters and position to the beginning.
func (r *Router) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.current = 0
	for i := range r.counts {
		r.counts[i] = 0
	}
}
