package checkpoint_test

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/checkpoint"
)

func TestConcurrentSaveAndLoad(t *testing.T) {
	store := checkpoint.New(filepath.Join(t.TempDir(), "cp.json"))

	var wg sync.WaitGroup
	const writers = 8

	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			e := checkpoint.Entry{
				CycleID:    fmt.Sprintf("cycle-%d", n),
				Timestamp:  time.Now(),
				Manifests:  n * 2,
				DriftCount: n,
			}
			if err := store.Save(e); err != nil {
				t.Errorf("Save(%d): %v", n, err)
			}
		}(i)
	}

	wg.Wait()

	entry, err := store.Load()
	if err != nil {
		t.Fatalf("Load after concurrent saves: %v", err)
	}
	if entry == nil {
		t.Fatal("expected a checkpoint entry, got nil")
	}
}

func TestPersistsAcrossReopen(t *testing.T) {
	p := filepath.Join(t.TempDir(), "cp.json")

	w := checkpoint.New(p)
	want := checkpoint.Entry{
		CycleID:    "persist-test",
		Timestamp:  time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		Manifests:  7,
		DriftCount: 2,
		Error:      "",
	}
	if err := w.Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Simulate daemon restart by creating a new Store instance.
	r := checkpoint.New(p)
	got, err := r.Load()
	if err != nil {
		t.Fatalf("Load (new instance): %v", err)
	}
	if got == nil {
		t.Fatal("expected entry after reopen, got nil")
	}
	if got.CycleID != want.CycleID || got.Manifests != want.Manifests {
		t.Errorf("mismatch: got %+v, want %+v", got, want)
	}
}
