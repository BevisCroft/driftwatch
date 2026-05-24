package trend

import (
	"testing"
	"time"
)

var epoch = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func newTracker(now func() time.Time) *Tracker {
	tr := New(10 * time.Minute)
	tr.now = now
	return tr
}

func TestSummaries_EmptyTracker(t *testing.T) {
	tr := newTracker(fixedNow(epoch))
	if got := tr.Summaries(); len(got) != 0 {
		t.Fatalf("expected empty summaries, got %d", len(got))
	}
}

func TestRecord_SingleService(t *testing.T) {
	tr := newTracker(fixedNow(epoch))
	tr.Record("svc-a", "spec.replicas")
	tr.Record("svc-a", "spec.image")

	sums := tr.Summaries()
	if len(sums) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(sums))
	}
	if sums[0].Count != 2 {
		t.Errorf("expected count 2, got %d", sums[0].Count)
	}
	if len(sums[0].Fields) != 2 {
		t.Errorf("expected 2 distinct fields, got %d", len(sums[0].Fields))
	}
}

func TestSummaries_SortedByCountDescending(t *testing.T) {
	tr := newTracker(fixedNow(epoch))
	tr.Record("svc-b", "spec.replicas")
	tr.Record("svc-a", "spec.replicas")
	tr.Record("svc-a", "spec.image")
	tr.Record("svc-a", "spec.tag")

	sums := tr.Summaries()
	if len(sums) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(sums))
	}
	if sums[0].Service != "svc-a" {
		t.Errorf("expected svc-a first, got %s", sums[0].Service)
	}
	if sums[1].Service != "svc-b" {
		t.Errorf("expected svc-b second, got %s", sums[1].Service)
	}
}

func TestSummaries_ExcludesExpiredEntries(t *testing.T) {
	tr := New(10 * time.Minute)
	old := epoch
	recent := epoch.Add(15 * time.Minute)

	tr.now = fixedNow(old)
	tr.Record("svc-a", "spec.replicas") // will be outside window

	tr.now = fixedNow(recent)
	tr.Record("svc-b", "spec.image") // within window

	sums := tr.Summaries()
	if len(sums) != 1 {
		t.Fatalf("expected 1 summary after expiry, got %d", len(sums))
	}
	if sums[0].Service != "svc-b" {
		t.Errorf("expected svc-b, got %s", sums[0].Service)
	}
}

func TestPurge_RemovesOldEntries(t *testing.T) {
	tr := New(10 * time.Minute)
	tr.now = fixedNow(epoch)
	tr.Record("svc-a", "spec.replicas")

	tr.now = fixedNow(epoch.Add(20 * time.Minute))
	tr.Purge()

	if len(tr.entries) != 0 {
		t.Errorf("expected entries purged, got %d", len(tr.entries))
	}
}

func TestRecord_DeduplicatesFieldsPerService(t *testing.T) {
	tr := newTracker(fixedNow(epoch))
	tr.Record("svc-a", "spec.replicas")
	tr.Record("svc-a", "spec.replicas") // duplicate field

	sums := tr.Summaries()
	if len(sums[0].Fields) != 1 {
		t.Errorf("expected 1 unique field, got %d", len(sums[0].Fields))
	}
	if sums[0].Count != 2 {
		t.Errorf("expected count 2 (observations, not unique fields), got %d", sums[0].Count)
	}
}
