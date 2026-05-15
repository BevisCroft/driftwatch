package snapshot_test

import (
	"errors"
	"os"
	"testing"

	"github.com/yourorg/driftwatch/internal/snapshot"
)

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	store, err := snapshot.NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	original := snapshot.Snapshot{
		Name: "web-deployment",
		Kind: "Deployment",
		Fields: map[string]interface{}{
			"replicas": float64(3),
			"image":    "nginx:1.25",
		},
	}

	if err := store.Save(original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := store.Load(original.Name)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.Name != original.Name {
		t.Errorf("Name: got %q, want %q", loaded.Name, original.Name)
	}
	if loaded.Kind != original.Kind {
		t.Errorf("Kind: got %q, want %q", loaded.Kind, original.Kind)
	}
	if loaded.CapturedAt.IsZero() {
		t.Error("CapturedAt should not be zero after Save")
	}
	if v, ok := loaded.Fields["replicas"]; !ok || v != float64(3) {
		t.Errorf("Fields[replicas]: got %v", v)
	}
}

func TestLoad_NotFound(t *testing.T) {
	dir := t.TempDir()
	store, err := snapshot.NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	_, err = store.Load("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing snapshot, got nil")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected os.ErrNotExist, got: %v", err)
	}
}

func TestSave_OverwritesPreviousSnapshot(t *testing.T) {
	dir := t.TempDir()
	store, _ := snapshot.NewStore(dir)

	first := snapshot.Snapshot{
		Name:   "svc",
		Kind:   "Service",
		Fields: map[string]interface{}{"port": float64(80)},
	}
	second := snapshot.Snapshot{
		Name:   "svc",
		Kind:   "Service",
		Fields: map[string]interface{}{"port": float64(443)},
	}

	if err := store.Save(first); err != nil {
		t.Fatalf("Save first: %v", err)
	}
	if err := store.Save(second); err != nil {
		t.Fatalf("Save second: %v", err)
	}

	loaded, err := store.Load("svc")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Fields["port"] != float64(443) {
		t.Errorf("expected port 443, got %v", loaded.Fields["port"])
	}
}
