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
// DEPRECATED: GOV-8 violation - uses name-based lookup
// TODO: Remove after all callers migrated to UUID-only
// Accepts either:
//   - internal DB id (uuid/int) OR
//   - DIS domain longname (domain.user.rick)
//
// Returns the INTERNAL ID and matching domain record data.
func (s *Server) resolveDomainInternalID(ctx context.Context, ref string) (string, error) {
	var id string

	// Try DB internal ID first (UUID-based - acceptable)
	err := s.DB().QueryRow(ctx, `SELECT id FROM domains WHERE id=$1`, ref).Scan(&id)
	if err == nil {
		return id, nil
	}

	// Try DIS longname style (domain_id)
	err = s.DB().QueryRow(ctx, `SELECT id FROM domains WHERE domain_id=$1`, ref).Scan(&id)
	if err == nil {
		return id, nil
	}

	// GOV-8: WARNING - Name-based lookup violates DIS-Invariant-001
	// This fallback path should be removed once all callers use UUID
	err = s.DB().QueryRow(ctx, `SELECT id FROM domains WHERE name=$1`, ref).Scan(&id)
	if err == nil {
		s.logger.Printf("⚠️  GOV-8 WARNING: Name-based domain lookup used for '%s'. Migrate to UUID.", ref)
		return id, nil
	}

	return "", fmt.Errorf("domain not found: %s", ref)
}

// GOV-8: resolveDomainByUUID enforces UUID-only domain resolution
// This is the ONLY acceptable method for domain lookups in authority operations
func (s *Server) resolveDomainByUUID(ctx context.Context, domainID string) (string, error) {
	var id string

	// Validate UUID format first
	err := s.DB().QueryRow(ctx, `SELECT id FROM domains WHERE id = $1`, domainID).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("GOV-8: domain UUID not found: %s", domainID)
		}
		return "", fmt.Errorf("GOV-8: domain validation failed: %w", err)
	}

	return id, nil
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

	// Phase 10J.4: Update flattened payload->css->content
	_, err = s.DB().Exec(ctx, `
        UPDATE domains
        SET payload = jsonb_set(
            payload,
            '{css}',
            jsonb_build_object(
                'content', $1::text,
                'hash', '',
                'verified', true,
                'updated_at', now()::text
            ),
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
	s.router.HandleFunc("/api/domain/", func(w http.ResponseWriter, r *http.Request) {
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
