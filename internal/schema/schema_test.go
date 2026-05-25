package schema_test

import (
	"testing"

	"github.com/driftwatch/driftwatch/internal/schema"
)

func baseManifest() map[string]any {
	return map[string]any{
		"kind": "Deployment",
		"name": "my-service",
		"spec": map[string]any{"replicas": 3},
	}
}

func TestValidate_ValidManifest(t *testing.T) {
	v := schema.New()
	res, err := v.Validate("svc", baseManifest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Valid() {
		t.Fatalf("expected valid, got violations: %v", res.Violations)
	}
}

func TestValidate_MissingRequiredField(t *testing.T) {
	v := schema.New()
	m := baseManifest()
	delete(m, "kind")
	res, err := v.Validate("svc", m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Valid() {
		t.Fatal("expected violation for missing kind")
	}
	if res.Violations[0].Field != "kind" {
		t.Errorf("expected field=kind, got %q", res.Violations[0].Field)
	}
}

func TestValidate_ForbiddenKey(t *testing.T) {
	v := schema.New(schema.WithForbiddenKeys("debug"))
	m := baseManifest()
	m["spec"].(map[string]any)["debug"] = true
	res, err := v.Validate("svc", m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Valid() {
		t.Fatal("expected violation for forbidden key")
	}
	if res.Violations[0].Field != "spec.debug" {
		t.Errorf("unexpected field: %q", res.Violations[0].Field)
	}
}

func TestValidate_EmptyServiceReturnsError(t *testing.T) {
	v := schema.New()
	_, err := v.Validate("", baseManifest())
	if err == nil {
		t.Fatal("expected error for empty service name")
	}
}

func TestValidate_CustomRequiredFields(t *testing.T) {
	v := schema.New(schema.WithRequiredFields("kind", "name", "owner"))
	res, err := v.Validate("svc", baseManifest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Valid() {
		t.Fatal("expected violation for missing owner field")
	}
	if res.Violations[0].Field != "owner" {
		t.Errorf("expected field=owner, got %q", res.Violations[0].Field)
	}
}

func TestValidate_BlankKind(t *testing.T) {
	v := schema.New()
	m := baseManifest()
	m["kind"] = "   "
	res, err := v.Validate("svc", m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// blank kind should produce a violation
	var found bool
	for _, viol := range res.Violations {
		if viol.Field == "kind" {
			found = true
		}
	}
	if !found {
		t.Error("expected violation for blank kind")
	}
}
