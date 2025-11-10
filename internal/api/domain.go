package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
)

// DomainResponse represents the GET response struct served to Finagler
type DomainResponse struct {
	DomainID string `json:"domain_id"`
	CSS      string `json:"css"`
	JSX      string `json:"jsx,omitempty"`
}

// ---- Domain ID Resolver (minimal tactical version) ----
// Accepts either:
//   - internal DB id (uuid/int) OR
//   - DIS domain longname (domain.user.rick)
//
// Returns the INTERNAL ID and matching domain record data.
func (s *Server) resolveDomainInternalID(ctx context.Context, ref string) (string, error) {
	var id string

	// Try DB internal ID first
	err := s.DB().QueryRow(ctx, `SELECT id FROM domains WHERE id=$1`, ref).Scan(&id)
	if err == nil {
		return id, nil
	}

	// Try DIS longname style (domain_id)
	err = s.DB().QueryRow(ctx, `SELECT id FROM domains WHERE domain_id=$1`, ref).Scan(&id)
	if err == nil {
		return id, nil
	}

	// Try the "name" column as last fallback
	err = s.DB().QueryRow(ctx, `SELECT id FROM domains WHERE name=$1`, ref).Scan(&id)
	if err == nil {
		return id, nil
	}

	return "", fmt.Errorf("domain not found: %s", ref)
}

// ---- PUT /api/domain/{id}/css ----
// Saves CSS for a domain (your editor calls this)
// POST /api/domain/{id}/css
// PUT /api/domain/{id}/css
func (s *Server) handleUpdateDomainCSS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ref := strings.TrimPrefix(r.URL.Path, "/api/domain/")
	ref = strings.TrimSuffix(ref, "/css")

	if ref == "" {
		JSONError(w, http.StatusBadRequest, "missing domain id")
		return
	}

	// Read raw text CSS (not JSON)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		JSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	css := strings.TrimSpace(string(body))

	if css == "" {
		JSONError(w, http.StatusBadRequest, "empty CSS not allowed")
		return
	}

	internalID, err := s.resolveDomainInternalID(ctx, ref)
	if err != nil {
		JSONError(w, http.StatusNotFound, err.Error())
		return
	}

	// 🔧 Update nested {meta,data,css} path and update domains.updated_at
	_, err = s.DB().Exec(ctx, `
        UPDATE domains
        SET data = jsonb_set(
            jsonb_set(
                jsonb_set(
                    COALESCE(data, '{}'::jsonb),
                    '{meta}',
                    COALESCE(data->'meta', '{}'::jsonb),
                    true
                ),
                '{meta,data}',
                COALESCE(data#>'{meta,data}', '{}'::jsonb),
                true
            ),
            '{meta,data,css}',
            to_jsonb($1::text),
            true
        ),
        updated_at = NOW()
        WHERE id = $2
    `, css, internalID)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "db error: "+err.Error())
		return
	}

	// TODO: Future canon sync - update canon table when domain CSS changes
	// This will ensure consistency between runtime domains table and canonical storage

	JSONOk(w, "CSS updated successfully")
}

// ---- ROUTER ----

func (s *Server) registerRuntimeDomainRoutes() {
	s.mux.HandleFunc("/api/domain/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// /api/domain/{ref}/cascade
		if strings.HasSuffix(path, "/cascade") && r.Method == http.MethodGet {
			s.handleDomainCascade(w, r)
			return
		}

		// /api/domain/{ref}/css  (PUT)
		if strings.HasSuffix(path, "/css") {
			if r.Method == http.MethodPut {
				s.handleUpdateDomainCSS(w, r)
				return
			}
			JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		// /api/domain/{ref}  (GET runtime theme)
		if r.Method == http.MethodGet {
			s.handleGetDomain(w, r)
			return
		}

		JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	})
}

// handleDisDomainGet handles GET /api/domain/dis/{domain_id}
// This returns the canonical DIS domain object from the canon table, not runtime theme CSS.
func (s *Server) handleDisDomainGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx := r.Context()
	domainID := strings.TrimPrefix(r.URL.Path, "/api/domain/dis/")
	domainID = strings.TrimSpace(domainID)
	if domainID == "" {
		JSONError(w, http.StatusBadRequest, "missing domain id")
		return
	}

	var raw json.RawMessage
	err := s.Ledger().DB.QueryRow(ctx,
		`SELECT content
		   FROM canon
		  WHERE type = 'domain'
		    AND content->'meta'->>'domain_id' = $1`,
		domainID,
	).Scan(&raw)

	switch {
	case err == pgx.ErrNoRows:
		JSONError(w, http.StatusNotFound, "domain not found")
		return

	case err != nil:
		JSONError(w, http.StatusInternalServerError, "db error: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

// GET /api/domain/theme/{id}
// Returns stacked/cascaded CSS for given domain
func (s *Server) handleDomainTheme(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ref := strings.TrimPrefix(r.URL.Path, "/api/domain/theme/")
	if ref == "" {
		JSONError(w, http.StatusBadRequest, "missing domain id")
		return
	}

	// Resolve DB row id first
	internalID, err := s.resolveDomainInternalID(ctx, ref)
	if err != nil {
		JSONError(w, http.StatusNotFound, err.Error())
		return
	}

	// TEMP fake cascade: null base + domain css
	var css string
	err = s.DB().QueryRow(ctx, `SELECT COALESCE(css,'') FROM domains WHERE id=$1`, internalID).Scan(&css)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "db error: "+err.Error())
		return
	}

	baseCSS := `
body {
  font-family: Inter, sans-serif;
  background: #0a0f1a;
  color: #e2e8f0;
}
`

	resp := map[string]string{
		"css": baseCSS + "\n/* domain */\n" + css,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
