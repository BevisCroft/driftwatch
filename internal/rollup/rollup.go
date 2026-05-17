// Package rollup aggregates drift results across multiple services,
// grouping related changes to reduce noise in reports and alerts.
package rollup

import (
	"sort"
	"time"

	"github.com/driftwatch/internal/drift"
)

// Summary holds aggregated drift information for a single service.
type Summary struct {
	Service    string
	DriftCount int
	Fields     []string
	Severity   string
	At         time.Time
}

// Aggregator groups drift results into per-service summaries.
type Aggregator struct {
	errorThreshold int
}

// New returns an Aggregator. errorThreshold controls how many drifted
// fields must be present before a summary is promoted to "error" severity.
func New(errorThreshold int) *Aggregator {
	if errorThreshold <= 0 {
		errorThreshold = 3
	}
	return &Aggregator{errorThreshold: errorThreshold}
}

// Aggregate converts a slice of drift.Result values into per-service
// summaries, deduplicating field names and computing severity.
func (a *Aggregator) Aggregate(results []drift.Result) []Summary {
	type entry struct {
		fields map[string]struct{}
		at     time.Time
	}

	index := make(map[string]*entry)

	for _, r := range results {
		if !r.HasDrift {
			continue
		}
		e, ok := index[r.Service]
		if !ok {
			e = &entry{fields: make(map[string]struct{}), at: r.DetectedAt}
			index[r.Service] = e
		}
		for _, d := range r.Diffs {
			e.fields[d.Field] = struct{}{}
		}
		if r.DetectedAt.After(e.at) {
			e.at = r.DetectedAt
		}
	}

	summaries := make([]Summary, 0, len(index))
	for svc, e := range index {
		fields := make([]string, 0, len(e.fields))
		for f := range e.fields {
			fields = append(fields, f)
		}
		sort.Strings(fields)

		severity := "warn"
		if len(fields) >= a.errorThreshold {
			severity = "error"
		}

		summaries = append(summaries, Summary{
			Service:    svc,
			DriftCount: len(fields),
			Fields:     fields,
			Severity:   severity,
			At:         e.at,
		})
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Service < summaries[j].Service
	})
	return summaries
}
