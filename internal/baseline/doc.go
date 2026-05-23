// Package baseline manages pinned, approved configuration states for services
// monitored by driftwatch.
//
// A baseline represents the last known-good or explicitly approved set of
// configuration field values for a service. During drift detection cycles,
// detected changes can be compared against the baseline to determine whether
// drift is expected (i.e. matches a pinned baseline) or genuinely anomalous.
//
// Baselines are stored as JSON files on disk and are safe for concurrent use.
package baseline
