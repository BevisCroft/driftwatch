package schema_test

import (
	"sync"
	"testing"

	"github.com/driftwatch/driftwatch/internal/schema"
)

// TestValidate_ConcurrentAccess ensures the validator is safe for concurrent use.
func TestValidate_ConcurrentAccess(t *testing.T) {
	v := schema.New(schema.WithForbiddenKeys("secret"))
	const workers = 20
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			m := baseManifest()
			res, err := v.Validate("svc", m)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !res.Valid() {
				t.Errorf("expected valid result, got %v", res.Violations)
			}
		}()
	}
	wg.Wait()
}

// TestValidate_MultipleViolations checks that all violations are collected in one pass.
func TestValidate_MultipleViolations(t *testing.T) {
	v := schema.New(
		schema.WithRequiredFields("kind", "name", "version"),
		schema.WithForbiddenKeys("debug", "trace"),
	)
	m := map[string]any{
		"kind": "Service",
		// name and version missing
		"spec": map[string]any{
			"debug": true,
			"trace": true,
		},
	}
	res, err := v.Validate("multi-svc", m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Expect: missing name, missing version, forbidden debug, forbidden trace = 4
	if len(res.Violations) != 4 {
		t.Errorf("expected 4 violations, got %d: %v", len(res.Violations), res.Violations)
	}
}
