package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"dis-core/internal/db"
)

// HandleIdentityRegister allows registering new identities via POST.
func HandleIdentityRegister(store *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var req struct {
				DISUID    string `json:"dis_uid"`
				Namespace string `json:"namespace"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			if req.DISUID == "" {
				http.Error(w, "Missing dis_uid", http.StatusBadRequest)
				return
			}

			id, err := db.InsertIdentity(store, req.DISUID, req.Namespace)
			if err != nil {
				http.Error(w, fmt.Sprintf("Insert failed: %v", err), http.StatusInternalServerError)
				return
			}

			resp := map[string]any{
				"status":     "created",
				"id":         id,
				"dis_uid":    req.DISUID,
				"namespace":  req.Namespace,
				"created_at": db.NowRFC3339Nano(),
			}
			writeJSON(w, http.StatusCreated, resp)

		default:
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		}
	}
}
