package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecordCycle_NoDrift(t *testing.T) {
	m := New()
	m.RecordCycle(false)

	s := m.Snapshot()
	if s.CyclesTotal != 1 {
		t.Fatalf("expected CyclesTotal=1, got %d", s.CyclesTotal)
	}
	if s.DriftsDetected != 0 {
		t.Fatalf("expected DriftsDetected=0, got %d", s.DriftsDetected)
	}
	if s.LastCycleDrift {
		t.Fatal("expected LastCycleDrift=false")
	}
}

func TestRecordCycle_WithDrift(t *testing.T) {
	m := New()
	m.RecordCycle(true)

	s := m.Snapshot()
	if s.CyclesTotal != 1 {
		t.Fatalf("expected CyclesTotal=1, got %d", s.CyclesTotal)
	}
	if s.DriftsDetected != 1 {
		t.Fatalf("expected DriftsDetected=1, got %d", s.DriftsDetected)
	}
	if !s.LastCycleDrift {
		t.Fatal("expected LastCycleDrift=true")
	}
}

func TestRecordAlert(t *testing.T) {
	m := New()
	m.RecordAlert()
	m.RecordAlert()

	if m.Snapshot().AlertsEmitted != 2 {
		t.Fatalf("expected AlertsEmitted=2, got %d", m.Snapshot().AlertsEmitted)
	}
}

func TestRecordError(t *testing.T) {
	m := New()
	m.RecordError()

	if m.Snapshot().Errors != 1 {
		t.Fatalf("expected Errors=1, got %d", m.Snapshot().Errors)
	}
}

func TestHandler_ContainsExpectedKeys(t *testing.T) {
	m := New()
	m.RecordCycle(true)
	m.RecordAlert()
	m.RecordError()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	m.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	body, _ := io.ReadAll(rr.Body)
	text := string(body)

	expected := []string{
		"driftwatch_cycles_total 1",
		"driftwatch_drifts_detected_total 1",
		"driftwatch_alerts_emitted_total 1",
		"driftwatch_errors_total 1",
		"driftwatch_last_cycle_drift 1",
	}
	for _, want := range expected {
		if !strings.Contains(text, want) {
			t.Errorf("response missing %q\nfull body:\n%s", want, text)
		}
	}
}

func TestHandler_LastCycleAt_Present(t *testing.T) {
	m := New()
	m.RecordCycle(false)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	m.Handler().ServeHTTP(rr, req)

	body, _ := io.ReadAll(rr.Body)
	if !strings.Contains(string(body), "# last_cycle_at") {
		t.Error("expected last_cycle_at comment in output")
	}
}
