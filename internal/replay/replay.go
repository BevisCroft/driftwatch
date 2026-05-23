// Package replay provides functionality for replaying historical audit log
// entries through the drift detection pipeline. This allows operators to
// retrospectively analyse configuration drift over a given time window
// without requiring a live polling cycle.
package replay

import (
	"context"
	"fmt"
	"time"

	"github.com/yourorg/driftwatch/internal/audit"
	"github.com/yourorg/driftwatch/internal/drift"
	"github.com/yourorg/driftwatch/internal/reporter"
)

// Entry represents a single replayed drift event reconstructed from an
// audit log record.
type Entry struct {
	// Timestamp is when the original drift event was recorded.
	Timestamp time.Time
	// Result holds the drift comparison outcome for the service.
	Result drift.Result
}

// Options controls which audit records are included in the replay window.
type Options struct {
	// From filters out records that were recorded before this time.
	// A zero value means no lower bound.
	From time.Time
	// To filters out records that were recorded after this time.
	// A zero value means no upper bound.
	To time.Time
	// Service restricts replay to a specific service name.
	// An empty string means all services are included.
	Service string
}

// Replayer reads from an audit log and re-emits drift results that fall
// within the configured time window.
type Replayer struct {
	reader  *audit.Reader
	reporter *reporter.Reporter
}

// New constructs a Replayer that sources records from the provided audit
// Reader and writes formatted output via the Reporter.
func New(r *audit.Reader, rep *reporter.Reporter) *Replayer {
	return &Replayer{
		reader:  r,
		reporter: rep,
	}
}

// Run executes the replay, collecting all audit entries that satisfy opts
// and writing the reconstructed drift results to the reporter. It returns
// the number of entries replayed and any terminal error encountered.
func (rp *Replayer) Run(ctx context.Context, opts Options) (int, error) {
	records, err := rp.reader.ReadAll()
	if err != nil {
		return 0, fmt.Errorf("replay: reading audit log: %w", err)
	}

	var results []drift.Result

	for _, rec := range records {
		if err := ctx.Err(); err != nil {
			return len(results), fmt.Errorf("replay: context cancelled: %w", err)
		}

		if !opts.From.IsZero() && rec.Timestamp.Before(opts.From) {
			continue
		}
		if !opts.To.IsZero() && rec.Timestamp.After(opts.To) {
			continue
		}
		if opts.Service != "" && rec.Service != opts.Service {
			continue
		}

		// Reconstruct a minimal drift.Result from the audit record so that
		// the existing reporter pipeline can render it without modification.
		result := drift.Result{
			Service:  rec.Service,
			HasDrift: rec.Event == "drift_detected",
		}
		results = append(results, result)
	}

	if err := rp.reporter.Write(results); err != nil {
		return len(results), fmt.Errorf("replay: writing report: %w", err)
	}

	return len(results), nil
}
