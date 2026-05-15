package healthcheck

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestServer() *Server {
	return New(":0")
}

func TestHealthz_DefaultHealthy(t *testing.T) {
	srv := newTestServer()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	srv.handleHealth(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var s Status
	if err := json.NewDecoder(rr.Body).Decode(&s); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if !s.Healthy {
		t.Error("expected healthy=true by default")
	}
}

func TestHealthz_AfterSuccessfulCycle(t *testing.T) {
	srv := newTestServer()
	srv.SetCycleResult(3, nil)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	srv.handleHealth(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var s Status
	_ = json.NewDecoder(rr.Body).Decode(&s)
	if s.DriftCount != 3 {
		t.Errorf("expected drift_count=3, got %d", s.DriftCount)
	}
	if s.LastError != "" {
		t.Errorf("expected no error, got %q", s.LastError)
	}
}

func TestHealthz_AfterFailedCycle(t *testing.T) {
	srv := newTestServer()
	srv.SetCycleResult(0, errors.New("manifest load failed"))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	srv.handleHealth(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}

	var s Status
	_ = json.NewDecoder(rr.Body).Decode(&s)
	if s.Healthy {
		t.Error("expected healthy=false after error")
	}
	if s.LastError == "" {
		t.Error("expected last_error to be populated")
	}
}

func TestHealthz_RecoverAfterError(t *testing.T) {
	srv := newTestServer()
	srv.SetCycleResult(0, errors.New("transient error"))
	srv.SetCycleResult(1, nil) // next cycle succeeds

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	srv.handleHealth(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 after recovery, got %d", rr.Code)
	}
}
