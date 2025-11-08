package api

import (
	"encoding/json"
	"log"
	"net/http"

	"dis-core/internal/policy"
	"dis-core/internal/schema"
)

// RegisterPolicyRoutes registers all /api/policy and /api/domain/{id}/policy routes.
func (s *Server) registerPolicyRoutes() {
	mux := s.mux
	mux.HandleFunc("POST /api/domain/{id}/policy/evaluate", s.handleEvaluateDomainPolicy)
	mux.HandleFunc("POST /api/policy/eval", s.handlePolicyEval)
}

// -----------------------------------------------------------------------------
// POST /api/domain/{id}/policy/evaluate
// Evaluates a policy set for a specific domain (with null-domain inheritance).
// -----------------------------------------------------------------------------
func (s *Server) handleEvaluateDomainPolicy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	domainID := r.PathValue("id")
	if domainID == "" {
		http.Error(w, "missing domain ID", http.StatusBadRequest)
		return
	}

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid input JSON", http.StatusBadRequest)
		return
	}

	// ✅ Step 1: Validate request payload against canonical schema
	if err := schema.ValidateAuthorityInput(input); err != nil {
		http.Error(w, "schema validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// ✅ Step 2: Load domain-specific and null-domain policy bundles
	domainBundle, err := policy.LoadDomainPolicyBundle(s.DB(), domainID)
	if err != nil {
		http.Error(w, "failed to load domain policy bundle: "+err.Error(), http.StatusInternalServerError)
		return
	}
	nullBundle, err := policy.LoadDomainPolicyBundle(s.DB(), "")
	if err != nil {
		http.Error(w, "failed to load null policy bundle: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// ✅ Step 3: Merge bundles into the runtime policy set
	mergedBundle := policy.MergeBundles(domainBundle, nullBundle)

	// ✅ Step 4: Build OPA engine and evaluate action
	engine, err := policy.NewEngine(mergedBundle.Modules)
	if err != nil {
		http.Error(w, "failed to build policy engine: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// ✅ Add this debug block here:
	for k := range mergedBundle.Modules {
		log.Printf("[policy] loaded module: %s", k)
	}

	decision, err := engine.EvaluateAction(input)
	if err != nil {
		http.Error(w, "policy evaluation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// ✅ Step 5: Respond with decision JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(decision)
}

// -----------------------------------------------------------------------------
// POST /api/policy/eval
// Evaluates policies for the NULL domain only (global evaluator).
// -----------------------------------------------------------------------------
func (s *Server) handlePolicyEval(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}

	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid JSON input", http.StatusBadRequest)
			return
		}
	} else {
		// Fallback example input (safe default for testing)
		input = map[string]interface{}{
			"actor_ref":      "id-rick",
			"action_ref":     "domain.freeze.v1",
			"policy_ref":     "policy.freeze.control.v1",
			"decision":       map[string]interface{}{"outcome": "deny"},
			"consent_status": "granted",
		}
	}

	// ✅ Step 1: Schema validation (constitutional guardrail)
	if err := schema.ValidateAuthorityInput(input); err != nil {
		http.Error(w, "schema validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// ✅ Step 2: Load null/global policy bundle
	nullBundle, err := policy.LoadDomainPolicyBundle(s.DB(), "")
	if err != nil {
		http.Error(w, "failed to load null policy bundle: "+err.Error(), http.StatusInternalServerError)
		return
	}

	engine, err := policy.NewEngine(nullBundle.Modules)
	if err != nil {
		http.Error(w, "failed to build policy engine: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// ✅ Step 3: Evaluate and respond
	decision, err := engine.EvaluateAction(input)
	if err != nil {
		http.Error(w, "policy evaluation error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(decision)
}
