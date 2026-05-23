package diff_test

import (
	"strings"
	"testing"

	"github.com/example/driftwatch/internal/diff"
)

func TestRenderText_NoDrift(t *testing.T) {
	var sb strings.Builder
	diff.RenderText(&sb, "svc-a", nil)
	got := sb.String()
	if !strings.Contains(got, "no drift detected") {
		t.Errorf("expected no-drift message, got: %s", got)
	}
}

func TestRenderText_WithChanges(t *testing.T) {
	changes := []diff.Change{
		{Field: "replicas", Kind: diff.Modified, OldValue: 2, NewValue: 4},
	}
	var sb strings.Builder
	diff.RenderText(&sb, "svc-b", changes)
	got := sb.String()

	if !strings.Contains(got, "svc-b") {
		t.Error("expected service name in output")
	}
	if !strings.Contains(got, "MODIFIED") {
		t.Error("expected MODIFIED label in output")
	}
	if !strings.Contains(got, "replicas") {
		t.Error("expected field name in output")
	}
}

func TestRenderMarkdown_NoDrift(t *testing.T) {
	var sb strings.Builder
	diff.RenderMarkdown(&sb, "svc-c", nil)
	got := sb.String()
	if !strings.Contains(got, "no drift detected") {
		t.Errorf("expected no-drift message, got: %s", got)
	}
}

func TestRenderMarkdown_WithChanges(t *testing.T) {
	changes := []diff.Change{
		{Field: "image", Kind: diff.Added, NewValue: "nginx:1.26"},
		{Field: "tag", Kind: diff.Removed, OldValue: "v1"},
	}
	var sb strings.Builder
	diff.RenderMarkdown(&sb, "svc-d", changes)
	got := sb.String()

	if !strings.Contains(got, "svc-d") {
		t.Error("expected service name")
	}
	if !strings.Contains(got, "| Field |") {
		t.Error("expected markdown table header")
	}
	if !strings.Contains(got, "image") {
		t.Error("expected field 'image' in output")
	}
	if !strings.Contains(got, "tag") {
		t.Error("expected field 'tag' in output")
	}
}
