// Package retrylog provides an in-memory log of per-service retry
// events with configurable age-based pruning.
//
// Typical usage:
//
//	log := retrylog.New(10 * time.Minute)
//	log.Record("auth-service", "connection refused", 1)
//	summaries := log.Summaries()
//
Entries older than the configured maxAge are excluded from Summaries
but are not immediately removed from memory; call Reset to clear a
service's history explicitly.
package retrylog
