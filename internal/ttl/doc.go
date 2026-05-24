// Package ttl provides a lightweight, thread-safe in-memory cache with
// per-entry time-to-live expiry. It is used by driftwatch components that
// need to hold transient state (e.g. recent alert keys, deduplication tokens)
// without an external dependency.
//
// Entries are evicted lazily on read and eagerly by a configurable background
// sweep goroutine. Call Stop on the Cache to release the goroutine when the
// cache is no longer needed.
package ttl
