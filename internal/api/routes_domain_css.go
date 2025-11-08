// routes_domain_css.go
package api

import (
	"database/sql"
	"io"
	"net/http"
)

// GET /api/domain/{id}/css
func (s *Server) handleDomainCSS(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var css sql.NullString
	err := s.db.QueryRow(`SELECT data->>'css' FROM domains WHERE id = $1`, id).Scan(&css)
	if err == sql.ErrNoRows {
		http.Error(w, "domain not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	if css.Valid && css.String != "" {
		io.WriteString(w, css.String)
	} else {
		// Default fallback theme
		io.WriteString(w, `body { background-color: #0f172a; color: #f1f5f9; }`)
	}
}
