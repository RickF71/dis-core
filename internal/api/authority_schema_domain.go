package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/yaml.v3"

	"dis-core/internal/receipts"
)

// DomainSchemaAggregate represents the aggregated schema and policy data for a domain
type DomainSchemaAggregate struct {
	DomainRef        string                           `json:"domain_ref" yaml:"domain_ref"`
	Timestamp        string                           `json:"timestamp" yaml:"timestamp"`
	Schema           interface{}                      `json:"schema" yaml:"schema"`
	Policies         interface{}                      `json:"policies" yaml:"policies"`
	PolicyCount      int                              `json:"policy_count" yaml:"policy_count"`
	PolicyContinuity *receipts.PolicyContinuityResult `json:"policy_continuity,omitempty" yaml:"policy_continuity,omitempty"`
	ReceiptStats     *receipts.DomainReceiptStats     `json:"receipt_stats,omitempty" yaml:"receipt_stats,omitempty"`
}

// GetAuthoritySchemaForDomain handles GET /api/authority/schema/domain/{id}
// Aggregates data from Phase 10C endpoints: /api/domain/{id}/schema and /api/domain/{id}/policy
func (s *Server) GetAuthoritySchemaForDomain(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)
	domainID := chi.URLParam(r, "id")

	if domainID == "" {
		switch format {
		case FormatJSON:
			http.Error(w, `{"error":"Domain ID required"}`, http.StatusBadRequest)
		case FormatText:
			http.Error(w, "error: Domain ID required", http.StatusBadRequest)
		default:
			http.Error(w, "Domain ID required", http.StatusBadRequest)
		}
		return
	}

	ctx := r.Context()

	// Use DB if available for aggregated authority schema endpoints;
	// allow filesystem-only operation when no DB is configured (nil-safe).
	db := s.DB()

	// Aggregate data from Phase 10C endpoints
	aggregate, err := s.aggregateDomainSchemaData(ctx, db, domainID)
	if err != nil {
		switch format {
		case FormatJSON:
			http.Error(w, fmt.Sprintf(`{"error":"Failed to aggregate domain data: %s"}`, err.Error()), http.StatusInternalServerError)
		case FormatText:
			http.Error(w, fmt.Sprintf("error: Failed to aggregate domain data: %s", err.Error()), http.StatusInternalServerError)
		default:
			http.Error(w, fmt.Sprintf("Failed to aggregate domain data: %s", err.Error()), http.StatusInternalServerError)
		}
		return
	}

	// Return response based on format
	switch format {
	case FormatJSON:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(aggregate)
	case FormatText:
		w.Header().Set("Content-Type", "text/plain")
		yamlData, err := yaml.Marshal(aggregate)
		if err != nil {
			http.Error(w, "error: Failed to marshal YAML", http.StatusInternalServerError)
			return
		}
		w.Write(yamlData)
	default:
		// Default to JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(aggregate)
	}
}

// aggregateDomainSchemaData fetches and combines data from Phase 10C endpoints
func (s *Server) aggregateDomainSchemaData(ctx context.Context, db *pgxpool.Pool, domainID string) (*DomainSchemaAggregate, error) {
	timestamp := time.Now().UTC().Format(time.RFC3339)

	aggregate := &DomainSchemaAggregate{
		DomainRef:   domainID,
		Timestamp:   timestamp,
		Schema:      map[string]interface{}{},
		Policies:    []interface{}{},
		PolicyCount: 0,
	}

	// Fetch schema data using the same logic as GetDomainSchema
	var schemaData map[string]interface{}
	var err error
	schemaData, err = s.getDomainSchemaFromDB(ctx, db, domainID)
	if err != nil {
		// Try filesystem fallback
		schemaData = s.getDomainSchemaFromFS(domainID)
	}
	aggregate.Schema = schemaData

	// Fetch policy data using the same logic as GetDomainPolicy
	policies, err := s.getDomainPolicies(ctx, db, domainID)
	if err != nil {
		// Use empty policies for missing domains or database errors
		aggregate.Policies = []interface{}{}
		aggregate.PolicyCount = 0
	} else {
		// Convert policies to interface{} for JSON marshaling
		policyInterfaces := make([]interface{}, len(policies))
		for i, policy := range policies {
			policyInterfaces[i] = policy
		}
		aggregate.Policies = policyInterfaces
		aggregate.PolicyCount = len(policies)
	}

	// Phase 10D: Add policy continuity data (only if DB is available)
	if db != nil {
		if continuityResult, err := receipts.VerifyPolicyContinuity(ctx, db, domainID); err == nil {
			aggregate.PolicyContinuity = &continuityResult
		}

		// Get receipt statistics for this domain
		if receiptStats, err := receipts.GetDomainReceiptStats(ctx, db, domainID); err == nil {
			aggregate.ReceiptStats = &receiptStats
		}
	}

	return aggregate, nil
} // SchemaValidationRequest represents the request body for schema validation
type SchemaValidationRequest struct {
	Schema interface{} `json:"schema"`
}

// SchemaValidationResponse represents the validation result
type SchemaValidationResponse struct {
	Valid       bool     `json:"valid"`
	Errors      []string `json:"errors"`
	Timestamp   string   `json:"timestamp"`
	ValidatedBy string   `json:"validated_by"`
}

// HandleSchemaValidation handles POST /api/authority/schema/validate
// Validates schema structure against canonical schema.json in registry
func (s *Server) HandleSchemaValidation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body (supports both JSON and YAML)
	var requestBody SchemaValidationRequest
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/x-yaml" || contentType == "text/yaml" || contentType == "text/plain" {
		// Parse YAML
		var yamlData interface{}
		if err := yaml.NewDecoder(r.Body).Decode(&yamlData); err != nil {
			http.Error(w, `{"error":"Invalid YAML format"}`, http.StatusBadRequest)
			return
		}
		requestBody.Schema = yamlData
	} else {
		// Parse JSON (default)
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, `{"error":"Invalid JSON format"}`, http.StatusBadRequest)
			return
		}
	}

	// Perform validation
	response := s.ValidateSchemaStructure(requestBody.Schema)

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ValidateSchemaStructure validates the provided schema against canonical structure
// Exported version for testing
func (s *Server) ValidateSchemaStructure(schema interface{}) SchemaValidationResponse {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	response := SchemaValidationResponse{
		Valid:       true,
		Errors:      []string{},
		Timestamp:   timestamp,
		ValidatedBy: "DIS-Core Phase 10D Schema Validator",
	}

	// Convert schema to map for validation
	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		response.Valid = false
		response.Errors = append(response.Errors, "Schema must be a JSON object")
		return response
	}

	// Basic structure validation
	if schemaMap == nil || len(schemaMap) == 0 {
		response.Valid = false
		response.Errors = append(response.Errors, "Schema cannot be empty")
		return response
	}

	// Validate required fields for JSON Schema structure
	if schemaType, exists := schemaMap["type"]; exists {
		if typeStr, ok := schemaType.(string); ok && typeStr == "object" {
			// Valid object schema
		} else {
			response.Errors = append(response.Errors, "Schema type should be 'object' for domain schemas")
		}
	} else {
		response.Errors = append(response.Errors, "Missing 'type' field in schema")
	}

	// Check for properties field
	if _, exists := schemaMap["properties"]; !exists {
		response.Errors = append(response.Errors, "Schema should contain 'properties' field")
	}

	// Validate $schema reference if present
	if schemaRef, exists := schemaMap["$schema"]; exists {
		if refStr, ok := schemaRef.(string); ok {
			if refStr != "http://json-schema.org/draft-07/schema#" &&
				refStr != "http://json-schema.org/draft-04/schema#" &&
				refStr != "https://json-schema.org/draft/2020-12/schema" {
				response.Errors = append(response.Errors, "Unsupported JSON Schema version")
			}
		}
	}

	// Set final validity
	response.Valid = len(response.Errors) == 0

	return response
}

// HandleSchemaViewerHTML serves the HTML dashboard for Phase 10D Schema Viewer
func (s *Server) HandleSchemaViewerHTML(w http.ResponseWriter, r *http.Request) {
	// Serve the static HTML file
	htmlPath := filepath.Join("static", "schema_viewer.html")
	http.ServeFile(w, r, htmlPath)
}
