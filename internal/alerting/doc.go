// Package alerting provides drift alerting capabilities for driftwatch.
//
// It consumes drift.Result values produced by the detector and emits
// human-readable alerts to a configured io.Writer (default: os.Stderr).
//
// Alert severity is determined by the number of drifted fields relative
// to a configurable threshold:
//
//	< threshold  → WARN
//	>= threshold → ERROR
//
// Services with no drift produce no output.
package alerting
