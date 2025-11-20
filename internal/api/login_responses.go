package api

// LoginGenesisResponse is the canonical response for the first human login flow.
type LoginGenesisResponse struct {
	Status           string `json:"status"`
	ActorID          string `json:"actor_id"`
	DomainID         string `json:"domain_id"`
	ReceiptID        string `json:"receipt_id"`
	PresentationName string `json:"presentation_name"`
}
