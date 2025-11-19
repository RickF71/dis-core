package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"dis-core/internal/authority"
	"dis-core/internal/seats"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Phase S4: Seats API Endpoints

// GetDomainSeats handles GET /api/domain/{id}/seats
func (s *Server) GetDomainSeats(w http.ResponseWriter, r *http.Request) {
	domainID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid domain ID")
		return
	}

	ctx := r.Context()
	seatsList, err := s.seatsRepo.GetActiveSeats(ctx, domainID)
	if err != nil {
		s.logger.Printf("[seats] Error getting seats for domain %s: %v", domainID, err)
		JSONError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get seats: %v", err))
		return
	}

	JSON(w, http.StatusOK, seatsList)
}

// AppointMemberSeat handles POST /api/domain/{id}/seats/appoint
func (s *Server) AppointMemberSeat(w http.ResponseWriter, r *http.Request) {
	domainID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid domain ID")
		return
	}

	var req seats.AppointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := r.Context()

	// Get Prime Seat
	pseat, err := s.seatsRepo.GetPrimeSeat(ctx, domainID)
	if err != nil {
		JSONError(w, http.StatusNotFound, "Prime Seat not found")
		return
	}

	// Appoint member
	seat, err := s.seatsRepo.AppointMemberSeat(
		ctx, domainID, pseat.ID, req.MemberID,
		req.Scope, req.RegoRef, req.PolicyVersion, req.AppointmentReceipt,
	)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to appoint member")
		return
	}

	JSON(w, http.StatusCreated, seat)
}

// FreezeSeat handles POST /api/domain/{id}/seats/{seatId}/freeze
func (s *Server) FreezeSeat(w http.ResponseWriter, r *http.Request) {
	domainIDStr := chi.URLParam(r, "id")
	seatID, err := uuid.Parse(chi.URLParam(r, "seatId"))
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid seat ID")
		return
	}

	ctx := r.Context()
	if err := s.seatsRepo.FreezeSeat(ctx, seatID); err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to freeze seat")
		return
	}

	// GOV-9: Record authority receipt for freeze action
	if s.db != nil {
		domainID, err := uuid.Parse(domainIDStr)
		if err == nil {
			receiptPayload := map[string]interface{}{
				"action":  "seat.freeze.v1",
				"seat_id": seatID.String(),
				"reason":  "admin_action",
				"scope":   "seat",
			}
			authority.RecordAuthorityReceipt(ctx, s.db, domainID, "seat.freeze.v1", receiptPayload, nil)
		}
	}

	JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Seat frozen",
		"seat_id": seatID,
	})
}

// UnfreezeSeat handles POST /api/domain/{id}/seats/{seatId}/unfreeze
func (s *Server) UnfreezeSeat(w http.ResponseWriter, r *http.Request) {
	domainIDStr := chi.URLParam(r, "id")
	seatID, err := uuid.Parse(chi.URLParam(r, "seatId"))
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid seat ID")
		return
	}

	ctx := r.Context()
	if err := s.seatsRepo.UnfreezeSeat(ctx, seatID); err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to unfreeze seat")
		return
	}

	// GOV-9: Record authority receipt for unfreeze action
	if s.db != nil {
		domainID, err := uuid.Parse(domainIDStr)
		if err == nil {
			receiptPayload := map[string]interface{}{
				"action":  "seat.unfreeze.v1",
				"seat_id": seatID.String(),
				"reason":  "admin_action",
				"scope":   "seat",
			}
			authority.RecordAuthorityReceipt(ctx, s.db, domainID, "seat.unfreeze.v1", receiptPayload, nil)
		}
	}

	JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Seat unfrozen",
		"seat_id": seatID,
	})
}

// UpdateSeatRego handles PUT /api/domain/{id}/seats/{seatId}/rego
func (s *Server) UpdateSeatRego(w http.ResponseWriter, r *http.Request) {
	seatID, err := uuid.Parse(chi.URLParam(r, "seatId"))
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid seat ID")
		return
	}

	var req seats.UpdateRegoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := r.Context()
	if err := s.seatsRepo.UpdateSeatRego(ctx, seatID, req.RegoText, req.PolicyVersion); err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to update REGO")
		return
	}

	JSON(w, http.StatusOK, map[string]interface{}{
		"success":        true,
		"message":        "REGO updated",
		"seat_id":        seatID,
		"policy_version": req.PolicyVersion,
	})
}

// GOV-7: CreatePrimeSeat handles POST /api/domain/{id}/seat/prime
// Creates or verifies a Prime Seat for corporeal domains with single-occupancy enforcement
func (s *Server) CreatePrimeSeat(w http.ResponseWriter, r *http.Request) {
	domainID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid domain ID")
		return
	}

	var req struct {
		MemberID string `json:"member_id"`
		RegoRef  string `json:"rego_ref"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := r.Context()

	// Check if Prime Seat already exists
	existingSeat, err := s.seatsRepo.GetPrimeSeat(ctx, domainID)
	if err == nil && existingSeat != nil {
		// Prime Seat exists - verify or return existing
		memberID := ""
		if existingSeat.MemberID != nil {
			memberID = *existingSeat.MemberID
		}
		JSON(w, http.StatusOK, map[string]interface{}{
			"status":    "ok",
			"seat_id":   existingSeat.ID.String(),
			"member_id": memberID,
			"message":   "Prime Seat already exists",
			"existing":  true,
		})
		return
	}

	// Create new Prime Seat
	pseat, err := s.seatsRepo.CreatePrimeSeat(ctx, domainID, req.MemberID, req.RegoRef)
	if err != nil {
		s.logger.Printf("[seats] GOV-7: Failed to create Prime Seat for domain %s: %v", domainID, err)
		JSONError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create Prime Seat: %v", err))
		return
	}

	s.logger.Printf("[seats] GOV-7: Created Prime Seat %s for domain %s with member %s",
		pseat.ID, domainID, req.MemberID)

	memberID := ""
	if pseat.MemberID != nil {
		memberID = *pseat.MemberID
	}
	JSON(w, http.StatusCreated, map[string]interface{}{
		"status":    "ok",
		"seat_id":   pseat.ID.String(),
		"member_id": memberID,
		"message":   "Prime Seat created successfully",
		"existing":  false,
	})
}
