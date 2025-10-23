package policy

// PolicyDecision is the output of an OPA evaluation.
type PolicyDecision struct {
	Allow     bool    `json:"allow"`
	RiskScore float64 `json:"risk_score"`
	Reason    string  `json:"reason,omitempty"`
}
