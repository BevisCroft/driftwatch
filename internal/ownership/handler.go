package ownership

import (
	"encoding/json"
	"net/http"
)

// Handler returns an http.Handler that exposes ownership CRUD endpoints.
//
//	GET    /ownership          — list all entries
//	GET    /ownership/{service} — get entry for a service
//	POST   /ownership          — add or replace an entry (JSON body)
//	DELETE /ownership/{service} — remove an entry
func Handler(r *Registry) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ownership", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			listHandler(w, r)
		case http.MethodPost:
			addHandler(w, req, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/ownership/", func(w http.ResponseWriter, req *http.Request) {
		service := req.URL.Path[len("/ownership/"):]
		switch req.Method {
		case http.MethodGet:
			getHandler(w, r, service)
		case http.MethodDelete:
			deleteHandler(w, r, service)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	return mux
}

func listHandler(w http.ResponseWriter, r *Registry) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(r.All())
}

func addHandler(w http.ResponseWriter, req *http.Request, r *Registry) {
	var e Entry
	if err := json.NewDecoder(req.Body).Decode(&e); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if err := r.Set(e); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func getHandler(w http.ResponseWriter, r *Registry, service string) {
	e, ok := r.Get(service)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e)
}

func deleteHandler(w http.ResponseWriter, r *Registry, service string) {
	if !r.Remove(service) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
