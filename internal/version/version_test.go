package version_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"

	"driftwatch/internal/version"
)

func TestGet_ReturnsRuntimeFields(t *testing.T) {
	info := version.Get()

	if info.GoVersion != runtime.Version() {
		t.Errorf("GoVersion = %q; want %q", info.GoVersion, runtime.Version())
	}
	if info.OS != runtime.GOOS {
		t.Errorf("OS = %q; want %q", info.OS, runtime.GOOS)
	}
	if info.Arch != runtime.GOARCH {
		t.Errorf("Arch = %q; want %q", info.Arch, runtime.GOARCH)
	}
}

func TestGet_DefaultsWhenNotInjected(t *testing.T) {
	info := version.Get()

	if info.Version == "" {
		t.Error("Version should not be empty")
	}
	if info.Commit == "" {
		t.Error("Commit should not be empty")
	}
	if info.BuildDate == "" {
		t.Error("BuildDate should not be empty")
	}
}

func TestHandler_ReturnsJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/version", nil)

	version.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q; want application/json", ct)
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if _, ok := payload["info"]; !ok {
		t.Error("response missing 'info' field")
	}
	if _, ok := payload["timestamp"]; !ok {
		t.Error("response missing 'timestamp' field")
	}
}

func TestHandler_SetsVersionHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/version", nil)

	version.Handler().ServeHTTP(rec, req)

	if rec.Header().Get("X-Driftwatch-Version") == "" {
		t.Error("X-Driftwatch-Version header not set")
	}
	if rec.Header().Get("X-Driftwatch-Commit") == "" {
		t.Error("X-Driftwatch-Commit header not set")
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodDelete, http.MethodPut} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(method, "/version", nil)

		version.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: status = %d; want 405", method, rec.Code)
		}
	}
}
