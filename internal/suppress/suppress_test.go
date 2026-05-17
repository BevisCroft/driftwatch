package suppress

import (
	"testing"
	"time"
)

func TestIsSuppressed_MatchesServiceAndField(t *testing.T) {
	l := New()
	l.Add(Entry{Service: "svc-a", Field: "spec.replicas", Reason: "known drift"})

	if !l.IsSuppressed("svc-a", "spec.replicas") {
		t.Fatal("expected suppression to match")
	}
	if l.IsSuppressed("svc-b", "spec.replicas") {
		t.Fatal("different service should not match")
	}
	if l.IsSuppressed("svc-a", "spec.image") {
		t.Fatal("different field should not match")
	}
}

func TestIsSuppressed_WildcardField(t *testing.T) {
	l := New()
	l.Add(Entry{Service: "svc-a", Field: "*", Reason: "all fields"})

	if !l.IsSuppressed("svc-a", "spec.replicas") {
		t.Fatal("wildcard should match any field")
	}
	if !l.IsSuppressed("svc-a", "spec.image") {
		t.Fatal("wildcard should match any field")
	}
}

func TestIsSuppressed_ExpiredEntryIgnored(t *testing.T) {
	l := New()
	past := time.Now().Add(-1 * time.Hour)
	l.Add(Entry{Service: "svc-a", Field: "spec.replicas", ExpiresAt: past})

	if l.IsSuppressed("svc-a", "spec.replicas") {
		t.Fatal("expired entry should not suppress")
	}
}

func TestIsSuppressed_NonExpiredEntryMatches(t *testing.T) {
	l := New()
	future := time.Now().Add(1 * time.Hour)
	l.Add(Entry{Service: "svc-a", Field: "spec.replicas", ExpiresAt: future})

	if !l.IsSuppressed("svc-a", "spec.replicas") {
		t.Fatal("non-expired entry should suppress")
	}
}

func TestPurge_RemovesExpiredEntries(t *testing.T) {
	l := New()
	past := time.Now().Add(-1 * time.Hour)
	future := time.Now().Add(1 * time.Hour)
	l.Add(Entry{Service: "svc-a", Field: "f1", ExpiresAt: past})
	l.Add(Entry{Service: "svc-b", Field: "f2", ExpiresAt: future})

	l.Purge()
	snap := l.Snapshot()
	if len(snap) != 1 {
		t.Fatalf("expected 1 entry after purge, got %d", len(snap))
	}
	if snap[0].Service != "svc-b" {
		t.Fatalf("expected svc-b to survive purge, got %s", snap[0].Service)
	}
}

func TestSnapshot_ExcludesExpired(t *testing.T) {
	l := New()
	past := time.Now().Add(-1 * time.Hour)
	l.Add(Entry{Service: "svc-a", Field: "f1", ExpiresAt: past})
	l.Add(Entry{Service: "svc-b", Field: "f2"})

	snap := l.Snapshot()
	if len(snap) != 1 {
		t.Fatalf("expected 1 snapshot entry, got %d", len(snap))
	}
}
