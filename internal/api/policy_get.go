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

// PolicyResponse represents the JSON response for policy GET endpoint
type PolicyResponse struct {
	DomainID      string `json:"domain_id"`
	DomainName    string `json:"domain_name,omitempty"`
	Mode          string `json:"mode"`
	ContentType   string `json:"content_type"`
	PolicyContent string `json:"policy_content"`
}

// GetDomainPolicyMode handles GET /api/policy/get/{id}?mode=effective|local|inherited
// Returns policy content as JSON by default:
// - local: Just this domain's policy
// - inherited: Parent policies concatenated (excludes current domain)
// - effective: Full hierarchy including current domain (root→parent→child)
// Format suffixes for backward compatibility:
// - /rego or /text: returns text/plain
// - /json: returns application/json (same as default)
func (s *Server) GetDomainPolicyMode(w http.ResponseWriter, r *http.Request) {
	domainID := chi.URLParam(r, "id")
	mode := r.URL.Query().Get("mode")
	format := r.URL.Query().Get("format")

	if domainID == "" {
		JSONError(w, http.StatusBadRequest, "Domain ID required")
		return
	}

	// Default to local if no mode specified
	if mode == "" {
		mode = "local"
	}

	ctx := r.Context()
	var policyContent string
	var err error

	// Require a DB for policy endpoints
	db := s.requireDB(w)
	if db == nil {
		return
	}

	// Get domain name for response (optional)
	var domainName string
	row := db.QueryRow(ctx, "SELECT name FROM domains WHERE id = $1", domainID)
	_ = row.Scan(&domainName) // Ignore error - name is optional

	switch mode {
	case "local":
		policyContent, err = s.getDomainRegoPolicy(ctx, db, domainID)
		if err != nil || policyContent == "" {
			policyContent = s.getDefaultPolicyTemplate()
		}

	case "inherited":
		policyContent, err = s.getInheritedPolicies(ctx, db, domainID)
		if err != nil {
			JSONError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get inherited policies: %v", err))
			return
		}

	case "effective":
		policyContent, err = s.getEffectivePolicies(ctx, db, domainID)
		if err != nil {
			JSONError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get effective policies: %v", err))
			return
		}

	default:
		JSONError(w, http.StatusBadRequest, "Invalid mode. Use: local, inherited, or effective")
		return
	}

	// Return format based on query parameter or default to JSON
	if format == "rego" || format == "text" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(policyContent))
		return
	}

	// Default: return JSON
	response := PolicyResponse{
		DomainID:      domainID,
		DomainName:    domainName,
		Mode:          mode,
		ContentType:   "text/rego",
		PolicyContent: policyContent,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getInheritedPolicies returns parent policies concatenated (excludes current domain)
func (s *Server) getInheritedPolicies(ctx context.Context, db *pgxpool.Pool, domainID string) (string, error) {
	lineage, err := s.getDomainLineage(ctx, db, domainID)
	if err != nil {
		return "", err
	}

	// Skip the first item (current domain) and collect parent policies
	var policies []string
	for i := len(lineage) - 1; i > 0; i-- {
		parentID := lineage[i]
		policy, err := s.getDomainRegoPolicy(ctx, db, parentID)
		if err == nil && policy != "" {
			comment := fmt.Sprintf("# Policy from domain: %s", parentID)
			policies = append(policies, comment, policy)
		}
	}

	if len(policies) == 0 {
		return "# No inherited policies", nil
	}

	return strings.Join(policies, "\n\n"), nil
}

// getEffectivePolicies returns full hierarchy including current domain (parent→child order)
func (s *Server) getEffectivePolicies(ctx context.Context, db *pgxpool.Pool, domainID string) (string, error) {
	lineage, err := s.getDomainLineage(ctx, db, domainID)
	if err != nil {
		return "", err
	}

	// Collect all policies in parent-first order
	var policies []string
	for i := len(lineage) - 1; i >= 0; i-- {
		domID := lineage[i]
		policy, err := s.getDomainRegoPolicy(ctx, db, domID)
		if err == nil && policy != "" {
			comment := fmt.Sprintf("# Policy from domain: %s", domID)
			policies = append(policies, comment, policy)
		}
	}

	if len(policies) == 0 {
		return s.getDefaultPolicyTemplate(), nil
	}

	return strings.Join(policies, "\n\n"), nil
}

// getDefaultPolicyTemplate returns the default Rego v1 policy template
func (s *Server) getDefaultPolicyTemplate() string {
	return `# No policy defined for this domain
package dis.policy

default allow := false
`
}
