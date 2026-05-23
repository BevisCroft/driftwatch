package lineage

import (
	"testing"
	"time"

	"github.com/driftwatch/internal/drift"
)

func makeResults(fields ...string) []drift.Result {
	results := make([]drift.Result, 0, len(fields))
	for _, f := range fields {
		results = append(results, drift.Result{
			Service: "svc",
			Field:   f,
			Wanted:  "a",
			Got:     "b",
		})
	}
	return results
}

func TestRecord_AppendsEntry(t *testing.T) {
	tr := New(time.Hour)
	tr.Record("api", makeResults("spec.replicas"))

	entries := tr.Get("api")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if len(entries[0].Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(entries[0].Results))
	}
}

func TestGet_UnknownService_ReturnsEmpty(t *testing.T) {
	tr := New(time.Hour)
	if entries := tr.Get("nonexistent"); len(entries) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(entries))
	}
}

func TestRecord_PrunesExpiredEntries(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	tr := New(5 * time.Minute)
	tr.now = func() time.Time { return now }

	// Record an old entry.
	tr.Record("svc", makeResults("f1"))

	// Advance time past the retention window.
	tr.now = func() time.Time { return now.Add(10 * time.Minute) }
	tr.Record("svc", makeResults("f2"))

	entries := tr.Get("svc")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after pruning, got %d", len(entries))
	}
	if entries[0].Results[0].Field != "f2" {
		t.Errorf("expected retained entry to have field f2, got %s", entries[0].Results[0].Field)
	}
}

func TestServices_ReturnsAllTracked(t *testing.T) {
	tr := New(time.Hour)
	tr.Record("alpha", makeResults("x"))
	tr.Record("beta", makeResults("y"))

	svcs := tr.Services()
	if len(svcs) != 2 {
		t.Errorf("expected 2 services, got %d", len(svcs))
	}
}

func TestSummary_NoHistory(t *testing.T) {
	tr := New(time.Hour)
	s := tr.Summary("ghost")
	if s != "ghost: no history" {
		t.Errorf("unexpected summary: %q", s)
	}
}

func TestSummary_WithHistory(t *testing.T) {
	tr := New(time.Hour)
	tr.Record("api", makeResults("a", "b"))
	tr.Record("api", makeResults("c"))

	s := tr.Summary("api")
	expected := "api: 3 entries over 2 observations"
	if s != expected {
		t.Errorf("expected %q, got %q", expected, s)
	}
}
