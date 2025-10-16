package scaffold

import "time"

type Spotlight struct {
	ID            string    `json:"id"`
	DomainID      string    `json:"domain_id"`
	ObservedState string    `json:"observed_state"`
	MoralSignal   float64   `json:"moral_signal"`
	Observer      string    `json:"observer"`
	UpdatedAt     time.Time `json:"updated_at"`
}
