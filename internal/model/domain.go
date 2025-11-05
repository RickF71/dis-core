package model

// Domain represents a full DIS domain record stored in canon.content (JSONB).
type Domain struct {
	Meta struct {
		DomainID      string `json:"domain_id"`
		Name          string `json:"name"`
		SchemaID      string `json:"schema_id,omitempty"`
		SchemaVersion string `json:"schema_version,omitempty"`
		Description   string `json:"description,omitempty"`
	} `json:"meta"`

	Consent struct {
		Gate   string `json:"gate,omitempty"`
		Quorum string `json:"quorum,omitempty"`
	} `json:"consent,omitempty"`

	CSS    string `json:"css,omitempty"`
	JSX    string `json:"jsx,omitempty"`
	Freeze struct {
		Active bool `json:"active,omitempty"`
		TTL    int  `json:"ttl,omitempty"`
	} `json:"freeze,omitempty"`

	// Optional extensions—safe to grow over time.
	Policy   map[string]interface{} `json:"policy,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
