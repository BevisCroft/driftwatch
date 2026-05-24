package sampling

import (
	"encoding/json"
	"net/http"
)

type statusResponse struct {
	Rate     float64  `json:"rate"`
	Strategy Strategy `json:"strategy"`
}

// Handler returns an http.Handler exposing sampler status and runtime
// rate updates via a simple JSON API.
//
//	GET  /sampling        → current rate and strategy
//	POST /sampling/reset  → reset all per-service counters
func Handler(s *Sampler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/sampling", func(w http.ResponseWriter, r *http.Request) {
		statusHandler(s, w, r)
	})
	mux.HandleFunc("/sampling/reset", func(w http.ResponseWriter, r *http.Request) {
		resetHandler(s, w, r)
	})
	return mux
}

func statusHandler(s *Sampler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(statusResponse{
		Rate:     s.Rate(),
		Strategy: s.cfg.Strategy,
	})
}

func resetHandler(s *Sampler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	service := r.URL.Query().Get("service")
	if service == "" {
		http.Error(w, "service query parameter required", http.StatusBadRequest)
		return
	}
	s.Reset(service)
	w.WriteHeader(http.StatusNoContent)
}
