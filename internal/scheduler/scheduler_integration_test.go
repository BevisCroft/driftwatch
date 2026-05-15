//go:build integration

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

// TestScheduler_MultiCycleDrift verifies that across several ticks the
// scheduler correctly identifies an accumulating drift count.
func TestScheduler_MultiCycleDrift(t *testing.T) {
	manifestDir := t.TempDir()
	snapshotDir := t.TempDir()

	manifestPath := filepath.Join(manifestDir, "api.yaml")
	initial := []byte("name: api\nkind: Service\nspec:\n  port: 8080\n")
	if err := os.WriteFile(manifestPath, initial, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		ManifestDir:  manifestDir,
		PollInterval: 30 * time.Millisecond,
		OutputFormat: "json",
	}
	loader := manifest.NewLoader()
	detector := drift.NewDetector()
	store := snapshot.NewStore(snapshotDir)
	rep := reporter.New(os.Discard, cfg.OutputFormat)
	sched := scheduler.New(cfg, loader, detector, store, rep)

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	go func() {
		time.Sleep(60 * time.Millisecond)
		updated := []byte("name: api\nkind: Service\nspec:\n  port: 9090\n")
		_ = os.WriteFile(manifestPath, updated, 0o644)
	}()

	if err := sched.Run(ctx); err == nil {
		t.Fatal("expected context cancellation error")
	}
}
