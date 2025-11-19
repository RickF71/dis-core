package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"dis-core/internal/identity"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// handleGetIdentitySchema returns the schema set for a domain
// GET /api/identity/schema/:domainId
func (s *Server) handleGetIdentitySchema(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	domainID := chi.URLParam(r, "domainId")

	if domainID == "" {
		http.Error(w, "domain_id required", http.StatusBadRequest)
		return
	}

	resolver := identity.NewSchemaResolver(s.db)
	schemaSet, err := resolver.ResolveSchemaSet(ctx, domainID)
	if err != nil {
		s.logger.Printf("Error resolving schema set for domain %s: %v", domainID, err)
		http.Error(w, fmt.Sprintf("failed to resolve schema set: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schemaSet)
}

// AdoptSchemaRequest represents a schema adoption request
type AdoptSchemaRequest struct {
	SchemaID      string `json:"schema_id"`
	SchemaVersion string `json:"schema_version"`
	ActorID       string `json:"actor_id,omitempty"`
}

// handleAdoptSchema adopts a schema for a domain
// POST /api/identity/schema/:domainId/adopt
func (s *Server) handleAdoptSchema(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	domainID := chi.URLParam(r, "domainId")

	if domainID == "" {
		http.Error(w, "domain_id required", http.StatusBadRequest)
		return
	}

	var req AdoptSchemaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.SchemaID == "" || req.SchemaVersion == "" {
		http.Error(w, "schema_id and schema_version required", http.StatusBadRequest)
		return
	}

	// Check compatibility
	resolver := identity.NewSchemaResolver(s.db)
	if err := resolver.CheckSchemaCompatibility(ctx, domainID, req.SchemaID, req.SchemaVersion); err != nil {
		s.logger.Printf("Schema compatibility check failed for %s: %v", domainID, err)
		http.Error(w, fmt.Sprintf("schema incompatible: %v", err), http.StatusConflict)
		return
	}

	// Adopt schema
	var adoptedBy *string
	if req.ActorID != "" {
		adoptedBy = &req.ActorID
	}

	err := s.db.QueryRow(ctx, `
		INSERT INTO domain_schemas (domain_id, schema_id, schema_version, adopted_by, adopted_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (domain_id, schema_id, schema_version) DO NOTHING
		RETURNING adopted_at
	`, domainID, req.SchemaID, req.SchemaVersion, adoptedBy, time.Now()).Scan(&time.Time{})
	if err != nil {
		s.logger.Printf("Error adopting schema for domain %s: %v", domainID, err)
		http.Error(w, fmt.Sprintf("failed to adopt schema: %v", err), http.StatusInternalServerError)
		return
	}

	// Record authority receipt: identity.schema.adopt.v1
	if s.identityReceiptRecorder != nil {
		domainUUID, err := uuid.Parse(domainID)
		if err != nil {
			s.logger.Printf("Warning: invalid domain UUID %s: %v", domainID, err)
		} else {
			metadata := map[string]interface{}{
				"schema_id":      req.SchemaID,
				"schema_version": req.SchemaVersion,
				"domain_id":      domainID,
			}

			var actorUUID *uuid.UUID
			if adoptedBy != nil {
				parsed, err := uuid.Parse(*adoptedBy)
				if err == nil {
					actorUUID = &parsed
					metadata["adopted_by"] = *adoptedBy
				}
			}

			receipt, err := s.identityReceiptRecorder.RecordSchemaAdoption(ctx, domainUUID, req.SchemaID, req.SchemaVersion, actorUUID, metadata)
			if err != nil {
				s.logger.Printf("Warning: failed to record schema adoption receipt: %v", err)
			} else {
				s.logger.Printf("Schema adoption receipt recorded: %s", receipt.ID.String())
			}
		}
	}

	// Return updated schema set
	schemaSet, err := resolver.ResolveSchemaSet(ctx, domainID)
	if err != nil {
		s.logger.Printf("Error resolving updated schema set: %v", err)
		http.Error(w, fmt.Sprintf("schema adopted but failed to retrieve updated set: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(schemaSet)
}

// SavePolicyRequest represents a policy save request
type SavePolicyRequest struct {
	Policy  map[string]interface{} `json:"policy"`
	ActorID string                 `json:"actor_id,omitempty"`
	Comment string                 `json:"comment,omitempty"`
}

// SavePolicyResponse represents the response after saving a policy
type SavePolicyResponse struct {
	Success         bool                       `json:"success"`
	Validation      *identity.ValidationResult `json:"validation,omitempty"`
	EffectivePolicy map[string]interface{}     `json:"effective_policy,omitempty"`
	UpdatedAt       string                     `json:"updated_at,omitempty"`
	ReceiptID       string                     `json:"receipt_id,omitempty"`
}

// handleSaveIdentityPolicy saves/updates identity policy for a domain
// POST /api/identity/policy/:domainId
func (s *Server) handleSaveIdentityPolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	domainID := chi.URLParam(r, "domainId")

	if domainID == "" {
		http.Error(w, "domain_id required", http.StatusBadRequest)
		return
	}

	var req SavePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Load effective schema
	resolver := identity.NewSchemaResolver(s.db)
	schemaSet, err := resolver.ResolveSchemaSet(ctx, domainID)
	if err != nil {
		s.logger.Printf("Error resolving schema set for domain %s: %v", domainID, err)
		http.Error(w, fmt.Sprintf("failed to resolve schema set: %v", err), http.StatusInternalServerError)
		return
	}

	// Validate policy against schema
	validator := identity.NewPolicyValidator(resolver)
	validationResult := validator.ValidatePolicy(domainID, req.Policy, schemaSet)

	if !validationResult.Valid {
		// Return validation errors
		response := SavePolicyResponse{
			Success:    false,
			Validation: validationResult,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Check editability - only allow editing fields from adopted schemas
	editableFields := validator.CheckEditableFields(schemaSet)
	for fieldName := range req.Policy {
		if !editableFields[fieldName] {
			// Field is from inherited-only schema, cannot be edited
			response := SavePolicyResponse{
				Success: false,
				Validation: &identity.ValidationResult{
					Valid: false,
					Errors: []identity.ValidationError{
						{
							Field:   fieldName,
							Message: fmt.Sprintf("Field '%s' cannot be edited (not in adopted schema)", fieldName),
							Type:    "unadopted_field",
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// Save policy to domain payload
	if err := s.saveLocalIdentityPolicy(ctx, domainID, req.Policy); err != nil {
		s.logger.Printf("Error saving policy for domain %s: %v", domainID, err)
		http.Error(w, fmt.Sprintf("failed to save policy: %v", err), http.StatusInternalServerError)
		return
	}

	// Compute effective policy (merge parent + local with shadowing)
	effectivePolicy, err := s.computeEffectiveIdentityPolicy(ctx, domainID)
	if err != nil {
		s.logger.Printf("Warning: failed to compute effective policy: %v", err)
		effectivePolicy = req.Policy // Fallback to local only
	}

	updatedAt := time.Now()

	// Record authority receipt: identity.policy.update.v1
	var receiptID string
	if s.identityReceiptRecorder != nil {
		domainUUID, err := uuid.Parse(domainID)
		if err != nil {
			s.logger.Printf("Warning: invalid domain UUID %s: %v", domainID, err)
		} else {
			metadata := map[string]interface{}{
				"domain_id": domainID,
				"comment":   req.Comment,
			}

			var actorUUID *uuid.UUID
			if req.ActorID != "" {
				parsed, err := uuid.Parse(req.ActorID)
				if err == nil {
					actorUUID = &parsed
					metadata["updated_by"] = req.ActorID
				}
			}

			receipt, err := s.identityReceiptRecorder.RecordPolicyUpdate(ctx, domainUUID, actorUUID, metadata)
			if err != nil {
				s.logger.Printf("Warning: failed to record policy update receipt: %v", err)
			} else {
				receiptID = receipt.ID.String()
				s.logger.Printf("Policy update receipt recorded: %s", receiptID)
			}
		}
	}

	// Return success response
	response := SavePolicyResponse{
		Success:         true,
		Validation:      validationResult,
		EffectivePolicy: effectivePolicy,
		UpdatedAt:       updatedAt.Format(time.RFC3339),
		ReceiptID:       receiptID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// saveLocalIdentityPolicy saves the local identity policy to domain payload
func (s *Server) saveLocalIdentityPolicy(ctx context.Context, domainID string, policy map[string]interface{}) error {
	// Fetch current payload
	var payloadJSON []byte
	err := s.db.QueryRow(ctx, "SELECT payload FROM domains WHERE id = $1", domainID).Scan(&payloadJSON)
	if err != nil {
		return fmt.Errorf("domain not found: %w", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	// Update policy.identity_v1
	if payload["policy"] == nil {
		payload["policy"] = make(map[string]interface{})
	}
	policyBlock := payload["policy"].(map[string]interface{})
	policyBlock["identity_v1"] = policy

	// Save back
	updatedPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	_, err = s.db.Exec(ctx, `
		UPDATE domains
		SET payload = $1, updated_at = $2
		WHERE id = $3
	`, updatedPayload, time.Now(), domainID)
	if err != nil {
		return fmt.Errorf("failed to update domain: %w", err)
	}

	return nil
}

// computeEffectiveIdentityPolicy computes merged effective policy with parent inheritance
func (s *Server) computeEffectiveIdentityPolicy(ctx context.Context, domainID string) (map[string]interface{}, error) {
	// Get local policy
	localPolicy, err := s.getDomainIdentityPolicy(ctx, domainID)
	if err != nil {
		return nil, err
	}

	// Get parent policy
	var parentID string
	err = s.db.QueryRow(ctx, "SELECT parent_id FROM domains WHERE id = $1", domainID).Scan(&parentID)
	if err != nil || parentID == "" {
		// No parent, return local only
		return localPolicy, nil
	}

	parentPolicy, err := s.getDomainIdentityPolicy(ctx, parentID)
	if err != nil {
		// Parent has no policy, return local only
		return localPolicy, nil
	}

	// Merge: parent first, local overrides (shadowing)
	effective := make(map[string]interface{})
	for k, v := range parentPolicy {
		effective[k] = v
	}
	for k, v := range localPolicy {
		effective[k] = v
	}

	return effective, nil
}
