package events

import "sync"

type Event struct {
	ID      string
	Type    string
	Source  string
	Target  string
	Context map[string]any
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
