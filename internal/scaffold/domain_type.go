package scaffold

type DomainType struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Class           string `json:"class"`
	Parent          string `json:"parent,omitempty"`
	Description     string `json:"description,omitempty"`
	GovernanceModel string `json:"governance_model,omitempty"`
}
