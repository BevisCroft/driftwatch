package tagindex_test

import (
	"sort"
	"testing"

	"github.com/example/driftwatch/internal/tagindex"
)

func sortedStrings(ss []string) []string {
	sort.Strings(ss)
	return ss
}

func TestAdd_And_Get_RoundTrip(t *testing.T) {
	idx := tagindex.New()
	tags := map[string]string{"env": "prod", "team": "platform"}
	if err := idx.Add("svc-a", tags); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e, ok := idx.Get("svc-a")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Tags["env"] != "prod" {
		t.Errorf("expected env=prod, got %q", e.Tags["env"])
	}
}

func TestAdd_EmptyService_ReturnsError(t *testing.T) {
	idx := tagindex.New()
	if err := idx.Add("", map[string]string{"k": "v"}); err == nil {
		t.Fatal("expected error for empty service name")
	}
}

func TestLookup_SingleTag(t *testing.T) {
	idx := tagindex.New()
	_ = idx.Add("svc-a", map[string]string{"env": "prod"})
	_ = idx.Add("svc-b", map[string]string{"env": "staging"})
	_ = idx.Add("svc-c", map[string]string{"env": "prod"})

	got := sortedStrings(idx.Lookup(map[string]string{"env": "prod"}))
	want := []string{"svc-a", "svc-c"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("Lookup = %v, want %v", got, want)
	}
}

func TestLookup_MultiTag_Intersection(t *testing.T) {
	idx := tagindex.New()
	_ = idx.Add("svc-a", map[string]string{"env": "prod", "team": "platform"})
	_ = idx.Add("svc-b", map[string]string{"env": "prod", "team": "data"})
	_ = idx.Add("svc-c", map[string]string{"env": "staging", "team": "platform"})

	got := idx.Lookup(map[string]string{"env": "prod", "team": "platform"})
	if len(got) != 1 || got[0] != "svc-a" {
		t.Errorf("expected [svc-a], got %v", got)
	}
}

func TestLookup_NoMatch_ReturnsEmpty(t *testing.T) {
	idx := tagindex.New()
	_ = idx.Add("svc-a", map[string]string{"env": "prod"})
	got := idx.Lookup(map[string]string{"env": "canary"})
	if len(got) != 0 {
		t.Errorf("expected empty result, got %v", got)
	}
}

func TestRemove_DeletesService(t *testing.T) {
	idx := tagindex.New()
	_ = idx.Add("svc-a", map[string]string{"env": "prod"})
	idx.Remove("svc-a")
	if _, ok := idx.Get("svc-a"); ok {
		t.Error("expected service to be removed")
	}
	got := idx.Lookup(map[string]string{"env": "prod"})
	if len(got) != 0 {
		t.Errorf("expected empty lookup after remove, got %v", got)
	}
}

func TestAdd_ReplacesExistingEntry(t *testing.T) {
	idx := tagindex.New()
	_ = idx.Add("svc-a", map[string]string{"env": "prod"})
	_ = idx.Add("svc-a", map[string]string{"env": "staging"})

	// Old tag must no longer be indexed.
	got := idx.Lookup(map[string]string{"env": "prod"})
	if len(got) != 0 {
		t.Errorf("old tag still indexed after replace: %v", got)
	}
	got = idx.Lookup(map[string]string{"env": "staging"})
	if len(got) != 1 || got[0] != "svc-a" {
		t.Errorf("expected [svc-a] under new tag, got %v", got)
	}
}
