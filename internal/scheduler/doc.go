// Package scheduler orchestrates periodic drift-detection cycles.
//
// A Scheduler is initialised with a Config that supplies the poll interval
// and manifest directory. On each tick it loads all manifests, compares them
// against the most-recent snapshots, and forwards any detected drift to the
// configured Reporter.
//
// Typical usage:
//
//	sched := scheduler.New(cfg, loader, detector, store, rep)
//	if err := sched.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
//		log.Fatal(err)
//	}
package scheduler
