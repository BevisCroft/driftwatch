package tagindex_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/example/driftwatch/internal/tagindex"
)

// TestConcurrentAddAndLookup verifies that concurrent writes and reads
// do not cause data races or panics.
func TestConcurrentAddAndLookup(t *testing.T) {
	idx := tagindex.New()
	const workers = 20
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			svc := fmt.Sprintf("svc-%d", n)
			env := "prod"
			if n%2 == 0 {
				env = "staging"
			}
			_ = idx.Add(svc, map[string]string{"env": env, "id": fmt.Sprintf("%d", n)})
		}(i)
	}

	// Concurrent lookups while adds are in flight.
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = idx.Lookup(map[string]string{"env": "prod"})
		}()
	}
	wg.Wait()

	// After all goroutines finish, prod services should be exactly those with odd n.
	got := idx.Lookup(map[string]string{"env": "prod"})
	if len(got) != workers/2 {
		t.Errorf("expected %d prod services, got %d: %v", workers/2, len(got), got)
	}
}

// TestConcurrentRemove verifies that concurrent removes are safe.
func TestConcurrentRemove(t *testing.T) {
	idx := tagindex.New()
	const n = 30
	for i := 0; i < n; i++ {
		_ = idx.Add(fmt.Sprintf("svc-%d", i), map[string]string{"env": "prod"})
	}

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			idx.Remove(fmt.Sprintf("svc-%d", id))
		}(i)
	}
	wg.Wait()

	got := idx.Lookup(map[string]string{"env": "prod"})
	if len(got) != 0 {
		t.Errorf("expected empty index after all removes, got %v", got)
	}
}
