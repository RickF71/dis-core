package receipts

import (
	"time"

	"github.com/google/uuid"
)

// ReceiptEnvelope is a semantically partitioned receipt with multiple panels
// that can be independently populated and later verified / hashed.
type ReceiptEnvelope struct {
	ID        string    `json:"id"`
	DomainID  string    `json:"domain_id"`
	ActorID   string    `json:"actor_id"`
	Timestamp time.Time `json:"timestamp"`

	ActionPanel    map[string]any `json:"action"`
	PolicyPanel    map[string]any `json:"policy"`
	IdentityPanel  map[string]any `json:"identity"`
	DimensionPanel map[string]any `json:"dimension"`
	LineagePanel   map[string]any `json:"lineage"`
	DomainPanel    map[string]any `json:"domain"`

	PrevHash string `json:"prev_hash"`
	Hash     string `json:"hash"`
}

// NewEnvelope constructs a new ReceiptEnvelope for the given origin info and actor.
// Accept origin ID and name to avoid package import cycles with domain types.
func NewEnvelope(originID string, originName string, actor string) *ReceiptEnvelope {
	return &ReceiptEnvelope{
		ID:        uuid.New().String(),
		DomainID:  originID,
		ActorID:   actor,
		Timestamp: time.Now().UTC(),

		ActionPanel:    map[string]any{},
		PolicyPanel:    map[string]any{},
		IdentityPanel:  map[string]any{},
		DimensionPanel: map[string]any{},
		LineagePanel:   map[string]any{},
		DomainPanel:    map[string]any{},
	}
}
