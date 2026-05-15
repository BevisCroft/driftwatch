package alerting_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/driftwatch/internal/alerting"
	"github.com/driftwatch/internal/drift"
)

func driftResult(service string, fields []drift.Difference) drift.Result {
	return drift.Result{
		Service:     service,
		HasDrift:    len(fields) > 0,
		Differences: fields,
	}
}

func TestNotify_NoDrift(t *testing.T) {
	var buf bytes.Buffer
	a := alerting.New(&buf, 3)
	results := []drift.Result{driftResult("svc-a", nil)}
	alerts := a.Notify(results)
	if len(alerts) != 0 {
		t.Fatalf("expected 0 alerts, got %d", len(alerts))
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output, got %q", buf.String())
	}
}

func TestNotify_WarnLevel(t *testing.T) {
	var buf bytes.Buffer
	a := alerting.New(&buf, 3)
	diffs := []drift.Difference{{Field: "spec.replicas", Expected: "3", Actual: "1"}}
	results := []drift.Result{driftResult("svc-b", diffs)}
	alerts := a.Notify(results)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].Level != alerting.LevelWarn {
		t.Errorf("expected WARN, got %s", alerts[0].Level)
	}
	if !strings.Contains(buf.String(), "svc-b") {
		t.Errorf("output missing service name: %q", buf.String())
	}
}

func TestNotify_ErrorLevel(t *testing.T) {
	var buf bytes.Buffer
	a := alerting.New(&buf, 2)
	diffs := []drift.Difference{
		{Field: "spec.replicas", Expected: "3", Actual: "1"},
		{Field: "spec.image", Expected: "v1", Actual: "v2"},
		{Field: "spec.env", Expected: "prod", Actual: "staging"},
	}
	results := []drift.Result{driftResult("svc-c", diffs)}
	alerts := a.Notify(results)
	if alerts[0].Level != alerting.LevelError {
		t.Errorf("expected ERROR, got %s", alerts[0].Level)
	}
}

func TestNotify_MultipleServices(t *testing.T) {
	var buf bytes.Buffer
	a := alerting.New(&buf, 3)
	results := []drift.Result{
		driftResult("svc-x", nil),
		driftResult("svc-y", []drift.Difference{{Field: "kind", Expected: "Deployment", Actual: "DaemonSet"}}),
	}
	alerts := a.Notify(results)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert for drifted service, got %d", len(alerts))
	}
	if alerts[0].Service != "svc-y" {
		t.Errorf("expected svc-y, got %s", alerts[0].Service)
	}
}

func TestNew_Defaults(t *testing.T) {
	a := alerting.New(nil, 0)
	if a == nil {
		t.Fatal("expected non-nil alerter")
	}
}
