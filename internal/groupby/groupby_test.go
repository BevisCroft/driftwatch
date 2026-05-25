package groupby_test

import (
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/groupby"
)

func makeResult(service, severity, namespace string) drift.Result {
	return drift.Result{
		Service:  service,
		Severity: drift.Severity(severity),
		Drifted:  true,
		Live:     map[string]any{"namespace": namespace},
	}
}

func TestBy_UnknownKey_ReturnsError(t *testing.T) {
	g := groupby.New()
	_, err := g.By("unknown", nil)
	if err == nil {
		t.Fatal("expected error for unknown key, got nil")
	}
}

func TestBy_Service_GroupsCorrectly(t *testing.T) {
	results := []drift.Result{
		makeResult("api", "warn", "prod"),
		makeResult("api", "error", "prod"),
		makeResult("worker", "warn", "staging"),
	}
	g := groupby.New()
	groups, err := g.By(groupby.KeyService, results)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if groups[0].Label != "api" || len(groups[0].Results) != 2 {
		t.Errorf("unexpected first group: %+v", groups[0])
	}
	if groups[1].Label != "worker" || len(groups[1].Results) != 1 {
		t.Errorf("unexpected second group: %+v", groups[1])
	}
}

func TestBy_Severity_GroupsCorrectly(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", "warn", "prod"),
		makeResult("svc-b", "error", "prod"),
		makeResult("svc-c", "warn", "prod"),
	}
	g := groupby.New()
	groups, err := g.By(groupby.KeySeverity, results)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
}

func TestBy_Namespace_FallsBackToDefault(t *testing.T) {
	result := drift.Result{Service: "svc", Severity: "warn", Drifted: true, Live: map[string]any{}}
	g := groupby.New()
	groups, err := g.By(groupby.KeyNamespace, []drift.Result{result})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 1 || groups[0].Label != "default" {
		t.Errorf("expected label 'default', got %q", groups[0].Label)
	}
}

func TestBy_CustomExtractor_Works(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", "warn", "prod"),
		makeResult("svc-b", "warn", "staging"),
	}
	g := groupby.New()
	g.Register("env", func(r drift.Result) string {
		if ns, ok := r.Live["namespace"]; ok && ns == "prod" {
			return "production"
		}
		return "non-production"
	})
	groups, err := g.By("env", results)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
}

func TestBy_EmptyResults_ReturnsEmptyGroups(t *testing.T) {
	g := groupby.New()
	groups, err := g.By(groupby.KeyService, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 0 {
		t.Errorf("expected 0 groups, got %d", len(groups))
	}
}
