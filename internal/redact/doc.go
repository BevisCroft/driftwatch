// Package redact provides a Redactor type that scrubs sensitive field
// values from manifests and drift results before they reach logs,
// reports, or external alerting systems.
//
// Patterns are registered as case-insensitive substrings matched
// against fully-qualified field paths such as "spec.env.DB_PASSWORD".
// Common defaults include "password", "secret", and "token", but
// callers may supply any set of patterns via New or AddPattern.
//
// Usage:
//
//	r := redact.New([]string{"password", "secret", "token"})
//	safeMap := r.ScrubMap(driftFields)
package redact
