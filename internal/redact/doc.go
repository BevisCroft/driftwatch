// Package redact provides a Redactor type that scrubs sensitive field
// values from manifests and drift results before they reach logs,
// reports, or external alerting systems.
//
// Patterns are registered as case-insensitive substrings matched
// against fully-qualified field paths such as "spec.env.DB_PASSWORD".
// Common defaults include "password", "secret", and "token", but
// callers may supply any set of patterns via New or AddPattern.
//
// The redaction replacement value is the string "[REDACTED]". Original
// values are never stored or returned once scrubbed.
//
// Thread safety: Redactor is safe for concurrent reads after
// construction. Calls to AddPattern must not overlap with any other
// method call; callers are responsible for external synchronisation if
// patterns are added after the Redactor is shared across goroutines.
//
// Usage:
//
//	r := redact.New([]string{"password", "secret", "token"})
//	safeMap := r.ScrubMap(driftFields)
//
// To extend an existing Redactor with additional patterns at setup time:
//
//	r.AddPattern("api_key")
//	r.AddPattern("private_key")
package redact
