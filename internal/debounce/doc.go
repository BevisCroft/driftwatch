// Package debounce provides a per-service quiet-window debouncer for
// drift notification events.
//
// When multiple drift cycles fire in quick succession for the same service,
// the Debouncer ensures that downstream alerting and notification paths are
// only triggered once per quiet window, reducing noise during transient
// configuration churn.
package debounce
