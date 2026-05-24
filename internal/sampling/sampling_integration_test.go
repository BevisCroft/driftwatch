package sampling_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/example/driftwatch/internal/sampling"
)

func TestAllow_ConcurrentRandom(t *testing.T) {
	s := sampling.New(sampling.Config{Rate: 0.5})

	const goroutines = 20
	const eventsEach = 500

	var allowed atomic.Int64
	var wg sync.WaitGroup

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < eventsEach; i++ {
				if s.Allow("svc") {
					allowed.Add(1)
				}
			}
		}()
	}
	wg.Wait()

	total := int64(goroutines * eventsEach)
	got := allowed.Load()

	// Expect roughly 50% ± 10% with high probability.
	lo := int64(float64(total) * 0.40)
	hi := int64(float64(total) * 0.60)
	if got < lo || got > hi {
		t.Fatalf("concurrent random sampling: allowed %d/%d, expected [%d, %d]", got, total, lo, hi)
	}
}

func TestAllow_ConcurrentDeterministic(t *testing.T) {
	s := sampling.New(sampling.Config{Rate: 0.5, Strategy: sampling.StrategyDeterministic})

	const goroutines = 10
	var wg sync.WaitGroup

	// Each goroutine uses a distinct service name so counters don't interfere.
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			svc := string(rune('a' + id))
			allowed := 0
			for i := 0; i < 10; i++ {
				if s.Allow(svc) {
					allowed++
				}
			}
			if allowed != 5 {
				t.Errorf("service %s: expected 5 allowed, got %d", svc, allowed)
			}
		}(g)
	}
	wg.Wait()
}
