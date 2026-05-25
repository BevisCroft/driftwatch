package watchlist

import (
	"encoding/json"
	"net/http"
)

// Handler returns an http.Handler exposing watchlist management endpoints.
//
//	GET  /watchlist        → list all entries
//	POST /watchlist        → add an entry (JSON body: Entry)
//	DELETE /watchlist/{id} → remove an entry by service name
func Handler(wl *Watchlist) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/watchlist", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listHandler(wl, w, r)
		case http.MethodPost:
			addHandler(wl, w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/watchlist/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		service := r.URL.Path[len("/watchlist/"):]
		if service == "" {
			http.Error(w, "service name required", http.StatusBadRequest)
			return
		}
		if !wl.Remove(service) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	return mux
}

func listHandler(wl *Watchlist, w http.ResponseWriter, _ *http.Request) {
	entries := wl.All()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(entries)
}

func addHandler(wl *Watchlist, w http.ResponseWriter, r *http.Request) {
	var e Entry
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := wl.Add(e); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
