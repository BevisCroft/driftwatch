package baseline_test

import (
	"os"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/baseline"
)

func tempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "baseline-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestPin_AndGet_RoundTrip(t *testing.T) {
	store, err := baseline.New(tempDir(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	entry := baseline.Entry{
		Service:    "auth-service",
		PinnedAt:   time.Now().UTC().Truncate(time.Second),
		ApprovedBy: "alice",
		Fields:     map[string]interface{}{"replicas": float64(3), "image": "auth:v1.2"},
	}

	if err := store.Pin(entry); err != nil {
		t.Fatalf("Pin: %v", err)
	}

	got, err := store.Get("auth-service")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got == nil {
		t.Fatal("expected entry, got nil")
	}
	if got.Service != entry.Service {
		t.Errorf("service: got %q want %q", got.Service, entry.Service)
	}
	if got.ApprovedBy != entry.ApprovedBy {
		t.Errorf("approved_by: got %q want %q", got.ApprovedBy, entry.ApprovedBy)
	}
	if got.Fields["replicas"] != entry.Fields["replicas"] {
		t.Errorf("fields[replicas]: got %v want %v", got.Fields["replicas"], entry.Fields["replicas"])
	}
}

func TestGet_NotFound_ReturnsNil(t *testing.T) {
	store, _ := baseline.New(tempDir(t))

	got, err := store.Get("nonexistent")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestPin_EmptyService_ReturnsError(t *testing.T) {
	store, _ := baseline.New(tempDir(t))

	err := store.Pin(baseline.Entry{Service: ""})
	if err == nil {
		t.Fatal("expected error for empty service name")
	}
}

func TestDelete_RemovesEntry(t *testing.T) {
	store, _ := baseline.New(tempDir(t))

	entry := baseline.Entry{
		Service:  "svc",
		PinnedAt: time.Now().UTC(),
		Fields:   map[string]interface{}{},
	}
	_ = store.Pin(entry)
	_ = store.Delete("svc")

	got, err := store.Get("svc")
	if err != nil {
		t.Fatalf("Get after delete: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil after delete, got %+v", got)
	}
}

func TestDelete_NonExistent_NoError(t *testing.T) {
	store, _ := baseline.New(tempDir(t))
	if err := store.Delete("ghost"); err != nil {
		t.Errorf("expected no error deleting nonexistent entry, got: %v", err)
	}
}
