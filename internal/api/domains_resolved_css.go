// domains_resolved_css.go implements CSS inheritance resolution
// This endpoint returns accumulated CSS from ancestor chain + current domain
package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"dis-core/internal/domain"
)

// HandleGetResolvedCSS handles GET /api/domain/{id}/resolved-css
// Returns CSS with full inheritance chain applied: ancestors → self
func (s *Server) HandleGetResolvedCSS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	domainID := chi.URLParam(r, "id")
	if domainID == "" {
		http.Error(w, "missing domain id", http.StatusBadRequest)
		return
	}

	// Parse domain ID as UUID
	domainUUID, err := uuid.Parse(domainID)
	if err != nil {
		http.Error(w, "invalid domain id format", http.StatusBadRequest)
		return
	}

	// 1. Get ancestor chain in order: [root, parent, ..., grandparent]
	db := s.requireDB(w)
	if db == nil {
		return
	}

	chain, err := domain.ResolveDomainLineage(ctx, db, domainUUID)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot compute ancestry: %v", err), http.StatusInternalServerError)
		return
	}

	// 2. Load accumulated CSS from ancestors (oldest to newest)
	var cssParts []string

	// Add CSS from each ancestor in the chain
	for _, ancestorID := range chain {
		css, err := s.getDomainCSSByUUID(ctx, db, ancestorID)
		if err == nil && css != "" {
			cssParts = append(cssParts, fmt.Sprintf("/* Domain: %s */\n%s", ancestorID.String(), css))
		}
	}

	// 3. Add active domain's CSS last (highest specificity)
	selfCSS, err := s.getDomainCSSByUUID(ctx, db, domainUUID)
	if err == nil && selfCSS != "" {
		cssParts = append(cssParts, fmt.Sprintf("/* Domain: %s (current) */\n%s", domainID, selfCSS))
	}

	// 4. Merge all CSS parts with separators
	merged := strings.Join(cssParts, "\n\n")

	// 5. Return as text/css
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Header().Set("X-CSS-Inheritance-Chain-Length", fmt.Sprintf("%d", len(cssParts)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(merged))
}

// getDomainCSSByUUID retrieves CSS content for a domain by UUID
func (s *Server) getDomainCSSByUUID(ctx context.Context, db *pgxpool.Pool, domainID uuid.UUID) (string, error) {
	var cssContent string
	err := db.QueryRow(ctx, `
		SELECT COALESCE(payload->'css'->>'content', '')
		FROM domains
		WHERE id = $1
	`, domainID).Scan(&cssContent)

	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("domain not found: %s", domainID)
		}
		return "", fmt.Errorf("failed to get CSS for domain %s: %w", domainID, err)
	}

	return cssContent, nil
}
