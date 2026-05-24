package backoff_test

import (
	"testing"
	"time"

	"github.com/driftwatch/internal/backoff"
)

func defaultStrategy() backoff.Strategy {
	return backoff.Strategy{
		InitialInterval: 100 * time.Millisecond,
		MaxInterval:     10 * time.Second,
		Multiplier:      2.0,
		JitterFraction:  0.0, // deterministic for tests
	}
}

func TestNext_FirstAttemptReturnsInitialInterval(t *testing.T) {
	b := backoff.New(defaultStrategy())
	d := b.Next("svc-a")
	if d != 100*time.Millisecond {
		t.Fatalf("expected 100ms, got %v", d)
	}
}

func TestNext_ExponentialGrowth(t *testing.T) {
	b := backoff.New(defaultStrategy())
	expected := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		400 * time.Millisecond,
		800 * time.Millisecond,
	}
	for i, want := range expected {
		got := b.Next("svc-a")
		if got != want {
			t.Fatalf("attempt %d: expected %v, got %v", i, want, got)
		}
	}
}

func TestNext_CapsAtMaxInterval(t *testing.T) {
	b := backoff.New(defaultStrategy())
	for i := 0; i < 20; i++ {
		b.Next("svc-a")
	}
	got := b.Next("svc-a")
	if got > 10*time.Second {
		t.Fatalf("expected duration capped at 10s, got %v", got)
	}
}

func TestNext_IndependentKeys(t *testing.T) {
	b := backoff.New(defaultStrategy())
	b.Next("svc-a")
	b.Next("svc-a")

	got := b.Next("svc-b")
	if got != 100*time.Millisecond {
		t.Fatalf("svc-b should start at initial interval, got %v", got)
	}
}

func TestReset_ClearsAttempts(t *testing.T) {
	b := backoff.New(defaultStrategy())
	b.Next("svc-a")
	b.Next("svc-a")

	if b.Attempts("svc-a") != 2 {
		t.Fatalf("expected 2 attempts before reset")
	}

	b.Reset("svc-a")

	if b.Attempts("svc-a") != 0 {
		t.Fatalf("expected 0 attempts after reset")
	}

	got := b.Next("svc-a")
	if got != 100*time.Millisecond {
		t.Fatalf("expected initial interval after reset, got %v", got)
	}
}

func TestNext_WithJitter_DoesNotExceedMax(t *testing.T) {
	s := backoff.Strategy{
		InitialInterval: 100 * time.Millisecond,
		MaxInterval:     2 * time.Second,
		Multiplier:      2.0,
		JitterFraction:  1.0,
	}
	b := backoff.New(s)
	for i := 0; i < 50; i++ {
		d := b.Next("svc-a")
		if d > 4*time.Second {
			t.Fatalf("jittered duration exceeded safe bound: %v", d)
		}
	}
}
