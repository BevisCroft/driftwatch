package debounce

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestAllow_FirstEventAlwaysAllowed(t *testing.T) {
	d := New(5 * time.Second)
	if !d.Allow("svc-a") {
		t.Fatal("expected first event to be allowed")
	}
}

func TestAllow_SecondEventWithinWindowSuppressed(t *testing.T) {
	base := time.Now()
	d := New(10 * time.Second)
	d.nowFn = fixedNow(base)
	d.Allow("svc-a")

	d.nowFn = fixedNow(base.Add(5 * time.Second))
	if d.Allow("svc-a") {
		t.Fatal("expected event within window to be suppressed")
	}
}

func TestAllow_EventAfterWindowIsAllowed(t *testing.T) {
	base := time.Now()
	d := New(10 * time.Second)
	d.nowFn = fixedNow(base)
	d.Allow("svc-a")

	d.nowFn = fixedNow(base.Add(11 * time.Second))
	if !d.Allow("svc-a") {
		t.Fatal("expected event after window to be allowed")
	}
}

func TestAllow_IndependentServicesDoNotInterfere(t *testing.T) {
	base := time.Now()
	d := New(10 * time.Second)
	d.nowFn = fixedNow(base)
	d.Allow("svc-a")

	d.nowFn = fixedNow(base.Add(2 * time.Second))
	if !d.Allow("svc-b") {
		t.Fatal("expected svc-b to be allowed independently of svc-a")
	}
}

func TestReset_AllowsImmediateRetrigger(t *testing.T) {
	base := time.Now()
	d := New(30 * time.Second)
	d.nowFn = fixedNow(base)
	d.Allow("svc-a")

	d.Reset("svc-a")
	d.nowFn = fixedNow(base.Add(1 * time.Second))
	if !d.Allow("svc-a") {
		t.Fatal("expected event to be allowed after reset")
	}
}

func TestPurge_RemovesStaleEntries(t *testing.T) {
	base := time.Now()
	d := New(5 * time.Second)
	d.nowFn = fixedNow(base)
	d.Allow("svc-a")
	d.Allow("svc-b")

	d.nowFn = fixedNow(base.Add(6 * time.Second))
	d.Purge()

	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.last) != 0 {
		t.Fatalf("expected all entries purged, got %d", len(d.last))
	}
}

func TestPurge_RetainsActiveEntries(t *testing.T) {
	base := time.Now()
	d := New(10 * time.Second)
	d.nowFn = fixedNow(base)
	d.Allow("svc-a")

	d.nowFn = fixedNow(base.Add(3 * time.Second))
	d.Allow("svc-b")

	d.nowFn = fixedNow(base.Add(11 * time.Second))
	d.Purge()

	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.last["svc-b"]; !ok {
		t.Fatal("expected svc-b to be retained after purge")
	}
	if _, ok := d.last["svc-a"]; ok {
		t.Fatal("expected svc-a to be purged")
	}
}
