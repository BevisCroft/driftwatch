package watchlist_test

import (
	"sync"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/watchlist"
)

func TestAdd_And_Get_RoundTrip(t *testing.T) {
	wl := watchlist.New()
	e := watchlist.Entry{Service: "api", Namespace: "prod", Labels: map[string]string{"tier": "backend"}}
	if err := wl.Add(e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := wl.Get("api")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if got.Namespace != "prod" {
		t.Errorf("namespace: want prod, got %s", got.Namespace)
	}
	if got.AddedAt.IsZero() {
		t.Error("AddedAt should be set automatically")
	}
}

func TestAdd_EmptyService_ReturnsError(t *testing.T) {
	wl := watchlist.New()
	err := wl.Add(watchlist.Entry{Service: ""})
	if err == nil {
		t.Fatal("expected error for empty service")
	}
}

func TestAdd_Duplicate_ReturnsError(t *testing.T) {
	wl := watchlist.New()
	e := watchlist.Entry{Service: "svc"}
	_ = wl.Add(e)
	if err := wl.Add(e); err == nil {
		t.Fatal("expected error for duplicate service")
	}
}

func TestRemove_ExistingEntry(t *testing.T) {
	wl := watchlist.New()
	_ = wl.Add(watchlist.Entry{Service: "svc"})
	if !wl.Remove("svc") {
		t.Fatal("expected Remove to return true")
	}
	if wl.Contains("svc") {
		t.Error("service should no longer be present")
	}
}

func TestRemove_NotFound_ReturnsFalse(t *testing.T) {
	wl := watchlist.New()
	if wl.Remove("ghost") {
		t.Fatal("expected false for unknown service")
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	wl := watchlist.New()
	_ = wl.Add(watchlist.Entry{Service: "a"})
	_ = wl.Add(watchlist.Entry{Service: "b"})
	if len(wl.All()) != 2 {
		t.Errorf("expected 2 entries, got %d", len(wl.All()))
	}
}

func TestContains_KnownAndUnknown(t *testing.T) {
	wl := watchlist.New()
	_ = wl.Add(watchlist.Entry{Service: "known", AddedAt: time.Now()})
	if !wl.Contains("known") {
		t.Error("expected Contains to return true")
	}
	if wl.Contains("unknown") {
		t.Error("expected Contains to return false")
	}
}

func TestAdd_ConcurrentSafe(t *testing.T) {
	wl := watchlist.New()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = wl.Add(watchlist.Entry{Service: fmt.Sprintf("svc-%d", n)})
		}(i)
	}
	wg.Wait()
}
