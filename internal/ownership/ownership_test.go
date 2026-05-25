package ownership

import (
	"testing"
)

func TestSet_And_Get_RoundTrip(t *testing.T) {
	reg := New()
	e := Entry{Service: "svc-a", Team: "alpha", Contacts: []string{"a@example.com"}}
	if err := reg.Set(e); err != nil {
		t.Fatalf("Set: unexpected error: %v", err)
	}
	got, ok := reg.Get("svc-a")
	if !ok {
		t.Fatal("Get: expected entry to exist")
	}
	if got.Team != "alpha" {
		t.Errorf("team: got %q, want %q", got.Team, "alpha")
	}
}

func TestSet_EmptyService_ReturnsError(t *testing.T) {
	reg := New()
	if err := reg.Set(Entry{Team: "alpha"}); err == nil {
		t.Fatal("expected error for empty service")
	}
}

func TestSet_EmptyTeam_ReturnsError(t *testing.T) {
	reg := New()
	if err := reg.Set(Entry{Service: "svc-a"}); err == nil {
		t.Fatal("expected error for empty team")
	}
}

func TestGet_UnknownService_ReturnsFalse(t *testing.T) {
	reg := New()
	_, ok := reg.Get("missing")
	if ok {
		t.Fatal("expected false for unknown service")
	}
}

func TestRemove_ExistingEntry(t *testing.T) {
	reg := New()
	_ = reg.Set(Entry{Service: "svc-b", Team: "beta"})
	if !reg.Remove("svc-b") {
		t.Fatal("Remove: expected true for existing entry")
	}
	_, ok := reg.Get("svc-b")
	if ok {
		t.Fatal("Get: expected entry to be deleted")
	}
}

func TestRemove_NotFound_ReturnsFalse(t *testing.T) {
	reg := New()
	if reg.Remove("ghost") {
		t.Fatal("expected false for non-existent service")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	reg := New()
	_ = reg.Set(Entry{Service: "svc-a", Team: "alpha"})
	_ = reg.Set(Entry{Service: "svc-b", Team: "beta"})
	all := reg.All()
	if len(all) != 2 {
		t.Errorf("All: got %d entries, want 2", len(all))
	}
}

func TestSet_Overwrites_PreviousEntry(t *testing.T) {
	reg := New()
	_ = reg.Set(Entry{Service: "svc-a", Team: "alpha"})
	_ = reg.Set(Entry{Service: "svc-a", Team: "gamma"})
	e, _ := reg.Get("svc-a")
	if e.Team != "gamma" {
		t.Errorf("expected team %q, got %q", "gamma", e.Team)
	}
}
