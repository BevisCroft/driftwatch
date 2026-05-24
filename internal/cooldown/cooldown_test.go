package cooldown

import (
	"testing"
	"time"
)

var fixedNow = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

func newTestTracker(d time.Duration) *Tracker {
	t := New(d)
	t.now = func() time.Time { return fixedNow }
	return t
}

func TestAllow_FirstEventAlwaysAllowed(t *testing.T) {
	tr := newTestTracker(5 * time.Minute)
	if !tr.Allow("svc-a") {
		t.Fatal("expected first event to be allowed")
	}
}

func TestAllow_SecondEventWithinWindowSuppressed(t *testing.T) {
	tr := newTestTracker(5 * time.Minute)
	tr.Allow("svc-a")
	if tr.Allow("svc-a") {
		t.Fatal("expected second event within window to be suppressed")
	}
}

func TestAllow_EventAfterWindowIsAllowed(t *testing.T) {
	tr := newTestTracker(5 * time.Minute)
	tr.Allow("svc-a")

	// Advance clock past the cooldown window.
	tr.now = func() time.Time { return fixedNow.Add(6 * time.Minute) }
	if !tr.Allow("svc-a") {
		t.Fatal("expected event after window expiry to be allowed")
	}
}

func TestAllow_IndependentServicesDoNotInterfere(t *testing.T) {
	tr := newTestTracker(5 * time.Minute)
	tr.Allow("svc-a")
	if !tr.Allow("svc-b") {
		t.Fatal("expected independent service to be allowed")
	}
}

func TestReset_AllowsImmediateRetrigger(t *testing.T) {
	tr := newTestTracker(5 * time.Minute)
	tr.Allow("svc-a")
	tr.Reset("svc-a")
	if !tr.Allow("svc-a") {
		t.Fatal("expected event to be allowed after reset")
	}
}

func TestPurge_RemovesExpiredEntries(t *testing.T) {
	tr := newTestTracker(5 * time.Minute)
	tr.Allow("svc-a")
	tr.Allow("svc-b")

	if tr.Active() != 2 {
		t.Fatalf("expected 2 active cooldowns, got %d", tr.Active())
	}

	// Advance clock so both entries expire.
	tr.now = func() time.Time { return fixedNow.Add(10 * time.Minute) }
	tr.Purge()

	if tr.Active() != 0 {
		t.Fatalf("expected 0 active cooldowns after purge, got %d", tr.Active())
	}
}

func TestActive_CountsOnlyNonExpired(t *testing.T) {
	tr := newTestTracker(5 * time.Minute)
	tr.Allow("svc-a")
	tr.Allow("svc-b")

	// Expire svc-a only.
	tr.now = func() time.Time { return fixedNow.Add(6 * time.Minute) }
	tr.Allow("svc-c") // starts a fresh window at the advanced time

	if got := tr.Active(); got != 1 {
		t.Fatalf("expected 1 active cooldown, got %d", got)
	}
}
