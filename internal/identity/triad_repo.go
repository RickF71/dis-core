package identity

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CASVersionMismatch is returned when a CAS operation fails due to state mismatch
type CASVersionMismatch struct {
	Expected string
	Actual   string
}

func (e *CASVersionMismatch) Error() string {
	return fmt.Sprintf("CAS version mismatch: expected state '%s', found '%s'", e.Expected, e.Actual)
}

// TriadRepository handles identity_seats table operations
type TriadRepository struct {
	db *pgxpool.Pool
}

// NewTriadRepository creates a new triad repository
func NewTriadRepository(db *pgxpool.Pool) *TriadRepository {
	return &TriadRepository{db: db}
}

// CreateIdentitySeat creates a single identity seat
func (r *TriadRepository) CreateIdentitySeat(ctx context.Context, identityID, seatType, state string) (*IdentitySeat, error) {
	var seat IdentitySeat
	now := time.Now()

	query := `
		INSERT INTO identity_seats (identity_id, seat_type, state, assigned_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (identity_id, seat_type) DO UPDATE
		SET state = EXCLUDED.state, assigned_at = EXCLUDED.assigned_at
		RETURNING id, identity_id, seat_type, state, assigned_at, occupied_at, frozen_at, metadata, created_at
	`

	var assignedAt *time.Time
	if state == SeatStateAssigned || state == SeatStateOccupied {
		assignedAt = &now
	}

	err := r.db.QueryRow(ctx, query, identityID, seatType, state, assignedAt, now).Scan(
		&seat.ID, &seat.IdentityID, &seat.SeatType, &seat.State,
		&seat.AssignedAt, &seat.OccupiedAt, &seat.FrozenAt, &seat.Metadata, &seat.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create identity seat: %w", err)
	}

	return &seat, nil
}

// InitializeTriad creates all three seats for an identity (idempotent)
func (r *TriadRepository) InitializeTriad(ctx context.Context, identityID string) (*IdentityTriad, error) {
	// Create terra, numen, lima seats
	seatTypes := []string{SeatTypeTerra, SeatTypeNumen, SeatTypeLima}
	triad := &IdentityTriad{}

	for _, seatType := range seatTypes {
		// Check if seat exists
		existing, err := r.GetIdentitySeat(ctx, identityID, seatType)
		if err == nil && existing != nil {
			// Seat already exists, assign to triad
			switch seatType {
			case SeatTypeTerra:
				triad.Terra = existing
			case SeatTypeNumen:
				triad.Numen = existing
			case SeatTypeLima:
				triad.Lima = existing
			}
			continue
		}

		// Create new seat in ASSIGNED state (GOV-1: auto-assignment)
		seat, err := r.CreateIdentitySeat(ctx, identityID, seatType, SeatStateAssigned)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize %s seat: %w", seatType, err)
		}

		switch seatType {
		case SeatTypeTerra:
			triad.Terra = seat
		case SeatTypeNumen:
			triad.Numen = seat
		case SeatTypeLima:
			triad.Lima = seat
		}
	}

	return triad, nil
}

// GetIdentitySeat retrieves a specific seat for an identity
func (r *TriadRepository) GetIdentitySeat(ctx context.Context, identityID, seatType string) (*IdentitySeat, error) {
	var seat IdentitySeat

	query := `
		SELECT id, identity_id, seat_type, state, assigned_at, occupied_at, frozen_at, metadata, created_at
		FROM identity_seats
		WHERE identity_id = $1 AND seat_type = $2
	`

	err := r.db.QueryRow(ctx, query, identityID, seatType).Scan(
		&seat.ID, &seat.IdentityID, &seat.SeatType, &seat.State,
		&seat.AssignedAt, &seat.OccupiedAt, &seat.FrozenAt, &seat.Metadata, &seat.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("identity seat not found: %s/%s", identityID, seatType)
		}
		return nil, fmt.Errorf("failed to get identity seat: %w", err)
	}

	return &seat, nil
}

// GetIdentityTriad retrieves all three seats for an identity
func (r *TriadRepository) GetIdentityTriad(ctx context.Context, identityID string) (*IdentityTriad, error) {
	terra, err := r.GetIdentitySeat(ctx, identityID, SeatTypeTerra)
	if err != nil {
		return nil, fmt.Errorf("terra seat not found: %w", err)
	}

	numen, err := r.GetIdentitySeat(ctx, identityID, SeatTypeNumen)
	if err != nil {
		return nil, fmt.Errorf("numen seat not found: %w", err)
	}

	lima, err := r.GetIdentitySeat(ctx, identityID, SeatTypeLima)
	if err != nil {
		return nil, fmt.Errorf("lima seat not found: %w", err)
	}

	return &IdentityTriad{
		Terra: terra,
		Numen: numen,
		Lima:  lima,
	}, nil
}

// UpdateSeatState updates the state of an identity seat
func (r *TriadRepository) UpdateSeatState(ctx context.Context, seatID uuid.UUID, newState string) error {
	now := time.Now()
	var stateField string

	switch newState {
	case SeatStateAssigned:
		stateField = "assigned_at"
	case SeatStateOccupied:
		stateField = "occupied_at"
	case SeatStateFrozen:
		stateField = "frozen_at"
	default:
		stateField = ""
	}

	var query string
	if stateField != "" {
		query = fmt.Sprintf(`
			UPDATE identity_seats
			SET state = $1, %s = $2
			WHERE id = $3
		`, stateField)
		_, err := r.db.Exec(ctx, query, newState, now, seatID)
		return err
	}

	query = `UPDATE identity_seats SET state = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, newState, seatID)
	return err
}

// GetAllIdentities retrieves all identity IDs for bulk operations
func (r *TriadRepository) GetAllIdentities(ctx context.Context) ([]string, error) {
	rows, err := r.db.Query(ctx, `SELECT id FROM identities`)
	if err != nil {
		return nil, fmt.Errorf("failed to query identities: %w", err)
	}
	defer rows.Close()

	identityIDs := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan identity: %w", err)
		}
		identityIDs = append(identityIDs, id)
	}

	return identityIDs, nil
}

// GetMissingTriads returns identities that don't have complete triads
func (r *TriadRepository) GetMissingTriads(ctx context.Context) ([]string, error) {
	query := `
		SELECT identity_id
		FROM identity_triad_status
		WHERE terra_state IS NULL OR numen_state IS NULL OR lima_state IS NULL
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query missing triads: %w", err)
	}
	defer rows.Close()

	missing := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan missing triad: %w", err)
		}
		missing = append(missing, id)
	}

	return missing, nil
}

// UpdateSeatStateCAS performs a compare-and-swap update on a seat state
// Returns the new state on success, or an error if the current state doesn't match 'from'
// CASUpdateResult contains the result of a Compare-And-Swap operation
type CASUpdateResult struct {
	NewState        string
	PreviousVersion int64
	NewVersion      int64
	Success         bool
}

// UpdateSeatStateCAS performs a compare-and-swap update on a seat state with version tracking
// Returns the new state and version information for audit logging
func (r *TriadRepository) UpdateSeatStateCAS(ctx context.Context, identityID, seatType, from, to string) (CASUpdateResult, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return CASUpdateResult{Success: false}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Lock the row and read current state + version
	var currentState string
	var currentVersion int64
	query := `SELECT state, version FROM identity_seats WHERE identity_id = $1 AND seat_type = $2 FOR UPDATE`
	err = tx.QueryRow(ctx, query, identityID, seatType).Scan(&currentState, &currentVersion)
	if err != nil {
		if err == pgx.ErrNoRows {
			return CASUpdateResult{Success: false}, fmt.Errorf("seat not found: identity=%s, type=%s", identityID, seatType)
		}
		return CASUpdateResult{Success: false}, fmt.Errorf("failed to read current state: %w", err)
	}

	// Compare current state with expected 'from' state
	if currentState != from {
		return CASUpdateResult{
			NewState:        currentState,
			PreviousVersion: currentVersion,
			NewVersion:      currentVersion,
			Success:         false,
		}, &CASVersionMismatch{Expected: from, Actual: currentState}
	}

	// Update to new state with incremented version and appropriate timestamp
	newVersion := currentVersion + 1
	now := time.Now()
	var updateQuery string
	var args []interface{}

	switch to {
	case SeatStateOccupied:
		updateQuery = `UPDATE identity_seats SET state = $1, version = $2, occupied_at = $3 WHERE identity_id = $4 AND seat_type = $5`
		args = []interface{}{to, newVersion, now, identityID, seatType}
	case SeatStateFrozen:
		updateQuery = `UPDATE identity_seats SET state = $1, version = $2, frozen_at = $3 WHERE identity_id = $4 AND seat_type = $5`
		args = []interface{}{to, newVersion, now, identityID, seatType}
	default:
		updateQuery = `UPDATE identity_seats SET state = $1, version = $2 WHERE identity_id = $3 AND seat_type = $4`
		args = []interface{}{to, newVersion, identityID, seatType}
	}

	_, err = tx.Exec(ctx, updateQuery, args...)
	if err != nil {
		return CASUpdateResult{Success: false}, fmt.Errorf("failed to update seat state: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return CASUpdateResult{Success: false}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return CASUpdateResult{
		NewState:        to,
		PreviousVersion: currentVersion,
		NewVersion:      newVersion,
		Success:         true,
	}, nil
}
