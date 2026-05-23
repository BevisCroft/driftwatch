// Package throttle implements per-service burst throttling for drift
// notifications. It ensures that a single noisy service cannot flood
// downstream alerting channels by capping the number of notifications
// emitted within a configurable sliding time window.
package throttle
