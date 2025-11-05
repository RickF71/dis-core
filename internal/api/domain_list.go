package api

import (
	"encoding/json"
	"net/http"
)

// DomainSummary represents a minimal domain record for selection lists.
type DomainSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ------------------------------------------------------------
//  /api/domain/list — Returns all known domains
// ------------------------------------------------------------

func (s *Server) handleDomainList(w http.ResponseWriter, r *http.Request) {
	rows, err := s.Ledger.DB.Query(`
		SELECT
			content->'meta'->>'domain_id' AS id,
			content->'meta'->>'name' AS name
		FROM canon
		WHERE type = 'domain'
		ORDER BY name
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []DomainSummary
	for rows.Next() {
		var d DomainSummary
		if err := rows.Scan(&d.ID, &d.Name); err == nil {
			if d.Name == "" {
				d.Name = d.ID
			}
			list = append(list, d)
		}
	}

	// ------------------------------------------------------------------
	//  Bootstrap fallback — add core domains if the database is sparse
	// ------------------------------------------------------------------
	if len(list) <= 1 {
		bootstrap := []DomainSummary{
			{ID: "domain.null", Name: "Root Null Domain"},
			{ID: "domain.terra", Name: "Terra"},
			{ID: "domain.usa", Name: "United States"},
		}
		for _, b := range bootstrap {
			// only append if not already in list
			exists := false
			for _, d := range list {
				if d.ID == b.ID {
					exists = true
					break
				}
			}
			if !exists {
				list = append(list, b)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}
