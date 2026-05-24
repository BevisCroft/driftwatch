package checkpoint_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/checkpoint"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "checkpoint.json")
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	store := checkpoint.New(tempPath(t))

	want := checkpoint.Entry{
		CycleID:    "abc-123",
		Timestamp:  time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
		Manifests:  10,
		DriftCount: 3,
	}

	if err := store.Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got == nil {
		t.Fatal("expected entry, got nil")
	}
	if got.CycleID != want.CycleID {
		t.Errorf("CycleID: got %q, want %q", got.CycleID, want.CycleID)
	}
	if got.DriftCount != want.DriftCount {
		t.Errorf("DriftCount: got %d, want %d", got.DriftCount, want.DriftCount)
	}
}

func TestLoad_NotFound_ReturnsNil(t *testing.T) {
	store := checkpoint.New(tempPath(t))

	entry, err := store.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry != nil {
		t.Errorf("expected nil, got %+v", entry)
	}
}

func TestSave_OverwritesPrevious(t *testing.T) {
	store := checkpoint.New(tempPath(t))

	first := checkpoint.Entry{CycleID: "first", DriftCount: 1}
	second := checkpoint.Entry{CycleID: "second", DriftCount: 5}

	_ = store.Save(first)
	_ = store.Save(second)

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.CycleID != "second" {
		t.Errorf("expected second, got %q", got.CycleID)
	}
}

func TestDelete_RemovesFile(t *testing.T) {
	p := tempPath(t)
	store := checkpoint.New(p)

	_ = store.Save(checkpoint.Entry{CycleID: "x"})
	if err := store.Delete(); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Error("expected file to be deleted")
	}
}

func TestDelete_NoFile_IsNoop(t *testing.T) {
	store := checkpoint.New(tempPath(t))
	if err := store.Delete(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}
