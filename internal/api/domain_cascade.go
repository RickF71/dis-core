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

type cascadeNode struct {
	ID       string      `json:"id"`
	DomainID string      `json:"domain_id"`
	ParentID pgtype.Text `json:"-"`
	CSS      string      `json:"css"`
}

type cascadeResp struct {
	Order []string `json:"order"` // domain_id from root→active
	CSS   string   `json:"css"`   // merged CSS (root→active)
}

// GET /api/domain/{ref}/cascade
// ref may be internal id, domain_id, or name (resolver handles)
func (s *Server) handleDomainCascade(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	ref := strings.TrimPrefix(r.URL.Path, "/api/domain/")
	ref = strings.TrimSuffix(ref, "/cascade")
	ref = strings.TrimSpace(ref)
	if ref == "" {
		http.Error(w, "missing domain id", http.StatusBadRequest)
		return
	}

	// Resolve ref → internal id
	internalID, err := s.resolveDomainInternalID(ctx, ref)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	chain, err := s.buildCascadeChain(ctx, internalID)
	if err != nil {
		http.Error(w, "cascade error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Build order and merged CSS (root → leaf)
	order := make([]string, 0, len(chain))
	var merged strings.Builder
	for _, n := range chain {
		order = append(order, n.DomainID)
		if strings.TrimSpace(n.CSS) != "" {
			merged.WriteString("\n/* ===== ")
			merged.WriteString(n.DomainID)
			merged.WriteString(" ===== */\n")
			merged.WriteString(n.CSS)
			merged.WriteString("\n")
		}
	}

	resp := cascadeResp{Order: order, CSS: merged.String()}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// buildCascadeChain returns nodes from ROOT→ACTIVE (domain.null at index 0)
func (s *Server) buildCascadeChain(ctx context.Context, internalID string) ([]cascadeNode, error) {
	// Walk up: active → root
	var up []cascadeNode
	curr := internalID
	for {
		var n cascadeNode
		err := s.DB().QueryRow(ctx, `
			SELECT id, COALESCE(domain_id, ''), parent_id, COALESCE(css, '')
			FROM domains
			WHERE id = $1
		`, curr).Scan(&n.ID, &n.DomainID, &n.ParentID, &n.CSS)
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("domain not found (id=%s)", curr)
		}
		if err != nil {
			return nil, err
		}
		up = append(up, n)
		if !n.ParentID.Valid || strings.TrimSpace(n.ParentID.String) == "" {
			break
		}
		curr = n.ParentID.String
	}

	// Reverse: root → active
	for i, j := 0, len(up)-1; i < j; i, j = i+1, j-1 {
		up[i], up[j] = up[j], up[i]
	}
	return up, nil
}
