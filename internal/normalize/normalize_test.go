package normalize

import (
	"testing"
)

func TestApply_NoMatchingFields_ReturnsUnchanged(t *testing.T) {
	n := New()
	spec := map[string]interface{}{
		"image": "nginx:latest",
		"port":  float64(8080),
	}
	out := n.Apply(spec)
	if out["image"] != "nginx:latest" {
		t.Errorf("unexpected image: %v", out["image"])
	}
	if out["port"] != float64(8080) {
		t.Errorf("unexpected port: %v", out["port"])
	}
}

func TestApply_ReplicasZero_DefaultsToOne(t *testing.T) {
	n := New()
	spec := map[string]interface{}{"replicas": float64(0)}
	out := n.Apply(spec)
	if out["replicas"] != float64(1) {
		t.Errorf("expected replicas=1, got %v", out["replicas"])
	}
}

func TestApply_ReplicasNonZero_Unchanged(t *testing.T) {
	n := New()
	spec := map[string]interface{}{"replicas": float64(3)}
	out := n.Apply(spec)
	if out["replicas"] != float64(3) {
		t.Errorf("expected replicas=3, got %v", out["replicas"])
	}
}

func TestApply_TagsSorted_AndLowercased(t *testing.T) {
	n := New()
	spec := map[string]interface{}{
		"tags": []interface{}{"Zebra", "apple", " Mango "},
	}
	out := n.Apply(spec)
	tags, ok := out["tags"].([]interface{})
	if !ok {
		t.Fatal("tags should be []interface{}")
	}
	expected := []string{"apple", "mango", "zebra"}
	for i, want := range expected {
		if tags[i] != want {
			t.Errorf("tags[%d]: want %q, got %q", i, want, tags[i])
		}
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	n := New()
	original := map[string]interface{}{"replicas": float64(0)}
	_ = n.Apply(original)
	if original["replicas"] != float64(0) {
		t.Error("Apply must not mutate the original spec")
	}
}

func TestAddRule_CustomTransformApplied(t *testing.T) {
	n := New()
	n.AddRule(Rule{
		Field: "image",
		Transform: func(v interface{}) interface{} {
			s, ok := v.(string)
			if !ok {
				return v
			}
			if s == "" {
				return "scratch"
			}
			return s
		},
	})
	spec := map[string]interface{}{"image": ""}
	out := n.Apply(spec)
	if out["image"] != "scratch" {
		t.Errorf("expected image=scratch, got %v", out["image"])
	}
}

func TestApply_TagsNonSlice_PassedThrough(t *testing.T) {
	n := New()
	spec := map[string]interface{}{"tags": "not-a-slice"}
	out := n.Apply(spec)
	if out["tags"] != "not-a-slice" {
		t.Errorf("non-slice tags should pass through unchanged, got %v", out["tags"])
	}
}
