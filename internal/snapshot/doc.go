// Package snapshot provides a lightweight store for persisting manifest
// snapshots to disk. Snapshots capture the name, kind, and field values
// of a manifest at a point in time and are used by the drift detector to
// compare the current deployed state against a known-good baseline.
//
// Snapshots are serialised as JSON files under a configurable directory,
// one file per manifest name. Saving a snapshot is idempotent — each
// call overwrites the previous file for that manifest.
package snapshot
