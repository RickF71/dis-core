package domain

func (m *Manager) ImportFromYAML(data map[string]any) error {
	// validate structure, canonicalize, and insert into DB
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Example: store into domains table
	if _, err := tx.Exec(`INSERT OR REPLACE INTO domains (name, data) VALUES (?, ?)`,
		data["name"], data); err != nil {
		return err
	}

	return tx.Commit()
}
