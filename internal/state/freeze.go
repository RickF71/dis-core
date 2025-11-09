package state

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// copilot: Implement FreezeState model and persistence.
// Add LoadFreezeStates(db), SaveFreezeState(db, state), and GetFreezeMap(db) helpers using pgxpool.Pool.
// Each record corresponds to freeze_states table.

// FreezeState represents a freeze state record in the database
type FreezeState struct {
	ID          string                 `json:"id" db:"id"`
	DomainID    string                 `json:"domain_id" db:"domain_id"`
	Version     string                 `json:"version" db:"version"`
	Active      bool                   `json:"active" db:"active"`
	FreezeData  map[string]interface{} `json:"freeze_data" db:"freeze_data"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	ActivatedAt *time.Time             `json:"activated_at" db:"activated_at"`
	ExpiresAt   *time.Time             `json:"expires_at" db:"expires_at"`
}

// LoadFreezeStates retrieves all freeze states from the database
func LoadFreezeStates(ctx context.Context, db *pgxpool.Pool) ([]FreezeState, error) {
	query := `
		SELECT id, domain_id, version, active, freeze_data, created_at, updated_at, activated_at, expires_at
		FROM freeze_states
		ORDER BY updated_at DESC`

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var states []FreezeState
	for rows.Next() {
		var state FreezeState
		var freezeDataBytes []byte

		err := rows.Scan(
			&state.ID,
			&state.DomainID,
			&state.Version,
			&state.Active,
			&freezeDataBytes,
			&state.CreatedAt,
			&state.UpdatedAt,
			&state.ActivatedAt,
			&state.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal freeze_data JSON
		if len(freezeDataBytes) > 0 {
			err = json.Unmarshal(freezeDataBytes, &state.FreezeData)
			if err != nil {
				return nil, err
			}
		}

		states = append(states, state)
	}

	return states, nil
}

// SaveFreezeState inserts or updates a freeze state in the database
func SaveFreezeState(ctx context.Context, db *pgxpool.Pool, state *FreezeState) error {
	// Marshal freeze_data to JSON
	freezeDataBytes, err := json.Marshal(state.FreezeData)
	if err != nil {
		return err
	}

	// Check if record exists
	var exists bool
	err = db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM freeze_states WHERE id = $1)", state.ID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		// Update existing record
		query := `
			UPDATE freeze_states
			SET domain_id = $2, version = $3, active = $4, freeze_data = $5, updated_at = $6, activated_at = $7, expires_at = $8
			WHERE id = $1`

		_, err = db.Exec(ctx, query,
			state.ID,
			state.DomainID,
			state.Version,
			state.Active,
			freezeDataBytes,
			time.Now(),
			state.ActivatedAt,
			state.ExpiresAt,
		)
	} else {
		// Insert new record
		query := `
			INSERT INTO freeze_states (id, domain_id, version, active, freeze_data, created_at, updated_at, activated_at, expires_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

		_, err = db.Exec(ctx, query,
			state.ID,
			state.DomainID,
			state.Version,
			state.Active,
			freezeDataBytes,
			time.Now(),
			time.Now(),
			state.ActivatedAt,
			state.ExpiresAt,
		)
	}

	return err
}

// GetFreezeMap returns a map of domain IDs to their freeze states
func GetFreezeMap(ctx context.Context, db *pgxpool.Pool) (map[string]interface{}, error) {
	states, err := LoadFreezeStates(ctx, db)
	if err != nil {
		return nil, err
	}

	freezeMap := make(map[string]interface{})

	for _, state := range states {
		if state.Active {
			freezeMap[state.DomainID] = map[string]interface{}{
				"id":           state.ID,
				"version":      state.Version,
				"activated_at": state.ActivatedAt,
				"expires_at":   state.ExpiresAt,
				"freeze_data":  state.FreezeData,
			}
		}
	}

	return freezeMap, nil
}

// GetActiveFreezeStates returns only active freeze states
func GetActiveFreezeStates(ctx context.Context, db *pgxpool.Pool) ([]FreezeState, error) {
	query := `
		SELECT id, domain_id, version, active, freeze_data, created_at, updated_at, activated_at, expires_at
		FROM freeze_states
		WHERE active = true
		ORDER BY activated_at DESC`

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var states []FreezeState
	for rows.Next() {
		var state FreezeState
		var freezeDataBytes []byte

		err := rows.Scan(
			&state.ID,
			&state.DomainID,
			&state.Version,
			&state.Active,
			&freezeDataBytes,
			&state.CreatedAt,
			&state.UpdatedAt,
			&state.ActivatedAt,
			&state.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal freeze_data JSON
		if len(freezeDataBytes) > 0 {
			err = json.Unmarshal(freezeDataBytes, &state.FreezeData)
			if err != nil {
				return nil, err
			}
		}

		states = append(states, state)
	}

	return states, nil
}

// ActivateFreezeState activates a freeze state by ID
func ActivateFreezeState(ctx context.Context, db *pgxpool.Pool, stateID string) error {
	now := time.Now()
	query := `
		UPDATE freeze_states
		SET active = true, activated_at = $2, updated_at = $3
		WHERE id = $1`

	_, err := db.Exec(ctx, query, stateID, &now, &now)
	return err
}

// DeactivateFreezeState deactivates a freeze state by ID
func DeactivateFreezeState(ctx context.Context, db *pgxpool.Pool, stateID string) error {
	now := time.Now()
	query := `
		UPDATE freeze_states
		SET active = false, updated_at = $2
		WHERE id = $1`

	_, err := db.Exec(ctx, query, stateID, &now)
	return err
}
