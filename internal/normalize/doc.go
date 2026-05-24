// Package normalize provides field-level normalization for service manifests.
//
// Before the drift detector compares a deployed manifest against its source,
// both spec maps are passed through a Normalizer so that semantically
// equivalent but syntactically different values — such as a zero replica count
// that defaults to one, or an unordered tag list — do not produce spurious
// drift results.
//
// Usage:
//
//	n := normalize.New()
//	normalizedSpec := n.Apply(rawSpec)
//
// Custom rules can be registered at runtime via AddRule to handle
// application-specific field semantics.
package normalize
