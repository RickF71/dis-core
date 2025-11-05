package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"dis-core/internal/model"
)

// ------------------------------------------------------------
//
//	/api/domain/dis/:id — Return full domain JSON + CSS
//
// ------------------------------------------------------------
func (s *Server) handleDomainGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	row := s.Ledger.DB.QueryRow(`
		SELECT content
		FROM canon
		WHERE type = 'domain'
		  AND content->'meta'->>'domain_id' = $1
	`, id)

	var raw []byte
	if err := row.Scan(&raw); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "domain not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	var d model.Domain
	if err := json.Unmarshal(raw, &d); err != nil {
		http.Error(w, "invalid domain json", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(d)
}
