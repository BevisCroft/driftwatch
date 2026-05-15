package reporter_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/reporter"
)

func sampleResults() []drift.Result {
	return []drift.Result{
		{
			Name:    "service-a",
			HasDrift: false,
			Diffs:   nil,
		},
		{
			Name:    "service-b",
			HasDrift: true,
			Diffs: []drift.Diff{
				{Field: "spec.replicas", Expected: 3, Actual: 1},
			},
		},
	}
}

func TestWrite_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	r := reporter.New(&buf, reporter.FormatText)
	if err := r.Write(sampleResults()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "[OK]    service-a") {
		t.Errorf("expected OK line for service-a, got:\n%s", out)
	}
	if !strings.Contains(out, "[DRIFT] service-b") {
		t.Errorf("expected DRIFT line for service-b, got:\n%s", out)
	}
	if !strings.Contains(out, "spec.replicas") {
		t.Errorf("expected field name in output, got:\n%s", out)
	}
}

func TestWrite_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	r := reporter.New(&buf, reporter.FormatJSON)
	if err := r.Write(sampleResults()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var report reporter.Report
	if err := json.Unmarshal(buf.Bytes(), &report); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}
	if report.TotalChecked != 2 {
		t.Errorf("expected TotalChecked=2, got %d", report.TotalChecked)
	}
	if report.DriftCount != 1 {
		t.Errorf("expected DriftCount=1, got %d", report.DriftCount)
	}
}

func TestWrite_EmptyResults(t *testing.T) {
	var buf bytes.Buffer
	r := reporter.New(&buf, reporter.FormatText)
	if err := r.Write([]drift.Result{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(buf.String(), "DRIFT") {
		t.Error("expected no DRIFT lines for empty results")
	}
}
