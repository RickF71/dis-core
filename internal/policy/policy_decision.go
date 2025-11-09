package policy

import "time"

// PolicyDecision is the result of a policy evaluation.
type PolicyDecision struct {
	Allow      bool                   `json:"allow"`
	Reason     string                 `json:"reason,omitempty"`
	BreakGlass bool                   `json:"break_glass,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}
