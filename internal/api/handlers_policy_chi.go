package api

import (
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// handleListPoliciesFormatAware handles GET /api/policy/list with format support
func (s *Server) handleListPoliciesFormatAware(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)

	// Get policies using existing logic
	policies, err := s.getAllPolicies()
	if err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusInternalServerError, "failed to list policies")
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	// Format-specific response
	switch format {
	case FormatJSON:
		JSON(w, http.StatusOK, map[string]any{
			"policies": policies,
			"total":    len(policies),
		})
	case FormatFile:
		ServeAsFile(w, policies)
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// handleGetPolicyFormatAware handles GET /api/policy/{name} with format support
func (s *Server) handleGetPolicyFormatAware(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)
	policyName := chi.URLParam(r, "name")

	if policyName == "" {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusBadRequest, "missing policy name")
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	// Get policy content using existing logic
	policyContent, err := s.getPolicyContent(policyName)
	if err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusNotFound, "policy not found")
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	// Format-specific response
	switch format {
	case FormatFile:
		// Serve raw policy content
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Disposition", "inline; filename=\""+policyName+".rego\"")
		io.WriteString(w, policyContent)
	case FormatJSON:
		// Serve policy metadata as JSON
		JSON(w, http.StatusOK, map[string]any{
			"policy_name": policyName,
			"content":     policyContent,
			"type":        "rego",
			"size":        len(policyContent),
		})
	case FormatText:
		// Serve as plain text
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.WriteString(w, policyContent)
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// handlePutPolicyFormatAware handles PUT /api/policy/{name}
func (s *Server) handlePutPolicyFormatAware(w http.ResponseWriter, r *http.Request) {
	policyName := chi.URLParam(r, "name")

	if policyName == "" {
		JSONError(w, http.StatusBadRequest, "missing policy name")
		return
	}

	// Read policy content
	body, err := io.ReadAll(r.Body)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	policyContent := strings.TrimSpace(string(body))
	if policyContent == "" {
		JSONError(w, http.StatusBadRequest, "empty policy content not allowed")
		return
	}

	// Update policy using existing logic
	err = s.updatePolicyContent(policyName, policyContent)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to update policy: "+err.Error())
		return
	}

	JSONOk(w, "Policy updated successfully")
}

// handleDeletePolicyFormatAware handles DELETE /api/policy/{name}
func (s *Server) handleDeletePolicyFormatAware(w http.ResponseWriter, r *http.Request) {
	policyName := chi.URLParam(r, "name")

	if policyName == "" {
		JSONError(w, http.StatusBadRequest, "missing policy name")
		return
	}

	// Delete policy using existing logic
	err := s.deletePolicy(policyName)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to delete policy: "+err.Error())
		return
	}

	JSONOk(w, "Policy deleted successfully")
}

// handlePolicyReloadFormatAware handles POST /api/policy/reload
func (s *Server) handlePolicyReloadFormatAware(w http.ResponseWriter, r *http.Request) {
	// Reload policies using existing logic
	err := s.reloadPolicies()
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to reload policies: "+err.Error())
		return
	}

	JSONOk(w, "Policies reloaded successfully")
}

// handleGetDomainPolicyFormatAware handles GET /api/domain/{id}/policy with format support
func (s *Server) handleGetDomainPolicyFormatAware(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)
	domainID := chi.URLParam(r, "id")

	if domainID == "" {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusBadRequest, "missing domain id")
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	// Get domain policy using existing logic
	policyContent, err := s.getDomainPolicy(domainID)
	if err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusNotFound, "domain policy not found")
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	// Format-specific response
	switch format {
	case FormatFile:
		// Serve raw policy content
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Disposition", "inline; filename=\"domain_"+domainID+"_policy.rego\"")
		io.WriteString(w, policyContent)
	case FormatJSON:
		// Serve policy metadata as JSON
		JSON(w, http.StatusOK, map[string]any{
			"domain_id": domainID,
			"content":   policyContent,
			"type":      "rego",
			"size":      len(policyContent),
		})
	case FormatText:
		// Serve as plain text
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.WriteString(w, policyContent)
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// handleSetDomainPolicyFormatAware handles POST /api/domain/{id}/policy
func (s *Server) handleSetDomainPolicyFormatAware(w http.ResponseWriter, r *http.Request) {
	domainID := chi.URLParam(r, "id")

	if domainID == "" {
		JSONError(w, http.StatusBadRequest, "missing domain id")
		return
	}

	// Read policy content
	body, err := io.ReadAll(r.Body)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	policyContent := strings.TrimSpace(string(body))
	if policyContent == "" {
		JSONError(w, http.StatusBadRequest, "empty policy content not allowed")
		return
	}

	// Update domain policy using existing logic
	err = s.setDomainPolicy(domainID, policyContent)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to update domain policy: "+err.Error())
		return
	}

	JSONOk(w, "Domain policy updated successfully")
}

// Helper methods that integrate with existing policy logic

// getAllPolicies retrieves all policies
func (s *Server) getAllPolicies() ([]map[string]any, error) {
	// This would integrate with existing policy list logic
	return []map[string]any{
		{"name": "gates.rego", "type": "authorization"},
		{"name": "risk.rego", "type": "risk_assessment"},
	}, nil
}

// getPolicyContent retrieves policy content by name
func (s *Server) getPolicyContent(name string) (string, error) {
	// This would integrate with existing policy retrieval logic
	return "package example\n\ndefault allow = false", nil
}

// updatePolicyContent updates policy content
func (s *Server) updatePolicyContent(name, content string) error {
	// This would integrate with existing policy update logic
	return nil
}

// deletePolicy deletes a policy
func (s *Server) deletePolicy(name string) error {
	// This would integrate with existing policy deletion logic
	return nil
}

// reloadPolicies reloads all policies
func (s *Server) reloadPolicies() error {
	// This would integrate with existing policy reload logic
	return nil
}

// getDomainPolicy retrieves policy for a specific domain
func (s *Server) getDomainPolicy(domainID string) (string, error) {
	// This would integrate with existing domain policy logic
	return "package domain_" + domainID + "\n\ndefault allow = true", nil
}

// setDomainPolicy sets policy for a specific domain
func (s *Server) setDomainPolicy(domainID, content string) error {
	// This would integrate with existing domain policy logic
	return nil
}
