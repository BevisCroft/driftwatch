// Package policy evaluates drift results against user-defined severity policies,
// allowing operators to promote or demote drift severity based on field patterns
// and service selectors.
package policy

import (
	"path"
	"strings"

	"github.com/driftwatch/internal/drift"
)

// Rule defines a single policy rule that overrides the severity of a drift
// result when both the service and field patterns match.
type Rule struct {
	// ServiceGlob is a glob pattern matched against the service name.
	// Use "*" to match all services.
	ServiceGlob string
	// FieldGlob is a glob pattern matched against the drifted field path.
	// Use "*" to match all fields.
	FieldGlob string
	// Severity is the severity to assign when the rule matches.
	Severity drift.Severity
}

// Evaluator applies a set of Rules to drift results, rewriting severity
// values where a rule matches.
type Evaluator struct {
	rules []Rule
}

// New returns a new Evaluator loaded with the provided rules.
// Rules are evaluated in order; the first match wins.
func New(rules []Rule) *Evaluator {
	return &Evaluator{rules: rules}
}

// Apply returns a copy of results with severities rewritten according to
// the configured rules. Results with no matching rule are returned unchanged.
func (e *Evaluator) Apply(results []drift.Result) []drift.Result {
	out := make([]drift.Result, len(results))
	for i, r := range results {
		out[i] = e.applyOne(r)
	}
	return out
}

func (e *Evaluator) applyOne(r drift.Result) drift.Result {
	for _, rule := range e.rules {
		if matchGlob(rule.ServiceGlob, r.Service) && fieldsMatch(rule.FieldGlob, r.Fields) {
			r.Severity = rule.Severity
			return r
		}
	}
	return r
}

func fieldsMatch(glob string, fields []string) bool {
	for _, f := range fields {
		if matchGlob(glob, f) {
			return true
		}
	}
	return false
}

func matchGlob(pattern, s string) bool {
	if pattern == "" {
		return false
	}
	matched, err := path.Match(strings.ToLower(pattern), strings.ToLower(s))
	return err == nil && matched
}
