package authority

import "time"

// SeatKind represents the kind of seat within a domain.
type SeatKind string

const (
	SeatKindCorporealRoot SeatKind = "corporeal.root"
	SeatKindMember        SeatKind = "member"
	SeatKindActorProxy    SeatKind = "actor.proxy"
)

// Seat models a membership seat inside a domain.
type Seat struct {
	ID         string    `json:"id"`
	DomainID   string    `json:"domain_id"`
	IdentityID string    `json:"identity_id"`
	Kind       SeatKind  `json:"kind"`
	CreatedAt  time.Time `json:"created_at"`
}

// SeatStatus provides runtime status information for a seat.
type SeatStatus struct {
	Seat       *Seat  `json:"seat"`
	LineageRef string `json:"lineage_ref,omitempty"`
	Active     bool   `json:"active"`
}

// SeatLineage reports the provenance and ancestor chain for a seat.
type SeatLineage struct {
	SeatID     string   `json:"seat_id"`
	DomainID   string   `json:"domain_id"`
	IdentityID string   `json:"identity_id"`
	Ancestors  []string `json:"ancestors"`
	Receipts   []string `json:"receipts"`
}
