package api

import (
	"encoding/json"
	"net/http"
	"time"

	"dis-core/internal/domain"
	"dis-core/internal/identity"
	"dis-core/internal/repo"
	"dis-core/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// GOV-12: Alias Canon & DSCI Integration - API Endpoints

// AliasDTO represents an alias in API responses
type AliasDTO struct {
	ID              uuid.UUID              `json:"id"`
	OwnerDomainID   uuid.UUID              `json:"owner_domain_id"`
	TargetDomainID  uuid.UUID              `json:"target_domain_id"`
	AliasName       string                 `json:"alias_name"`
	AliasType       string                 `json:"alias_type"`
	IsCorporealAuto bool                   `json:"is_corporeal_auto"`
	DSCIContractID  *uuid.UUID             `json:"dsci_contract_id,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	RetiredAt       *time.Time             `json:"retired_at,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Status          string                 `json:"status"`
}

// CreateRelationshipAliasRequest represents request to create a RELATIONSHIP alias
type CreateRelationshipAliasRequest struct {
	TargetDomainID string                 `json:"target_domain_id"`
	AliasName      string                 `json:"alias_name"`
	DSCIContractID *string                `json:"dsci_contract_id,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// CreateMaskAliasRequest represents request to create a MASK alias
type CreateMaskAliasRequest struct {
	RequestedName string                 `json:"requested_name,omitempty"`
	TTLDays       *int                   `json:"ttl_days,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// AliasListResponse represents response for alias listing
type AliasListResponse struct {
	Aliases []AliasDTO `json:"aliases"`
	Count   int        `json:"count"`
}

// toAliasDTO converts domain.Alias to AliasDTO
func toAliasDTO(alias *domain.Alias) AliasDTO {
	dto := AliasDTO{
		ID:              alias.ID,
		OwnerDomainID:   alias.OwnerDomainID,
		TargetDomainID:  alias.TargetDomainID,
		AliasName:       alias.AliasName,
		AliasType:       string(alias.AliasType),
		IsCorporealAuto: alias.IsCorporealAuto,
		DSCIContractID:  alias.DSCIContractID,
		CreatedAt:       alias.CreatedAt,
		RetiredAt:       alias.RetiredAt,
		Metadata:        alias.Metadata,
		Status:          string(alias.Status()),
	}
	return dto
}

// HandleGetDomainAliases retrieves all aliases for a domain (as owner or target)
// GET /api/domain/:id/aliases?includeRetired=false&type=MASK
func (s *Server) HandleGetDomainAliases(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	domainIDStr := chi.URLParam(r, "id")

	domainID, err := uuid.Parse(domainIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid domain ID"}`, http.StatusBadRequest)
		return
	}

	// Parse query parameters
	includeRetired := r.URL.Query().Get("includeRetired") == "true"
	typeFilter := r.URL.Query().Get("type")

	// Create alias service (lazy initialization pattern)
	aliasRepo := repo.NewAliasRepository(s.db)
	aliasService := services.NewAliasService(aliasRepo)

	// Retrieve aliases
	aliases, err := aliasService.GetAliasesForDomain(ctx, domainID, includeRetired)
	if err != nil {
		http.Error(w, `{"error":"failed to retrieve aliases"}`, http.StatusInternalServerError)
		return
	}

	// Filter by type if specified
	var filteredAliases []domain.Alias
	if typeFilter != "" {
		aliasType := domain.AliasType(typeFilter)
		if !aliasType.IsValid() {
			http.Error(w, `{"error":"invalid alias type"}`, http.StatusBadRequest)
			return
		}
		for _, alias := range aliases {
			if alias.AliasType == aliasType {
				filteredAliases = append(filteredAliases, alias)
			}
		}
	} else {
		filteredAliases = aliases
	}

	// Convert to DTOs
	dtos := make([]AliasDTO, len(filteredAliases))
	for i, alias := range filteredAliases {
		dtos[i] = toAliasDTO(&alias)
	}

	response := AliasListResponse{
		Aliases: dtos,
		Count:   len(dtos),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleCreateRelationshipAlias creates a RELATIONSHIP alias
// POST /api/domain/:id/alias/relationship
//
// GOV-12: RELATIONSHIP aliases are volitional, domain-scoped identities.
// The :id is the owner domain ID (user's identity/corporeal domain).
// Requires alias_name and target_domain_id.
func (s *Server) HandleCreateRelationshipAlias(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerDomainIDStr := chi.URLParam(r, "id")

	ownerDomainID, err := uuid.Parse(ownerDomainIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid owner domain ID"}`, http.StatusBadRequest)
		return
	}

	// Parse request body
	var req CreateRelationshipAliasRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.AliasName == "" {
		http.Error(w, `{"error":"alias_name is required"}`, http.StatusBadRequest)
		return
	}
	if req.TargetDomainID == "" {
		http.Error(w, `{"error":"target_domain_id is required"}`, http.StatusBadRequest)
		return
	}

	targetDomainID, err := uuid.Parse(req.TargetDomainID)
	if err != nil {
		http.Error(w, `{"error":"invalid target_domain_id"}`, http.StatusBadRequest)
		return
	}

	// Parse optional DSCI contract ID
	var contractID *uuid.UUID
	if req.DSCIContractID != nil && *req.DSCIContractID != "" {
		parsedContractID, err := uuid.Parse(*req.DSCIContractID)
		if err != nil {
			http.Error(w, `{"error":"invalid dsci_contract_id"}`, http.StatusBadRequest)
			return
		}
		contractID = &parsedContractID
	}

	// TODO: Authorization check - verify caller has authority to create alias for owner domain
	// This should follow the same pattern as other domain-scoped endpoints.
	// For now, we allow the operation (development phase).

	// Create alias service
	aliasRepo := repo.NewAliasRepository(s.db)
	aliasService := services.NewAliasService(aliasRepo)

	// Create RELATIONSHIP alias
	metadata := domain.AliasMetadata(req.Metadata)
	alias, err := aliasService.CreateRelationshipAlias(
		ctx,
		ownerDomainID,
		targetDomainID,
		req.AliasName,
		contractID,
		metadata,
	)
	if err != nil {
		// Check for constraint violations (name conflict)
		if err.Error() == "alias already exists for this owner, target, and name" {
			http.Error(w, `{"error":"alias name already exists for this domain pair"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error":"failed to create alias"}`, http.StatusInternalServerError)
		return
	}

	// Record identity receipt (GOV-10 integration)
	if s.identityReceiptRecorder != nil {
		receipt := &identity.IdentityReceipt{
			DomainID:  ownerDomainID,
			ActorID:   ownerDomainID, // TODO: Use actual seat/actor from auth context
			Action:    identity.IdentityAliasAddV1,
			ConsentBy: ownerDomainID, // TODO: Use actual consent seat
			Payload: map[string]interface{}{
				"alias_id":         alias.ID.String(),
				"owner_domain_id":  alias.OwnerDomainID.String(),
				"target_domain_id": alias.TargetDomainID.String(),
				"alias_type":       string(alias.AliasType),
				"alias_name":       alias.AliasName,
			},
		}
		if alias.DSCIContractID != nil {
			receipt.Payload["dsci_contract_id"] = alias.DSCIContractID.String()
		}
		// Non-blocking: log error but don't fail the request
		if err := s.identityReceiptRecorder.RecordIdentityReceipt(ctx, receipt); err != nil {
			s.logger.Printf("Warning: Failed to record alias creation receipt: %v", err)
		}
	}

	// Respond with created alias
	dto := toAliasDTO(alias)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto)
}

// HandleCreateMaskAlias creates a MASK alias
// POST /api/domain/:id/alias/mask
//
// GOV-12: MASK aliases are ephemeral browsing personas with minimal authority.
// The :id is the owner domain ID (self-targeting: owner = target).
func (s *Server) HandleCreateMaskAlias(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ownerDomainIDStr := chi.URLParam(r, "id")

	ownerDomainID, err := uuid.Parse(ownerDomainIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid owner domain ID"}`, http.StatusBadRequest)
		return
	}

	// Parse request body
	var req CreateMaskAliasRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Extract TTL days (if provided)
	ttlDays := 0
	if req.TTLDays != nil {
		ttlDays = *req.TTLDays
	}

	// TODO: Authorization check - verify caller has authority to create mask for owner domain

	// Create alias service
	aliasRepo := repo.NewAliasRepository(s.db)
	aliasService := services.NewAliasService(aliasRepo)

	// Create MASK alias
	metadata := domain.AliasMetadata(req.Metadata)
	alias, err := aliasService.CreateMaskAlias(
		ctx,
		ownerDomainID,
		req.RequestedName,
		ttlDays,
		metadata,
	)
	if err != nil {
		http.Error(w, `{"error":"failed to create mask alias"}`, http.StatusInternalServerError)
		return
	}

	// Record identity receipt (GOV-10 integration)
	if s.identityReceiptRecorder != nil {
		receipt := &identity.IdentityReceipt{
			DomainID:  ownerDomainID,
			ActorID:   ownerDomainID, // TODO: Use actual seat/actor from auth context
			Action:    identity.IdentityAliasAddV1,
			ConsentBy: ownerDomainID, // TODO: Use actual consent seat
			Payload: map[string]interface{}{
				"alias_id":         alias.ID.String(),
				"owner_domain_id":  alias.OwnerDomainID.String(),
				"target_domain_id": alias.TargetDomainID.String(),
				"alias_type":       string(alias.AliasType),
				"alias_name":       alias.AliasName,
			},
		}
		if ttlDays > 0 {
			receipt.Payload["ttl_days"] = ttlDays
		}
		// Non-blocking: log error but don't fail the request
		if err := s.identityReceiptRecorder.RecordIdentityReceipt(ctx, receipt); err != nil {
			s.logger.Printf("Warning: Failed to record mask creation receipt: %v", err)
		}
	}

	// Respond with created alias
	dto := toAliasDTO(alias)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto)
}
