// Package audit provides an append-only, newline-delimited JSON audit log
// for recording drift detection events per service.
//
// Usage:
//
//	logger, err := audit.New("/var/log/driftwatch/audit.log")
//	if err != nil { ... }
//	defer logger.Close()
//
//	logger.Record("my-service", "drift_detected", "spec.replicas changed 2 → 3")
//
// Entries can be replayed with audit.ReadAll.
package audit
