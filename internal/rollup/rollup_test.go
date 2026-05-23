package rollup_test

import (
	"testing"
	"time"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/rollup"
)

var now = time.Now()

func makeResult(service string, hasDrift bool, fields ...string) drift.Result {
	diffs := make([]drift.Diff, 0, len(fields))
	for _, f := range fields {
		diffs = append(diffs, drift.Diff{Field: f, Want: "a", Got: "b"})
	}
	return drift.Result{
		Service:    service,
		HasDrift:   hasDrift,
		Diffs:      diffs,
		DetectedAt: now,
	}
}

func TestAggregate_NoDrift(t *testing.T) {
	a := rollup.New(3)
	results := []drift.Result{
		makeResult("svc-a", false),
		makeResult("svc-b", false),
	}
	summaries := a.Aggregate(results)
	if len(summaries) != 0 {
		t.Fatalf("expected 0 summaries, got %d", len(summaries))
	}
}

func TestAggregate_SingleService_WarnSeverity(t *testing.T) {
	a := rollup.New(3)
	results := []drift.Result{makeResult("svc-a", true, "replicas", "image")}
	summaries := a.Aggregate(results)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	s := summaries[0]
	if s.Service != "svc-a" {
		t.Errorf("unexpected service: %s", s.Service)
	}
	if s.Severity != "warn" {
		t.Errorf("expected warn, got %s", s.Severity)
	}
	if s.DriftCount != 2 {
		t.Errorf("expected drift count 2, got %d", s.DriftCount)
	}
}

func TestAggregate_SingleService_ErrorSeverity(t *testing.T) {
	a := rollup.New(3)
	results := []drift.Result{makeResult("svc-b", true, "replicas", "image", "env")}
	summaries := a.Aggregate(results)
	if summaries[0].Severity != "error" {
		t.Errorf("expected error severity, got %s", summaries[0].Severity)
	}
}

func TestAggregate_DeduplicatesFields(t *testing.T) {
	a := rollup.New(5)
	results := []drift.Result{
		makeResult("svc-c", true, "replicas", "image"),
		makeResult("svc-c", true, "image", "env"),
	}
	summaries := a.Aggregate(results)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if summaries[0].DriftCount != 3 {
		t.Errorf("expected 3 unique fields, got %d", summaries[0].DriftCount)
	}
}

func TestAggregate_MultipleServices_Sorted(t *testing.T) {
	a := rollup.New(3)
	results := []drift.Result{
		makeResult("zebra-svc", true, "replicas"),
		makeResult("alpha-svc", true, "image"),
	}
	summaries := a.Aggregate(results)
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}
	if summaries[0].Service != "alpha-svc" {
		t.Errorf("expected alpha-svc first, got %s", summaries[0].Service)
	}
}

func TestAggregate_EmptyResults(t *testing.T) {
	a := rollup.New(3)
	summaries := a.Aggregate([]drift.Result{})
	if len(summaries) != 0 {
		t.Fatalf("expected 0 summaries for empty input, got %d", len(summaries))
	}
}

func TestAggregate_NoDriftMixedWithDrift(t *testing.T) {
	a := rollup.New(3)
	results := []drift.Result{
		makeResult("svc-a", false),
		makeResult("svc-b", true, "replicas"),
		makeResult("svc-c", false),
	}
	summaries := a.Aggregate(results)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if summaries[0].Service != "svc-b" {
		t.Errorf("expected svc-b, got %s", summaries[0].Service)
	}
}
