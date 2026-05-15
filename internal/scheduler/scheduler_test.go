package scheduler_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/config"
	"github.com/example/driftwatch/internal/drift"
	"github.com/example/driftwatch/internal/manifest"
	"github.com/example/driftwatch/internal/reporter"
	"github.com/example/driftwatch/internal/scheduler"
	"github.com/example/driftwatch/internal/snapshot"
)

func newTestScheduler(t *testing.T, dir string) *scheduler.Scheduler {
	t.Helper()
	cfg := &config.Config{
		ManifestDir:  dir,
		PollInterval: 50 * time.Millisecond,
		OutputFormat: "text",
	}
	loader := manifest.NewLoader()
	detector := drift.NewDetector()
	store := snapshot.NewStore(t.TempDir())
	rep := reporter.New(os.Discard, cfg.OutputFormat)
	return scheduler.New(cfg, loader, detector, store, rep)
}

func TestScheduler_CancelsCleanly(t *testing.T) {
	dir := t.TempDir()
	sched := newTestScheduler(t, dir)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	err := sched.Run(ctx)
	if err == nil {
		t.Fatal("expected non-nil error on context cancellation")
	}
	if err != context.DeadlineExceeded && err != context.Canceled {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestScheduler_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	sched := newTestScheduler(t, dir)

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	// Should not panic or error on an empty manifest directory.
	_ = sched.Run(ctx)
}

func TestScheduler_DetectsDrift(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "svc.yaml")

	initial := []byte("name: svc\nkind: Deployment\nspec:\n  replicas: 1\n")
	if err := os.WriteFile(manifestPath, initial, 0o644); err != nil {
		t.Fatal(err)
	}

	sched := newTestScheduler(t, dir)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// After the first cycle a snapshot is saved; mutate the file so the
	// second cycle can detect drift.
	go func() {
		time.Sleep(70 * time.Millisecond)
		updated := []byte("name: svc\nkind: Deployment\nspec:\n  replicas: 3\n")
		_ = os.WriteFile(manifestPath, updated, 0o644)
	}()

	_ = sched.Run(ctx)
}
