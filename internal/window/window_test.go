package window

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestAdd_FirstEventReturnsOne(t *testing.T) {
	c := New(time.Minute)
	if got := c.Add("svc-a"); got != 1 {
		t.Fatalf("expected 1, got %d", got)
	}
}

func TestAdd_AccumulatesWithinWindow(t *testing.T) {
	base := time.Now()
	c := New(time.Minute)
	c.now = fixedNow(base)

	c.Add("svc-a")
	c.Add("svc-a")
	if got := c.Add("svc-a"); got != 3 {
		t.Fatalf("expected 3, got %d", got)
	}
}

func TestAdd_EvictsExpiredEvents(t *testing.T) {
	base := time.Now()
	c := New(30 * time.Second)
	c.now = fixedNow(base)

	c.Add("svc-a")
	c.Add("svc-a")

	// Advance past the window.
	c.now = fixedNow(base.Add(31 * time.Second))
	if got := c.Add("svc-a"); got != 1 {
		t.Fatalf("expected 1 after eviction, got %d", got)
	}
}

func TestCount_DoesNotAddEvent(t *testing.T) {
	base := time.Now()
	c := New(time.Minute)
	c.now = fixedNow(base)

	c.Add("svc-a")
	count1 := c.Count("svc-a")
	count2 := c.Count("svc-a")
	if count1 != 1 || count2 != 1 {
		t.Fatalf("Count should not grow: got %d, %d", count1, count2)
	}
}

func TestCount_UnknownServiceReturnsZero(t *testing.T) {
	c := New(time.Minute)
	if got := c.Count("unknown"); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestReset_ClearsService(t *testing.T) {
	c := New(time.Minute)
	c.Add("svc-a")
	c.Add("svc-a")
	c.Reset("svc-a")
	if got := c.Count("svc-a"); got != 0 {
		t.Fatalf("expected 0 after reset, got %d", got)
	}
}

func TestServices_ReturnsTrackedKeys(t *testing.T) {
	c := New(time.Minute)
	c.Add("alpha")
	c.Add("beta")

	svcs := c.Services()
	if len(svcs) != 2 {
		t.Fatalf("expected 2 services, got %d", len(svcs))
	}
}

func TestAdd_IndependentServices(t *testing.T) {
	c := New(time.Minute)
	c.Add("svc-a")
	c.Add("svc-a")
	c.Add("svc-b")

	if got := c.Count("svc-a"); got != 2 {
		t.Fatalf("svc-a: expected 2, got %d", got)
	}
	if got := c.Count("svc-b"); got != 1 {
		t.Fatalf("svc-b: expected 1, got %d", got)
	}
}
