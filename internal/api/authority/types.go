package authorityapi

import (
	"time"
)

// IdentitySeatDTO is an API-safe representation of an identity seat
type IdentitySeatDTO struct {
	SeatID    string `json:"seat_id"`
	Layer     string `json:"layer"`      // terra | numen | lima
	State     string `json:"state"`      // EMPTY|ASSIGNED|OCCUPIED|FROZEN
	UpdatedAt string `json:"updated_at"` // ISO8601
}

// IdentityTriadDTO contains all three identity seats
type IdentityTriadDTO struct {
	IdentityID string            `json:"identity_id"`
	Seats      []IdentitySeatDTO `json:"seats"`
	Complete   bool              `json:"complete"` // true if all 3 seats present
}

// FlowPreview contains sample authority flow scenarios
type FlowPreview struct {
	Scenarios []FlowScenario `json:"scenarios"`
	Timestamp string         `json:"timestamp"`
}

// FlowScenario represents a sample authority flow case
type FlowScenario struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Direction   string `json:"direction"` // upward|downward|lateral
	SeatState   string `json:"seat_state"`
	Expected    string `json:"expected"` // allow|deny
}

// FlowEvalInput is the input for authority flow evaluation
type FlowEvalInput struct {
	IdentityID     string         `json:"identity_id"`
	DomainID       string         `json:"domain_id"`
	Action         string         `json:"action"`
	Direction      string         `json:"direction"` // upward|downward|lateral
	ActionDomain   string         `json:"action_domain,omitempty"`
	SeatDomain     string         `json:"seat_domain,omitempty"`
	ParentApproved bool           `json:"parent_approved"`
	Context        map[string]any `json:"context,omitempty"`
}

// FlowEvalResult contains the authority evaluation result
type FlowEvalResult struct {
	Allow         bool              `json:"allow"`
	Reason        string            `json:"reason"`
	TriadSeats    []TriadSeatStatus `json:"triad_seats"`
	Details       map[string]any    `json:"details,omitempty"`
	EvaluatedAt   string            `json:"evaluated_at"`
	PolicyVersion string            `json:"policy_version,omitempty"`
}

// TriadSeatStatus represents the status of one triad seat during evaluation
type TriadSeatStatus struct {
	Layer  string `json:"layer"` // terra|numen|lima
	State  string `json:"state"`
	Frozen bool   `json:"frozen"`
}

// GetTimestamp returns current UTC timestamp in RFC3339 format
func GetTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
