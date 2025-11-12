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
	ctx := r.Context()

	var css sql.NullString
	// Phase 10J.4: Fetch from flattened payload->css->content
	err := s.DB().QueryRow(ctx,
		`SELECT payload->'css'->>'content' FROM domains WHERE id = $1`, id,
	).Scan(&css)

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
		// CSS is stored as raw text, no need to remove quotes
		io.WriteString(w, css.String)
	} else {
		io.WriteString(w, `body { background-color: #0f172a; color: #f1f5f9; }`)
	}
}
