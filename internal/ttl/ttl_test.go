package ttl

import (
	"testing"
	"time"
)

// fixedNow returns a function that returns t unchanged, used to freeze time.
func fixedNow(t time.Time) func() time.Time { return func() time.Time { return t } }

func newTestCache(ttlDur time.Duration) *Cache {
	c := &Cache{
		items:  make(map[string]entry),
		ttl:    ttlDur,
		now:    time.Now,
		stopCh: make(chan struct{}),
	}
	return c
}

func TestSet_And_Get_WithinTTL(t *testing.T) {
	now := time.Now()
	c := newTestCache(5 * time.Minute)
	c.now = fixedNow(now)

	c.Set("k", "hello")
	v, ok := c.Get("k")
	if !ok {
		t.Fatal("expected entry to be found")
	}
	if v.(string) != "hello" {
		t.Fatalf("got %q, want %q", v, "hello")
	}
}

func TestGet_ExpiredEntry_NotFound(t *testing.T) {
	now := time.Now()
	c := newTestCache(1 * time.Second)
	c.now = fixedNow(now)
	c.Set("k", 42)

	// advance past TTL
	c.now = fixedNow(now.Add(2 * time.Second))
	_, ok := c.Get("k")
	if ok {
		t.Fatal("expected expired entry to be absent")
	}
}

func TestDelete_RemovesEntry(t *testing.T) {
	c := newTestCache(time.Minute)
	c.Set("x", true)
	c.Delete("x")
	_, ok := c.Get("x")
	if ok {
		t.Fatal("expected deleted entry to be absent")
	}
}

func TestLen_CountsOnlyLiveEntries(t *testing.T) {
	now := time.Now()
	c := newTestCache(10 * time.Second)
	c.now = fixedNow(now)
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	// expire "c" manually by backdating its entry
	c.mu.Lock()
	c.items["c"] = entry{value: 3, expiresAt: now.Add(-1 * time.Second)}
	c.mu.Unlock()

	if got := c.Len(); got != 2 {
		t.Fatalf("Len() = %d, want 2", got)
	}
}

func TestSet_OverwriteResetsExpiry(t *testing.T) {
	now := time.Now()
	c := newTestCache(5 * time.Second)
	c.now = fixedNow(now)
	c.Set("k", "first")

	// advance close to expiry then overwrite
	c.now = fixedNow(now.Add(4 * time.Second))
	c.Set("k", "second")

	// advance past original expiry but within new TTL
	c.now = fixedNow(now.Add(7 * time.Second))
	v, ok := c.Get("k")
	if !ok {
		t.Fatal("expected refreshed entry to still be alive")
	}
	if v.(string) != "second" {
		t.Fatalf("got %q, want %q", v, "second")
	}
}

func TestPurge_RemovesExpiredEntries(t *testing.T) {
	now := time.Now()
	c := newTestCache(1 * time.Second)
	c.now = fixedNow(now)
	c.Set("alive", true)
	c.Set("dead", false)

	c.mu.Lock()
	c.items["dead"] = entry{value: false, expiresAt: now.Add(-time.Second)}
	c.mu.Unlock()

	c.now = fixedNow(now.Add(2 * time.Second))
	c.purge()

	c.mu.RLock()
	_, hasAlive := c.items["alive"]
	_, hasDead := c.items["dead"]
	c.mu.RUnlock()

	if !hasAlive {
		t.Error("expected 'alive' entry to remain after purge")
	}
	if hasDead {
		t.Error("expected 'dead' entry to be removed by purge")
	}
}
