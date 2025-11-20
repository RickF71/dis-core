package authority

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	authpkg "dis-core/internal/auth"
	dbstore "dis-core/internal/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// InstantiateSeat creates a new seat in the target domain for the given identity.
// This implementation persists the seat into the database and returns the
// constructed Seat object.
func (e *Engine) InstantiateSeat(ctx context.Context, domainID string, identityID string) (*Seat, error) {
	// Policy check: ensure caller and contract allow seat instantiation
	allowed, perr := e.CanInstantiateSeat(ctx, domainID, identityID, "")
	if perr != nil {
		return nil, fmt.Errorf("policy: %w", perr)
	}
	if !allowed {
		return nil, fmt.Errorf("instantiate seat: unauthorized by policy in domain %s", domainID)
	}

	id := uuid.NewString()
	s := &Seat{
		ID:         id,
		DomainID:   domainID,
		IdentityID: identityID,
		Kind:       SeatKindMember,
		CreatedAt:  time.Now().UTC(),
	}

	if e.DB == nil {
		return nil, fmt.Errorf("instantiate seat: db pool not configured")
	}

	// Begin transaction to ensure seat insertion and receipt recording are atomic.
	tx, err := e.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("instantiate seat: begin tx: %w", err)
	}
	// Ensure rollback on error/path exit if not committed
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := dbstore.InsertSeatTx(ctx, tx, s.ID, s.DomainID, s.IdentityID, string(s.Kind)); err != nil {
		_ = tx.Rollback(ctx)
		return nil, fmt.Errorf("instantiate seat: %w", err)
	}

	rcptPayload := map[string]interface{}{
		"policy_ref":     "policy:CanInstantiateSeat",
		"policy_allowed": true,
	}

	if err := e.RecordSeatReceiptTx(ctx, tx, s.ID, "seat.instantiate.v1", rcptPayload); err != nil {
		_ = tx.Rollback(ctx)
		return nil, fmt.Errorf("instantiate seat: failed to record receipt: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("instantiate seat: commit tx: %w", err)
	}

	return s, nil
}

// SeatStatus fetches seat information from the DB and returns runtime status.
// If the seat is not found, (nil, nil) is returned.
func (e *Engine) SeatStatus(ctx context.Context, seatID string) (*SeatStatus, error) {
	if e.DB == nil {
		return nil, fmt.Errorf("seat status: db pool not configured")
	}

	// Policy check: visibility for this seat
	allowed, perr := e.CanViewSeat(ctx, seatID)
	if perr != nil {
		return nil, fmt.Errorf("policy: %w", perr)
	}
	if !allowed {
		// soft-deny: caller should get no visibility
		return nil, nil
	}

	row, err := dbstore.GetSeat(ctx, e.DB, seatID)
	if err != nil {
		return nil, fmt.Errorf("seat status: %w", err)
	}
	if row == nil {
		return nil, nil
	}

	// Convert map to Seat
	s := &Seat{ID: seatID}
	if v, ok := row["domain_id"].(string); ok {
		s.DomainID = v
	}
	if v, ok := row["identity_id"].(string); ok {
		s.IdentityID = v
	}
	if v, ok := row["kind"].(string); ok {
		s.Kind = SeatKind(v)
	}
	if v, ok := row["created_at"].(time.Time); ok {
		s.CreatedAt = v
	}

	status := &SeatStatus{Seat: s, LineageRef: "", Active: true}
	if st, ok := row["status"].(string); ok {
		// interpret status if needed; for now Active = status == "active"
		status.Active = (st == "active")
	}

	// Best-effort: record a view receipt for status lookups (synchronous but non-fatal)
	_ = e.RecordSeatReceipt(ctx, seatID, "seat.status.view.v1", map[string]interface{}{"policy_ref": "policy:CanViewSeat", "policy_allowed": true})

	return status, nil
}

// ListDomainSeats returns seats persisted under a domain.
func (e *Engine) ListDomainSeats(ctx context.Context, domainID string) ([]*Seat, error) {
	if e.DB == nil {
		return nil, fmt.Errorf("list domain seats: db pool not configured")
	}

	// Policy check: can caller list seats in this domain?
	allowed, perr := e.CanListDomainSeats(ctx, domainID)
	if perr != nil {
		return nil, fmt.Errorf("policy: %w", perr)
	}
	if !allowed {
		// unauthorized -> return empty slice (no visibility)
		return []*Seat{}, nil
	}

	rows, err := dbstore.ListSeatsByDomain(ctx, e.DB, domainID)
	if err != nil {
		return nil, fmt.Errorf("list domain seats: %w", err)
	}

	var out []*Seat
	for _, r := range rows {
		s := &Seat{}
		if v, ok := r["id"].(string); ok {
			s.ID = v
		}
		if v, ok := r["domain_id"].(string); ok {
			s.DomainID = v
		}
		if v, ok := r["identity_id"].(string); ok {
			s.IdentityID = v
		}
		if v, ok := r["kind"].(string); ok {
			s.Kind = SeatKind(v)
		}
		if v, ok := r["created_at"].(time.Time); ok {
			s.CreatedAt = v
		}
		out = append(out, s)
	}

	// Best-effort: record list receipts per-seat (non-fatal)
	go func(ctx context.Context, seats []*Seat) {
		for _, ss := range seats {
			_ = e.RecordSeatReceipt(ctx, ss.ID, "seat.list.v1", map[string]interface{}{"policy_ref": "policy:CanListDomainSeats", "policy_allowed": true})
		}
	}(context.Background(), out)
	return out, nil
}

// ResolveSeatLineage currently returns a basic lineage containing receipts
// related to the seat. This is intentionally minimal and will be extended in
// subsequent MX iterations.
func (e *Engine) ResolveSeatLineage(ctx context.Context, seatID string) (*SeatLineage, error) {
	if e.DB == nil {
		return nil, fmt.Errorf("resolve seat lineage: db pool not configured")
	}

	// Attempt to fetch seat to gather domain/identity fields
	row, err := dbstore.GetSeat(ctx, e.DB, seatID)
	if err != nil {
		return nil, fmt.Errorf("resolve seat lineage: %w", err)
	}
	if row == nil {
		return nil, nil
	}

	sl := &SeatLineage{SeatID: seatID, Ancestors: []string{}, Receipts: []string{}}
	if v, ok := row["domain_id"].(string); ok {
		sl.DomainID = v
	}
	if v, ok := row["identity_id"].(string); ok {
		sl.IdentityID = v
	}

	// Policy check: can caller view the lineage for this seat?
	allowed, perr := e.CanViewSeatLineage(ctx, sl.DomainID, sl.IdentityID)
	if perr != nil {
		return nil, fmt.Errorf("policy: %w", perr)
	}
	if !allowed {
		// soft-deny: return an empty SeatLineage (no receipts)
		return &SeatLineage{SeatID: seatID}, nil
	}

	receipts, err := dbstore.ListSeatReceipts(ctx, e.DB, seatID)
	if err != nil {
		return nil, fmt.Errorf("resolve seat lineage: %w", err)
	}
	for _, r := range receipts {
		if id, ok := r["id"].(string); ok {
			sl.Receipts = append(sl.Receipts, id)
		}
	}

	// Best-effort: record lineage view receipt
	_ = e.RecordSeatReceipt(ctx, seatID, "seat.lineage.view.v1", map[string]interface{}{"policy_ref": "policy:CanViewSeatLineage", "policy_allowed": true})

	return sl, nil
}

// RecordSeatReceipt persists or records a receipt related to a seat lifecycle
// event. It computes a deterministic hash chain (prev_hash, hash) and stores
// the receipt in the generic receipts table using internal/db.SaveReceipt.
func (e *Engine) RecordSeatReceipt(ctx context.Context, seatID string, typ string, payload any) error {
	if e.DB == nil {
		return fmt.Errorf("record seat receipt: db pool not configured")
	}

	// Fetch seat meta (domain/identity) to enrich the payload and set domain
	var domainID, identityID string
	if row, err := dbstore.GetSeat(ctx, e.DB, seatID); err == nil && row != nil {
		if v, ok := row["domain_id"].(string); ok {
			domainID = v
		}
		if v, ok := row["identity_id"].(string); ok {
			identityID = v
		}
	}

	// Normalize payload into a map[string]interface{}
	var pl map[string]interface{}
	switch t := payload.(type) {
	case map[string]interface{}:
		pl = t
	default:
		// Marshal then unmarshal to a generic map
		b, err := json.Marshal(t)
		if err != nil {
			return fmt.Errorf("record seat receipt: failed to marshal payload: %w", err)
		}
		if err := json.Unmarshal(b, &pl); err != nil {
			return fmt.Errorf("record seat receipt: failed to unmarshal payload: %w", err)
		}
	}

	// Attach canonical fields
	pl["seat_id"] = seatID
	if domainID != "" {
		pl["domain_id"] = domainID
	}
	if identityID != "" {
		pl["identity_id"] = identityID
	}

	// Determine previous hash (if any)
	var prevHash string
	receipts, err := dbstore.ListSeatReceipts(ctx, e.DB, seatID)
	if err == nil && len(receipts) > 0 {
		last := receipts[len(receipts)-1]
		if payloadObj, ok := last["payload"].(map[string]interface{}); ok {
			if h, ok := payloadObj["hash"].(string); ok {
				prevHash = h
			}
		}
	}
	pl["prev_hash"] = prevHash

	// Compute deterministic hash: sha256 of (type + '|' + prevHash + '|' + json(payload))
	b, err := json.Marshal(pl)
	if err != nil {
		return fmt.Errorf("record seat receipt: failed to marshal payload for hash: %w", err)
	}
	h := sha256.Sum256(append([]byte(typ+"|"+prevHash+"|"), b...))
	hashStr := hex.EncodeToString(h[:])
	pl["hash"] = hashStr

	// Who issued this receipt?
	var actor string
	if au := authpkg.GetActiveUserFromCtx(ctx); au != nil {
		actor = au.CorporealDomainUID
	}

	// Persist via generic SaveReceipt
	rcpt := &dbstore.Receipt{
		ID:        uuid.NewString(),
		Type:      typ,
		Actor:     actor,
		Target:    seatID,
		Domain:    domainID,
		Payload:   pl,
		CreatedAt: time.Now().UTC(),
	}
	if err := dbstore.SaveReceipt(ctx, e.DB, rcpt); err != nil {
		// Some installations use a different receipts schema (receipt_type / metadata, etc.).
		// Attempt a compatible fallback insert to support both schemas.
		_, err2 := e.DB.Exec(ctx, `
			INSERT INTO receipts (id, receipt_type, issued_by, event_id, policy_ref, metadata, issued_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, rcpt.ID, typ, rcpt.Actor, rcpt.Target, pl["policy_ref"], pl, rcpt.CreatedAt)
		if err2 != nil {
			return fmt.Errorf("record seat receipt: failed to save receipt: %w (fallback: %v)", err, err2)
		}
		return nil
	}

	return nil
}

// RecordSeatReceiptTx is the strict, transactional variant which records a
// seat-related receipt using the provided pgx.Tx. It is deterministic and
// returns an error on any failure so callers can rollback the transaction.
func (e *Engine) RecordSeatReceiptTx(ctx context.Context, tx pgx.Tx, seatID string, typ string, payload map[string]any) error {
	if tx == nil {
		return fmt.Errorf("record seat receipt tx: tx is nil")
	}

	// Fetch seat meta using the provided transaction so uncommitted inserts are visible.
	var domainID, identityID string
	if row, err := dbstore.GetSeatTx(ctx, tx, seatID); err == nil && row != nil {
		if v, ok := row["domain_id"].(string); ok {
			domainID = v
		}
		if v, ok := row["identity_id"].(string); ok {
			identityID = v
		}
	}

	// Attach canonical fields
	payload["seat_id"] = seatID
	if domainID != "" {
		payload["domain_id"] = domainID
	}
	if identityID != "" {
		payload["identity_id"] = identityID
	}

	// Determine previous hash via tx. First detect receipts schema safely to
	// avoid running a failing SELECT that would abort the transaction.
	var prevHash string
	var hasPayloadCol bool
	var err error
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'receipts' AND column_name = 'payload'
		)
	`).Scan(&hasPayloadCol)
	if err != nil {
		return fmt.Errorf("record seat receipt tx: failed to inspect receipts schema: %w", err)
	}

	if hasPayloadCol {
		rows, err := tx.Query(ctx, `
			SELECT payload
			FROM receipts
			WHERE target = $1 OR (payload->>'seat_id') = $1
			ORDER BY created_at ASC
		`, seatID)
		if err != nil {
			return fmt.Errorf("record seat receipt tx: failed to query receipts: %w", err)
		}
		defer rows.Close()
		var lastPayload map[string]interface{}
		for rows.Next() {
			var p map[string]interface{}
			if err := rows.Scan(&p); err != nil {
				return fmt.Errorf("record seat receipt tx: scan payload: %w", err)
			}
			lastPayload = p
		}
		if lastPayload != nil {
			if h, ok := lastPayload["hash"].(string); ok {
				prevHash = h
			}
		}
	} else {
		rows2, err := tx.Query(ctx, `
			SELECT metadata
			FROM receipts
			WHERE event_id = $1 OR (metadata->>'seat_id') = $1
			ORDER BY issued_at ASC
		`, seatID)
		if err != nil {
			return fmt.Errorf("record seat receipt tx: failed to query receipts (alt): %w", err)
		}
		defer rows2.Close()
		var last map[string]interface{}
		for rows2.Next() {
			var metadata map[string]interface{}
			if err := rows2.Scan(&metadata); err != nil {
				return fmt.Errorf("record seat receipt tx: scan metadata: %w", err)
			}
			last = metadata
		}
		if last != nil {
			if h, ok := last["hash"].(string); ok {
				prevHash = h
			}
		}
	}
	payload["prev_hash"] = prevHash

	// Compute hash
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("record seat receipt tx: failed to marshal payload for hash: %w", err)
	}
	h := sha256.Sum256(append([]byte(typ+"|"+prevHash+"|"), b...))
	hashStr := hex.EncodeToString(h[:])
	payload["hash"] = hashStr

	// actor
	var actor string
	if au := authpkg.GetActiveUserFromCtx(ctx); au != nil {
		actor = au.CorporealDomainUID
	}

	// Detect which receipts schema to use (type/payload vs receipt_type/metadata)
	var hasTypeCol bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'receipts' AND column_name = 'type'
		)
	`).Scan(&hasTypeCol); err != nil {
		return fmt.Errorf("record seat receipt tx: failed to inspect receipts schema for type column: %w", err)
	}

	id := uuid.NewString()
	if hasTypeCol {
		_, err = tx.Exec(ctx, `
			INSERT INTO receipts (id, type, actor, target, domain, payload, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, id, typ, actor, seatID, domainID, payload, time.Now().UTC())
		if err != nil {
			return fmt.Errorf("record seat receipt tx: failed to insert receipt (type schema): %w", err)
		}
		return nil
	}

	// Use alternate schema
	_, err = tx.Exec(ctx, `
		INSERT INTO receipts (id, receipt_type, issued_by, event_id, policy_ref, metadata, issued_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, id, typ, actor, seatID, payload["policy_ref"], payload, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("record seat receipt tx: failed to insert receipt (alt schema): %w", err)
	}
	return nil
}
