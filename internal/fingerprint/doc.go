// Package fingerprint provides lightweight structural fingerprinting for
// service manifests. It computes deterministic SHA-256 hashes over a
// manifest's service name, kind, and spec fields, allowing the scheduler
// to skip full drift detection when nothing has changed between cycles.
//
// The Store is safe for concurrent use.
package fingerprint
