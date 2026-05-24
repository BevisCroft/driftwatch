package sampling

import (
	"testing"
)

func TestAllow_RateZero_BlocksAll(t *testing.T) {
	s := New(Config{Rate: 0})
	for i := 0; i < 100; i++ {
		if s.Allow("svc") {
			t.Fatal("expected no events to pass with rate=0")
		}
	}
}

func TestAllow_RateOne_AllowsAll(t *testing.T) {
	s := New(Config{Rate: 1})
	for i := 0; i < 100; i++ {
		if !s.Allow("svc") {
			t.Fatal("expected all events to pass with rate=1")
		}
	}
}

func TestAllow_RateClamped(t *testing.T) {
	s := New(Config{Rate: 1.5})
	if s.Rate() != 1.0 {
		t.Fatalf("expected rate clamped to 1.0, got %f", s.Rate())
	}
	s2 := New(Config{Rate: -0.5})
	if s2.Rate() != 0.0 {
		t.Fatalf("expected rate clamped to 0.0, got %f", s2.Rate())
	}
}

func TestAllow_Deterministic_EveryOther(t *testing.T) {
	s := New(Config{Rate: 0.5, Strategy: StrategyDeterministic})

	// With rate=0.5, period=2, events 1,3,5... should pass (n%2==1).
	allowed := 0
	for i := 0; i < 10; i++ {
		if s.Allow("svc") {
			allowed++
		}
	}
	if allowed != 5 {
		t.Fatalf("expected 5 allowed events, got %d", allowed)
	}
}

func TestAllow_Deterministic_IndependentServices(t *testing.T) {
	s := New(Config{Rate: 0.5, Strategy: StrategyDeterministic})

	// Each service has its own counter.
	if !s.Allow("alpha") {
		t.Fatal("first event for alpha should be allowed")
	}
	if !s.Allow("beta") {
		t.Fatal("first event for beta should be allowed")
	}
}

func TestReset_ClearsCounter(t *testing.T) {
	s := New(Config{Rate: 0.5, Strategy: StrategyDeterministic})

	s.Allow("svc") // counter=1, allowed
	s.Allow("svc") // counter=2, blocked
	s.Reset("svc")

	// After reset counter restarts at 0, so next call is counter=1 → allowed.
	if !s.Allow("svc") {
		t.Fatal("expected first event after reset to be allowed")
	}
}

func TestAllow_DefaultStrategy_IsRandom(t *testing.T) {
	s := New(Config{Rate: 0.5})
	if s.cfg.Strategy != StrategyRandom {
		t.Fatalf("expected default strategy to be random, got %s", s.cfg.Strategy)
	}
}
