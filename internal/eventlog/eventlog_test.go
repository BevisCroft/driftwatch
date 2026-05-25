package eventlog

import (
	"testing"
	"time"

	"github.com/org/driftwatch/internal/drift"
)

func makeResults(fields ...string) []drift.Result {
	var rs []drift.Result
	for _, f := range fields {
		rs = append(rs, drift.Result{
			Service: "svc",
			Field:   f,
			Want:    "a",
			Got:     "b",
		})
	}
	return rs
}

func TestNew_ZeroMaxSize_ReturnsError(t *testing.T) {
	_, err := New(0)
	if err == nil {
		t.Fatal("expected error for maxSize=0")
	}
}

func TestRecord_And_Query_RoundTrip(t *testing.T) {
	l, _ := New(10)
	if err := l.Record("api", makeResults("spec.replicas")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entries := l.Query("api")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Service != "api" {
		t.Errorf("expected service=api, got %s", entries[0].Service)
	}
}

func TestRecord_EmptyService_ReturnsError(t *testing.T) {
	l, _ := New(10)
	if err := l.Record("", makeResults("f")); err == nil {
		t.Fatal("expected error for empty service")
	}
}

func TestQuery_AllEntries_WhenServiceEmpty(t *testing.T) {
	l, _ := New(10)
	_ = l.Record("alpha", makeResults("f1"))
	_ = l.Record("beta", makeResults("f2"))
	all := l.Query("")
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestRecord_EvictsOldestWhenFull(t *testing.T) {
	l, _ := New(3)
	for i := 0; i < 5; i++ {
		_ = l.Record("svc", makeResults("field"))
	}
	if l.Len() != 3 {
		t.Errorf("expected 3 entries after eviction, got %d", l.Len())
	}
}

func TestRecord_TimestampIsSet(t *testing.T) {
	before := time.Now()
	l, _ := New(5)
	_ = l.Record("svc", makeResults("f"))
	after := time.Now()

	entries := l.Query("svc")
	ts := entries[0].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp %v outside expected range [%v, %v]", ts, before, after)
	}
}

func TestClear_RemovesAllEntries(t *testing.T) {
	l, _ := New(10)
	_ = l.Record("svc", makeResults("f"))
	l.Clear()
	if l.Len() != 0 {
		t.Errorf("expected 0 entries after Clear, got %d", l.Len())
	}
}
