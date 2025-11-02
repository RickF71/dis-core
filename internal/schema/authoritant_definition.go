package schema

// AuthoritantDefinition is the common base for all Authorization Domain specs.
// It is conceptually defined by authoritant.definition.v1
type AuthoritantDefinition struct {
	Type        string `json:"type" yaml:"type"`
	Version     string `json:"version" yaml:"version"`
	Domain      string `json:"domain" yaml:"domain"`
	Scope       string `json:"scope,omitempty" yaml:"scope,omitempty"`
	Status      string `json:"status,omitempty" yaml:"status,omitempty"`
	CreatedAt   string `json:"created_at" yaml:"created_at"`
	UpdatedAt   string `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

func (d AuthoritantDefinition) ValidateBase() error {
	if err := requireNonEmpty("type", d.Type); err != nil {
		return err
	}
	if err := requireNonEmpty("version", d.Version); err != nil {
		return err
	}
	if err := requireNonEmpty("domain", d.Domain); err != nil {
		return err
	}
	if err := requireNonEmpty("created_at", d.CreatedAt); err != nil {
		return err
	}
	if err := ISO8601(d.CreatedAt); err != nil {
		return err
	}
	if d.Status != "" {
		if err := oneOf("status", d.Status, "active", "suspended", "expired"); err != nil {
			return err
		}
	}
	return nil
}
