// Package drift implements configuration drift detection for driftwatch.
//
// It compares a source manifest (the desired state defined in version control)
// against a deployed manifest (the live state of a running service) and
// reports any discrepancies as a list of [DriftResult] values.
//
// # Basic Usage
//
//	detector := drift.NewDetector()
//	results := detector.Compare(sourceManifest, deployedManifest)
//	for _, r := range results {
//		fmt.Println(r)
//	}
//
// Each [DriftResult] describes a single field that has drifted, including
// the field path, expected value, actual value, and the [DriftType]
// (added, removed, or changed).
package drift
