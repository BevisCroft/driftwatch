package routing

import (
	"encoding/json"
	"net/http"
)

// Handler returns an http.Handler that exposes read-only routing information.
func Handler(r *Router) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/routing/endpoints", listHandler(r))
	mux.HandleFunc("/routing/next", nextHandler(r))
	return mux
}

func listHandler(r *Router) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"endpoints": r.All(),
		})
	}
}

func nextHandler(r *Router) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		ep := r.Next()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ep) //nolint:errcheck
	}
}
