package events

import (
	"dis-core/internal/model"
	"encoding/json"
	"fmt"
	"time"
)

// GenerateLiveDisEvent creates a demo DisEvent for live streaming
func GenerateLiveDisEvent(idx int) model.DisEvent {
	states := []model.DisEventType{"domain.freeze.v1", "domain.unfreeze.v1"}

	return model.DisEvent{
		ID:        idx,
		TS:        time.Now(),
		Type:      states[idx%len(states)],
		Actor:     "domain.usa",
		Payload:   []byte(fmt.Sprintf(`{"domain":"domain.usa","state":"%s"}`, states[idx%len(states)])),
		Signature: "",
	}
}

// FormatSSEDisEvent formats a DisEvent as Server-Sent Event data
func FormatSSEDisEvent(ev model.DisEvent) (string, error) {
	bytes, err := json.Marshal(ev)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("event: domain.update\ndata: %s\n\n", string(bytes)), nil
}
