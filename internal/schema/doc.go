// Package schema validates the structural correctness of service manifests
// before they are processed by the drift detector.
//
// Validation rules include:
//   - Required top-level fields (kind, name by default)
//   - Forbidden spec keys that should never appear in a deployed manifest
//   - Non-blank string constraints on critical fields
//
// Use [New] to create a [Validator], optionally customised with [Option]
// functions, then call [Validator.Validate] for each manifest.
package schema
