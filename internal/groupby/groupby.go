// Package groupby provides utilities for grouping drift results by
// arbitrary keys such as namespace, team, severity, or custom labels.
package groupby

import (
	"fmt"
	"sort"

	"github.com/driftwatch/internal/drift"
)

// Key represents the dimension by which results are grouped.
type Key string

const (
	KeyNamespace Key = "namespace"
	KeySeverity  Key = "severity"
	KeyService   Key = "service"
)

// Group holds a label and the drift results that belong to it.
type Group struct {
	Label   string
	Results []drift.Result
}

// Grouper groups drift results by a chosen key.
type Grouper struct {
	extractors map[Key]func(drift.Result) string
}

// New returns a Grouper with built-in extractors registered.
func New() *Grouper {
	g := &Grouper{
		extractors: make(map[Key]func(drift.Result) string),
	}
	g.extractors[KeyService] = func(r drift.Result) string { return r.Service }
	g.extractors[KeySeverity] = func(r drift.Result) string { return string(r.Severity) }
	g.extractors[KeyNamespace] = func(r drift.Result) string {
		if ns, ok := r.Live["namespace"]; ok {
			return fmt.Sprintf("%v", ns)
		}
		return "default"
	}
	return g
}

// Register adds a custom extractor for the given key.
func (g *Grouper) Register(k Key, fn func(drift.Result) string) {
	g.extractors[k] = fn
}

// By groups the provided results using the named key. It returns an
// error if the key has no registered extractor. Groups are returned
// sorted by label for deterministic output.
func (g *Grouper) By(k Key, results []drift.Result) ([]Group, error) {
	fn, ok := g.extractors[k]
	if !ok {
		return nil, fmt.Errorf("groupby: unknown key %q", k)
	}

	buckets := make(map[string][]drift.Result)
	for _, r := range results {
		label := fn(r)
		buckets[label] = append(buckets[label], r)
	}

	groups := make([]Group, 0, len(buckets))
	for label, rs := range buckets {
		groups = append(groups, Group{Label: label, Results: rs})
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Label < groups[j].Label
	})
	return groups, nil
}
