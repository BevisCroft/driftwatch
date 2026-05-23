package digest_test

import (
	"testing"

	"github.com/driftwatch/internal/digest"
	"github.com/driftwatch/internal/manifest"
)

func baseManifest() manifest.Manifest {
	return manifest.Manifest{
		Kind:      "Deployment",
		Name:      "api-server",
		Namespace: "production",
		Spec: map[string]any{
			"replicas": 3,
			"image":    "api:v1.2.0",
		},
	}
}

func TestCompute_ReturnsSHA256Hex(t *testing.T) {
	d := digest.New()
	h, err := d.Compute(baseManifest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(h) != 64 {
		t.Errorf("expected 64-char hex digest, got %d chars: %s", len(h), h)
	}
}

func TestCompute_Deterministic(t *testing.T) {
	d := digest.New()
	m := baseManifest()

	h1, _ := d.Compute(m)
	h2, _ := d.Compute(m)

	if h1 != h2 {
		t.Errorf("expected same digest on repeated calls; got %s and %s", h1, h2)
	}
}

func TestEqual_IdenticalManifests(t *testing.T) {
	d := digest.New()

	eq, err := d.Equal(baseManifest(), baseManifest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !eq {
		t.Error("expected identical manifests to be equal")
	}
}

func TestEqual_DifferentSpec(t *testing.T) {
	d := digest.New()
	a := baseManifest()
	b := baseManifest()
	b.Spec["replicas"] = 5

	eq, err := d.Equal(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eq {
		t.Error("expected manifests with different spec to be unequal")
	}
}

func TestEqual_DifferentKind(t *testing.T) {
	d := digest.New()
	a := baseManifest()
	b := baseManifest()
	b.Kind = "StatefulSet"

	eq, err := d.Equal(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eq {
		t.Error("expected manifests with different kind to be unequal")
	}
}

func TestEqual_DifferentNamespace(t *testing.T) {
	d := digest.New()
	a := baseManifest()
	b := baseManifest()
	b.Namespace = "staging"

	eq, err := d.Equal(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eq {
		t.Error("expected manifests with different namespace to be unequal")
	}
}
