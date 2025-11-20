package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// IdentityPolicyResponse represents the effective identity policy for a domain
type IdentityPolicyResponse struct {
	DomainID        string                 `json:"domain_id"`
	DomainName      string                 `json:"domain_name"`
	PolicyVersion   string                 `json:"policy_version"`
	LocalPolicy     map[string]interface{} `json:"local_policy"`
	ParentPolicyRef *ParentPolicyRef       `json:"parent_policy_ref,omitempty"`
	EffectivePolicy map[string]interface{} `json:"effective_policy"`
	Digest          string                 `json:"digest"`
	UpdatedAt       string                 `json:"updated_at"`
	Source          string                 `json:"source"` // "direct" or "mock"
}

// ParentPolicyRef contains reference to parent domain's identity policy
type ParentPolicyRef struct {
	DomainID   string `json:"domain_id"`
	DomainName string `json:"domain_name"`
	Digest     string `json:"digest"`
	Version    string `json:"version"`
}

// handleGetIdentityPolicy returns the effective identity policy for a domain
// GET /api/policy/identity/:domainId
func (s *Server) handleGetIdentityPolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	domainID := chi.URLParam(r, "domainId")

	if domainID == "" {
		http.Error(w, "domain_id required", http.StatusBadRequest)
		return
	}

	db := s.requireDB(w)
	if db == nil {
		return
	}

	policy, err := s.getIdentityPolicy(ctx, db, domainID)
	if err != nil {
		s.logger.Printf("Error fetching identity policy for domain %s: %v", domainID, err)
		http.Error(w, fmt.Sprintf("failed to fetch identity policy: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policy)
}

// getIdentityPolicy fetches and merges identity policy for a domain
func (s *Server) getIdentityPolicy(ctx context.Context, db *pgxpool.Pool, domainID string) (*IdentityPolicyResponse, error) {
	// Fetch domain info
	var domainName, parentID string
	var updatedAt time.Time
	err := db.QueryRow(ctx, `
		SELECT name, parent_id, updated_at
		FROM domains
		WHERE id = $1
	`, domainID).Scan(&domainName, &parentID, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}

	// Fetch local identity policy from domain payload
	localPolicy, err := s.getDomainIdentityPolicy(ctx, db, domainID)
	if err != nil {
		s.logger.Printf("Warning: failed to fetch local identity policy for %s: %v", domainID, err)
		localPolicy = make(map[string]interface{})
	}

	var parentRef *ParentPolicyRef
	effectivePolicy := make(map[string]interface{})

	// If domain has a parent, fetch parent's identity policy
	if parentID != "" {
		parentPolicy, err := s.getDomainIdentityPolicy(ctx, db, parentID)
		if err != nil {
			s.logger.Printf("Warning: failed to fetch parent identity policy for %s: %v", parentID, err)
		} else {
			// Fetch parent domain name
			var parentName string
			err := db.QueryRow(ctx, "SELECT name FROM domains WHERE id = $1", parentID).Scan(&parentName)
			if err == nil {
				// Compute parent policy digest
				parentDigest := computePolicyDigest(parentPolicy)

				parentRef = &ParentPolicyRef{
					DomainID:   parentID,
					DomainName: parentName,
					Digest:     parentDigest,
					Version:    "domain.policy.v1",
				}

				// Merge parent policy into effective policy (parent first, local overrides)
				effectivePolicy = mergePolicies(parentPolicy, localPolicy)
			}
		}
	}

	// If no parent or parent merge failed, effective = local
	if len(effectivePolicy) == 0 {
		for k, v := range localPolicy {
			effectivePolicy[k] = v
		}
	}

	// Compute digest of effective policy
	digest := computePolicyDigest(effectivePolicy)

	return &IdentityPolicyResponse{
		DomainID:        domainID,
		DomainName:      domainName,
		PolicyVersion:   "domain.policy.v1",
		LocalPolicy:     localPolicy,
		ParentPolicyRef: parentRef,
		EffectivePolicy: effectivePolicy,
		Digest:          digest,
		UpdatedAt:       updatedAt.Format(time.RFC3339),
		Source:          "direct",
	}, nil
}

// getDomainIdentityPolicy extracts identity policy from domain payload
func (s *Server) getDomainIdentityPolicy(ctx context.Context, db *pgxpool.Pool, domainID string) (map[string]interface{}, error) {
	var payloadJSON []byte
	err := db.QueryRow(ctx, `
		SELECT payload
		FROM domains
		WHERE id = $1
	`, domainID).Scan(&payloadJSON)
	if err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return nil, fmt.Errorf("invalid payload JSON: %w", err)
	}

	// Look for policy.identity or policy.identity_v1
	policyBlock, ok := payload["policy"].(map[string]interface{})
	if !ok {
		return make(map[string]interface{}), nil
	}

	// Try identity_v1 first, then identity
	if identityPolicy, ok := policyBlock["identity_v1"].(map[string]interface{}); ok {
		return identityPolicy, nil
	}
	if identityPolicy, ok := policyBlock["identity"].(map[string]interface{}); ok {
		return identityPolicy, nil
	}

	return make(map[string]interface{}), nil
}

// mergePolicies performs shallow merge (parent then local override)
func mergePolicies(parent, local map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy parent
	for k, v := range parent {
		result[k] = v
	}

	// Override with local
	for k, v := range local {
		result[k] = v
	}

	return result
}

// computePolicyDigest calculates SHA-256 digest of policy map
func computePolicyDigest(policy map[string]interface{}) string {
	if len(policy) == 0 {
		return "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" // empty digest
	}

	// Serialize to stable JSON
	jsonBytes, err := json.Marshal(policy)
	if err != nil {
		return "sha256:error"
	}

	hash := sha256.Sum256(jsonBytes)
	return "sha256:" + hex.EncodeToString(hash[:])
}
