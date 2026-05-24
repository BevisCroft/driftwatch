package checkpoint

import (
	"encoding/json"
	"net/http"
)

// Handler returns an http.Handler that exposes the current checkpoint over HTTP.
//
//	GET  /checkpoint       — returns the latest checkpoint as JSON (204 if none)
//	DELETE /checkpoint     — removes the checkpoint file
func Handler(s *Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/checkpoint", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getHandler(s, w, r)
		case http.MethodDelete:
			deleteHandler(s, w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	return mux
}

func getHandler(s *Store, w http.ResponseWriter, _ *http.Request) {
	entry, err := s.Load()
	if err != nil {
		http.Error(w, "failed to load checkpoint: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if entry == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(entry)
}

func deleteHandler(s *Store, w http.ResponseWriter, _ *http.Request) {
	if err := s.Delete(); err != nil {
		http.Error(w, "failed to delete checkpoint: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
