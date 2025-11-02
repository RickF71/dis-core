package schema

// AuthoritantDelegated — authoritant.delegated.v1
type AuthoritantDelegated struct {
	AuthoritantDefinition `json:",inline" yaml:",inline"`
	Delegator             string                 `json:"delegator" yaml:"delegator"`
	Delegate              string                 `json:"delegate" yaml:"delegate"`
	TTL                   string                 `json:"ttl" yaml:"ttl"`
	Scope                 string                 `json:"scope" yaml:"scope"`
	Constraints           map[string]interface{} `json:"constraints,omitempty" yaml:"constraints,omitempty"`
	ExpiresAt             string                 `json:"expires_at,omitempty" yaml:"expires_at,omitempty"`
	Status                string                 `json:"status,omitempty" yaml:"status,omitempty"`
}

func (d AuthoritantDelegated) Validate() error {
	if err := d.ValidateBase(); err != nil {
		return err
	}
	if err := requireNonEmpty("delegator", d.Delegator); err != nil {
		return err
	}
	if err := requireNonEmpty("delegate", d.Delegate); err != nil {
		return err
	}
	if err := requireNonEmpty("ttl", d.TTL); err != nil {
		return err
	}
	if err := requireNonEmpty("scope", d.Scope); err != nil {
		return err
	}
	// RFC3339 optional for ExpiresAt if provided
	if d.ExpiresAt != "" {
		if err := ISO8601(d.ExpiresAt); err != nil {
			return err
		}
	}
	if d.Status != "" {
		if err := oneOf("status", d.Status, "active", "expired", "revoked"); err != nil {
			return err
		}
	}
	return nil
}
