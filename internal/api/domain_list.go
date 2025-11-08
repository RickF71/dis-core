package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// DomainSummary represents a minimal domain record for selection lists.
type DomainSummary struct {
	ID   string `json:"id"`   // UUID
	Name string `json:"name"` // e.g. "domain.terra"
}

// ------------------------------------------------------------
//
//	GET /api/domains — Returns all known domains (JSONB model)
//
// ------------------------------------------------------------
func (s *Server) handleDomainList(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB().Query(`
		SELECT
			id::text AS id,
			name
		FROM domains
		ORDER BY name
	`)
	if err != nil {
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	list := []DomainSummary{}
	for rows.Next() {
		var d DomainSummary
		if err := rows.Scan(&d.ID, &d.Name); err != nil {
			http.Error(w, "scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, d)
	}
	if err := rows.Err(); err != nil && err != sql.ErrNoRows {
		http.Error(w, "iteration error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(list); err != nil {
		http.Error(w, "encode error: "+err.Error(), http.StatusInternalServerError)
	}
}
