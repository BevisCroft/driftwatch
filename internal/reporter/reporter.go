// Package reporter formats and outputs drift detection results.
package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/driftwatch/internal/drift"
)

// Format defines the output format for drift reports.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// Report holds a collection of drift results with metadata.
type Report struct {
	GeneratedAt time.Time      `json:"generated_at"`
	TotalChecked int           `json:"total_checked"`
	DriftCount   int           `json:"drift_count"`
	Results      []drift.Result `json:"results"`
}

// Reporter writes drift reports to a given writer.
type Reporter struct {
	w      io.Writer
	format Format
}

// New creates a new Reporter with the specified output writer and format.
func New(w io.Writer, format Format) *Reporter {
	return &Reporter{w: w, format: format}
}

// Write renders the report to the configured writer.
func (r *Reporter) Write(results []drift.Result) error {
	report := Report{
		GeneratedAt:  time.Now().UTC(),
		TotalChecked: len(results),
		Results:      results,
	}
	for _, res := range results {
		if res.HasDrift {
			report.DriftCount++
		}
	}

	switch r.format {
	case FormatJSON:
		return r.writeJSON(report)
	default:
		return r.writeText(report)
	}
}

func (r *Reporter) writeJSON(report Report) error {
	enc := json.NewEncoder(r.w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

func (r *Reporter) writeText(report Report) error {
	fmt.Fprintf(r.w, "Drift Report — %s\n", report.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(r.w, "Checked: %d | Drifted: %d\n\n", report.TotalChecked, report.DriftCount)
	for _, res := range report.Results {
		if !res.HasDrift {
			fmt.Fprintf(r.w, "[OK]    %s\n", res.Name)
			continue
		}
		fmt.Fprintf(r.w, "[DRIFT] %s\n", res.Name)
		for _, d := range res.Diffs {
			fmt.Fprintf(r.w, "        field=%s expected=%v actual=%v\n", d.Field, d.Expected, d.Actual)
		}
	}
	return nil
}
