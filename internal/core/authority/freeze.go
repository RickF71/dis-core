package authority

// FreezeState is the returned structure for freeze operations.
type FreezeState struct {
	DomainID string   `json:"domain_id"`
	Frozen   bool     `json:"frozen"`
	Notes    []string `json:"notes,omitempty"`
}

// Freeze applies a domain-level freeze from the governance engine.
// Populate with logic extracted from freeze endpoints once present.
func (e *Engine) Freeze(domainID string) (*FreezeState, error) {
	return &FreezeState{
		DomainID: domainID,
		Frozen:   true,
		Notes:    []string{"MX-2.1 placeholder: real freeze logic needed"},
	}, nil
}

// Unfreeze clears a domain freeze.
func (e *Engine) Unfreeze(domainID string) (*FreezeState, error) {
	return &FreezeState{
		DomainID: domainID,
		Frozen:   false,
		Notes:    []string{"MX-2.1 placeholder: real unfreeze logic needed"},
	}, nil
}
