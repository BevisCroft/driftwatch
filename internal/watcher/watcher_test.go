package watcher_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/reporter"
	"github.com/example/driftwatch/internal/watcher"
)

func writeTempManifest(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeTempManifest: %v", err)
	}
}

func TestWatcher_RunCancels(t *testing.T) {
	dir := t.TempDir()
	writeTempManifest(t, dir, "svc.yaml", `kind: Service
metadata:
  name: svc
spec:
  port: 8080
`)

	var buf strings.Builder
	r := reporter.New(&buf, "text")
	w := watcher.New(dir, 50*time.Millisecond, r)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	err := w.Run(ctx)
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

func TestWatcher_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	var buf strings.Builder
	r := reporter.New(&buf, "text")
	w := watcher.New(dir, 30*time.Millisecond, r)

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	err := w.Run(ctx)
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
	// No output expected for empty directory.
	if buf.Len() != 0 {
		t.Errorf("expected empty output, got: %s", buf.String())
	}
}
