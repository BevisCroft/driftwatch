package alerting_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/driftwatch/internal/alerting"
	"github.com/driftwatch/internal/drift"
)

// TestNotify_FullCycle exercises the alerter with a realistic multi-service
// scenario to verify correct level assignment and output formatting end-to-end.
func TestNotify_FullCycle(t *testing.T) {
	var buf bytes.Buffer
	const threshold = 2
	a := alerting.New(&buf, threshold)

	results := []drift.Result{
		{
			Service:  "api-gateway",
			HasDrift: false,
		},
		{
			Service:  "auth-service",
			HasDrift: true,
			Differences: []drift.Difference{
				{Field: "spec.replicas", Expected: "2", Actual: "1"},
			},
		},
		{
			Service:  "payment-service",
			HasDrift: true,
			Differences: []drift.Difference{
				{Field: "spec.image", Expected: "pay:v3", Actual: "pay:v1"},
				{Field: "spec.replicas", Expected: "5", Actual: "2"},
				{Field: "spec.env", Expected: "production", Actual: "staging"},
			},
		},
	}

	alerts := a.Notify(results)

	if len(alerts) != 2 {
		t.Fatalf("expected 2 alerts, got %d", len(alerts))
	}

	// auth-service has 1 diff < threshold=2, expect WARN
	if alerts[0].Level != alerting.LevelWarn {
		t.Errorf("auth-service: expected WARN, got %s", alerts[0].Level)
	}

	// payment-service has 3 diffs >= threshold=2, expect ERROR
	if alerts[1].Level != alerting.LevelError {
		t.Errorf("payment-service: expected ERROR, got %s", alerts[1].Level)
	}

	out := buf.String()
	for _, svc := range []string{"auth-service", "payment-service"} {
		if !strings.Contains(out, svc) {
			t.Errorf("output missing service %q", svc)
		}
	}
	if strings.Contains(out, "api-gateway") {
		t.Errorf("output should not mention api-gateway (no drift)")
	}

	for _, al := range alerts {
		if al.Timestamp.IsZero() {
			t.Errorf("alert for %s has zero timestamp", al.Service)
		}
	}
}
