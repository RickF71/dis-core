package identity

import (
	"time"

	"github.com/google/uuid"
)

// IdentitySeat represents terra/numen/lima seats (GOV-1 identity triad)
type IdentitySeat struct {
	ID         uuid.UUID              `json:"id"`
	IdentityID string                 `json:"identity_id"`
	SeatType   string                 `json:"seat_type"` // terra, numen, lima
	State      string                 `json:"state"`     // EMPTY, ASSIGNED, OCCUPIED, FROZEN
	AssignedAt *time.Time             `json:"assigned_at,omitempty"`
	OccupiedAt *time.Time             `json:"occupied_at,omitempty"`
	FrozenAt   *time.Time             `json:"frozen_at,omitempty"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"created_at"`
}

// IdentityTriad holds all three universal seats for an identity
type IdentityTriad struct {
	Terra *IdentitySeat `json:"terra"`
	Numen *IdentitySeat `json:"numen"`
	Lima  *IdentitySeat `json:"lima"`
}

// SeatState constants define the lifecycle states
const (
	SeatStateEmpty    = "EMPTY"
	SeatStateAssigned = "ASSIGNED"
	SeatStateOccupied = "OCCUPIED"
	SeatStateFrozen   = "FROZEN"
)

// SeatType constants define the three identity triad seats
const (
	SeatTypeTerra = "terra" // Existence seat
	SeatTypeNumen = "numen" // Meaning seat
	SeatTypeLima  = "lima"  // Consent seat
)

// HasAuthority checks if a seat has authority based on GOV-1 rules
func (s *IdentitySeat) HasAuthority() bool {
	return s.State == SeatStateAssigned || s.State == SeatStateOccupied
}

// IsOccupied checks if a seat is in OCCUPIED state (full authority)
func (s *IdentitySeat) IsOccupied() bool {
	return s.State == SeatStateOccupied
}

// IsFrozen checks if a seat is frozen (no authority)
func (s *IdentitySeat) IsFrozen() bool {
	return s.State == SeatStateFrozen
}

// IsComplete checks if the triad has all three seats in valid states
func (t *IdentityTriad) IsComplete() bool {
	return t.Terra != nil && t.Numen != nil && t.Lima != nil &&
		t.Terra.State != SeatStateEmpty &&
		t.Numen.State != SeatStateEmpty &&
		t.Lima.State != SeatStateEmpty
}
