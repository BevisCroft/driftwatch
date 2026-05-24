package cache

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func newTestCache(ttl time.Duration, maxSize int) *Cache {
	c := New(ttl, maxSize)
	c.now = fixedNow(time.Unix(1_000_000, 0))
	return c
}

func TestSet_And_Get_HitWithinTTL(t *testing.T) {
	c := newTestCache(5*time.Minute, 0)
	c.Set("svc-a", "value-a")

	v, ok := c.Get("svc-a")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if v != "value-a" {
		t.Fatalf("got %v, want value-a", v)
	}
}

func TestGet_ExpiredEntry_ReturnsMiss(t *testing.T) {
	base := time.Unix(1_000_000, 0)
	c := New(1*time.Minute, 0)
	c.now = fixedNow(base)
	c.Set("svc-b", "value-b")

	// advance clock past TTL
	c.now = fixedNow(base.Add(2 * time.Minute))

	_, ok := c.Get("svc-b")
	if ok {
		t.Fatal("expected cache miss for expired entry")
	}
}

func TestGet_MissingKey_ReturnsMiss(t *testing.T) {
	c := newTestCache(time.Minute, 0)
	_, ok := c.Get("nonexistent")
	if ok {
		t.Fatal("expected miss for unknown key")
	}
}

func TestDelete_RemovesEntry(t *testing.T) {
	c := newTestCache(time.Minute, 0)
	c.Set("svc-c", 42)
	c.Delete("svc-c")

	_, ok := c.Get("svc-c")
	if ok {
		t.Fatal("entry should have been deleted")
	}
}

func TestFlush_ClearsAllEntries(t *testing.T) {
	c := newTestCache(time.Minute, 0)
	c.Set("a", 1)
	c.Set("b", 2)
	c.Flush()

	if c.Len() != 0 {
		t.Fatalf("expected 0 entries after flush, got %d", c.Len())
	}
}

func TestSet_MaxSize_EvictsOldest(t *testing.T) {
	base := time.Unix(1_000_000, 0)
	c := New(10*time.Minute, 2)

	// insert first entry at t=0
	c.now = fixedNow(base)
	c.Set("first", "v1")

	// insert second entry at t+1s (later expiry)
	c.now = fixedNow(base.Add(time.Second))
	c.Set("second", "v2")

	// third insert should evict "first" (earliest expiry)
	c.now = fixedNow(base.Add(2 * time.Second))
	c.Set("third", "v3")

	if c.Len() != 2 {
		t.Fatalf("expected 2 entries, got %d", c.Len())
	}
	_, firstOk := c.Get("first")
	if firstOk {
		t.Fatal("expected 'first' to have been evicted")
	}
	_, thirdOk := c.Get("third")
	if !thirdOk {
		t.Fatal("expected 'third' to be present")
	}
}

func TestLen_ReflectsCurrentCount(t *testing.T) {
	c := newTestCache(time.Minute, 0)
	if c.Len() != 0 {
		t.Fatalf("expected 0, got %d", c.Len())
	}
	c.Set("x", true)
	c.Set("y", true)
	if c.Len() != 2 {
		t.Fatalf("expected 2, got %d", c.Len())
	}
}
