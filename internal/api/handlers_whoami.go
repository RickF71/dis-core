package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"dis-core/internal/auth"

	"github.com/google/uuid"
)

// handleWhoAmI returns a minimal bound identity surface for Finagler integration
// GET /api/whoami
func (s *Server) handleWhoAmI(w http.ResponseWriter, r *http.Request) {
	user := auth.GetActiveUser(r)
	// Accept either the canonical IsBound() (which relies on CorporealDomainID)
	// or the session-bound path which sets CorporealDomainUID. Some auth
	// middleware attaches bound state via CorporealDomainUID (string) while
	// leaving CorporealDomainID == 0 for tests; accept either form.
	if user == nil || !(user.Bound && (user.CorporealDomainID > 0 || user.CorporealDomainUID != "")) {
		JSONError(w, http.StatusUnauthorized, "No bound active user")
		return
	}

	// domain_id: prefer the session-populated CorporealDomainUID when present
	domainID := user.CorporealDomainUID

	// seat_id and actor_id may be available via the active actor context
	seatID, hasActive := auth.GetActiveActor(r)
	var actorID string

	if hasActive && seatID != "" && s.seatsRepo != nil {
		if sid, err := uuid.Parse(seatID); err == nil {
			if seat, err := s.seatsRepo.GetSeat(r.Context(), sid); err == nil && seat != nil {
				if seat.MemberID != nil {
					m := *seat.MemberID
					// member_id may be stored as plain uuid, or prefixed like actor.<id>
					if strings.HasPrefix(m, "actor.") {
						actorID = strings.TrimPrefix(m, "actor.")
					} else if _, err := uuid.Parse(m); err == nil {
						actorID = m
					}
				}
			}
		}
	}

	resp := map[string]string{
		"domain_id": domainID,
		"actor_id":  actorID,
		"seat_id":   seatID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
