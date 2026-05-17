// Package healthcheck provides a simple HTTP health endpoint for the
// driftwatch daemon, exposing liveness and last-cycle status.
package healthcheck

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Status holds the current health state of the daemon.
type Status struct {
	Healthy     bool      `json:"healthy"`
	LastCycleAt time.Time `json:"last_cycle_at,omitempty"`
	LastError   string    `json:"last_error,omitempty"`
	DriftCount  int       `json:"drift_count"`
}

// Server exposes an HTTP endpoint that reports daemon health.
type Server struct {
	mu     sync.RWMutex
	status Status
	addr   string
}

// New creates a new health-check Server listening on addr (e.g. ":9090").
func New(addr string) *Server {
	return &Server{
		addr:   addr,
		status: Status{Healthy: true},
	}
}

// SetCycleResult updates the health status after each drift-detection cycle.
func (s *Server) SetCycleResult(driftCount int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.status.LastCycleAt = time.Now().UTC()
	s.status.DriftCount = driftCount

	if err != nil {
		s.status.Healthy = false
		s.status.LastError = err.Error()
	} else {
		s.status.Healthy = true
		s.status.LastError = ""
	}
}

// GetStatus returns a snapshot of the current health status.
// It is safe to call concurrently.
func (s *Server) GetStatus() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

// ListenAndServe starts the HTTP server. It blocks until the server exits.
func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	return http.ListenAndServe(s.addr, mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	snap := s.GetStatus()

	w.Header().Set("Content-Type", "application/json")
	if !snap.Healthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	_ = json.NewEncoder(w).Encode(snap)
}
