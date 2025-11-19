package api

import (
	"encoding/json"
	"net/http"

	"dis-core/internal/auth"
)

// ActorView represents a single actor/seat for the authenticated user
type ActorView struct {
	SeatID     string `json:"seat_id"`     // UUID of the seat
	DomainID   string `json:"domain_id"`   // UUID of the domain
	DomainName string `json:"domain_name"` // Human-readable domain name
	SeatType   string `json:"seat_type"`   // "prime" | "member"
	IsPrime    bool   `json:"is_prime"`    // True if seat_type == "prime"
	MemberID   string `json:"member_id"`   // The member_id (e.g., "human.rick")
	Status     string `json:"status"`      // "active" | "frozen" | "detached"
}

// MeActorsResponse returns all actors/seats for the authenticated user
type MeActorsResponse struct {
	Actors []ActorView `json:"actors"`
}

// handleMeActors returns all seats/actors for the authenticated user
// GET /api/me/actors
func (s *Server) handleMeActors(w http.ResponseWriter, r *http.Request) {
	user := auth.GetActiveUser(r)

	if user == nil {
		JSONError(w, http.StatusUnauthorized, "No active user context")
		return
	}

	if !user.IsAuthenticated() {
		JSONError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	ctx := r.Context()

	// Query all seats where member_id is related to this user
	// Patterns: human.{externalUID}, actor.{externalUID}.*, identity patterns
	query := `
		SELECT
			s.id as seat_id,
			s.domain_id,
			d.name as domain_name,
			s.seat_type,
			s.member_id,
			s.status
		FROM domain_seats s
		JOIN domains d ON d.id = s.domain_id
		WHERE s.member_id LIKE $1
		   OR s.member_id LIKE $2
		   OR s.member_id = $3
		ORDER BY
			CASE WHEN s.seat_type = 'prime' THEN 0 ELSE 1 END,
			d.name ASC
	`

	// Build search patterns based on external UID
	externalUID := user.ExternalUID
	humanPattern := "human." + externalUID + "%"
	actorPattern := "actor." + externalUID + "%"
	directMatch := externalUID

	rows, err := s.db.Query(ctx, query, humanPattern, actorPattern, directMatch)
	if err != nil {
		s.logger.Printf("[me/actors] Query error: %v", err)
		JSONError(w, http.StatusInternalServerError, "Failed to query actors")
		return
	}
	defer rows.Close()

	actors := []ActorView{}

	for rows.Next() {
		var av ActorView
		var memberID *string

		err := rows.Scan(
			&av.SeatID,
			&av.DomainID,
			&av.DomainName,
			&av.SeatType,
			&memberID,
			&av.Status,
		)
		if err != nil {
			s.logger.Printf("[me/actors] Scan error: %v", err)
			continue
		}

		if memberID != nil {
			av.MemberID = *memberID
		}

		av.IsPrime = av.SeatType == "prime"

		actors = append(actors, av)
	}

	resp := MeActorsResponse{
		Actors: actors,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
