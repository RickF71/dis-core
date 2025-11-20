package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PolicyValidationResponse represents the result of hierarchical Rego validation
type PolicyValidationResponse struct {
	Success bool             `json:"success"`
	Error   string           `json:"error,omitempty"`
	Details string           `json:"details,omitempty"`
	Hints   []ValidationHint `json:"hints,omitempty"`
}

// ValidationHint represents a specific error or suggestion for policy correction
type ValidationHint struct {
	Line         int    `json:"line"`
	Column       int    `json:"column,omitempty"`
	Message      string `json:"message"`
	File         string `json:"file,omitempty"`
	SuggestedFix string `json:"suggested_fix,omitempty"`
}

// ValidateDomainPolicy handles POST /api/policy/validate/{id}
// Validates the domain's Rego policy in hierarchical context (parent → child)
func (s *Server) ValidateDomainPolicy(w http.ResponseWriter, r *http.Request) {
	domainID := chi.URLParam(r, "id")

	if domainID == "" {
		JSONError(w, http.StatusBadRequest, "Domain ID required")
		return
	}

	// Accept optional policy content in request body (for validation before save)
	var req struct {
		Content string `json:"content"`
	}
	json.NewDecoder(r.Body).Decode(&req) // Ignore decode errors - content is optional

	ctx := r.Context()

	// Require DB for policy validation
	db := s.requireDB(w)
	if db == nil {
		return
	}

	// Build and validate the hierarchical policy bundle
	result, err := s.validateHierarchicalPolicy(ctx, db, domainID, req.Content)
	if err != nil {
		JSON(w, http.StatusOK, PolicyValidationResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	JSON(w, http.StatusOK, result)
}

// validateHierarchicalPolicy builds a Rego bundle with parent-first ordering and validates it
// If newContent is provided, it will be used for the target domain instead of database value
func (s *Server) validateHierarchicalPolicy(ctx context.Context, db *pgxpool.Pool, domainID string, newContent string) (PolicyValidationResponse, error) {
	// Get domain lineage (child → parent → ... → null)
	lineage, err := s.getDomainLineage(ctx, db, domainID)
	if err != nil {
		return PolicyValidationResponse{}, fmt.Errorf("failed to get domain lineage: %w", err)
	}

	// Create temp directory for bundle
	bundleDir, err := os.MkdirTemp("", "rego-bundle-*")
	if err != nil {
		return PolicyValidationResponse{}, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(bundleDir)

	// Collect policies in parent-first order (reverse lineage)
	policies := make(map[string]string)
	for i := len(lineage) - 1; i >= 0; i-- {
		domainIDInLineage := lineage[i]

		var policy string
		// Use newContent for the target domain if provided
		if domainIDInLineage == domainID && newContent != "" {
			policy = newContent
		} else {
			policy, err = s.getDomainRegoPolicy(ctx, db, domainIDInLineage)
			if err != nil {
				continue // Skip domains without policy
			}
		}

		if policy != "" {
			// Use index to ensure parent-first ordering
			fileName := fmt.Sprintf("%d_%s.rego", len(lineage)-1-i, domainIDInLineage[:8])
			policies[fileName] = policy
		}
	}

	// If no policies found, that's valid (empty policy = default deny)
	if len(policies) == 0 {
		return PolicyValidationResponse{
			Success: true,
			Details: "No policies defined (default deny)",
		}, nil
	}

	// Write all policy files to bundle directory
	for fileName, content := range policies {
		filePath := filepath.Join(bundleDir, fileName)
		if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
			return PolicyValidationResponse{}, fmt.Errorf("failed to write policy file %s: %w", fileName, err)
		}
	}

	// Run OPA validation with --fail and --format=json flags
	cmd := exec.Command("opa", "eval",
		"--fail",
		"--format", "json",
		"--bundle", bundleDir,
		"data")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	// Try to parse structured errors from stdout (OPA returns JSON even on error)
	var opaOutput struct {
		Errors []struct {
			Code     string `json:"code"`
			Message  string `json:"message"`
			Location struct {
				File string `json:"file"`
				Row  int    `json:"row"`
				Col  int    `json:"col"`
			} `json:"location"`
		} `json:"errors"`
	}

	hints := []ValidationHint{}
	parseErr := json.Unmarshal(stdout.Bytes(), &opaOutput)

	if err != nil {
		// Validation failed - extract hints from structured errors
		if parseErr == nil && len(opaOutput.Errors) > 0 {
			for _, opaErr := range opaOutput.Errors {
				hint := ValidationHint{
					Line:    opaErr.Location.Row,
					Column:  opaErr.Location.Col,
					Message: opaErr.Message,
					File:    filepath.Base(opaErr.Location.File),
				}

				// Add suggested fixes for common errors
				hint.SuggestedFix = s.suggestFix(opaErr.Code, opaErr.Message)

				// Detect missing parent policy imports
				if s.isMissingImportError(opaErr.Message) {
					hint.Message += " (Hint: Parent policy may not be loaded in bundle)"
				}

				hints = append(hints, hint)
			}
		}

		errorMsg := stderr.String()
		if errorMsg == "" {
			errorMsg = err.Error()
		}

		return PolicyValidationResponse{
			Success: false,
			Error:   "Rego validation failed",
			Details: errorMsg,
			Hints:   hints,
		}, nil
	}

	// Validation passed
	return PolicyValidationResponse{
		Success: true,
		Details: fmt.Sprintf("Validated %d policy files in hierarchy", len(policies)),
	}, nil
}

// suggestFix returns a suggested fix for common Rego syntax errors
func (s *Server) suggestFix(code string, message string) string {
	// Missing assignment operator
	if code == "rego_parse_error" {
		if bytes.Contains([]byte(message), []byte("expected")) && bytes.Contains([]byte(message), []byte(":=")) {
			return "Use := for assignment instead of ="
		}
		if bytes.Contains([]byte(message), []byte("`if` keyword is required")) {
			return "Add 'if' keyword before rule body: rule_name if { ... }"
		}
		if bytes.Contains([]byte(message), []byte("unexpected")) {
			return "Check syntax - ensure all rules use Rego v1 syntax with := and if"
		}
	}

	// Undefined reference
	if code == "rego_type_error" && bytes.Contains([]byte(message), []byte("undefined")) {
		return "Ensure all referenced packages are imported or defined in parent policies"
	}

	return ""
}

// isMissingImportError detects if an error is related to missing parent policy imports
func (s *Server) isMissingImportError(message string) bool {
	return bytes.Contains([]byte(message), []byte("undefined")) &&
		(bytes.Contains([]byte(message), []byte("data.")) || bytes.Contains([]byte(message), []byte("import")))
}
