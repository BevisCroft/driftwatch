// Package cache provides a lightweight, thread-safe in-memory key/value
// cache used by driftwatch to avoid redundant manifest loads and repeated
// drift computations within a single poll cycle.
//
// Entries are given a configurable TTL; expired entries are not returned by
// Get but are only physically removed on the next Set when the cache is at
// capacity, or when Flush is called explicitly.
//
// An optional maxSize cap ensures the cache does not grow unboundedly in
// environments with large numbers of tracked services.
package cache
