// Package scheduler provides periodic execution of drift detection cycles.
package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/example/driftwatch/internal/config"
	"github.com/example/driftwatch/internal/drift"
	"github.com/example/driftwatch/internal/manifest"
	"github.com/example/driftwatch/internal/reporter"
	"github.com/example/driftwatch/internal/snapshot"
)

// Scheduler runs drift detection on a fixed interval derived from config.
type Scheduler struct {
	cfg      *config.Config
	loader   *manifest.Loader
	detector *drift.Detector
	store    *snapshot.Store
	rep      *reporter.Reporter
}

// New creates a Scheduler wired to the provided dependencies.
func New(cfg *config.Config, loader *manifest.Loader, detector *drift.Detector, store *snapshot.Store, rep *reporter.Reporter) *Scheduler {
	return &Scheduler{
		cfg:      cfg,
		loader:   loader,
		detector: detector,
		store:    store,
		rep:      rep,
	}
}

// Run starts the periodic detection loop and blocks until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.cfg.PollInterval)
	defer ticker.Stop()

	log.Printf("scheduler: starting with interval %s", s.cfg.PollInterval)

	// Run an immediate cycle before waiting for the first tick.
	if err := s.cycle(ctx); err != nil {
		log.Printf("scheduler: cycle error: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("scheduler: context cancelled, stopping")
			return ctx.Err()
		case <-ticker.C:
			if err := s.cycle(ctx); err != nil {
				log.Printf("scheduler: cycle error: %v", err)
			}
		}
	}
}

// cycle loads manifests, compares against snapshots, and reports results.
func (s *Scheduler) cycle(ctx context.Context) error {
	manifests, err := s.loader.LoadAll(s.cfg.ManifestDir)
	if err != nil {
		return err
	}

	var results []drift.Result
	for _, m := range manifests {
		prev, err := s.store.Load(m.Name)
		if err != nil {
			// No previous snapshot — save current as baseline.
			_ = s.store.Save(m.Name, m)
			continue
		}
		res := s.detector.Compare(prev, m)
		results = append(results, res)
		_ = s.store.Save(m.Name, m)
	}

	return s.rep.Write(results)
}
