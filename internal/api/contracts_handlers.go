package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"dis-core/internal/contracts"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// GOV-13: Contract API Endpoints

// ContractDTO represents a contract in API responses
type ContractDTO struct {
	ID              uuid.UUID  `json:"id"`
	DomainID        uuid.UUID  `json:"domain_id"`
	SubjectDomainID uuid.UUID  `json:"subject_domain_id"`
	AliasID         *uuid.UUID `json:"alias_id,omitempty"`
	ContractType    string     `json:"contract_type"`
	DSCIChannel     string     `json:"dsci_channel"`
	DSCIReference   string     `json:"dsci_reference"`
	DSCIVersion     string     `json:"dsci_version"`
	EffectiveAt     time.Time  `json:"effective_at"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	RevokedAt       *time.Time `json:"revoked_at,omitempty"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	CreatedBy       *uuid.UUID `json:"created_by,omitempty"`
}

// CreateContractRequest represents request to create a contract
type CreateContractRequest struct {
	SubjectDomainID string  `json:"subject_domain_id"`
	AliasID         *string `json:"alias_id,omitempty"`
	ContractType    string  `json:"contract_type"`
	DSCIChannel     string  `json:"dsci_channel"`
	DSCIReference   string  `json:"dsci_reference"`
	DSCIVersion     string  `json:"dsci_version"`
	EffectiveAt     string  `json:"effective_at"`
	ExpiresAt       *string `json:"expires_at,omitempty"`
	CreatedBy       *string `json:"created_by,omitempty"`
}

// ContractListResponse represents response for contract listing
type ContractListResponse struct {
	Contracts []ContractDTO `json:"contracts"`
	Count     int           `json:"count"`
}

// RevokeContractRequest represents request to revoke a contract
type RevokeContractRequest struct {
	Reason  string `json:"reason"`
	ActorID string `json:"actor_id"`
}

// toContractDTO converts a domain Contract to a ContractDTO
func toContractDTO(contract *contracts.Contract) ContractDTO {
	return ContractDTO{
		ID:              contract.ID,
		DomainID:        contract.DomainID,
		SubjectDomainID: contract.SubjectDomainID,
		AliasID:         contract.AliasID,
		ContractType:    contract.ContractType,
		DSCIChannel:     contract.DSCIChannel,
		DSCIReference:   contract.DSCIReference,
		DSCIVersion:     contract.DSCIVersion,
		EffectiveAt:     contract.EffectiveAt,
		ExpiresAt:       contract.ExpiresAt,
		RevokedAt:       contract.RevokedAt,
		Status:          string(contract.Status),
		CreatedAt:       contract.CreatedAt,
		UpdatedAt:       contract.UpdatedAt,
		CreatedBy:       contract.CreatedBy,
	}
}

// HandleCreateContract handles POST /api/domain/:id/contracts
// Creates a new DSCI contract issued by the domain
func (s *Server) HandleCreateContract(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	domainIDStr := chi.URLParam(r, "id")

	// Validate domain ID
	domainID, err := uuid.Parse(domainIDStr)
	if err != nil {
		http.Error(w, "invalid domain ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req CreateContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.SubjectDomainID == "" {
		http.Error(w, "subject_domain_id is required", http.StatusBadRequest)
		return
	}
	if req.ContractType == "" {
		http.Error(w, "contract_type is required", http.StatusBadRequest)
		return
	}
	if req.DSCIChannel == "" {
		http.Error(w, "dsci_channel is required", http.StatusBadRequest)
		return
	}
	if req.DSCIReference == "" {
		http.Error(w, "dsci_reference is required", http.StatusBadRequest)
		return
	}
	if req.DSCIVersion == "" {
		http.Error(w, "dsci_version is required", http.StatusBadRequest)
		return
	}
	if req.EffectiveAt == "" {
		http.Error(w, "effective_at is required", http.StatusBadRequest)
		return
	}

	// Parse subject_domain_id
	subjectDomainID, err := uuid.Parse(req.SubjectDomainID)
	if err != nil {
		http.Error(w, "invalid subject_domain_id format", http.StatusBadRequest)
		return
	}

	// Parse effective_at
	effectiveAt, err := time.Parse(time.RFC3339, req.EffectiveAt)
	if err != nil {
		http.Error(w, "invalid effective_at format (must be RFC3339)", http.StatusBadRequest)
		return
	}

	// Parse optional fields
	var aliasID *uuid.UUID
	if req.AliasID != nil && *req.AliasID != "" {
		parsed, err := uuid.Parse(*req.AliasID)
		if err != nil {
			http.Error(w, "invalid alias_id format", http.StatusBadRequest)
			return
		}
		aliasID = &parsed
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		parsed, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			http.Error(w, "invalid expires_at format (must be RFC3339)", http.StatusBadRequest)
			return
		}
		expiresAt = &parsed
	}

	var createdBy *uuid.UUID
	if req.CreatedBy != nil && *req.CreatedBy != "" {
		parsed, err := uuid.Parse(*req.CreatedBy)
		if err != nil {
			http.Error(w, "invalid created_by format", http.StatusBadRequest)
			return
		}
		createdBy = &parsed
	}

	// Build input
	input := contracts.CreateContractInput{
		DomainID:        domainID,
		SubjectDomainID: subjectDomainID,
		AliasID:         aliasID,
		ContractType:    req.ContractType,
		DSCIChannel:     req.DSCIChannel,
		DSCIReference:   req.DSCIReference,
		DSCIVersion:     req.DSCIVersion,
		EffectiveAt:     effectiveAt,
		ExpiresAt:       expiresAt,
		CreatedBy:       createdBy,
	}

	// Create contract via service
	contract, err := s.contractService.CreateContract(ctx, input)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create contract: %v", err), http.StatusInternalServerError)
		return
	}

	// Return created contract
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toContractDTO(contract))
}

// HandleGetDomainContracts handles GET /api/domain/:id/contracts
// Lists all contracts issued by a specific domain
func (s *Server) HandleGetDomainContracts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	domainIDStr := chi.URLParam(r, "id")

	// Validate domain ID
	if _, err := uuid.Parse(domainIDStr); err != nil {
		http.Error(w, "invalid domain ID", http.StatusBadRequest)
		return
	}

	// Get optional status filter
	statusFilter := r.URL.Query().Get("status")

	// Retrieve contracts via service
	contractList, err := s.contractService.ListDomainContracts(ctx, domainIDStr, statusFilter)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to retrieve contracts: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to DTOs
	dtos := make([]ContractDTO, 0, len(contractList))
	for _, contract := range contractList {
		dtos = append(dtos, toContractDTO(contract))
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ContractListResponse{
		Contracts: dtos,
		Count:     len(dtos),
	})
}

// HandleGetContract handles GET /api/domain/:id/contracts/:contractId
// Retrieves a single contract by ID
func (s *Server) HandleGetContract(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	contractIDStr := chi.URLParam(r, "contractId")

	// Validate contract ID
	if _, err := uuid.Parse(contractIDStr); err != nil {
		http.Error(w, "invalid contract ID", http.StatusBadRequest)
		return
	}

	// Retrieve contract via service
	contract, err := s.contractService.GetContract(ctx, contractIDStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("contract not found: %v", err), http.StatusNotFound)
		return
	}

	// Return contract
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toContractDTO(contract))
}

// HandleRevokeContract handles POST /api/domain/:id/contracts/:contractId/revoke
// Revokes an active contract
func (s *Server) HandleRevokeContract(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	contractIDStr := chi.URLParam(r, "contractId")

	// Validate contract ID
	if _, err := uuid.Parse(contractIDStr); err != nil {
		http.Error(w, "invalid contract ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req RevokeContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Reason == "" {
		http.Error(w, "reason is required", http.StatusBadRequest)
		return
	}
	if req.ActorID == "" {
		http.Error(w, "actor_id is required", http.StatusBadRequest)
		return
	}

	// Revoke contract via service
	contract, err := s.contractService.RevokeContract(ctx, contractIDStr, req.Reason, req.ActorID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to revoke contract: %v", err), http.StatusInternalServerError)
		return
	}

	// Return revoked contract
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toContractDTO(contract))
}
