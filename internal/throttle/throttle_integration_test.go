package throttle_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/throttle"
)

func TestThrottle_ConcurrentServices(t *testing.T) {
	th := throttle.New(time.Second, 5, nil)

	var allowed atomic.Int64
	var wg sync.WaitGroup

	services := []string{"alpha", "beta", "gamma"}
	for _, svc := range services {
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(s string) {
				defer wg.Done()
				if th.Allow(s) {
					allowed.Add(1)
				}
			}(svc)
		}
	}
	wg.Wait()

	// Each service allows at most 5; 3 services × 5 = 15 max.
	if got := allowed.Load(); got > 15 {
		t.Fatalf("expected at most 15 allowed events, got %d", got)
	}
}

func TestThrottle_RealTimeWindowExpiry(t *testing.T) {
	th := throttle.New(100*time.Millisecond, 1, nil)

	if !th.Allow("svc") {
		t.Fatal("first event should be allowed")
	}
	if th.Allow("svc") {
		t.Fatal("second event within window should be throttled")
	}

	time.Sleep(150 * time.Millisecond)

	if !th.Allow("svc") {
		t.Fatal("event after window expiry should be allowed")
	}
}
