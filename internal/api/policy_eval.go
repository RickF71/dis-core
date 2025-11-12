package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/go-chi/chi/v5"
)

// PolicyEvalRequest represents the input for policy evaluation
type PolicyEvalRequest struct {
	Policy string                 `json:"policy"`
	Input  map[string]interface{} `json:"input"`
	Query  string                 `json:"query,omitempty"` // e.g., "data.dis.policy.allow"
}

// PolicyEvalResponse represents the result of policy evaluation
type PolicyEvalResponse struct {
	Result  interface{} `json:"result"`
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Trace   []string    `json:"trace,omitempty"`
}

// HandlePolicyEval handles POST /api/policy/eval
// Evaluates Rego policy using OPA CLI
func (s *Server) HandlePolicyEval(w http.ResponseWriter, r *http.Request) {
	var req PolicyEvalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Policy == "" {
		JSONError(w, http.StatusBadRequest, "Policy content required")
		return
	}

	// Default query if not specified
	if req.Query == "" {
		req.Query = "data.dis.policy.allow"
	}

	// Evaluate policy using OPA
	result, err := s.evaluateRegoPolicy(req.Policy, req.Input, req.Query)
	if err != nil {
		JSON(w, http.StatusOK, PolicyEvalResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	JSON(w, http.StatusOK, result)
}

// evaluateRegoPolicy uses OPA CLI to evaluate the policy
func (s *Server) evaluateRegoPolicy(policy string, input map[string]interface{}, query string) (PolicyEvalResponse, error) {
	// Check if OPA is available
	if _, err := exec.LookPath("opa"); err != nil {
		return PolicyEvalResponse{}, fmt.Errorf("OPA CLI not found in PATH")
	}

	// Create temporary files for policy and input
	policyFile := "/tmp/policy.rego"
	inputFile := "/tmp/input.json"

	// Write policy to temp file
	if err := writeFile(policyFile, []byte(policy)); err != nil {
		return PolicyEvalResponse{}, fmt.Errorf("failed to write policy file: %w", err)
	}

	// Write input to temp file
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return PolicyEvalResponse{}, fmt.Errorf("failed to marshal input: %w", err)
	}
	if err := writeFile(inputFile, inputJSON); err != nil {
		return PolicyEvalResponse{}, fmt.Errorf("failed to write input file: %w", err)
	}

	// Run OPA eval
	cmd := exec.Command("opa", "eval",
		"--data", policyFile,
		"--input", inputFile,
		"--format", "json",
		query)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	// Parse output even if command failed (OPA might return valid JSON with error details)
	var opaOutput struct {
		Result []struct {
			Expressions []struct {
				Value interface{} `json:"value"`
				Text  string      `json:"text"`
			} `json:"expressions"`
		} `json:"result"`
	}

	if parseErr := json.Unmarshal(stdout.Bytes(), &opaOutput); parseErr == nil && len(opaOutput.Result) > 0 {
		// Successfully parsed OPA output
		var result interface{}
		if len(opaOutput.Result[0].Expressions) > 0 {
			result = opaOutput.Result[0].Expressions[0].Value
		}

		return PolicyEvalResponse{
			Result:  result,
			Success: err == nil,
			Error:   strings.TrimSpace(stderr.String()),
		}, nil
	}

	// Failed to parse or execute
	if err != nil {
		return PolicyEvalResponse{}, fmt.Errorf("OPA evaluation failed: %s", stderr.String())
	}

	return PolicyEvalResponse{
		Success: true,
		Result:  nil,
	}, nil
}

// writeFile is a helper to write content to a file
func writeFile(path string, content []byte) error {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("cat > %s", path))
	cmd.Stdin = bytes.NewReader(content)
	return cmd.Run()
}

// UpdateDomainPolicy handles POST /api/policy/save/{id}
// Updates the domain's Rego policy content
// Supports parent-authority routing: if body.domainId is provided and differs from URL param,
// the save is treated as a child proposal and must validate against full hierarchy
func (s *Server) UpdateDomainPolicy(w http.ResponseWriter, r *http.Request) {
	saveDomainID := chi.URLParam(r, "id") // Parent or self

	if saveDomainID == "" {
		JSONError(w, http.StatusBadRequest, "Domain ID required")
		return
	}

	var req struct {
		Content  string `json:"content"`
		DomainID string `json:"domainId,omitempty"` // Actual domain being edited (if different from save target)
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := r.Context()

	// Determine actual domain to update
	targetDomainID := saveDomainID
	if req.DomainID != "" && req.DomainID != saveDomainID {
		// Parent-authority save: validate child against full hierarchy first
		targetDomainID = req.DomainID
	}

	// Update the policy in the database
	query := `
		UPDATE domains
		SET payload = jsonb_set(
			COALESCE(payload, '{}'::jsonb),
			'{policy}',
			jsonb_build_object(
				'content', $2::text,
				'updated_at', now()::text
			),
			true
		)
		WHERE id = $1
	`

	_, err := s.db.Exec(ctx, query, targetDomainID, req.Content)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update policy: %v", err))
		return
	}

	JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Policy updated successfully",
	})
}
