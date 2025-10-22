package api

import (
	"encoding/json"
	"net/http"
	"os"

	"dis-core/internal/db"

	"github.com/google/uuid"
)

// handleIdentities manages GET and POST requests for /identities and /api/identity/list.
func (s *Server) handleIdentities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {

	// ------------------------------------------------------------------------
	// GET: list all active identities
	// ------------------------------------------------------------------------
	case http.MethodGet:
		dataDir := resolveDataDir() // environment-driven path
		idents, err := db.ListIdentities(s.db, 100, 0)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{
				"error":   "failed to list identities",
				"details": err.Error(),
				"dataDir": dataDir,
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"status":     "ok",
			"count":      len(idents),
			"dataDir":    dataDir,
			"identities": idents,
		})
		return

	// ------------------------------------------------------------------------
	// POST: create new identity
	// ------------------------------------------------------------------------
	case http.MethodPost:
		var input struct {
			Namespace string `json:"namespace"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
			return
		}
		if input.Namespace == "" {
			input.Namespace = "default"
		}

		uid := uuid.NewString()
		if _, err := db.InsertIdentity(s.db, uid, input.Namespace); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error":   "failed to insert identity",
				"details": err.Error(),
			})
			return
		}

		// Optional: log to stdout for confirmation
		repoRoot := os.Getenv("DIS_REPO_ROOT")
		writeJSON(w, http.StatusCreated, map[string]string{
			"status":    "created",
			"dis_uid":   uid,
			"namespace": input.Namespace,
			"repo_root": repoRoot,
		})
		return

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}
