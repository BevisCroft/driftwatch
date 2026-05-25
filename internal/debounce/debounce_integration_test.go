package debounce_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/driftwatch/driftwatch/internal/debounce"
)

func TestDebounce_ConcurrentServices(t *testing.T) {
	d := debounce.New(50 * time.Millisecond)

	var allowed atomic.Int64
	var wg sync.WaitGroup

	services := []string{"alpha", "beta", "gamma", "delta"}
	for _, svc := range services {
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(s string) {
				defer wg.Done()
				if d.Allow(s) {
					allowed.Add(1)
				}
			}(svc)
		}
	}
	wg.Wait()

	// Each service should have allowed exactly 1 event out of 10 concurrent calls.
	if got := allowed.Load(); got > int64(len(services)) {
		t.Fatalf("expected at most %d allowed events, got %d", len(services), got)
	}
}

func TestDebounce_WindowExpiry_RealTime(t *testing.T) {
	window := 80 * time.Millisecond
	d := debounce.New(window)

	if !d.Allow("svc-x") {
		t.Fatal("expected first allow")
	}
	if d.Allow("svc-x") {
		t.Fatal("expected suppression within window")
	}

	time.Sleep(window + 20*time.Millisecond)

	if !d.Allow("svc-x") {
		t.Fatal("expected allow after window expiry")
	}
}

// TestDebounce_IndependentKeys verifies that debounce windows are tracked
// independently per key, so suppressing one key does not affect another.
func TestDebounce_IndependentKeys(t *testing.T) {
	d := debounce.New(100 * time.Millisecond)

	if !d.Allow("key-a") {
		t.Fatal("expected first allow for key-a")
	}
	if !d.Allow("key-b") {
		t.Fatal("expected first allow for key-b to be independent of key-a")
	}
	if d.Allow("key-a") {
		t.Fatal("expected key-a to be suppressed within window")
	}
	if d.Allow("key-b") {
		t.Fatal("expected key-b to be suppressed within window")
	}
}
