package policy

import (
	"database/sql"
	"fmt"
)

func LoadDomainPolicy(db *sql.DB, domainID string) (map[string]string, error) {
	q := `SELECT name, content FROM policies WHERE domain_id = $1`
	var rows *sql.Rows
	var err error

	if domainID == "" {
		// Load NULL domain (global policies)
		q = `SELECT name, content FROM policies WHERE domain_id IS NULL`
		rows, err = db.Query(q)
	} else {
		rows, err = db.Query(q, domainID)
	}

	if err != nil {
		return nil, fmt.Errorf("load domain.policy: %w", err)
	}
	defer rows.Close()

	modules := map[string]string{}
	for rows.Next() {
		var name, content string
		if err := rows.Scan(&name, &content); err != nil {
			return nil, err
		}
		modules[name+".rego"] = content
	}
	return modules, nil
}

func LoadDomainPolicyBundle(db *sql.DB, domainID string) (*DomainPolicyBundle, error) {
	var rows *sql.Rows
	var err error

	if domainID == "" {
		rows, err = db.Query(`SELECT name, content FROM policies WHERE domain_id IS NULL`)
		domainID = "domain.null"
	} else {
		rows, err = db.Query(`SELECT name, content FROM policies WHERE domain_id = $1`, domainID)
	}

	if err != nil {
		return nil, fmt.Errorf("load domain.policy bundle: %w", err)
	}
	defer rows.Close()

	modules := make(map[string]string)
	for rows.Next() {
		var name, content string
		if err := rows.Scan(&name, &content); err != nil {
			return nil, err
		}
		modules[name+".rego"] = content
	}

	return &DomainPolicyBundle{
		DomainID: domainID,
		Modules:  modules,
		Version:  "v1",
		Locked:   false,
	}, nil
}
