package suppress

import (
	"encoding/json"
	"net/http"
	"time"
)

// addRequest is the JSON body accepted by the add endpoint.
type addRequest struct {
	Service   string `json:"service"`
	Field     string `json:"field"`
	Reason    string `json:"reason"`
	ExpiresIn string `json:"expires_in"` // e.g. "2h", "30m"
}

// Handler returns an http.ServeMux pre-registered with suppression endpoints.
//
//	GET  /suppressions       — list active entries
//	POST /suppressions       — add a new entry
//	DELETE /suppressions     — purge expired entries
func Handler(l *List) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/suppressions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listHandler(l, w)
		case http.MethodPost:
			addHandler(l, w, r)
		case http.MethodDelete:
			purgeHandler(l, w)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return mux
}

func listHandler(l *List, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(l.Snapshot())
}

func addHandler(l *List, w http.ResponseWriter, r *http.Request) {
	var req addRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Service == "" || req.Field == "" {
		http.Error(w, "service and field are required", http.StatusBadRequest)
		return
	}
	e := Entry{Service: req.Service, Field: req.Field, Reason: req.Reason}
	if req.ExpiresIn != "" {
		d, err := time.ParseDuration(req.ExpiresIn)
		if err != nil {
			http.Error(w, "invalid expires_in duration", http.StatusBadRequest)
			return
		}
		e.ExpiresAt = time.Now().Add(d)
	}
	l.Add(e)
	w.WriteHeader(http.StatusCreated)
}

func purgeHandler(l *List, w http.ResponseWriter) {
	l.Purge()
	w.WriteHeader(http.StatusNoContent)
}
