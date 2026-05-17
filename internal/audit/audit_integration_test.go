package audit_test

import (
	"sync"
	"testing"

	"github.com/example/driftwatch/internal/audit"
)

func TestRecord_ConcurrentWrites(t *testing.T) {
	path := tempLog(t)
	l, err := audit.New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer l.Close()

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()
			if err := l.Record("svc", "concurrent_event", ""); err != nil {
				t.Errorf("goroutine %d Record: %v", i, err)
			}
		}(i)
	}
	wg.Wait()

	entries, err := audit.ReadAll(path)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(entries) != goroutines {
		t.Errorf("expected %d entries, got %d", goroutines, len(entries))
	}
}

func TestRecord_PersistsAcrossReopen(t *testing.T) {
	path := tempLog(t)

	l1, err := audit.New(path)
	if err != nil {
		t.Fatalf("New first: %v", err)
	}
	_ = l1.Record("svc", "first_open", "")
	l1.Close()

	l2, err := audit.New(path)
	if err != nil {
		t.Fatalf("New second: %v", err)
	}
	_ = l2.Record("svc", "second_open", "")
	l2.Close()

	entries, err := audit.ReadAll(path)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries after reopen, got %d", len(entries))
	}
	if entries[0].Event != "first_open" || entries[1].Event != "second_open" {
		t.Errorf("unexpected events: %v, %v", entries[0].Event, entries[1].Event)
	}
}
