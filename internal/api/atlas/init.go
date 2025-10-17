package atlas

import "database/sql"

// InitAtlasStore initializes the Atlas store and (optionally) runs schema checks.
// Keeping this wrapper lets api/server.go stay clean and future-proof.
func InitAtlasStore(db *sql.DB) (*AtlasStore, error) {
	// TODO: add schema/migration here if needed (ensureAtlasSchema(db))
	return NewAtlasStore(db), nil
}
