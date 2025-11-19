package seats

import (
	"time"

	"github.com/google/uuid"
)

// Seat represents a domain seat (Prime Seat or member)
type Seat struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	DomainID           uuid.UUID  `json:"domain_id" db:"domain_id"`
	ParentSeatID       *uuid.UUID `json:"parent_seat_id,omitempty" db:"parent_seat_id"`
	SeatType           string     `json:"seat_type" db:"seat_type"` // "prime" | "member"
	MemberID           *string    `json:"member_id,omitempty" db:"member_id"`
	AppointedBy        *uuid.UUID `json:"appointed_by,omitempty" db:"appointed_by"`
	AppointmentReceipt *string    `json:"appointment_receipt,omitempty" db:"appointment_receipt"`
	RegoRef            *string    `json:"rego_ref,omitempty" db:"rego_ref"`
	RegoText           *string    `json:"rego_text,omitempty" db:"rego_text"`
	PolicyVersion      *string    `json:"policy_version,omitempty" db:"policy_version"`
	Scope              *string    `json:"scope,omitempty" db:"scope"`
	Status             string     `json:"status" db:"status"` // "active" | "frozen" | "detached"
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
}

// SeatContext represents OPA input context for seat evaluation
type SeatContext struct {
	ActiveRoot   bool `json:"active_root"`
	ActiveMember bool `json:"active_member"`
	Frozen       bool `json:"frozen"`
	Detached     bool `json:"detached"`
}

// AppointRequest represents a member seat appointment request
type AppointRequest struct {
	MemberID           string `json:"member_id"`
	Scope              string `json:"scope"`
	RegoRef            string `json:"rego_ref"`
	PolicyVersion      string `json:"policy_version"`
	AppointmentReceipt string `json:"appointment_receipt"`
}

// UpdateRegoRequest represents a per-seat REGO update
type UpdateRegoRequest struct {
	RegoText      string `json:"rego_text"`
	PolicyVersion string `json:"policy_version"`
}
