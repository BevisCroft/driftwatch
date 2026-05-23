package throttle

import (
	"encoding/json"
	"net/http"
	"time"
)

// Handler returns an http.Handler that exposes throttle management endpoints.
//
//	GET  /throttle/status          — returns current window and burst config
//	DELETE /throttle/reset?service= — resets throttle state for a service
//	POST /throttle/purge            — removes all expired records
func Handler(th *Throttler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/throttle/status", statusHandler(th))
	mux.HandleFunc("/throttle/reset", resetHandler(th))
	mux.HandleFunc("/throttle/purge", purgeHandler(th))
	return mux
}

func statusHandler(th *Throttler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		payload := map[string]any{
			"window_ms": th.window / time.Millisecond,
			"max_burst": th.maxBurst,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload) //nolint:errcheck
	}
}

func resetHandler(th *Throttler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		svc := r.URL.Query().Get("service")
		if svc == "" {
			http.Error(w, "service query parameter required", http.StatusBadRequest)
			return
		}
		th.Reset(svc)
		w.WriteHeader(http.StatusNoContent)
	}
}

func purgeHandler(th *Throttler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		th.Purge()
		w.WriteHeader(http.StatusNoContent)
	}
}
