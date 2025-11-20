package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InsertSeat inserts a new row into domain_seats. It expects caller to provide
// a valid UUID for id and domainID. identityID may be empty (NULL member).
func InsertSeat(ctx context.Context, db *pgxpool.Pool, id, domainID, identityID, kind string) error {
	_, err := db.Exec(ctx, `
		INSERT INTO domain_seats (id, domain_id, member_id, seat_type, created_at)
		VALUES ($1::uuid, $2::uuid, NULLIF($3, '')::uuid, $4, NOW())
	`, id, domainID, identityID, kind)
	if err != nil {
		return fmt.Errorf("failed to insert seat: %w", err)
	}
	return nil
}

// InsertSeatTx inserts a new row into domain_seats using an existing transaction.
func InsertSeatTx(ctx context.Context, tx pgx.Tx, id, domainID, identityID, kind string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO domain_seats (id, domain_id, member_id, seat_type, created_at)
		VALUES ($1::uuid, $2::uuid, NULLIF($3, '')::uuid, $4, NOW())
	`, id, domainID, identityID, kind)
	if err != nil {
		return fmt.Errorf("failed to insert seat (tx): %w", err)
	}
	return nil
}

// GetSeat returns a map representation of a seat by id. Returns (nil, nil) if not found.
func GetSeat(ctx context.Context, db *pgxpool.Pool, id string) (map[string]interface{}, error) {
	row := db.QueryRow(ctx, `
		SELECT id::text, domain_id::text, COALESCE(member_id::text, '') as member_id, seat_type, status, created_at
		FROM domain_seats WHERE id = $1::uuid
	`, id)

	var sid, domainID, memberID, seatType, status string
	var createdAt time.Time
	if err := row.Scan(&sid, &domainID, &memberID, &seatType, &status, &createdAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get seat: %w", err)
	}

	out := map[string]interface{}{
		"id":          sid,
		"domain_id":   domainID,
		"identity_id": memberID,
		"kind":        seatType,
		"status":      status,
		"created_at":  createdAt,
	}
	return out, nil
}

// GetSeatTx is the tx-aware variant of GetSeat which runs inside an existing transaction.
// It mirrors GetSeat but uses the provided pgx.Tx so uncommitted inserts/updates in the
// same transaction are visible to the caller.
func GetSeatTx(ctx context.Context, tx pgx.Tx, id string) (map[string]interface{}, error) {
	row := tx.QueryRow(ctx, `
		SELECT id::text, domain_id::text, COALESCE(member_id::text, '') as member_id, seat_type, status, created_at
		FROM domain_seats WHERE id = $1::uuid
	`, id)

	var sid, domainID, memberID, seatType, status string
	var createdAt time.Time
	if err := row.Scan(&sid, &domainID, &memberID, &seatType, &status, &createdAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get seat (tx): %w", err)
	}

	out := map[string]interface{}{
		"id":          sid,
		"domain_id":   domainID,
		"identity_id": memberID,
		"kind":        seatType,
		"status":      status,
		"created_at":  createdAt,
	}
	return out, nil
}

// ListSeatsByDomain returns a list of seats belonging to a domain (ordered by created_at).
func ListSeatsByDomain(ctx context.Context, db *pgxpool.Pool, domainID string) ([]map[string]interface{}, error) {
	rows, err := db.Query(ctx, `
		SELECT id::text, domain_id::text, COALESCE(member_id::text, '') as member_id, seat_type, status, created_at
		FROM domain_seats
		WHERE domain_id = $1::uuid
		ORDER BY created_at ASC
	`, domainID)
	if err != nil {
		return nil, fmt.Errorf("failed to list seats by domain: %w", err)
	}
	defer rows.Close()

	var out []map[string]interface{}
	for rows.Next() {
		var sid, did, memberID, seatType, status string
		var createdAt time.Time
		if err := rows.Scan(&sid, &did, &memberID, &seatType, &status, &createdAt); err != nil {
			return nil, err
		}
		out = append(out, map[string]interface{}{
			"id":          sid,
			"domain_id":   did,
			"identity_id": memberID,
			"kind":        seatType,
			"status":      status,
			"created_at":  createdAt,
		})
	}
	return out, rows.Err()
}

// ListSeatReceipts queries the receipts table for receipts related to a seat.
// It searches receipts.target and payload->>'seat_id' for matches.
func ListSeatReceipts(ctx context.Context, db *pgxpool.Pool, seatID string) ([]map[string]interface{}, error) {
	rows, err := db.Query(ctx, `
		SELECT id, type, actor, target, domain, payload, created_at
		FROM receipts
		WHERE target = $1 OR (payload->>'seat_id') = $1
		ORDER BY created_at ASC
	`, seatID)
	if err != nil {
		// Support alternate receipts schema (e.g., receipt_type/metadata).
		// Try a compatible query and map results into the same shape.
		rows2, err2 := db.Query(ctx, `
			SELECT id, receipt_type, issued_by, event_id, policy_ref, metadata, issued_at
			FROM receipts
			WHERE event_id = $1 OR (metadata->>'seat_id') = $1
			ORDER BY issued_at ASC
		`, seatID)
		if err2 != nil {
			return nil, fmt.Errorf("failed to list seat receipts: %w", err)
		}
		defer rows2.Close()

		var out2 []map[string]interface{}
		for rows2.Next() {
			var id, rtype, issuedBy, eventID, policyRef string
			var metadata map[string]interface{}
			var createdAt time.Time
			if err := rows2.Scan(&id, &rtype, &issuedBy, &eventID, &policyRef, &metadata, &createdAt); err != nil {
				return nil, err
			}
			// Map to unified shape used by callers
			payload := metadata
			if payload == nil {
				payload = map[string]interface{}{}
			}
			payload["policy_ref"] = policyRef
			out2 = append(out2, map[string]interface{}{
				"id":         id,
				"type":       rtype,
				"actor":      issuedBy,
				"target":     eventID,
				"domain":     "",
				"payload":    payload,
				"created_at": createdAt,
			})
		}
		return out2, rows2.Err()
	}
	defer rows.Close()

	var out []map[string]interface{}
	for rows.Next() {
		var id, typ, actor, target, domain string
		var payload map[string]interface{}
		var createdAt time.Time
		if err := rows.Scan(&id, &typ, &actor, &target, &domain, &payload, &createdAt); err != nil {
			return nil, err
		}
		out = append(out, map[string]interface{}{
			"id":         id,
			"type":       typ,
			"actor":      actor,
			"target":     target,
			"domain":     domain,
			"payload":    payload,
			"created_at": createdAt,
		})
	}
	return out, rows.Err()
}

// FreezeSeatTx records a freeze for a seat inside the provided transaction.
func FreezeSeatTx(ctx context.Context, tx pgx.Tx, seatID int64, scope string) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO seat_freeze (seat_id, scope, frozen_at) VALUES ($1, $2, now())`,
		seatID, scope)
	if err != nil {
		return fmt.Errorf("failed to freeze seat (tx): %w", err)
	}
	return nil
}

// UnfreezeSeatTx removes a freeze record for a seat inside the provided transaction.
func UnfreezeSeatTx(ctx context.Context, tx pgx.Tx, seatID int64, scope string) error {
	_, err := tx.Exec(ctx,
		`DELETE FROM seat_freeze WHERE seat_id=$1 AND scope=$2`,
		seatID, scope)
	if err != nil {
		return fmt.Errorf("failed to unfreeze seat (tx): %w", err)
	}
	return nil
}

// OverrideFreezeTx records an override event for a seat freeze inside tx.
func OverrideFreezeTx(ctx context.Context, tx pgx.Tx, seatID int64, scope, reason string) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO seat_freeze_override (seat_id, scope, reason, created_at) VALUES ($1, $2, $3, now())`,
		seatID, scope, reason)
	if err != nil {
		return fmt.Errorf("failed to insert freeze override (tx): %w", err)
	}
	return nil
}
