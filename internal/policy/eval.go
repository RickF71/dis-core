package policy

import (
	"database/sql"
)

// LoadNullPolicies loads all NULL domain policies into the OPA engine.
func (e *OPAEngine) LoadNullPolicies(db *sql.DB) error {
	rows, err := db.Query(`SELECT name, content FROM policies WHERE domain_id IS NULL`)
	if err != nil {
		return err
	}
	defer rows.Close()

	modules := map[string]string{}
	for rows.Next() {
		var name, content string
		if err := rows.Scan(&name, &content); err != nil {
			return err
		}
		modules[name+".rego"] = content
	}

	eng, err := NewEngine(modules)
	if err != nil {
		return err
	}

	*e = *eng
	return nil
}
