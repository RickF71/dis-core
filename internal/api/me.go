package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"dis-core/internal/auth"
)

// MeResponse represents the safe, minimal identity surface for the current user
type MeResponse struct {
	Authenticated      bool   `json:"authenticated"`        // Whether user has external UID
	Bound              bool   `json:"bound"`                // Whether mapped to corporeal domain
	CorporealDomainUID string `json:"corporeal_domain_uid"` // Sovereign DIS identity name (if bound)
	CorporealDomainID  string `json:"corporeal_domain_id"`  // UUID of corporeal domain (if bound)
	PrimeSeatID        string `json:"prime_seat_id"`        // UUID of Prime Seat (if found)
	DisplayName        string `json:"display_name"`         // Human-readable name
}

// handleMe returns the active user's identity context
// GET /api/me
func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	user := auth.GetActiveUser(r)

	if user == nil {
		JSONError(w, http.StatusUnauthorized, "No active user context")
		return
	}

	resp := MeResponse{
		Authenticated:      user.IsAuthenticated(),
		Bound:              user.Bound,
		CorporealDomainUID: user.CorporealDomainUID,
		DisplayName:        user.CorporealDomainUID, // Use UID as display name for now
	}

	// If bound, retrieve the domain UUID
	if user.Bound && user.CorporealDomainID > 0 {
		// Query corporeal domain to get UUID
		var domainUUID string
		err := s.db.QueryRow(r.Context(), `
			SELECT id FROM domains WHERE id = $1 OR name = $2 LIMIT 1
		`, user.CorporealDomainID, user.CorporealDomainUID).Scan(&domainUUID)

		if err == nil {
			resp.CorporealDomainID = domainUUID
			s.logger.Printf("[me] Found domain UUID %s for user %s", domainUUID, user.CorporealDomainUID)

			// Try to get Prime Seat for this domain
			if s.seatsRepo != nil {
				resp.PrimeSeatID = fmt.Sprintf("pseat-%s", domainUUID[:8])
			}
		} else {
			s.logger.Printf("[me] Warning: Could not resolve domain UUID for user %s: %v", user.CorporealDomainUID, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
