package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DomainPolicy represents a policy associated with a domain
type DomainPolicy struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Version    string `json:"version"`
	DomainID   string `json:"domain_id"`
	PolicyType string `json:"policy_type"`
	Content    string `json:"content,omitempty"`
	Active     bool   `json:"active"`
}

// GetDomainPolicy handles GET /api/domain/{id}/policy with format support
func (s *Server) GetDomainPolicy(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)
	domainID := chi.URLParam(r, "id")

	if domainID == "" {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusBadRequest, "Domain ID required")
		case FormatText:
			http.Error(w, "Domain ID required", http.StatusBadRequest)
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	ctx := r.Context()

	// Prefer database-backed policies when present, but allow
	// filesystem-only operation when DB is not configured. Use s.DB()
	// (nil-safe) instead of requireDB which would write a 503.
	db := s.DB()

	policies, err := s.getDomainPolicies(ctx, db, domainID)
	if err != nil {
		policies = []DomainPolicy{}
	}
	// Format-specific response
	switch format {
	case FormatJSON:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(policies)
	case FormatText:
		w.Header().Set("Content-Type", "text/plain")
		if len(policies) > 0 {
			var textLines []string
			for _, policy := range policies {
				status := "inactive"
				if policy.Active {
					status = "active"
				}
				line := fmt.Sprintf("%s v%s (%s) - %s",
					policy.Name, policy.Version, policy.PolicyType, status)
				textLines = append(textLines, line)
			}
			response := strings.Join(textLines, "\n")
			w.Write([]byte(response))
		} else {
			w.Write([]byte("No policies found for domain"))
		}
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// getDomainPolicies queries the policies table filtered by domain_id
func (s *Server) getDomainPolicies(ctx context.Context, db *pgxpool.Pool, domainID string) ([]DomainPolicy, error) {
	query := `
		SELECT id, name, COALESCE(version, '1.0') as version,
		       domain_id, COALESCE(policy_type, 'unknown') as policy_type,
		       COALESCE(active, true) as active
		FROM policies
		WHERE domain_id = $1
		ORDER BY name, version
	`

	if db == nil {
		// No DB configured: return empty policy list (filesystem-only callers
		// should use alternative logic). This avoids panics when db is nil.
		return []DomainPolicy{}, nil
	}

	rows, err := db.Query(ctx, query, domainID)
	if err != nil {
		// If policies table doesn't exist or query fails, return empty array
		return []DomainPolicy{}, nil
	}
	defer rows.Close()

	var policies []DomainPolicy
	for rows.Next() {
		var p DomainPolicy
		err := rows.Scan(&p.ID, &p.Name, &p.Version, &p.DomainID, &p.PolicyType, &p.Active)
		if err != nil {
			continue // Skip invalid rows
		}
		policies = append(policies, p)
	}

	return policies, nil
}

// GetDomainPolicyInherited handles GET /api/domain/{id}/policy/inherited
// Returns the concatenated Rego policies from parent domains up to null
func (s *Server) GetDomainPolicyInherited(w http.ResponseWriter, r *http.Request) {
	domainID := chi.URLParam(r, "id")

	if domainID == "" {
		JSONError(w, http.StatusBadRequest, "Domain ID required")
		return
	}

	ctx := r.Context()

	// Require DB for lineage lookup
	db := s.requireDB(w)
	if db == nil {
		return
	}

	// Get domain lineage
	lineage, err := s.getDomainLineage(ctx, db, domainID)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to get domain lineage")
		return
	}

	// Collect policies from all parent domains (excluding current domain)
	var inheritedPolicies []string
	for i := len(lineage) - 1; i > 0; i-- { // Start from root, skip current domain
		parentID := lineage[i]
		policy, err := s.getDomainRegoPolicy(ctx, db, parentID)
		if err == nil && policy != "" {
			inheritedPolicies = append(inheritedPolicies, fmt.Sprintf("# Policy from domain: %s\n%s", parentID, policy))
		}
	}

	inherited := strings.Join(inheritedPolicies, "\n\n")
	if inherited == "" {
		inherited = "# No inherited policies"
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(inherited))
}

// GetDomainPolicyCurrent handles GET /api/domain/{id}/policy/current
// Returns only the current domain's Rego policy
func (s *Server) GetDomainPolicyCurrent(w http.ResponseWriter, r *http.Request) {
	domainID := chi.URLParam(r, "id")

	if domainID == "" {
		JSONError(w, http.StatusBadRequest, "Domain ID required")
		return
	}

	ctx := r.Context()
	db := s.requireDB(w)
	if db == nil {
		return
	}

	policy, err := s.getDomainRegoPolicy(ctx, db, domainID)
	if err != nil {
		JSONError(w, http.StatusNotFound, "Policy not found")
		return
	}

	if policy == "" {
		policy = "# No policy defined for this domain\npackage dis.policy\n\ndefault allow := false\n"
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(policy))
}

// getDomainLineage returns the chain of domain IDs from current up to null
func (s *Server) getDomainLineage(ctx context.Context, db *pgxpool.Pool, domainID string) ([]string, error) {
	if db == nil {
		return []string{domainID}, nil
	}

	var lineage []string
	currentID := domainID

	for currentID != "" && len(lineage) < 20 { // Safety limit
		lineage = append(lineage, currentID)

		var parentID *string
		err := db.QueryRow(ctx, "SELECT parent_id FROM domains WHERE id = $1", currentID).Scan(&parentID)
		if err != nil || parentID == nil {
			break
		}
		currentID = *parentID
	}

	return lineage, nil
}

// getDomainRegoPolicy fetches the Rego policy content for a specific domain
func (s *Server) getDomainRegoPolicy(ctx context.Context, db *pgxpool.Pool, domainID string) (string, error) {
	if db == nil {
		return "", fmt.Errorf("database not available")
	}

	var policy string
	err := db.QueryRow(ctx, `
		SELECT COALESCE(payload->'policy'->>'content', '')
		FROM domains
		WHERE id = $1
	`, domainID).Scan(&policy)

	return policy, err
}
