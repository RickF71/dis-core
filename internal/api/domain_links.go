package api

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5"
)

// ------------------------------------------------------------
//
//	GET /api/domain/links — returns parent→child relationships
//	using the new JSONB domains table
//
// ------------------------------------------------------------
func (s *Server) handleDomainLinks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Ensure DB is available
	db := s.requireDB(w)
	if db == nil {
		return
	}

	rows, err := db.Query(ctx, `
		SELECT
			payload->>'domain_id' AS child,
			payload->>'parent_name' AS parent
		FROM domains
		WHERE (payload->>'parent_name') IS NOT NULL
		      AND (payload->>'parent_name') <> ''
	`)
	if err != nil {
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Link struct {
		Source string `json:"source"`
		Target string `json:"target"`
	}

	links := []Link{}
	for rows.Next() {
		var child, parent string
		if err := rows.Scan(&child, &parent); err != nil {
			http.Error(w, "scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if child != "" && parent != "" {
			links = append(links, Link{Source: child, Target: parent})
		}
	}
	if err := rows.Err(); err != nil && err != pgx.ErrNoRows {
		http.Error(w, "iteration error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(links); err != nil {
		http.Error(w, "encode error: "+err.Error(), http.StatusInternalServerError)
	}
}
