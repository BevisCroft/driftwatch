package metrics_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/yourusername/driftwatch/internal/metrics"
)

// TestMetrics_ConcurrentRecording verifies that concurrent writes do not
// cause data races (run with -race).
func TestMetrics_ConcurrentRecording(t *testing.T) {
	m := metrics.New()

	const workers = 20
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()
			m.RecordCycle(i%2 == 0)
			m.RecordAlert()
			m.RecordError()
		}(i)
	}
	wg.Wait()

	s := m.Snapshot()
	if s.CyclesTotal != workers {
		t.Fatalf("expected %d cycles, got %d", workers, s.CyclesTotal)
	}
	if s.AlertsEmitted != workers {
		t.Fatalf("expected %d alerts, got %d", workers, s.AlertsEmitted)
	}
	if s.Errors != workers {
		t.Fatalf("expected %d errors, got %d", workers, s.Errors)
	}
}

// TestMetrics_HTTPServer verifies the handler works when mounted on a real
// test HTTP server.
func TestMetrics_HTTPServer(t *testing.T) {
	m := metrics.New()
	m.RecordCycle(false)

	srv := httptest.NewServer(m.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "driftwatch_cycles_total 1") {
		t.Error("expected cycle counter in HTTP response")
	}
}
