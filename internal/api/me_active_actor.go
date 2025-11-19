package api

import (
	"encoding/json"
	"net/http"

	"dis-core/internal/auth"

	"github.com/google/uuid"
)

// SetActiveActorRequest is the request body for setting the active actor
type SetActiveActorRequest struct {
	SeatID string `json:"seat_id"` // UUID of the seat to activate
}

// SetActiveActorResponse is the response for setting the active actor
type SetActiveActorResponse struct {
	OK           bool   `json:"ok"`
	ActiveSeatID string `json:"active_seat_id"`
	Message      string `json:"message,omitempty"`
}

// handleSetActiveActor sets the active actor/seat for the authenticated user
// POST /api/me/active-actor
func (s *Server) handleSetActiveActor(w http.ResponseWriter, r *http.Request) {
	user := auth.GetActiveUser(r)

	if user == nil {
		JSONError(w, http.StatusUnauthorized, "No active user context")
		return
	}

	if !user.IsAuthenticated() {
		JSONError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Decode request
	var req SetActiveActorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.SeatID == "" {
		JSONError(w, http.StatusBadRequest, "seat_id is required")
		return
	}

	// Parse UUID
	seatUUID, err := uuid.Parse(req.SeatID)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid seat_id format")
		return
	}

	ctx := r.Context()

	// Verify seat ownership
	err = s.seatsRepo.VerifySeatOwnership(ctx, seatUUID, user.ExternalUID)
	if err != nil {
		s.logger.Printf("[me/active-actor] Ownership verification failed for seat %s, user %s: %v",
			req.SeatID, user.ExternalUID, err)
		JSONError(w, http.StatusForbidden, "Seat not owned by user or does not exist")
		return
	}

	// Set active actor in context
	// Note: This stores it in the request context for this request
	// For persistent storage across requests, we'd need session/token storage
	newReq := auth.SetActiveActor(r, req.SeatID)

	s.logger.Printf("[me/active-actor] Set active actor to seat %s for user %s",
		req.SeatID, user.ExternalUID)

	resp := SetActiveActorResponse{
		OK:           true,
		ActiveSeatID: req.SeatID,
		Message:      "Active actor set successfully",
	}

	// Update request context for downstream handlers
	*r = *newReq

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleGetActiveActor returns the currently active actor/seat
// GET /api/me/active-actor
func (s *Server) handleGetActiveActor(w http.ResponseWriter, r *http.Request) {
	user := auth.GetActiveUser(r)

	if user == nil {
		JSONError(w, http.StatusUnauthorized, "No active user context")
		return
	}

	activeSeatID, hasActive := auth.GetActiveActor(r)

	resp := struct {
		HasActiveActor bool   `json:"has_active_actor"`
		ActiveSeatID   string `json:"active_seat_id,omitempty"`
	}{
		HasActiveActor: hasActive,
		ActiveSeatID:   activeSeatID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
