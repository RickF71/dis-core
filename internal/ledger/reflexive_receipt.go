package ledger

import (
	"encoding/json"
	"log"
	"time"

	"dis-core/internal/model"
	"dis-core/internal/rules"
)

type ReflexiveReceipt struct {
	DomainID   string                 `json:"domain_id"`
	EventRef   int                    `json:"event_ref"`
	ActionType string                 `json:"action_type"`
	Value      float64                `json:"value"`
	Time       time.Time              `json:"time"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

func EmitReflexiveReceipt(domainID string, e model.DisEvent, a rules.Action) error {
	r := ReflexiveReceipt{
		DomainID:   domainID,
		EventRef:   e.ID,
		ActionType: a.Type,
		Value:      a.TrustDelta,
		Time:       time.Now().UTC(),
		Context:    a.Context,
	}

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	log.Printf("🪞 ReflexiveReceipt [%s] — %s", domainID, string(data))

	receipt := &Receipt{
		By:     domainID,
		Action: string(e.Type) + ":" + a.Type,
	}

	return SaveReceipt(receipt)
}
