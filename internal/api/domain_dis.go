package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// DisBundle represents the resolved visual layer for a domain.
type DisBundle struct {
	CSS string `json:"css"`
	JSX string `json:"jsx"`
}

// ------------------------------------------------------------
//  /api/domain/dis/{id} — Returns resolved DIS.css / DIS.jsx
// ------------------------------------------------------------

// handleDomainDIS resolves canonical DIS domain objects
// GET /api/domain/dis/{domain_id}
// This pulls the authoritative DIS domain definition from canon.
func (s *Server) handleDomainDIS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	domainID := strings.TrimPrefix(r.URL.Path, "/api/domain/dis/")
	domainID = strings.TrimSpace(domainID)
	if domainID == "" {
		http.Error(w, "missing domain id", http.StatusBadRequest)
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
		http.Error(w, "domain not found", http.StatusNotFound)
		return

	case err != nil:
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

// ------------------------------------------------------------
//  Recursive resolver: follows interface->dis_css/dis_jsx refs
// ------------------------------------------------------------

type ifaceRef struct {
	Ref    string `json:"ref"`
	Append string `json:"append"`
}

func (s *Server) resolveInterface(domainID string) (string, string, error) {
	const maxDepth = 10
	seen := map[string]bool{}
	return s.resolveInterfaceRecursive(domainID, seen, maxDepth)
}

func (s *Server) resolveInterfaceRecursive(domainID string, seen map[string]bool, depth int) (string, string, error) {
	ctx := context.Background() // TODO: pass context through
	if depth <= 0 {
		return "", "", fmt.Errorf("too many nested refs (possible loop)")
	}
	if seen[domainID] {
		return "", "", fmt.Errorf("cyclic ref in interface chain: %s", domainID)
	}
	seen[domainID] = true

	var cssRaw, jsxRaw pgtype.Text
	err := s.Ledger().DB.QueryRow(ctx, `
		SELECT
		  content->'interface'->'dis_css',
		  content->'interface'->'dis_jsx'
		FROM canon
		WHERE type='domain'
		  AND content->'meta'->>'domain_id' = $1
	`, domainID).Scan(&cssRaw, &jsxRaw)

	if err == pgx.ErrNoRows {
		return "", "", fmt.Errorf("domain not found: %s", domainID)
	}
	if err != nil {
		return "", "", fmt.Errorf("db error: %v", err)
	}

	var cssRef, jsxRef ifaceRef
	_ = json.Unmarshal([]byte(cssRaw.String), &cssRef)
	_ = json.Unmarshal([]byte(jsxRaw.String), &jsxRef)

	baseCSS, baseJSX := cssRef.Append, jsxRef.Append

	if cssRef.Ref != "" {
		parentCSS, _, err := s.resolveInterfaceRecursive(cssRef.Ref, seen, depth-1)
		if err == nil {
			baseCSS = parentCSS + "\n" + baseCSS
		}
	}
	if jsxRef.Ref != "" {
		parentJSX, _, err := s.resolveInterfaceRecursive(jsxRef.Ref, seen, depth-1)
		if err == nil {
			baseJSX = parentJSX + "\n" + baseJSX
		}
	}
	return baseCSS, baseJSX, nil
}
