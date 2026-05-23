package fingerprint_test

import (
	"testing"

	"github.com/example/driftwatch/internal/fingerprint"
)

// stubManifest implements fingerprint.Manifest for testing.
type stubManifest struct {
	service string
	kind    string
	spec    map[string]any
}

func (s stubManifest) ServiceName() string       { return s.service }
func (s stubManifest) Kind() string              { return s.kind }
func (s stubManifest) Spec() map[string]any      { return s.spec }

func baseManifest() stubManifest {
	return stubManifest{
		service: "api-server",
		kind:    "Deployment",
		spec:    map[string]any{"replicas": 3, "image": "nginx:1.25"},
	}
}

func TestCompute_ReturnsSHA256Hex(t *testing.T) {
	m := baseManifest()
	fp, err := fingerprint.Compute(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fp) != 64 {
		t.Errorf("expected 64-char hex string, got len=%d", len(fp))
	}
}

func TestCompute_Deterministic(t *testing.T) {
	m := baseManifest()
	a, _ := fingerprint.Compute(m)
	b, _ := fingerprint.Compute(m)
	if a != b {
		t.Errorf("fingerprints differ across calls: %s vs %s", a, b)
	}
}

func TestCompute_DifferentSpec(t *testing.T) {
	m1 := baseManifest()
	m2 := baseManifest()
	m2.spec = map[string]any{"replicas": 5, "image": "nginx:1.25"}

	a, _ := fingerprint.Compute(m1)
	b, _ := fingerprint.Compute(m2)
	if a == b {
		t.Error("expected different fingerprints for different specs")
	}
}

func TestStore_ChangedOnFirstSeen(t *testing.T) {
	s := fingerprint.New()
	m := baseManifest()

	changed, err := s.Changed(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Error("expected Changed=true for unseen service")
	}
}

func TestStore_NoChangeAfterUpdate(t *testing.T) {
	s := fingerprint.New()
	m := baseManifest()

	if err := s.Update(m); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	changed, err := s.Changed(m)
	if err != nil {
		t.Fatalf("Changed failed: %v", err)
	}
	if changed {
		t.Error("expected Changed=false after Update with same manifest")
	}
}

func TestStore_DetectsSpecChange(t *testing.T) {
	s := fingerprint.New()
	m1 := baseManifest()

	_ = s.Update(m1)

	m2 := baseManifest()
	m2.spec = map[string]any{"replicas": 1, "image": "nginx:1.26"}

	changed, err := s.Changed(m2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Error("expected Changed=true after spec mutation")
	}
}
