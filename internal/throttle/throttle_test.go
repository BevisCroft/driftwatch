package throttle

import (
	"testing"
	"time"
)

var epoch = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestAllow_FirstEventAlwaysAllowed(t *testing.T) {
	th := New(time.Minute, 3, fixedNow(epoch))
	if !th.Allow("svc-a") {
		t.Fatal("expected first event to be allowed")
	}
}

func TestAllow_BurstExhausted(t *testing.T) {
	th := New(time.Minute, 2, fixedNow(epoch))
	if !th.Allow("svc-a") {
		t.Fatal("event 1 should be allowed")
	}
	if !th.Allow("svc-a") {
		t.Fatal("event 2 should be allowed")
	}
	if th.Allow("svc-a") {
		t.Fatal("event 3 should be throttled")
	}
}

func TestAllow_WindowReset(t *testing.T) {
	current := epoch
	nowFn := func() time.Time { return current }
	th := New(time.Minute, 1, nowFn)

	th.Allow("svc-a") // consume burst
	if th.Allow("svc-a") {
		t.Fatal("should be throttled within window")
	}

	current = epoch.Add(2 * time.Minute) // advance past window
	if !th.Allow("svc-a") {
		t.Fatal("should be allowed after window expires")
	}
}

func TestAllow_IndependentServices(t *testing.T) {
	th := New(time.Minute, 1, fixedNow(epoch))
	th.Allow("svc-a") // exhaust svc-a

	if !th.Allow("svc-b") {
		t.Fatal("svc-b should be independent of svc-a")
	}
}

func TestReset_RestoresCapacity(t *testing.T) {
	th := New(time.Minute, 1, fixedNow(epoch))
	th.Allow("svc-a") // exhaust
	th.Reset("svc-a")
	if !th.Allow("svc-a") {
		t.Fatal("should be allowed after reset")
	}
}

func TestPurge_RemovesExpiredRecords(t *testing.T) {
	current := epoch
	nowFn := func() time.Time { return current }
	th := New(time.Minute, 3, nowFn)

	th.Allow("svc-a")
	th.Allow("svc-b")

	current = epoch.Add(2 * time.Minute)
	th.Purge()

	th.mu.Lock()
	defer th.mu.Unlock()
	if len(th.records) != 0 {
		t.Fatalf("expected 0 records after purge, got %d", len(th.records))
	}
}
