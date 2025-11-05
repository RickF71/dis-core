package domain

import (
	"dis-core/internal/model"
	"encoding/json"
)

type payloadGenesis struct {
	Domains []struct {
		DomainID    string `json:"domain_id"`
		Label       string `json:"label"`
		Description string `json:"description"`
	} `json:"domains"`
}

type payloadIntent struct {
	ParentID string `json:"parent_id"`
	Desired  struct {
		DomainID    string `json:"domain_id"`
		Label       string `json:"label"`
		Description string `json:"description"`
	} `json:"desired"`
}

type payloadApprove struct {
	IntentEventID int `json:"intent_event_id"`
}

type payloadEmergence struct {
	ApprovedEventID int    `json:"approved_event_id"`
	DomainID        string `json:"domain_id"`
	ParentID        string `json:"parent_id"`
	Status          Status `json:"status"`
}

// Replay builds the world from scratch into an in-memory map.
func Replay(evts []model.DisEvent) map[string]*DomainState {
	state := make(map[string]*DomainState)
	intents := make(map[int]payloadIntent)

	for _, e := range evts {
		switch e.Type {

		case "genesis.initialize":
			var p payloadGenesis
			_ = json.Unmarshal(e.Payload, &p)
			for _, d := range p.Domains {
				state[d.DomainID] = &DomainState{
					DomainID:         d.DomainID,
					ParentID:         "",
					Status:           DomainActive,
					Label:            d.Label,
					Description:      d.Description,
					CreatedFromEvent: e.ID,
				}
			}

		case "domain.intent.totelevate":
			var p payloadIntent
			_ = json.Unmarshal(e.Payload, &p)
			intents[e.ID] = p

		case "domain.totelevate.approve":
			var p payloadApprove
			_ = json.Unmarshal(e.Payload, &p)
			// no state change yet — just record that intent is approved
			// emergence handles creation

		case "domain.emergence":
			var p payloadEmergence
			_ = json.Unmarshal(e.Payload, &p)
			intent := intents[p.ApprovedEventID-1] // off-by-one depending on DB auto-inc; may adjust
			state[p.DomainID] = &DomainState{
				DomainID:         p.DomainID,
				ParentID:         p.ParentID,
				Status:           p.Status,
				Label:            intent.Desired.Label,
				Description:      intent.Desired.Description,
				CreatedFromEvent: e.ID,
			}
		}
	}

	return state
}
