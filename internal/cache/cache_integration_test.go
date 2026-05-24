package cache_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/cache"
)

func TestCache_ConcurrentSetAndGet(t *testing.T) {
	c := cache.New(5*time.Second, 0)
	const workers = 20
	const ops = 50

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < ops; j++ {
				key := fmt.Sprintf("svc-%d-%d", id, j)
				c.Set(key, j)
				c.Get(key)
			}
		}(i)
	}
	wg.Wait()
	// no race detector errors is the primary assertion
}

func TestCache_TTLExpiry_RealTime(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping real-time TTL test in short mode")
	}

	c := cache.New(50*time.Millisecond, 0)
	c.Set("live", "data")

	_, ok := c.Get("live")
	if !ok {
		t.Fatal("expected hit immediately after set")
	}

	time.Sleep(100 * time.Millisecond)

	_, ok = c.Get("live")
	if ok {
		t.Fatal("expected miss after TTL expiry")
	}
}

func TestCache_MaxSize_ConcurrentEviction(t *testing.T) {
	c := cache.New(10*time.Second, 10)
	const goroutines = 8
	const insertsEach = 30

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < insertsEach; j++ {
				c.Set(fmt.Sprintf("%d:%d", id, j), id*j)
			}
		}(i)
	}
	wg.Wait()

	if got := c.Len(); got > 10 {
		t.Fatalf("cache exceeded maxSize: len=%d", got)
	}
}
