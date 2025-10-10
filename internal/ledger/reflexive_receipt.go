package ledger

import (
	"encoding/json"
	"log"
	"time"

	"dis-core/internal/events"
	"dis-core/internal/rules"
)

// ReflexiveReceipt represents a self-issued receipt triggered by domain cognition.
type ReflexiveReceipt struct {
	DomainID string                 `json:"domain_id"`
	EventRef string                 `json:"event_ref"`
	Type     string                 `json:"type"`
	Value    float64                `json:"value"`
	Time     time.Time              `json:"time"`
	Context  map[string]interface{} `json:"context"`
}

func EmitReflexiveReceipt(domainID string, e events.Event, a rules.Action) error {
	r := ReflexiveReceipt{
		DomainID: domainID,
		EventRef: e.ID,
		Type:     a.Type,
		Value:    a.TrustDelta,
		Time:     time.Now(),
		Context:  a.Context,
	}

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}

	log.Printf("🪞 ReflexiveReceipt [%s] — %s", domainID, string(data))
	return SaveReceipt(data)
}
