package ttl

import (
	"encoding/json"
	"net/http"
)

// Handler returns an http.Handler exposing cache diagnostics and management
// endpoints under the given prefix.
//
//	GET  /ttl/stats          – returns live entry count and configured TTL
//	DELETE /ttl/entry?key=k  – evicts a single key immediately
func Handler(c *Cache) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ttl/stats", statsHandler(c))
	mux.HandleFunc("/ttl/entry", entryHandler(c))
	return mux
}

func statsHandler(c *Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		payload := map[string]interface{}{
			"live_entries": c.Len(),
			"ttl_seconds":  c.ttl.Seconds(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload) //nolint:errcheck
	}
}

func entryHandler(c *Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "missing key parameter", http.StatusBadRequest)
			return
		}
		c.Delete(key)
		w.WriteHeader(http.StatusNoContent)
	}
}
