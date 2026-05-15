package drift

// Diff describes a single field-level difference between the expected
// manifest value and the value observed in the live service.
type Diff struct {
	// Field is the dot-separated path to the differing field (e.g. "spec.replicas").
	Field string `json:"field"`
	// Expected is the value declared in the source manifest.
	Expected interface{} `json:"expected"`
	// Actual is the value observed in the deployed service.
	Actual interface{} `json:"actual"`
}

// Result holds the drift comparison outcome for a single service manifest.
type Result struct {
	// Name is the identifier of the manifest / service being compared.
	Name string `json:"name"`
	// HasDrift is true when one or more field differences were detected.
	HasDrift bool `json:"has_drift"`
	// Diffs contains the list of individual field differences.
	// It is empty when HasDrift is false.
	Diffs []Diff `json:"diffs,omitempty"`
}

// Summary returns a short human-readable status string for the result.
func (r Result) Summary() string {
	if !r.HasDrift {
		return r.Name + ": OK"
	}
	return r.Name + ": DRIFT DETECTED"
}
