package retrylog

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestRecord_EmptyService_ReturnsError(t *testing.T) {
	l := New(time.Minute)
	if err := l.Record("", "timeout", 1); err == nil {
		t.Fatal("expected error for empty service, got nil")
	}
}

func TestRecord_And_Summaries_RoundTrip(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	l := New(time.Hour)
	l.now = fixedNow(base)

	if err := l.Record("svc-a", "timeout", 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := l.Record("svc-a", "connection refused", 2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	summaries := l.Summaries()
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	s := summaries[0]
	if s.Service != "svc-a" {
		t.Errorf("expected svc-a, got %s", s.Service)
	}
	if s.TotalRetries != 2 {
		t.Errorf("expected 2 retries, got %d", s.TotalRetries)
	}
	if s.LastReason != "connection refused" {
		t.Errorf("unexpected last reason: %s", s.LastReason)
	}
}

func TestSummaries_ExcludesExpiredEntries(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	l := New(30 * time.Minute)

	// Record an old entry
	l.now = fixedNow(base.Add(-1 * time.Hour))
	_ = l.Record("svc-b", "old error", 1)

	// Record a recent entry
	l.now = fixedNow(base)
	_ = l.Record("svc-b", "recent error", 2)

	summaries := l.Summaries()
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if summaries[0].TotalRetries != 1 {
		t.Errorf("expected 1 recent retry, got %d", summaries[0].TotalRetries)
	}
	if summaries[0].LastReason != "recent error" {
		t.Errorf("unexpected reason: %s", summaries[0].LastReason)
	}
}

func TestSummaries_AllExpired_ReturnsEmpty(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	l := New(5 * time.Minute)
	l.now = fixedNow(base.Add(-1 * time.Hour))
	_ = l.Record("svc-c", "old", 1)
	l.now = fixedNow(base)

	if summaries := l.Summaries(); len(summaries) != 0 {
		t.Errorf("expected 0 summaries, got %d", len(summaries))
	}
}

func TestReset_ClearsService(t *testing.T) {
	l := New(time.Hour)
	_ = l.Record("svc-d", "err", 1)
	l.Reset("svc-d")
	if summaries := l.Summaries(); len(summaries) != 0 {
		t.Errorf("expected 0 summaries after reset, got %d", len(summaries))
	}
}

func TestMultipleServices_TrackedIndependently(t *testing.T) {
	l := New(time.Hour)
	_ = l.Record("alpha", "timeout", 1)
	_ = l.Record("beta", "refused", 1)
	_ = l.Record("beta", "refused", 2)

	summaries := l.Summaries()
	counts := make(map[string]int)
	for _, s := range summaries {
		counts[s.Service] = s.TotalRetries
	}
	if counts["alpha"] != 1 {
		t.Errorf("alpha: expected 1, got %d", counts["alpha"])
	}
	if counts["beta"] != 2 {
		t.Errorf("beta: expected 2, got %d", counts["beta"])
	}
}
