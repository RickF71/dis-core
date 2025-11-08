package policy

func MergePolicies(domain, fallback *DomainPolicy) map[string]string {
	modules := make(map[string]string)

	if fallback == nil {
		fallback = &DomainPolicy{}
	}

	modules["gates.rego"] = choose(domain.Gates, fallback.Gates)
	modules["freeze.rego"] = choose(domain.Freeze, fallback.Freeze)
	modules["risk.rego"] = choose(domain.Risk, fallback.Risk)
	modules["ci_rules.rego"] = choose(domain.CIRules, fallback.CIRules)
	modules["redaction.rego"] = choose(domain.Redaction, fallback.Redaction)
	modules["inherit.rego"] = choose(domain.Inherit, fallback.Inherit)

	return modules
}

func choose(local, fallback string) string {
	if local != "" {
		return local
	}
	return fallback
}
