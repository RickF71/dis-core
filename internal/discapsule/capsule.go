package discapsule

// BedrockChallenge is the minimal challenge DIS-Core sends to the capsule
// when it needs human approval to establish structural bedrock (domain `null`).
type BedrockChallenge struct {
	InstanceID string `json:"instance_id"`
	Nonce      string `json:"nonce"`
	Timestamp  string `json:"timestamp"`
}

// BedrockGrant is the response from the capsule, indicating whether the
// human approved the creation of the 1D root domain `null`.
type BedrockGrant struct {
	GrantID    string `json:"grant_id"`
	InstanceID string `json:"instance_id"`
	Nonce      string `json:"nonce"`
	Approved   bool   `json:"approved"`
}

// Capsule is the abstraction boundary between DIS-Core and the human-held capsule.
// Right now we have a simple LocalCapsule implementation, but later this can
// be swapped for an external binary, hardware key, implant, etc.
type Capsule interface {
	PerformBedrockAuth(ch BedrockChallenge) (BedrockGrant, error)
}
