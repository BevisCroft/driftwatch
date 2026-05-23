package diff_test

import (
	"testing"

	"github.com/example/driftwatch/internal/diff"
)

func TestCompute_NoDiff(t *testing.T) {
	baseline := map[string]any{"replicas": 3, "image": "nginx:1.25"}
	current := map[string]any{"replicas": 3, "image": "nginx:1.25"}

	changes := diff.Compute(baseline, current)
	if len(changes) != 0 {
		t.Fatalf("expected no changes, got %d", len(changes))
	}
}

func TestCompute_ModifiedField(t *testing.T) {
	baseline := map[string]any{"replicas": 3}
	current := map[string]any{"replicas": 5}

	changes := diff.Compute(baseline, current)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Kind != diff.Modified {
		t.Errorf("expected Modified, got %s", changes[0].Kind)
	}
	if changes[0].Field != "replicas" {
		t.Errorf("unexpected field: %s", changes[0].Field)
	}
}

func TestCompute_AddedField(t *testing.T) {
	baseline := map[string]any{"image": "nginx:1.25"}
	current := map[string]any{"image": "nginx:1.25", "resources": "limits"}

	changes := diff.Compute(baseline, current)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Kind != diff.Added {
		t.Errorf("expected Added, got %s", changes[0].Kind)
	}
}

func TestCompute_RemovedField(t *testing.T) {
	baseline := map[string]any{"image": "nginx:1.25", "resources": "limits"}
	current := map[string]any{"image": "nginx:1.25"}

	changes := diff.Compute(baseline, current)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Kind != diff.Removed {
		t.Errorf("expected Removed, got %s", changes[0].Kind)
	}
}

func TestCompute_ResultsAreSorted(t *testing.T) {
	baseline := map[string]any{"z": 1, "a": 2, "m": 3}
	current := map[string]any{"z": 9, "a": 9, "m": 9}

	changes := diff.Compute(baseline, current)
	for i := 1; i < len(changes); i++ {
		if changes[i].Field < changes[i-1].Field {
			t.Errorf("results not sorted: %s before %s", changes[i-1].Field, changes[i].Field)
		}
	}
}

func TestChange_String(t *testing.T) {
	cases := []struct {
		c    diff.Change
		want string
	}{
		{diff.Change{Field: "replicas", Kind: diff.Added, NewValue: 3}, "replicas: added 3"},
		{diff.Change{Field: "image", Kind: diff.Removed, OldValue: "old"}, "image: removed (was old)"},
		{diff.Change{Field: "tag", Kind: diff.Modified, OldValue: "v1", NewValue: "v2"}, "tag: v1 -> v2"},
	}
	for _, tc := range cases {
		if got := tc.c.String(); got != tc.want {
			t.Errorf("String() = %q, want %q", got, tc.want)
		}
	}
}
