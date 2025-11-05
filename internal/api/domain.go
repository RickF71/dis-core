package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
func (s *Server) resolveDomainInternalID(ref string) (string, error) {
	var id string

	// Try DB internal ID first
	err := s.db.QueryRow(`SELECT id FROM domains WHERE id=$1`, ref).Scan(&id)
	if err == nil {
		return id, nil
	}

	// Try DIS longname style (domain_id)
	err = s.db.QueryRow(`SELECT id FROM domains WHERE domain_id=$1`, ref).Scan(&id)
	if err == nil {
		return id, nil
	}

	// Try the "name" column as last fallback
	err = s.db.QueryRow(`SELECT id FROM domains WHERE name=$1`, ref).Scan(&id)
	if err == nil {
		return id, nil
	}

	return "", fmt.Errorf("domain not found: %s", ref)
}

// ---- GET /api/domain/{id} ----
// Returns CSS + JSX for Finagler to apply the domain theme
func (s *Server) handleGetDomain(w http.ResponseWriter, r *http.Request) {
	ref := strings.TrimPrefix(r.URL.Path, "/api/domain/")
	if ref == "" {
		http.Error(w, "missing domain id", http.StatusBadRequest)
		return
	}

	// Resolve internal UUID/id first
	internalID, err := s.resolveDomainInternalID(ref)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var resp DomainResponse
	err = s.db.QueryRow(
		`SELECT domain_id, COALESCE(css, ''), COALESCE(jsx, '')
		   FROM domains WHERE id=$1`,
		internalID,
	).Scan(&resp.DomainID, &resp.CSS, &resp.JSX)

	if err == sql.ErrNoRows {
		http.Error(w, "domain not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("db error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ---- PUT /api/domain/{id}/css ----
// Saves CSS for a domain (your editor calls this)
func (s *Server) handleUpdateDomainCSS(w http.ResponseWriter, r *http.Request) {
	ref := strings.TrimPrefix(r.URL.Path, "/api/domain/")
	ref = strings.TrimSuffix(ref, "/css")

	if ref == "" {
		http.Error(w, "missing domain id", http.StatusBadRequest)
		return
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body error", http.StatusBadRequest)
		return
	}

	var payload struct {
		CSS string `json:"css"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Resolve to internal ID
	internalID, err := s.resolveDomainInternalID(ref)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Update row
	_, err = s.db.Exec(
		`UPDATE domains SET css=$1 WHERE id=$2`,
		payload.CSS, internalID,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("db error: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
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
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// /api/domain/{ref}  (GET runtime theme)
		if r.Method == http.MethodGet {
			s.handleGetDomain(w, r)
			return
		}

		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
}

// handleDisDomainGet handles GET /api/domain/dis/{domain_id}
// This returns the canonical DIS domain object from the canon table, not runtime theme CSS.
func (s *Server) handleDisDomainGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	domainID := strings.TrimPrefix(r.URL.Path, "/api/domain/dis/")
	domainID = strings.TrimSpace(domainID)
	if domainID == "" {
		http.Error(w, "missing domain id", http.StatusBadRequest)
		return
	}

	var raw json.RawMessage
	err := s.Ledger.DB.QueryRow(
		`SELECT content
		   FROM canon
		  WHERE type = 'domain'
		    AND content->'meta'->>'domain_id' = $1`,
		domainID,
	).Scan(&raw)

	switch {
	case err == sql.ErrNoRows:
		http.Error(w, "domain not found", http.StatusNotFound)
		return

	case err != nil:
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

// GET /api/domain/theme/{id}
// Returns stacked/cascaded CSS for given domain
func (s *Server) handleDomainTheme(w http.ResponseWriter, r *http.Request) {
	ref := strings.TrimPrefix(r.URL.Path, "/api/domain/theme/")
	if ref == "" {
		http.Error(w, "missing domain id", 400)
		return
	}

	// Resolve DB row id first
	internalID, err := s.resolveDomainInternalID(ref)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	// TEMP fake cascade: null base + domain css
	var css string
	err = s.db.QueryRow(`SELECT COALESCE(css,'') FROM domains WHERE id=$1`, internalID).Scan(&css)
	if err != nil {
		http.Error(w, "db error: "+err.Error(), 500)
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
