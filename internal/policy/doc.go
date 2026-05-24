// Package policy provides rule-based severity overrides for drift results.
//
// Operators define rules in a YAML file that map service/field glob patterns
// to a desired severity level. The Evaluator walks each drift.Result and
// rewrites its Severity field when the first matching rule is found.
//
// Example policy file:
//
//	rules:
//	  - service: "payments-*"
//	    field:   "spec.replicas"
//	    severity: warn
//	  - service: "*"
//	    field:   "spec.image"
//	    severity: error
package policy
