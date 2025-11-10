package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"dis-core/internal/policy"
)

// GET /api/domain/{id}/policy
func (s *Server) handleGetDomainPolicy(w http.ResponseWriter, r *http.Request) {
	domainID := r.PathValue("id")
	if domainID == "" {
		http.Error(w, "missing domain ID", http.StatusBadRequest)
		return
	}

	// Try to read domain-specific policy first
	domainPolicyPath := filepath.Join("domains", domainID, "policy", "gates.rego")
	content, err := os.ReadFile(domainPolicyPath)

	if err != nil {
		// Fall back to internal default policy
		internalPolicyPath := filepath.Join("internal", "policy", "gates.rego")
		content, err = os.ReadFile(internalPolicyPath)
		if err != nil {
			http.Error(w, "policy not found", http.StatusNotFound)
			return
		}
	}

	response := map[string]string{
		"content": string(content),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// POST /api/domain/{id}/policy
func (s *Server) handleSetDomainPolicy(w http.ResponseWriter, r *http.Request) {
	domainID := r.PathValue("id")
	if domainID == "" {
		http.Error(w, "missing domain ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Ensure the domain policy directory exists
	policyDir := filepath.Join("domains", domainID, "policy")
	if err := os.MkdirAll(policyDir, 0755); err != nil {
		http.Error(w, "failed to create policy directory", http.StatusInternalServerError)
		return
	}

	// Write the policy content to gates.rego
	policyPath := filepath.Join(policyDir, "gates.rego")
	if err := os.WriteFile(policyPath, []byte(req.Content), 0644); err != nil {
		http.Error(w, "failed to write policy file", http.StatusInternalServerError)
		return
	}

	// Reload the policy engine if available
	if s.policy != nil {
		if opaEngine, ok := s.policy.(*policy.OPAEngine); ok {
			if err := opaEngine.Reload(r.Context(), domainID); err != nil {
				http.Error(w, "failed to reload policy engine", http.StatusInternalServerError)
				return
			}
		}
	}

	response := map[string]string{
		"status": "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
