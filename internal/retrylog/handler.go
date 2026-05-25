package retrylog

import (
	"encoding/json"
	"net/http"
)

// Handler returns an http.Handler exposing retry log endpoints.
//
//	GET  /retrylog          - list summaries for all services
//	DELETE /retrylog?service=<name> - reset a service's retry history
func Handler(l *Log) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/retrylog", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listHandler(l, w, r)
		case http.MethodDelete:
			resetHandler(l, w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	return mux
}

func listHandler(l *Log, w http.ResponseWriter, _ *http.Request) {
	summaries := l.Summaries()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(summaries); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func resetHandler(l *Log, w http.ResponseWriter, r *http.Request) {
	svc := r.URL.Query().Get("service")
	if svc == "" {
		http.Error(w, "missing service query parameter", http.StatusBadRequest)
		return
	}
	l.Reset(svc)
	w.WriteHeader(http.StatusNoContent)
}
