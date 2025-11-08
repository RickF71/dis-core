package policy

type DomainPolicyBundle struct {
	DomainID string
	Modules  map[string]string
	Version  string
	Locked   bool
}

// MergeBundles merges the actual loaded policy text of a domain
// with its inherited (fallback) global policies.
func MergeBundles(domain, fallback *DomainPolicyBundle) *DomainPolicyBundle {
	out := &DomainPolicyBundle{
		DomainID: domain.DomainID,
		Version:  domain.Version,
		Locked:   domain.Locked,
		Modules:  make(map[string]string),
	}

	// fallback (null/global) first
	for k, v := range fallback.Modules {
		out.Modules[k] = v
	}
	// then domain-specific overrides
	for k, v := range domain.Modules {
		out.Modules[k] = v
	}

	return out
}
