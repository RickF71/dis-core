package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"dis-core/internal/api/middleware"
	"dis-core/internal/receipts"
)

// AuthorityStatusResponse represents the response for /api/authority/status
type AuthorityStatusResponse struct {
	Ready            bool    `json:"ready"`
	Schemas          int     `json:"schemas"`
	Policies         int     `json:"policies"`
	Receipts         int     `json:"receipts"`
	ContinuityRate   float64 `json:"continuity_rate,omitempty"`
	ContinuityRisk   string  `json:"continuity_risk,omitempty"`
	OrphanReceipts   int     `json:"orphan_receipts,omitempty"`
	RemediationReady bool    `json:"remediation_ready,omitempty"`
}

// PolicyEvaluationRequest represents the request for /api/policy/evaluate
type PolicyEvaluationRequest struct {
	Action  string                 `json:"action"`
	Context map[string]interface{} `json:"context"`
}

// PolicyEvaluationResponse represents the response for /api/policy/evaluate
type PolicyEvaluationResponse struct {
	Allow   bool                   `json:"allow"`
	Reason  string                 `json:"reason"`
	Details map[string]interface{} `json:"details"`
}

// handleAuthoritySchemaPhase7 returns the loaded Authority Console schema for Phase 7
func (s *Server) handleAuthoritySchemaPhase7(w http.ResponseWriter, r *http.Request) {
	// Get database from middleware context
	db := middleware.FromDB(r.Context())

	if db == nil {
		http.Error(w, "Database unavailable", http.StatusInternalServerError)
		return
	}

	ctx := context.Background()

	// Retrieve schema.json from canon table
	var content []byte
	err := db.QueryRow(ctx, "SELECT content FROM canon WHERE type = 'authority.console.schema' LIMIT 1").Scan(&content)
	if err != nil {
		// Fallback: try to read from filesystem
		schemaPath := "internal/schema/schema.json"
		content, err = os.ReadFile(schemaPath)
		if err != nil {
			http.Error(w, "Schema not found", http.StatusNotFound)
			return
		}
	}

	// Validate that it's proper JSON
	var schemaData map[string]interface{}
	if err := json.Unmarshal(content, &schemaData); err != nil {
		http.Error(w, "Invalid schema JSON", http.StatusInternalServerError)
		return
	}

	// Phase 9C: Add receipts field to schema response
	schemaData["receipts"] = []string{"ci.call.v1", "ci.import.v1"}

	// Phase 10D: Add policy continuity section
	if s.db != nil {
		// Get global policy continuity statistics
		if globalStats, err := s.getGlobalPolicyContinuityStats(context.Background()); err == nil {
			// Phase 10E: Add risk assessment and threshold information
			thresholds := receipts.DefaultContinuityThresholds()
			riskLevel := receipts.GetContinuityRiskLevel(globalStats.ContinuityRate)

			schemaData["policy_continuity"] = map[string]interface{}{
				"supported":            true,
				"continuity_rate":      globalStats.ContinuityRate,
				"total_receipts":       globalStats.TotalReceipts,
				"valid_refs":           globalStats.ValidRefs,
				"orphan_refs":          globalStats.OrphanRefs,
				"risk_level":           riskLevel,
				"thresholds":           thresholds,
				"endpoint":             "/api/policy/continuity/{domain}",
				"global_endpoint":      "/api/policy/continuity",
				"remediation_endpoint": "/api/policy/continuity/remediate",
				"dashboard_endpoint":   "/api/dashboard/continuity",
			}
		} else {
			schemaData["policy_continuity"] = map[string]interface{}{
				"supported": true,
				"error":     "statistics unavailable",
				"endpoint":  "/api/policy/continuity/{domain}",
			}
		}
	} else {
		schemaData["policy_continuity"] = map[string]interface{}{
			"supported": false,
			"reason":    "database not available",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schemaData)
}

// handlePolicyEvaluatePhase7 evaluates a policy against the given action and context for Phase 7
func (s *Server) handlePolicyEvaluatePhase7(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req PolicyEvaluationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	// Get policy engine from middleware context
	policyEngine := middleware.FromPolicyEngine(r.Context())

	if policyEngine == nil {
		http.Error(w, "Policy engine unavailable", http.StatusInternalServerError)
		return
	}

	// Create evaluation input for the policy engine
	input := map[string]interface{}{
		"action":   req.Action,
		"resource": "authority.console",
		"subject":  "system", // Could be extracted from authentication
		"context":  req.Context,
	}

	// Evaluate policy using the PolicyEngine interface
	decision, err := policyEngine.EvaluateAction(r.Context(), input)
	if err != nil {
		response := PolicyEvaluationResponse{
			Allow:  false,
			Reason: fmt.Sprintf("Policy evaluation error: %v", err),
			Details: map[string]interface{}{
				"error":   err.Error(),
				"action":  req.Action,
				"context": req.Context,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Handle case where decision is nil (no matching policy)
	if decision == nil {
		response := PolicyEvaluationResponse{
			Allow:  false,
			Reason: "No policy decision available",
			Details: map[string]interface{}{
				"action":  req.Action,
				"context": req.Context,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Build response
	response := PolicyEvaluationResponse{
		Allow:  decision.Allow,
		Reason: decision.Reason,
		Details: map[string]interface{}{
			"action":      req.Action,
			"context":     req.Context,
			"break_glass": decision.BreakGlass,
			"details":     decision.Details,
			"timestamp":   decision.Timestamp,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// PolicyContinuityOverview represents a simplified view for authority schema
type PolicyContinuityOverview struct {
	TotalReceipts  int     `json:"total_receipts"`
	ValidRefs      int     `json:"valid_refs"`
	OrphanRefs     int     `json:"orphan_refs"`
	ContinuityRate float64 `json:"continuity_rate"`
}

// getGlobalPolicyContinuityOverview returns a simplified overview of policy continuity
func (s *Server) getGlobalPolicyContinuityOverview() (PolicyContinuityOverview, error) {
	ctx := context.Background()
	overview := PolicyContinuityOverview{}

	if s.db == nil {
		return overview, fmt.Errorf("database not available")
	}

	// Get total receipts
	err := s.db.QueryRow(ctx, "SELECT COUNT(*) FROM receipts").Scan(&overview.TotalReceipts)
	if err != nil {
		return overview, fmt.Errorf("failed to count total receipts: %w", err)
	}

	// Count receipts with valid policy refs
	err = s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM receipts
		WHERE policy_ref IS NOT NULL AND policy_ref != ''
	`).Scan(&overview.ValidRefs)
	if err != nil {
		return overview, fmt.Errorf("failed to count valid refs: %w", err)
	}

	// Count orphan receipts
	overview.OrphanRefs = overview.TotalReceipts - overview.ValidRefs

	// Calculate continuity rate
	if overview.TotalReceipts > 0 {
		overview.ContinuityRate = float64(overview.ValidRefs) / float64(overview.TotalReceipts)
	}

	return overview, nil
}
