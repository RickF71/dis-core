package domain

type Status string

const (
	DomainEmerging Status = "emerging"
	DomainActive   Status = "active"
	DomainFrozen   Status = "frozen"
)

type DomainState struct {
	DomainID         string `db:"domain_id" json:"domain_id"`
	ParentID         string `db:"parent_id" json:"parent_id"`
	Status           Status `db:"status" json:"status"`
	CreatedFromEvent int    `db:"created_from_event" json:"created_from_event"`
	Label            string `db:"label" json:"label"`
	Description      string `db:"description" json:"description"`
}
