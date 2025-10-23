package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type Domain struct {
	ID                   string  `json:"id"`
	ParentID             *string `json:"parent_id,omitempty"`
	Name                 string  `json:"name"`
	IsNotech             bool    `json:"is_notech"`
	RequiresInsideDomain bool    `json:"requires_inside_domain"`
	CreatedAt            string  `json:"created_at"`
}

// RegisterDomainRoutes attaches /api/domain/* handlers
func RegisterDomainRoutes(mux *http.ServeMux, db *sql.DB) {
	mux.HandleFunc("/api/domain/list", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		rows, err := db.Query(`SELECT id, parent_id, name, is_notech, requires_inside_domain, created_at FROM domains`)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()

		var list []Domain
		for rows.Next() {
			var d Domain
			err := rows.Scan(&d.ID, &d.ParentID, &d.Name, &d.IsNotech, &d.RequiresInsideDomain, &d.CreatedAt)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			list = append(list, d)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"count": len(list), "domains": list})
	})

	mux.HandleFunc("/api/domain/create", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var d Domain
		if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
			http.Error(w, "bad json", 400)
			return
		}
		_, err := db.Exec(`
			INSERT INTO domains (name, is_notech, requires_inside_domain)
			VALUES (?, ?, ?)`,
			d.Name, d.IsNotech, d.RequiresInsideDomain)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "id": d.ID})
	})
}
