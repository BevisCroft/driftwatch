package policy

import (
	"testing"

	"github.com/driftwatch/internal/drift"
)

func baseResult(service string, fields []string, sev drift.Severity) drift.Result {
	return drift.Result{Service: service, Fields: fields, Severity: sev}
}

func TestApply_NoRules_ReturnsUnchanged(t *testing.T) {
	e := New(nil)
	in := []drift.Result{baseResult("api", []string{"spec.replicas"}, drift.SeverityWarn)}
	out := e.Apply(in)
	if out[0].Severity != drift.SeverityWarn {
		t.Fatalf("expected warn, got %s", out[0].Severity)
	}
}

func TestApply_ExactServiceMatch(t *testing.T) {
	rules := []Rule{{ServiceGlob: "api", FieldGlob: "*", Severity: drift.SeverityError}}
	e := New(rules)
	out := e.Apply([]drift.Result{baseResult("api", []string{"spec.image"}, drift.SeverityWarn)})
	if out[0].Severity != drift.SeverityError {
		t.Fatalf("expected error, got %s", out[0].Severity)
	}
}

func TestApply_GlobServiceMatch(t *testing.T) {
	rules := []Rule{{ServiceGlob: "payments-*", FieldGlob: "*", Severity: drift.SeverityInfo}}
	e := New(rules)
	out := e.Apply([]drift.Result{baseResult("payments-eu", []string{"spec.replicas"}, drift.SeverityError)})
	if out[0].Severity != drift.SeverityInfo {
		t.Fatalf("expected info, got %s", out[0].Severity)
	}
}

func TestApply_FieldGlobNoMatch_Unchanged(t *testing.T) {
	rules := []Rule{{ServiceGlob: "*", FieldGlob: "spec.image", Severity: drift.SeverityError}}
	e := New(rules)
	out := e.Apply([]drift.Result{baseResult("api", []string{"spec.replicas"}, drift.SeverityWarn)})
	if out[0].Severity != drift.SeverityWarn {
		t.Fatalf("expected warn (no match), got %s", out[0].Severity)
	}
}

func TestApply_FirstRuleWins(t *testing.T) {
	rules := []Rule{
		{ServiceGlob: "*", FieldGlob: "*", Severity: drift.SeverityInfo},
		{ServiceGlob: "*", FieldGlob: "*", Severity: drift.SeverityError},
	}
	e := New(rules)
	out := e.Apply([]drift.Result{baseResult("svc", []string{"spec.image"}, drift.SeverityWarn)})
	if out[0].Severity != drift.SeverityInfo {
		t.Fatalf("expected info (first rule), got %s", out[0].Severity)
	}
}

func TestApply_MultipleResults_IndependentMatching(t *testing.T) {
	rules := []Rule{{ServiceGlob: "critical", FieldGlob: "*", Severity: drift.SeverityError}}
	e := New(rules)
	in := []drift.Result{
		baseResult("critical", []string{"spec.image"}, drift.SeverityWarn),
		baseResult("other", []string{"spec.image"}, drift.SeverityWarn),
	}
	out := e.Apply(in)
	if out[0].Severity != drift.SeverityError {
		t.Fatalf("result[0]: expected error, got %s", out[0].Severity)
	}
	if out[1].Severity != drift.SeverityWarn {
		t.Fatalf("result[1]: expected warn, got %s", out[1].Severity)
	}
}
