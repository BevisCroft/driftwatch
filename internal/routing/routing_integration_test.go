package routing_test

import (
	"sync"
	"testing"

	"github.com/driftwatch/driftwatch/internal/routing"
)

func TestNext_ConcurrentAccess(t *testing.T) {
	r, err := routing.New([]routing.Endpoint{
		{Name: "alpha", URL: "http://alpha", Weight: 2},
		{Name: "beta", URL: "http://beta", Weight: 3},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var wg sync.WaitGroup
	counts := make(map[string]int)
	var mu sync.Mutex

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ep := r.Next()
			mu.Lock()
			counts[ep.Name]++
			mu.Unlock()
		}()
	}
	wg.Wait()

	total := counts["alpha"] + counts["beta"]
	if total != 50 {
		t.Errorf("expected 50 total selections, got %d", total)
	}
}

func TestNext_ResetUnderConcurrentLoad(t *testing.T) {
	r, _ := routing.New([]routing.Endpoint{
		{Name: "a", URL: "http://a", Weight: 1},
		{Name: "b", URL: "http://b", Weight: 1},
	})

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if idx%5 == 0 {
				r.Reset()
			} else {
				r.Next()
			}
		}(i)
	}
	wg.Wait()
	// No panic or race condition is the success criterion.
}
