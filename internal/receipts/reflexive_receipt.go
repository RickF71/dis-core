package receipts

import (
	"encoding/json"
	"log"
	"time"

	"dis-core/internal/events"
	"dis-core/internal/rules"
)

// ReflexiveReceipt represents a self-issued moral action within a domain’s cognition loop.
// It mirrors internal evaluation — actions, moral deltas, or domain state adjustments.
type ReflexiveReceipt struct {
	DomainID   string                 `json:"domain_id"`
	EventRef   string                 `json:"event_ref"`
	ActionType string                 `json:"action_type"`
	Value      float64                `json:"value"`
	Time       time.Time              `json:"time"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// EmitReflexiveReceipt creates and saves a self-issued receipt
// tied to moral or trust feedback events. It integrates directly
// with the unified v0.8.8 receipt structure.
func EmitReflexiveReceipt(domainID string, e events.Event, a rules.Action) error {
	r := ReflexiveReceipt{
		DomainID:   domainID,
		EventRef:   e.ID,
		ActionType: a.Type,
		Value:      a.TrustDelta,
		Time:       time.Now().UTC(),
		Context:    a.Context,
	}

	// Serialize for log visualization
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	log.Printf("🪞 ReflexiveReceipt [%s] — %s", domainID, string(data))

	// Build the new authoritative ledger receipt
	// Use canonical Receipt struct fields only
	receipt := &Receipt{
		By:     domainID,
		Action: e.Type + ":" + a.Type,
		// Fill other canonical fields as needed, e.g. Provenance, Metadata, etc.
	}

	// Store with full integrity (UUID, hash, timestamp)
	return SaveReceipt(receipt)
}
