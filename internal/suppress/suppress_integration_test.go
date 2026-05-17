package suppress_test

import (
	"sync"
	"testing"
	"time"

	"driftwatch/internal/suppress"
)

func TestConcurrentAddAndCheck(t *testing.T) {
	l := suppress.New()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			l.Add(suppress.Entry{
				Service: "svc",
				Field:   "spec.replicas",
				Reason:  "concurrent",
			})
		}(i)
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = l.IsSuppressed("svc", "spec.replicas")
		}()
	}

	wg.Wait()
	snap := l.Snapshot()
	if len(snap) == 0 {
		t.Fatal("expected at least one entry after concurrent adds")
	}
}

func TestPurge_RunsConcurrentlyWithAdd(t *testing.T) {
	l := suppress.New()
	past := time.Now().Add(-1 * time.Hour)
	for i := 0; i < 20; i++ {
		l.Add(suppress.Entry{Service: "svc", Field: "f", ExpiresAt: past})
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		l.Purge()
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		l.Add(suppress.Entry{Service: "svc2", Field: "f2"})
	}()
	wg.Wait()

	// svc2/f2 has no expiry so must survive
	if !l.IsSuppressed("svc2", "f2") {
		t.Fatal("newly added entry should still be present after concurrent purge")
	}
}
