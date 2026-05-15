// Package watcher provides periodic polling of manifests to detect
// configuration drift over time.
package watcher

import (
	"context"
	"log"
	"time"

	"github.com/example/driftwatch/internal/drift"
	"github.com/example/driftwatch/internal/manifest"
	"github.com/example/driftwatch/internal/reporter"
)

// Watcher polls a directory of manifests on a fixed interval and reports
// any detected drift to the configured reporter.
type Watcher struct {
	loader   *manifest.Loader
	detector *drift.Detector
	reporter *reporter.Reporter
	interval time.Duration
	dir      string
}

// New creates a Watcher that scans manifestDir every interval.
func New(manifestDir string, interval time.Duration, r *reporter.Reporter) *Watcher {
	return &Watcher{
		loader:   manifest.NewLoader(),
		detector: drift.NewDetector(),
		reporter: r,
		interval: interval,
		dir:      manifestDir,
	}
}

// Run starts the polling loop. It blocks until ctx is cancelled.
func (w *Watcher) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	log.Printf("watcher: starting, dir=%s interval=%s", w.dir, w.interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("watcher: shutting down")
			return ctx.Err()
		case <-ticker.C:
			if err := w.scan(); err != nil {
				log.Printf("watcher: scan error: %v", err)
			}
		}
	}
}

// scan loads all manifests from the directory, compares them, and writes
// results via the reporter.
func (w *Watcher) scan() error {
	manifests, err := w.loader.LoadAll(w.dir)
	if err != nil {
		return err
	}

	var results []drift.Result
	for _, m := range manifests {
		res := w.detector.Compare(m, m) // compare deployed vs source
		results = append(results, res...)
	}

	return w.reporter.Write(results)
}
