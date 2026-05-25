package schema

import (
	"encoding/json"
	"net/http"
)

// Handler returns an http.Handler exposing schema validation over HTTP.
//
// POST /validate  — accepts a JSON body {"service":"...","manifest":{...}}
//                   and returns {"valid":bool,"violations":[...]}
func Handler(v *Validator) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", validateHandler(v))
	return mux
}

type validateRequest struct {
	Service  string         `json:"service"`
	Manifest map[string]any `json:"manifest"`
}

type validateResponse struct {
	Valid      bool       `json:"valid"`
	Violations []violation `json:"violations"`
}

type violation struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func validateHandler(v *Validator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req validateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		res, err := v.Validate(req.Service, req.Manifest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp := validateResponse{Valid: res.Valid()}
		for _, viol := range res.Violations {
			resp.Violations = append(resp.Violations, violation{Field: viol.Field, Message: viol.Message})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}
