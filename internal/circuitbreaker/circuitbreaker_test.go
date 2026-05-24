package circuitbreaker_test

import (
	"testing"
	"time"

	"github.com/driftwatch/internal/circuitbreaker"
)

func newBreaker(threshold int, cooldown time.Duration) *circuitbreaker.Breaker {
	return circuitbreaker.New(threshold, cooldown)
}

func TestAllow_ClosedByDefault(t *testing.T) {
	b := newBreaker(3, time.Second)
	if err := b.Allow("svc-a"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRecordFailure_OpensAfterThreshold(t *testing.T) {
	b := newBreaker(2, time.Minute)
	b.RecordFailure("svc-a")
	if b.StateOf("svc-a") != circuitbreaker.StateClosed {
		t.Fatal("expected closed after one failure")
	}
	b.RecordFailure("svc-a")
	if b.StateOf("svc-a") != circuitbreaker.StateOpen {
		t.Fatal("expected open after threshold")
	}
}

func TestAllow_BlocksWhenOpen(t *testing.T) {
	b := newBreaker(1, time.Minute)
	b.RecordFailure("svc-a")
	if err := b.Allow("svc-a"); err != circuitbreaker.ErrOpen {
		t.Fatalf("expected ErrOpen, got %v", err)
	}
}

func TestAllow_HalfOpenAfterCooldown(t *testing.T) {
	now := time.Now()
	b := circuitbreaker.New(1, 50*time.Millisecond)
	// Manually advance time via real sleep for half-open transition.
	b.RecordFailure("svc-a")
	time.Sleep(60 * time.Millisecond)
	if err := b.Allow("svc-a"); err != nil {
		t.Fatalf("expected nil in half-open, got %v (elapsed: %v)", err, time.Since(now))
	}
	if b.StateOf("svc-a") != circuitbreaker.StateHalfOpen {
		t.Fatalf("expected half-open state")
	}
}

func TestRecordSuccess_ClosesCirucit(t *testing.T) {
	b := newBreaker(1, time.Minute)
	b.RecordFailure("svc-a")
	b.RecordSuccess("svc-a")
	if b.StateOf("svc-a") != circuitbreaker.StateClosed {
		t.Fatal("expected closed after success")
	}
	if err := b.Allow("svc-a"); err != nil {
		t.Fatalf("expected nil after recovery, got %v", err)
	}
}

func TestIndependentServices_DoNotInterfere(t *testing.T) {
	b := newBreaker(1, time.Minute)
	b.RecordFailure("svc-a")
	if err := b.Allow("svc-b"); err != nil {
		t.Fatalf("svc-b should not be affected by svc-a failures: %v", err)
	}
}

func TestState_String(t *testing.T) {
	cases := []struct {
		s    circuitbreaker.State
		want string
	}{
		{circuitbreaker.StateClosed, "closed"},
		{circuitbreaker.StateOpen, "open"},
		{circuitbreaker.StateHalfOpen, "half-open"},
	}
	for _, tc := range cases {
		if got := tc.s.String(); got != tc.want {
			t.Errorf("State(%d).String() = %q, want %q", tc.s, got, tc.want)
		}
	}
}
