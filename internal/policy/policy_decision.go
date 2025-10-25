package policy

// PolicyDecision is the result of a policy evaluation.
type PolicyDecision struct {
	Allow      bool                   `json:"allow"`
	RiskScore  float64                `json:"risk_score"`
	Reason     string                 `json:"reason,omitempty"`
	BreakGlass bool                   `json:"break_glass,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
}
