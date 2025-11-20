package api

import (
	"encoding/json"
	"net/http"
)

// handleGetDefaultDomain returns the default startup domain (domain.void)
func (s *Server) handleGetDefaultDomain(w http.ResponseWriter, r *http.Request) {
	db := s.requireDB(w)
	if db == nil {
		return
	}
	ctx := r.Context()
	// Look up domain.void by name
	row := db.QueryRow(ctx, `SELECT id, parent_id, payload, created_at, updated_at FROM domains WHERE name = 'domain.void' LIMIT 1`)

	var d Domain
	if err := row.Scan(&d.ID, &d.ParentID, &d.Payload, &d.CreatedAt, &d.UpdatedAt); err != nil {
		// fallback: try domain.terra if void missing
		row2 := db.QueryRow(ctx, `SELECT id, parent_id, payload, created_at, updated_at FROM domains WHERE name = 'domain.terra' LIMIT 1`)
		if err2 := row2.Scan(&d.ID, &d.ParentID, &d.Payload, &d.CreatedAt, &d.UpdatedAt); err2 != nil {
			http.Error(w, "no default domain found", http.StatusNotFound)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id":         d.ID,
		"parent_id":  d.ParentID,
		"payload":    d.Payload, // Phase 10J.4: renamed from data
		"created_at": d.CreatedAt,
		"updated_at": d.UpdatedAt,
	})
}
