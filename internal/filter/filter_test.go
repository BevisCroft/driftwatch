package filter_test

import (
	"testing"

	"github.com/driftwatch/internal/filter"
	"github.com/driftwatch/internal/manifest"
)

func sampleManifests() []manifest.Manifest {
	return []manifest.Manifest{
		{Name: "api", Namespace: "production", Kind: "Deployment", Labels: map[string]string{"team": "backend"}},
		{Name: "worker", Namespace: "production", Kind: "Deployment", Labels: map[string]string{"team": "data"}},
		{Name: "cache", Namespace: "staging", Kind: "StatefulSet", Labels: map[string]string{"team": "backend"}},
		{Name: "gateway", Namespace: "staging", Kind: "Deployment"},
	}
}

func TestFilter_NoOptions(t *testing.T) {
	result := filter.Filter(sampleManifests(), filter.Options{})
	if len(result) != 4 {
		t.Fatalf("expected 4 manifests, got %d", len(result))
	}
}

func TestFilter_ByNamespace(t *testing.T) {
	opts := filter.Options{Namespaces: []string{"production"}}
	result := filter.Filter(sampleManifests(), opts)
	if len(result) != 2 {
		t.Fatalf("expected 2 manifests, got %d", len(result))
	}
	for _, m := range result {
		if m.Namespace != "production" {
			t.Errorf("unexpected namespace %q", m.Namespace)
		}
	}
}

func TestFilter_ByLabel(t *testing.T) {
	opts := filter.Options{LabelSelector: "team=backend"}
	result := filter.Filter(sampleManifests(), opts)
	if len(result) != 2 {
		t.Fatalf("expected 2 manifests, got %d", len(result))
	}
}

func TestFilter_ByNamespaceAndLabel(t *testing.T) {
	opts := filter.Options{
		Namespaces:    []string{"staging"},
		LabelSelector: "team=backend",
	}
	result := filter.Filter(sampleManifests(), opts)
	if len(result) != 1 {
		t.Fatalf("expected 1 manifest, got %d", len(result))
	}
	if result[0].Name != "cache" {
		t.Errorf("expected 'cache', got %q", result[0].Name)
	}
}

func TestFilter_NoMatch(t *testing.T) {
	opts := filter.Options{Namespaces: []string{"canary"}}
	result := filter.Filter(sampleManifests(), opts)
	if len(result) != 0 {
		t.Fatalf("expected 0 manifests, got %d", len(result))
	}
}

func TestFilter_InvalidLabelSelector(t *testing.T) {
	opts := filter.Options{LabelSelector: "malformed"}
	result := filter.Filter(sampleManifests(), opts)
	if len(result) != 0 {
		t.Fatalf("expected 0 manifests for invalid selector, got %d", len(result))
	}
}

func TestFilter_LabelSelectorNilLabels(t *testing.T) {
	manifests := []manifest.Manifest{
		{Name: "nolabels", Namespace: "production", Kind: "Job"},
	}
	opts := filter.Options{LabelSelector: "team=backend"}
	result := filter.Filter(manifests, opts)
	if len(result) != 0 {
		t.Fatalf("expected 0 manifests, got %d", len(result))
	}
}
