// Package checkpoint provides a lightweight persistent store for driftwatch
// scan-cycle checkpoints.
//
// A checkpoint captures the outcome of each completed drift detection cycle —
// including timestamp, manifest count, and drift count — and writes it
// atomically to a JSON file. On daemon restart the scheduler can read the
// last checkpoint to emit accurate "time since last scan" metrics and to
// skip redundant alerting for already-known drift.
//
// Usage:
//
//	store := checkpoint.New("/var/lib/driftwatch/checkpoint.json")
//
//	// after a cycle completes:
//	store.Save(checkpoint.Entry{
//		CycleID:    uuid,
//		Timestamp:  time.Now(),
//		Manifests:  12,
//		DriftCount: 2,
//	})
//
//	// on startup:
//	entry, _ := store.Load()
package checkpoint
