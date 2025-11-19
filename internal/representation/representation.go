package representation

type Kind string

const (
	KindSovereign Kind = "sovereign"
	KindActor     Kind = "actor"
)

// Representation captures “how” an identity is present in a domain.
type Representation struct {
	Kind              Kind   `json:"kind"`
	IdentityID        string `json:"identity_id"`
	CorporealDomainID string `json:"corporeal_domain_id,omitempty"`
	DomainID          string `json:"domain_id"`          // domain being entered
	SeatID            string `json:"seat_id"`            // seat used in this domain
	ActorID           string `json:"actor_id,omitempty"` // empty for sovereign mode
}

func (r *Representation) IsSovereign() bool {
	return r.Kind == KindSovereign
}

func (r *Representation) IsActor() bool {
	return r.Kind == KindActor
}
