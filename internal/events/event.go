package events

import (
	"sync"
	"time"
)

type Event struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Source      string                 `json:"source"`
	Target      string                 `json:"target,omitempty"` // <— new
	CreatedAt   time.Time              `json:"created_at"`
	Context     map[string]interface{} `json:"context,omitempty"`
	ConsentRef  string                 `json:"consent_ref,omitempty"`
	FeedbackRef string                 `json:"feedback_ref,omitempty"`
}

var (
	eventLog = []Event{}
	mu       sync.Mutex
)

// Emit adds a new event to the log.
func Emit(e Event) {
	mu.Lock()
	defer mu.Unlock()
	eventLog = append(eventLog, e)
}

// FetchRecent retrieves events relevant to a given domain ID.
// For simplicity, it returns all events targeting or sourced by the domain.
func FetchRecent(domainID string) []Event {
	mu.Lock()
	defer mu.Unlock()
	var relevant []Event
	for _, e := range eventLog {
		if e.Source == domainID || e.Target == domainID {
			relevant = append(relevant, e)
		}
	}
	return relevant
}
