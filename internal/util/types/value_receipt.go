package types

import "time"

// ValueReceipt represents a coherence delta event recorded in the moral field.
type ValueReceipt struct {
	ID             string             `json:"id" db:"id"`
	By             string             `json:"by" db:"by"`
	ActionRef      string             `json:"action_ref" db:"action_ref"`
	SubstrateRef   string             `json:"substrate_ref,omitempty" db:"substrate_ref"`
	CoherenceDelta float64            `json:"coherence_delta" db:"coherence_delta"`
	ValueVector    map[string]float64 `json:"value_vector" db:"value_vector"` // JSON column
	ObserverField  string             `json:"observer_field,omitempty" db:"observer_field"`
	Notes          string             `json:"notes,omitempty" db:"notes"`
	Timestamp      time.Time          `json:"timestamp" db:"timestamp"`
}
