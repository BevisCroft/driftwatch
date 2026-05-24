package labelindex

import (
	"testing"
)

func TestSet_And_Lookup_SingleLabel(t *testing.T) {
	idx := New()
	_ = idx.Set("svc-a", map[string]string{"env": "prod"})

	got := idx.Lookup(map[string]string{"env": "prod"})
	if len(got) != 1 || got[0] != "svc-a" {
		t.Fatalf("expected [svc-a], got %v", got)
	}
}

func TestLookup_NoMatch_ReturnsNil(t *testing.T) {
	idx := New()
	_ = idx.Set("svc-a", map[string]string{"env": "prod"})

	got := idx.Lookup(map[string]string{"env": "staging"})
	if len(got) != 0 {
		t.Fatalf("expected empty result, got %v", got)
	}
}

func TestLookup_Intersection_MultipleLabels(t *testing.T) {
	idx := New()
	_ = idx.Set("svc-a", map[string]string{"env": "prod", "team": "platform"})
	_ = idx.Set("svc-b", map[string]string{"env": "prod", "team": "infra"})
	_ = idx.Set("svc-c", map[string]string{"env": "staging", "team": "platform"})

	got := idx.Lookup(map[string]string{"env": "prod", "team": "platform"})
	if len(got) != 1 || got[0] != "svc-a" {
		t.Fatalf("expected [svc-a], got %v", got)
	}
}

func TestLookup_EmptySelector_ReturnsAll(t *testing.T) {
	idx := New()
	_ = idx.Set("svc-a", map[string]string{"env": "prod"})
	_ = idx.Set("svc-b", map[string]string{"env": "staging"})

	got := idx.Lookup(nil)
	if len(got) != 2 {
		t.Fatalf("expected 2 services, got %v", got)
	}
}

func TestSet_ReplacesOldLabels(t *testing.T) {
	idx := New()
	_ = idx.Set("svc-a", map[string]string{"env": "prod"})
	_ = idx.Set("svc-a", map[string]string{"env": "staging"})

	prod := idx.Lookup(map[string]string{"env": "prod"})
	if len(prod) != 0 {
		t.Fatalf("old label should be gone, got %v", prod)
	}
	staging := idx.Lookup(map[string]string{"env": "staging"})
	if len(staging) != 1 || staging[0] != "svc-a" {
		t.Fatalf("expected [svc-a] under staging, got %v", staging)
	}
}

func TestRemove_DeletesService(t *testing.T) {
	idx := New()
	_ = idx.Set("svc-a", map[string]string{"env": "prod"})
	idx.Remove("svc-a")

	got := idx.Lookup(map[string]string{"env": "prod"})
	if len(got) != 0 {
		t.Fatalf("expected empty after remove, got %v", got)
	}
}

func TestSet_EmptyService_ReturnsError(t *testing.T) {
	idx := New()
	err := idx.Set("", map[string]string{"env": "prod"})
	if err == nil {
		t.Fatal("expected error for empty service name")
	}
}

func TestLookup_ResultsAreSorted(t *testing.T) {
	idx := New()
	_ = idx.Set("svc-z", map[string]string{"env": "prod"})
	_ = idx.Set("svc-a", map[string]string{"env": "prod"})
	_ = idx.Set("svc-m", map[string]string{"env": "prod"})

	got := idx.Lookup(map[string]string{"env": "prod"})
	expected := []string{"svc-a", "svc-m", "svc-z"}
	for i, v := range expected {
		if got[i] != v {
			t.Fatalf("expected sorted %v, got %v", expected, got)
		}
	}
}
