package api

import (
	"encoding/json"
	"net/http"
)

// GET /api/domain/links
func (s *Server) handleDomainLinks(w http.ResponseWriter, r *http.Request) {
	rows, err := s.Ledger.DB.Query(`
		SELECT content->'meta'->>'domain_id' AS child,
		       content->'meta'->>'parent' AS parent
		FROM canon
		WHERE type='domain' AND content->'meta' ? 'parent'
	`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var links []map[string]string
	for rows.Next() {
		var child, parent string
		rows.Scan(&child, &parent)
		if parent != "" {
			links = append(links, map[string]string{"source": child, "target": parent})
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(links)
}
