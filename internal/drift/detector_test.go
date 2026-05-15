package drift_test

import (
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/manifest"
)

func baseManifest() *manifest.Manifest {
	return &manifest.Manifest{
		Kind: "Service",
		Name: "api",
		Spec: map[string]interface{}{
			"replicas": 3,
			"image":    "api:v1.2.0",
		},
	}
}

func TestCompare_NoDrift(t *testing.T) {
	d := drift.NewDetector()
	src := baseManifest()
	dep := baseManifest()

	results := d.Compare(src, dep)
	if len(results) != 0 {
		t.Errorf("expected no drift, got %d result(s): %v", len(results), results)
	}
}

func TestCompare_KindChanged(t *testing.T) {
	d := drift.NewDetector()
	src := baseManifest()
	dep := baseManifest()
	dep.Kind = "Deployment"

	results := d.Compare(src, dep)
	if len(results) != 1 {
		t.Fatalf("expected 1 drift result, got %d", len(results))
	}
	if results[0].Field != "kind" || results[0].Type != drift.DriftTypeChanged {
		t.Errorf("unexpected drift result: %v", results[0])
	}
}

func TestCompare_SpecFieldChanged(t *testing.T) {
	d := drift.NewDetector()
	src := baseManifest()
	dep := baseManifest()
	dep.Spec["replicas"] = 1

	results := d.Compare(src, dep)
	if len(results) != 1 {
		t.Fatalf("expected 1 drift result, got %d", len(results))
	}
	if results[0].Field != "spec.replicas" || results[0].Type != drift.DriftTypeChanged {
		t.Errorf("unexpected drift result: %v", results[0])
	}
}

func TestCompare_SpecFieldAdded(t *testing.T) {
	d := drift.NewDetector()
	src := baseManifest()
	dep := baseManifest()
	dep.Spec["port"] = 8080

	results := d.Compare(src, dep)
	if len(results) != 1 {
		t.Fatalf("expected 1 drift result, got %d", len(results))
	}
	if results[0].Type != drift.DriftTypeAdded {
		t.Errorf("expected DriftTypeAdded, got %s", results[0].Type)
	}
}

func TestCompare_SpecFieldRemoved(t *testing.T) {
	d := drift.NewDetector()
	src := baseManifest()
	dep := baseManifest()
	delete(dep.Spec, "image")

	results := d.Compare(src, dep)
	if len(results) != 1 {
		t.Fatalf("expected 1 drift result, got %d", len(results))
	}
	if results[0].Type != drift.DriftTypeRemoved {
		t.Errorf("expected DriftTypeRemoved, got %s", results[0].Type)
	}
}
