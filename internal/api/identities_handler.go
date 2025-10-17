package api

import (
	"encoding/json"
	"net/http"

	"dis-core/internal/db"

	"github.com/google/uuid"
)

// handleIdentities manages GET and POST requests for /identities.
func (s *Server) handleIdentities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		// --- GET: list all active identities ---
		idents, err := db.ListIdentities(s.store, 100, 0)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(idents)

	case http.MethodPost:
		// --- POST: create new identity ---
		var input struct {
			Namespace string `json:"namespace"`
		}
		_ = json.NewDecoder(r.Body).Decode(&input)
		if input.Namespace == "" {
			input.Namespace = "default"
		}

		uid := uuid.NewString()
		_, err := db.InsertIdentity(s.store, uid, input.Namespace)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := map[string]string{
			"status":    "created",
			"dis_uid":   uid,
			"namespace": input.Namespace,
		}
		json.NewEncoder(w).Encode(resp)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
