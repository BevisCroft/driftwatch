package scorecard

import (
	"testing"
	"time"

	"github.com/example/driftwatch/internal/drift"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func cleanResult(svc string) drift.Result {
	return drift.Result{Service: svc, Drifted: false}
}

func driftedResult(svc string) drift.Result {
	return drift.Result{Service: svc, Drifted: true}
}

func TestGet_UnknownService_ReturnsFalse(t *testing.T) {
	sc := New(time.Minute)
	_, ok := sc.Get("unknown")
	if ok {
		t.Fatal("expected false for unknown service")
	}
}

func TestRecord_EmptyService_ReturnsError(t *testing.T) {
	sc := New(time.Minute)
	err := sc.Record(drift.Result{Service: ""})
	if err == nil {
		t.Fatal("expected error for empty service name")
	}
}

func TestScore_AllClean_ReturnsOne(t *testing.T) {
	sc := New(time.Minute)
	base := time.Now()
	sc.now = fixedNow(base)

	for i := 0; i < 5; i++ {
		if err := sc.Record(cleanResult("svc-a")); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	score, ok := sc.Get("svc-a")
	if !ok {
		t.Fatal("expected score to exist")
	}
	if score.Score != 1.0 {
		t.Errorf("expected 1.0, got %f", score.Score)
	}
	if score.DriftRuns != 0 || score.TotalRuns != 5 {
		t.Errorf("unexpected counts: drift=%d total=%d", score.DriftRuns, score.TotalRuns)
	}
}

func TestScore_AllDrifted_ReturnsZero(t *testing.T) {
	sc := New(time.Minute)
	sc.now = fixedNow(time.Now())

	for i := 0; i < 4; i++ {
		_ = sc.Record(driftedResult("svc-b"))
	}
	score, _ := sc.Get("svc-b")
	if score.Score != 0.0 {
		t.Errorf("expected 0.0, got %f", score.Score)
	}
}

func TestScore_MixedResults(t *testing.T) {
	sc := New(time.Minute)
	sc.now = fixedNow(time.Now())

	_ = sc.Record(cleanResult("svc-c"))
	_ = sc.Record(cleanResult("svc-c"))
	_ = sc.Record(driftedResult("svc-c"))
	_ = sc.Record(driftedResult("svc-c"))

	score, _ := sc.Get("svc-c")
	const want = 0.5
	if score.Score != want {
		t.Errorf("expected %f, got %f", want, score.Score)
	}
}

func TestEviction_OldEntriesExcluded(t *testing.T) {
	sc := New(10 * time.Second)
	old := time.Now().Add(-20 * time.Second)
	recent := time.Now()

	sc.now = fixedNow(old)
	_ = sc.Record(driftedResult("svc-d"))
	_ = sc.Record(driftedResult("svc-d"))

	sc.now = fixedNow(recent)
	_ = sc.Record(cleanResult("svc-d"))

	score, ok := sc.Get("svc-d")
	if !ok {
		t.Fatal("expected score")
	}
	if score.TotalRuns != 1 {
		t.Errorf("expected 1 run after eviction, got %d", score.TotalRuns)
	}
	if score.Score != 1.0 {
		t.Errorf("expected 1.0 after eviction, got %f", score.Score)
	}
}

func TestServices_ReturnsAllTracked(t *testing.T) {
	sc := New(time.Minute)
	sc.now = fixedNow(time.Now())

	_ = sc.Record(cleanResult("alpha"))
	_ = sc.Record(driftedResult("beta"))

	svcs := sc.Services()
	if len(svcs) != 2 {
		t.Errorf("expected 2 services, got %d", len(svcs))
	}
}
