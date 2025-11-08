package policy

import (
	"database/sql"
	"fmt"
)

func LoadDomainPolicyModel(db *sql.DB, domainID string) (*DomainPolicy, error) {
	var p DomainPolicy
	var err error

	if domainID == "" {
		err = db.QueryRow(`SELECT gates, freeze, risk, ci_rules, redaction, inherit
		                   FROM domain_policies WHERE domain_id IS NULL`).
			Scan(&p.Gates, &p.Freeze, &p.Risk, &p.CIRules, &p.Redaction, &p.Inherit)
	} else {
		err = db.QueryRow(`SELECT gates, freeze, risk, ci_rules, redaction, inherit
		                   FROM domain_policies WHERE domain_id = $1`, domainID).
			Scan(&p.Gates, &p.Freeze, &p.Risk, &p.CIRules, &p.Redaction, &p.Inherit)
	}

	if err != nil {
		return nil, fmt.Errorf("load domain.policy model: %w", err)
	}

	return &p, nil
}
