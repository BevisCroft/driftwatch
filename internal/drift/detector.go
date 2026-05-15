// Package drift provides functionality for detecting configuration drift
// between deployed service states and their source manifests.
package drift

import (
	"fmt"

	"github.com/driftwatch/internal/manifest"
)

// DriftType categorizes the kind of drift detected.
type DriftType string

const (
	DriftTypeAdded   DriftType = "added"
	DriftTypeRemoved DriftType = "removed"
	DriftTypeChanged DriftType = "changed"
)

// DriftResult holds the result of comparing two manifests.
type DriftResult struct {
	Field    string
	Expected interface{}
	Actual   interface{}
	Type     DriftType
}

// String returns a human-readable description of the drift.
func (d DriftResult) String() string {
	return fmt.Sprintf("[%s] field=%q expected=%v actual=%v", d.Type, d.Field, d.Expected, d.Actual)
}

// Detector compares manifests and reports drift.
type Detector struct{}

// NewDetector creates a new Detector instance.
func NewDetector() *Detector {
	return &Detector{}
}

// Compare compares a deployed manifest against the source manifest and
// returns a list of detected drift results.
func (d *Detector) Compare(source, deployed *manifest.Manifest) []DriftResult {
	var results []DriftResult

	if source.Kind != deployed.Kind {
		results = append(results, DriftResult{
			Field:    "kind",
			Expected: source.Kind,
			Actual:   deployed.Kind,
			Type:     DriftTypeChanged,
		})
	}

	if source.Name != deployed.Name {
		results = append(results, DriftResult{
			Field:    "name",
			Expected: source.Name,
			Actual:   deployed.Name,
			Type:     DriftTypeChanged,
		})
	}

	results = append(results, compareFields(source.Spec, deployed.Spec)...)

	return results
}

// compareFields performs a shallow comparison of two spec maps.
func compareFields(source, deployed map[string]interface{}) []DriftResult {
	var results []DriftResult

	for k, sv := range source {
		dv, ok := deployed[k]
		if !ok {
			results = append(results, DriftResult{Field: "spec." + k, Expected: sv, Actual: nil, Type: DriftTypeRemoved})
			continue
		}
		if fmt.Sprintf("%v", sv) != fmt.Sprintf("%v", dv) {
			results = append(results, DriftResult{Field: "spec." + k, Expected: sv, Actual: dv, Type: DriftTypeChanged})
		}
	}

	for k, dv := range deployed {
		if _, ok := source[k]; !ok {
			results = append(results, DriftResult{Field: "spec." + k, Expected: nil, Actual: dv, Type: DriftTypeAdded})
		}
	}

	return results
}
