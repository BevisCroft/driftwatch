// Package redact provides utilities for scrubbing sensitive fields
// from manifests and drift results before they are logged, reported,
// or transmitted to external alerting systems.
package redact

import (
	"strings"
	"sync"
)

const redactedValue = "[REDACTED]"

// Redactor holds a set of field name patterns that should be scrubbed.
type Redactor struct {
	mu       sync.RWMutex
	patterns []string
}

// New returns a Redactor pre-loaded with the supplied field patterns.
// Patterns are matched case-insensitively as substring matches against
// fully-qualified field paths (e.g. "spec.env.SECRET_KEY").
func New(patterns []string) *Redactor {
	norm := make([]string, len(patterns))
	for i, p := range patterns {
		norm[i] = strings.ToLower(p)
	}
	return &Redactor{patterns: norm}
}

// AddPattern appends a new pattern to the redactor at runtime.
func (r *Redactor) AddPattern(pattern string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.patterns = append(r.patterns, strings.ToLower(pattern))
}

// IsSensitive reports whether the given field path matches any
// registered pattern.
func (r *Redactor) IsSensitive(field string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	lower := strings.ToLower(field)
	for _, p := range r.patterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

// ScrubMap returns a shallow copy of m with sensitive values replaced
// by the redacted placeholder. Keys are field path strings.
func (r *Redactor) ScrubMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		if r.IsSensitive(k) {
			out[k] = redactedValue
		} else {
			out[k] = v
		}
	}
	return out
}

// ScrubValue returns the redacted placeholder if field is sensitive,
// otherwise it returns value unchanged.
func (r *Redactor) ScrubValue(field, value string) string {
	if r.IsSensitive(field) {
		return redactedValue
	}
	return value
}
