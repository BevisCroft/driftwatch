package audit_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/example/driftwatch/internal/audit"
)

func tempLog(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "audit.log")
}

func TestRecord_WritesEntry(t *testing.T) {
	path := tempLog(t)
	l, err := audit.New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer l.Close()

	if err := l.Record("svc-a", "drift_detected", "replicas changed"); err != nil {
		t.Fatalf("Record: %v", err)
	}

	entries, err := audit.ReadAll(path)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.Service != "svc-a" || e.Event != "drift_detected" || e.Details != "replicas changed" {
		t.Errorf("unexpected entry: %+v", e)
	}
}

func TestRecord_MultipleEntries(t *testing.T) {
	path := tempLog(t)
	l, err := audit.New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer l.Close()

	for i := 0; i < 5; i++ {
		if err := l.Record("svc", "event", ""); err != nil {
			t.Fatalf("Record %d: %v", i, err)
		}
	}

	entries, err := audit.ReadAll(path)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(entries) != 5 {
		t.Errorf("expected 5 entries, got %d", len(entries))
	}
}

func TestReadAll_FileNotFound(t *testing.T) {
	entries, err := audit.ReadAll("/nonexistent/audit.log")
	if err != nil {
		t.Fatalf("expected nil error for missing file, got %v", err)
	}
	if entries != nil {
		t.Errorf("expected nil entries for missing file")
	}
}

func TestReadAll_EmptyFile(t *testing.T) {
	path := tempLog(t)
	if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	entries, err := audit.ReadAll(path)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}
