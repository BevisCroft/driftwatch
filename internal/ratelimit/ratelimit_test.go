package ratelimit_test

import (
	"testing"
	"time"

	"github.com/driftwatch/internal/ratelimit"
)

func TestAllow_FirstEventAlwaysAllowed(t *testing.T) {
	l := ratelimit.New(100*time.Millisecond, 3)
	if !l.Allow("svc-a") {
		t.Fatal("expected first event to be allowed")
	}
}

func TestAllow_BurstExhausted(t *testing.T) {
	l := ratelimit.New(10*time.Second, 2)

	if !l.Allow("svc-b") {
		t.Fatal("expected 1st event allowed")
	}
	if !l.Allow("svc-b") {
		t.Fatal("expected 2nd event allowed")
	}
	if l.Allow("svc-b") {
		t.Fatal("expected 3rd event to be denied (burst exhausted)")
	}
}

func TestAllow_TokensRefillOverTime(t *testing.T) {
	rate := 50 * time.Millisecond
	l := ratelimit.New(rate, 1)

	if !l.Allow("svc-c") {
		t.Fatal("expected first event allowed")
	}
	if l.Allow("svc-c") {
		t.Fatal("expected second event denied before refill")
	}

	time.Sleep(rate + 10*time.Millisecond)

	if !l.Allow("svc-c") {
		t.Fatal("expected event allowed after token refill")
	}
}

func TestAllow_IndependentServicesDoNotInterfere(t *testing.T) {
	l := ratelimit.New(10*time.Second, 1)

	if !l.Allow("svc-x") {
		t.Fatal("expected svc-x allowed")
	}
	if l.Allow("svc-x") {
		t.Fatal("expected svc-x denied after burst")
	}
	// svc-y should be unaffected by svc-x's bucket.
	if !l.Allow("svc-y") {
		t.Fatal("expected svc-y allowed independently")
	}
}

func TestReset_RestoresCapacity(t *testing.T) {
	l := ratelimit.New(10*time.Second, 1)

	l.Allow("svc-d") // consume the only token
	if l.Allow("svc-d") {
		t.Fatal("expected denied before reset")
	}

	l.Reset("svc-d")

	if !l.Allow("svc-d") {
		t.Fatal("expected allowed after reset")
	}
}

func TestNew_MinBurstOfOne(t *testing.T) {
	// Passing maxBurst=0 should be clamped to 1.
	l := ratelimit.New(time.Second, 0)
	if !l.Allow("svc-e") {
		t.Fatal("expected at least one event allowed with burst=0 clamp")
	}
}
