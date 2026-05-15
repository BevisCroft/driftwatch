// Package metrics provides Prometheus-compatible metrics collection
// for driftwatch daemon cycles, drift detections, and alerting events.
package metrics

import (
	"net/http"
	"sync"
	"time"
)

// Metrics holds counters and gauges collected during daemon operation.
type Metrics struct {
	mu sync.RWMutex

	CyclesTotal    int64
	DriftsDetected int64
	AlertsEmitted  int64
	LastCycleAt    time.Time
	LastCycleDrift bool
	Errors         int64
}

// New returns an initialised Metrics instance.
func New() *Metrics {
	return &Metrics{}
}

// RecordCycle records a completed poll cycle.
// driftFound indicates whether any drift was detected in this cycle.
func (m *Metrics) RecordCycle(driftFound bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CyclesTotal++
	m.LastCycleAt = time.Now().UTC()
	m.LastCycleDrift = driftFound
	if driftFound {
		m.DriftsDetected++
	}
}

// RecordAlert increments the alert counter.
func (m *Metrics) RecordAlert() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.AlertsEmitted++
}

// RecordError increments the error counter.
func (m *Metrics) RecordError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Errors++
}

// Snapshot returns a point-in-time copy of current metrics.
func (m *Metrics) Snapshot() Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return Metrics{
		CyclesTotal:    m.CyclesTotal,
		DriftsDetected: m.DriftsDetected,
		AlertsEmitted:  m.AlertsEmitted,
		LastCycleAt:    m.LastCycleAt,
		LastCycleDrift: m.LastCycleDrift,
		Errors:         m.Errors,
	}
}

// Handler returns an http.Handler that exposes metrics in plain text.
func (m *Metrics) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		s := m.Snapshot()
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		lastCycle := ""
		if !s.LastCycleAt.IsZero() {
			lastCycle = s.LastCycleAt.Format(time.RFC3339)
		}
		driftFlag := 0
		if s.LastCycleDrift {
			driftFlag = 1
		}
		_, _ = w.Write([]byte("# driftwatch metrics\n"))
		_, _ = w.Write([]byte(formatLine("driftwatch_cycles_total", s.CyclesTotal)))
		_, _ = w.Write([]byte(formatLine("driftwatch_drifts_detected_total", s.DriftsDetected)))
		_, _ = w.Write([]byte(formatLine("driftwatch_alerts_emitted_total", s.AlertsEmitted)))
		_, _ = w.Write([]byte(formatLine("driftwatch_errors_total", s.Errors)))
		_, _ = w.Write([]byte(formatLine("driftwatch_last_cycle_drift", int64(driftFlag))))
		if lastCycle != "" {
			_, _ = w.Write([]byte("# last_cycle_at " + lastCycle + "\n"))
		}
	})
}

func formatLine(name string, val int64) string {
	return name + " " + itoa(val) + "\n"
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := make([]byte, 0, 20)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
