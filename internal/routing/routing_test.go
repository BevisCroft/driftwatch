package routing

import (
	"testing"
)

func TestNew_EmptyEndpoints(t *testing.T) {
	_, err := New(nil)
	if err == nil {
		t.Fatal("expected error for empty endpoints")
	}
}

func TestNew_ZeroWeight(t *testing.T) {
	_, err := New([]Endpoint{{Name: "a", URL: "http://a", Weight: 0}})
	if err == nil {
		t.Fatal("expected error for zero weight")
	}
}

func TestNext_SingleEndpoint(t *testing.T) {
	r, err := New([]Endpoint{{Name: "a", URL: "http://a", Weight: 3}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 0; i < 6; i++ {
		ep := r.Next()
		if ep.Name != "a" {
			t.Fatalf("expected 'a', got %q", ep.Name)
		}
	}
}

func TestNext_WeightedDistribution(t *testing.T) {
	r, err := New([]Endpoint{
		{Name: "a", URL: "http://a", Weight: 2},
		{Name: "b", URL: "http://b", Weight: 1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	counts := map[string]int{}
	for i := 0; i < 9; i++ {
		ep := r.Next()
		counts[ep.Name]++
	}
	if counts["a"] != 6 {
		t.Errorf("expected 6 selections for 'a', got %d", counts["a"])
	}
	if counts["b"] != 3 {
		t.Errorf("expected 3 selections for 'b', got %d", counts["b"])
	}
}

func TestReset_RestoresState(t *testing.T) {
	r, _ := New([]Endpoint{
		{Name: "a", URL: "http://a", Weight: 1},
		{Name: "b", URL: "http://b", Weight: 1},
	})
	r.Next() // a
	r.Next() // b
	r.Reset()
	ep := r.Next()
	if ep.Name != "a" {
		t.Errorf("expected 'a' after reset, got %q", ep.Name)
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	r, _ := New([]Endpoint{
		{Name: "x", URL: "http://x", Weight: 1},
	})
	all := r.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(all))
	}
	all[0].Name = "mutated"
	if r.All()[0].Name != "x" {
		t.Error("All() should return a copy, not a reference")
	}
}
