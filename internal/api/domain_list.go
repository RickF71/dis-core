package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

// DomainSummary represents lightweight domain info for API listings
type DomainSummary struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	DisplayName string     `json:"display_name,omitempty"`
	ParentID    *string    `json:"parent_id,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
}

// ------------------------------------------------------------
// GET /api/domains — Returns all known domains (JSON list)
// ------------------------------------------------------------
func (s *Server) handleDomainList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rows, err := s.DB().Query(ctx, `
		SELECT
			id::text AS id,
			name,
			parent_id::text AS parent_id,
			created_at
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
		if err := rows.Scan(&d.ID, &d.Name, &d.ParentID, &d.CreatedAt); err != nil {
			http.Error(w, "scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, d)
	}

	if err := rows.Err(); err != nil && err != pgx.ErrNoRows {
		http.Error(w, "iteration error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(list); err != nil {
		http.Error(w, "encode error: "+err.Error(), http.StatusInternalServerError)
	}
}
