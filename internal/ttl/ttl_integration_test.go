package ttl_test

import (
	"sync"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/ttl"
)

func TestNew_SweepEvictsExpiredEntries(t *testing.T) {
	c := ttl.New(50*time.Millisecond, 20*time.Millisecond)
	defer c.Stop()

	c.Set("ephemeral", "value")

	_, ok := c.Get("ephemeral")
	if !ok {
		t.Fatal("expected entry immediately after Set")
	}

	time.Sleep(120 * time.Millisecond)

	_, ok = c.Get("ephemeral")
	if ok {
		t.Fatal("expected entry to be expired after TTL")
	}
}

func TestCache_ConcurrentSetAndGet(t *testing.T) {
	c := ttl.New(500*time.Millisecond, 100*time.Millisecond)
	defer c.Stop()

	const workers = 20
	const ops = 50
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			key := "service-" + string(rune('A'+id%26))
			for j := 0; j < ops; j++ {
				c.Set(key, j)
				c.Get(key)
			}
		}(i)
	}
	wg.Wait()
}

func TestCache_StopHaltsSweep(t *testing.T) {
	c := ttl.New(1*time.Second, 10*time.Millisecond)
	c.Set("k", "v")
	c.Stop()
	// After Stop the cache is still readable; no panic should occur.
	v, ok := c.Get("k")
	if !ok || v.(string) != "v" {
		t.Fatal("expected entry to be readable after Stop")
	}
}
