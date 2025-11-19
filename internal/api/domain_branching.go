package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"dis-core/internal/authority"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// GOV-11H: Domain Branching API Endpoints

// BranchDomainRequest represents a request to create a branch of a domain
type BranchDomainRequest struct {
	NewParentID uuid.UUID              `json:"new_parent_id"`
	BranchName  string                 `json:"branch_name"`
	Reason      string                 `json:"reason"`
	CreatedBy   uuid.UUID              `json:"created_by"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// BranchDomainResponse represents the response from branch creation
type BranchDomainResponse struct {
	Success        bool                  `json:"success"`
	BranchDomainID uuid.UUID             `json:"branch_domain_id"`
	OriginalDomain uuid.UUID             `json:"original_domain_id"`
	NewParent      uuid.UUID             `json:"new_parent_id"`
	BranchDepth    int                   `json:"branch_depth"`
	ReceiptID      string                `json:"receipt_id,omitempty"`
	Message        string                `json:"message,omitempty"`
	BranchInfo     *authority.BranchInfo `json:"branch_info,omitempty"`
}

// handleBranchDomain creates a new domain as a branch of an existing domain
// POST /api/domain/:id/branch
//
// GOV-11H: Branching is the ONLY mechanism for domain realignment.
// This creates a new domain with:
// - branch_of pointing to original domain
// - branch_depth = original.branch_depth + 1
// - new parent_id as specified
// - emits domain.branch.v1 receipt
func (s *Server) handleBranchDomain(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	domainIDStr := chi.URLParam(r, "id")

	originalDomainID, err := uuid.Parse(domainIDStr)
	if err != nil {
		http.Error(w, "invalid domain ID", http.StatusBadRequest)
		return
	}

	var req BranchDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.NewParentID == uuid.Nil {
		http.Error(w, "new_parent_id is required", http.StatusBadRequest)
		return
	}
	if req.BranchName == "" {
		http.Error(w, "branch_name is required", http.StatusBadRequest)
		return
	}
	if req.Reason == "" {
		http.Error(w, "reason is required", http.StatusBadRequest)
		return
	}
	if req.CreatedBy == uuid.Nil {
		http.Error(w, "created_by is required", http.StatusBadRequest)
		return
	}

	// Verify original domain exists
	var exists bool
	err = s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM domains WHERE id = $1)", originalDomainID).Scan(&exists)
	if err != nil || !exists {
		http.Error(w, "original domain not found", http.StatusNotFound)
		return
	}

	// Verify new parent exists
	err = s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM domains WHERE id = $1)", req.NewParentID).Scan(&exists)
	if err != nil || !exists {
		http.Error(w, "new parent domain not found", http.StatusNotFound)
		return
	}

	// Create branch
	branchID, err := authority.CreateBranch(ctx, s.db, authority.CreateBranchRequest{
		OriginalDomainID: originalDomainID,
		NewParentID:      req.NewParentID,
		BranchName:       req.BranchName,
		Reason:           req.Reason,
		CreatedBy:        req.CreatedBy,
		Metadata:         req.Metadata,
	})
	if err != nil {
		s.logger.Printf("Failed to create branch: %v", err)
		http.Error(w, fmt.Sprintf("failed to create branch: %v", err), http.StatusInternalServerError)
		return
	}

	// Get branch info
	branchInfo, err := authority.GetBranchInfo(ctx, s.db, branchID)
	if err != nil {
		s.logger.Printf("Warning: failed to get branch info: %v", err)
	}

	// Return success response
	response := BranchDomainResponse{
		Success:        true,
		BranchDomainID: branchID,
		OriginalDomain: originalDomainID,
		NewParent:      req.NewParentID,
		Message:        fmt.Sprintf("Branch created successfully: %s", req.BranchName),
		BranchInfo:     branchInfo,
	}

	if branchInfo != nil {
		response.BranchDepth = branchInfo.BranchDepth
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// InstantiateSeatRequest represents a request to check DSCI eligibility
type InstantiateSeatRequest struct {
	UserID           uuid.UUID `json:"user_id"`
	OriginalDomainID uuid.UUID `json:"original_domain_id"`
	BranchDomainID   uuid.UUID `json:"branch_domain_id"`
}

// InstantiateSeatResponse indicates DSCI eligibility
type InstantiateSeatResponse struct {
	Eligible         bool      `json:"eligible"`
	Reason           string    `json:"reason"`
	UserID           uuid.UUID `json:"user_id"`
	OriginalDomainID uuid.UUID `json:"original_domain_id"`
	BranchDomainID   uuid.UUID `json:"branch_domain_id"`
	Message          string    `json:"message"`
}

// handleInstantiateSeat checks if a user can instantiate a seat in a branch domain via DSCI
// POST /api/domain/:id/seat/instantiate
//
// GOV-11H DSCI: Domain-Signed Contract Inheritance
// Checks if user has contract with original domain and branch_inheritance=true
// Does NOT actually create the seat (stub for now)
func (s *Server) handleInstantiateSeat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	domainIDStr := chi.URLParam(r, "id")

	branchDomainID, err := uuid.Parse(domainIDStr)
	if err != nil {
		http.Error(w, "invalid domain ID", http.StatusBadRequest)
		return
	}

	var req InstantiateSeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.UserID == uuid.Nil {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}
	if req.OriginalDomainID == uuid.Nil {
		http.Error(w, "original_domain_id is required", http.StatusBadRequest)
		return
	}

	// Override branch domain ID with path parameter if provided
	if req.BranchDomainID == uuid.Nil {
		req.BranchDomainID = branchDomainID
	}

	// Check DSCI eligibility
	eligible, reason, err := authority.CanInstantiateSeatFromContract(
		ctx,
		s.db,
		req.UserID,
		req.OriginalDomainID,
		req.BranchDomainID,
	)
	if err != nil {
		s.logger.Printf("Error checking DSCI eligibility: %v", err)
		http.Error(w, fmt.Sprintf("failed to check eligibility: %v", err), http.StatusInternalServerError)
		return
	}

	response := InstantiateSeatResponse{
		Eligible:         eligible,
		Reason:           reason,
		UserID:           req.UserID,
		OriginalDomainID: req.OriginalDomainID,
		BranchDomainID:   req.BranchDomainID,
	}

	if eligible {
		response.Message = "User is eligible to instantiate seat via DSCI"
	} else {
		response.Message = fmt.Sprintf("User is not eligible: %s", reason)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetBranchInfoResponse wraps branch info for API response
type GetBranchInfoResponse struct {
	*authority.BranchInfo
}

// handleGetBranchInfo retrieves branching metadata for a domain
// GET /api/domain/:id/branch/info
func (s *Server) handleGetBranchInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	domainIDStr := chi.URLParam(r, "id")

	domainID, err := uuid.Parse(domainIDStr)
	if err != nil {
		http.Error(w, "invalid domain ID", http.StatusBadRequest)
		return
	}

	branchInfo, err := authority.GetBranchInfo(ctx, s.db, domainID)
	if err != nil {
		s.logger.Printf("Error getting branch info: %v", err)
		http.Error(w, fmt.Sprintf("failed to get branch info: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GetBranchInfoResponse{BranchInfo: branchInfo})
}
