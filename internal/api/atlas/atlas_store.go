package atlas

import (
	"database/sql"
	"fmt"
	"time"
)

// AtlasStore handles Postgres persistence for Atlas receipts.
type AtlasStore struct {
	db *sql.DB
}

// NewAtlasStore creates a new store instance.
func NewAtlasStore(db *sql.DB) *AtlasStore {
	return &AtlasStore{db: db}
}

// InsertLocationReceipt inserts a new receipt into the database.
func (s *AtlasStore) InsertLocationReceipt(r *LocationReceipt) error {
	query := `
		INSERT INTO atlas_receipts (
			id, entity_id, location_id, issued_by, method,
			confidence, issued_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := s.db.Exec(query,
		r.ID,
		r.EntityID,
		r.LocationID,
		r.IssuedBy,
		r.Method,
		r.Confidence,
		r.IssuedAt,
	)
	if err != nil {
		return fmt.Errorf("insert receipt: %w", err)
	}
	return nil
}

// GetLocationReceipt retrieves a single receipt by ID.
func (s *AtlasStore) GetLocationReceipt(id string) (*LocationReceipt, error) {
	query := `
		SELECT id, entity_id, location_id, issued_by, method,
		       confidence, issued_at
		FROM atlas_receipts WHERE id = $1
	`
	row := s.db.QueryRow(query, id)
	var r LocationReceipt
	if err := row.Scan(
		&r.ID,
		&r.EntityID,
		&r.LocationID,
		&r.IssuedBy,
		&r.Method,
		&r.Confidence,
		&r.IssuedAt,
	); err != nil {
		return nil, fmt.Errorf("get receipt: %w", err)
	}
	return &r, nil
}

// ListLocationReceipts returns all stored receipts.
func (s *AtlasStore) ListLocationReceipts(limit int) ([]*LocationReceipt, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id, entity_id, location_id, issued_by, method,
		       confidence, issued_at
		FROM atlas_receipts
		ORDER BY issued_at DESC
		LIMIT $1
	`
	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("list receipts: %w", err)
	}
	defer rows.Close()

	var receipts []*LocationReceipt
	for rows.Next() {
		var r LocationReceipt
		if err := rows.Scan(
			&r.ID,
			&r.EntityID,
			&r.LocationID,
			&r.IssuedBy,
			&r.Method,
			&r.Confidence,
			&r.IssuedAt,
		); err != nil {
			return nil, err
		}
		receipts = append(receipts, &r)
	}
	return receipts, nil
}

// HealthCheck ensures DB connectivity for Atlas.
func (s *AtlasStore) HealthCheck() error {
	var now time.Time
	err := s.db.QueryRow("SELECT NOW()").Scan(&now)
	if err != nil {
		return fmt.Errorf("atlas store ping failed: %w", err)
	}
	return nil
}
