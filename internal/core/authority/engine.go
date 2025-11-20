package authority

import (
	"context"
	"fmt"
	"time"

	dbstore "dis-core/internal/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds minimal engine configuration passed during bootstrap.
// Expand as needed in later MX steps.
type Config struct {
	// Intentionally minimal for MX-3.4 — add fields later (e.g., log level)
}

// Engine is the structural governance engine for DIS.
type Engine struct {
	// DB is the backing Postgres connection pool used by engine operations.
	// It's set during NewEngine and used by newly-introduced seat methods.
	DB *pgxpool.Pool
}

// FreezeSeat atomically freezes a seat and records a receipt within the same tx.
func (e *Engine) FreezeSeat(ctx context.Context, seatID int64, scope string) error {
	if e.DB == nil {
		return fmt.Errorf("freeze seat: db pool not configured")
	}
	tx, err := e.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("freeze seat: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := dbstore.FreezeSeatTx(ctx, tx, seatID, scope); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("freeze seat: %w", err)
	}

	payload := map[string]any{"seat_id": seatID, "scope": scope}
	if err := e.RecordSeatReceiptTx(ctx, tx, fmt.Sprintf("%d", seatID), "seat.freeze.v1", payload); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("freeze seat: failed to record receipt: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("freeze seat: commit tx: %w", err)
	}
	return nil
}

// UnfreezeSeat atomically unfreezes a seat and records a receipt within the same tx.
func (e *Engine) UnfreezeSeat(ctx context.Context, seatID int64, scope string) error {
	if e.DB == nil {
		return fmt.Errorf("unfreeze seat: db pool not configured")
	}
	tx, err := e.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("unfreeze seat: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := dbstore.UnfreezeSeatTx(ctx, tx, seatID, scope); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("unfreeze seat: %w", err)
	}

	payload := map[string]any{"seat_id": seatID, "scope": scope}
	if err := e.RecordSeatReceiptTx(ctx, tx, fmt.Sprintf("%d", seatID), "seat.unfreeze.v1", payload); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("unfreeze seat: failed to record receipt: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("unfreeze seat: commit tx: %w", err)
	}
	return nil
}

// OverrideFreeze atomically records a freeze override and its receipt.
func (e *Engine) OverrideFreeze(ctx context.Context, seatID int64, scope string, reason string) error {
	if e.DB == nil {
		return fmt.Errorf("override freeze: db pool not configured")
	}
	tx, err := e.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("override freeze: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := dbstore.OverrideFreezeTx(ctx, tx, seatID, scope, reason); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("override freeze: %w", err)
	}

	payload := map[string]any{"seat_id": seatID, "scope": scope, "reason": reason}
	if err := e.RecordSeatReceiptTx(ctx, tx, fmt.Sprintf("%d", seatID), "seat.freeze.override.v1", payload); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("override freeze: failed to record receipt: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("override freeze: commit tx: %w", err)
	}
	return nil
}

// --- Domain-level freeze lifecycle ---

// FreezeDomain inserts or replaces an active freeze record for a domain scope and records a receipt.
func (e *Engine) FreezeDomain(ctx context.Context, domainID uuid.UUID, scope string, reason string, createdBy uuid.UUID, ttl *time.Time) (uuid.UUID, error) {
	if e.DB == nil {
		return uuid.Nil, fmt.Errorf("freeze domain: db pool not configured")
	}
	tx, err := e.DB.Begin(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("freeze domain: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Deactivate existing active freeze for same domain+scope
	if err := dbstore.DeactivateDomainFreezeTx(ctx, tx, domainID.String(), scope); err != nil {
		_ = tx.Rollback(ctx)
		return uuid.Nil, fmt.Errorf("freeze domain: deactivate existing: %w", err)
	}

	// Insert new freeze
	id, err := dbstore.InsertDomainFreezeTx(ctx, tx, domainID.String(), scope, reason, createdBy.String(), ttl, nil)
	if err != nil {
		_ = tx.Rollback(ctx)
		return uuid.Nil, fmt.Errorf("freeze domain: insert: %w", err)
	}

	// Record receipt inside tx
	payload := map[string]any{"domain_id": domainID.String(), "scope": scope, "reason": reason}
	if ttl != nil {
		payload["ttl_until"] = ttl.Format(time.RFC3339)
	}
	if err := e.RecordDomainReceiptTx(ctx, tx, domainID.String(), "domain.freeze.v1", payload); err != nil {
		_ = tx.Rollback(ctx)
		return uuid.Nil, fmt.Errorf("freeze domain: failed to record receipt: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("freeze domain: commit tx: %w", err)
	}

	uid, _ := uuid.Parse(id)
	return uid, nil
}

// UnfreezeDomain marks active freezes for the domain+scope as inactive and records a receipt.
func (e *Engine) UnfreezeDomain(ctx context.Context, domainID uuid.UUID, scope string, createdBy uuid.UUID, reason string) error {
	if e.DB == nil {
		return fmt.Errorf("unfreeze domain: db pool not configured")
	}
	tx, err := e.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("unfreeze domain: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := dbstore.DeactivateDomainFreezeTx(ctx, tx, domainID.String(), scope); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("unfreeze domain: %w", err)
	}

	payload := map[string]any{"domain_id": domainID.String(), "scope": scope, "reason": reason}
	if err := e.RecordDomainReceiptTx(ctx, tx, domainID.String(), "domain.unfreeze.v1", payload); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("unfreeze domain: failed to record receipt: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("unfreeze domain: commit tx: %w", err)
	}
	return nil
}

// OverrideFreezeDomain records an override freeze referencing a previous freeze and records a receipt.
func (e *Engine) OverrideFreezeDomain(ctx context.Context, domainID uuid.UUID, scope string, createdBy uuid.UUID, reason string, priorFreezeID uuid.UUID) (uuid.UUID, error) {
	if e.DB == nil {
		return uuid.Nil, fmt.Errorf("override freeze domain: db pool not configured")
	}
	tx, err := e.DB.Begin(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("override freeze domain: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Mark prior freeze inactive
	if err := dbstore.DeactivateDomainFreezeTx(ctx, tx, domainID.String(), scope); err != nil {
		_ = tx.Rollback(ctx)
		return uuid.Nil, fmt.Errorf("override freeze domain: deactivate prior: %w", err)
	}

	id, err := dbstore.InsertDomainFreezeOverrideTx(ctx, tx, domainID.String(), scope, reason, createdBy.String(), priorFreezeID.String())
	if err != nil {
		_ = tx.Rollback(ctx)
		return uuid.Nil, fmt.Errorf("override freeze domain: insert override: %w", err)
	}

	payload := map[string]any{"domain_id": domainID.String(), "scope": scope, "reason": reason, "override_of": priorFreezeID.String()}
	if err := e.RecordDomainReceiptTx(ctx, tx, domainID.String(), "domain.freeze.override.v1", payload); err != nil {
		_ = tx.Rollback(ctx)
		return uuid.Nil, fmt.Errorf("override freeze domain: failed to record receipt: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("override freeze domain: commit tx: %w", err)
	}
	uid, _ := uuid.Parse(id)
	return uid, nil
}

// GetActiveFreezes returns active freeze records for the domain (in-memory mapping)
func (e *Engine) GetActiveFreezes(ctx context.Context, domainID uuid.UUID) ([]dbstore.FreezeRecord, error) {
	if e.DB == nil {
		return nil, fmt.Errorf("get active freezes: db pool not configured")
	}
	frs, err := dbstore.GetActiveFreezes(ctx, e.DB, domainID.String())
	if err != nil {
		return nil, fmt.Errorf("get active freezes: %w", err)
	}
	return frs, nil
}

// RecordDomainReceiptTx records a domain-related receipt inside the provided tx.
func (e *Engine) RecordDomainReceiptTx(ctx context.Context, tx pgx.Tx, domainID string, typ string, payload map[string]any) error {
	// Attach domain_id
	payload["domain_id"] = domainID

	// Determine previous hash on domain receipts (simple deterministic approach)
	var prevHash string
	// Inspect schema
	var hasPayloadCol bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'receipts' AND column_name = 'payload'
		)
	`).Scan(&hasPayloadCol); err != nil {
		return fmt.Errorf("record domain receipt tx: inspect receipts schema: %w", err)
	}

	if hasPayloadCol {
		rows, err := tx.Query(ctx, `SELECT payload FROM receipts WHERE domain = $1 ORDER BY created_at ASC`, domainID)
		if err == nil {
			defer rows.Close()
			var last map[string]any
			for rows.Next() {
				if err := rows.Scan(&last); err != nil {
					return fmt.Errorf("record domain receipt tx: scan payload: %w", err)
				}
			}
			if last != nil {
				if h, ok := last["hash"].(string); ok {
					prevHash = h
				}
			}
		}
	}
	payload["prev_hash"] = prevHash

	// Compute simple hash and insert similarly to seat receipts
	// (use same type/payload schema detection below)
	// actor
	var actor string
	// best-effort: we don't import authpkg here to avoid cycles; leave actor empty

	// Detect which receipts schema to use
	var hasTypeCol bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'receipts' AND column_name = 'type'
		)
	`).Scan(&hasTypeCol); err != nil {
		return fmt.Errorf("record domain receipt tx: inspect receipts schema for type column: %w", err)
	}

	id := uuid.NewString()
	if hasTypeCol {
		if _, err := tx.Exec(ctx, `
			INSERT INTO receipts (id, type, actor, target, domain, payload, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, id, typ, actor, domainID, domainID, payload, time.Now().UTC()); err != nil {
			return fmt.Errorf("record domain receipt tx: insert (type schema): %w", err)
		}
		return nil
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO receipts (id, receipt_type, issued_by, event_id, policy_ref, metadata, issued_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, id, typ, actor, domainID, payload["policy_ref"], payload, time.Now().UTC()); err != nil {
		return fmt.Errorf("record domain receipt tx: insert (alt schema): %w", err)
	}
	return nil
}

// NewEngine initializes a new authority engine with configuration and DB pool.
func NewEngine(cfg *Config, db *pgxpool.Pool) *Engine {
	// TODO: initialize internal stores, caches, policy clients using cfg + db
	_ = cfg
	return &Engine{DB: db}
}

// StatusSummary mirrors the real data returned by the old status handler.
type StatusSummary struct {
	OK            bool     `json:"ok"`
	ActiveDomains int      `json:"active_domains,omitempty"`
	Frozen        []string `json:"frozen,omitempty"`
	Notes         []string `json:"notes,omitempty"`
}

// Status returns a structural summary of the current authority state.
// Logic from internal/api/authority_status.go is ported here.
func (e *Engine) Status() *StatusSummary {
	// TODO: replace placeholder with real logic extracted from old handler.

	return &StatusSummary{
		OK:            true,
		ActiveDomains: 1,
		Frozen:        []string{},
		Notes:         []string{"MX-2.1 placeholder"},
	}
}

// GetStatus returns the new canonical AuthorityStatus used by the
// API delegators. This is the MX-3.1 migration point. Real logic will
// evolve in subsequent MX-3 iterations.
func (e *Engine) GetStatus(ctx context.Context) (*AuthorityStatus, error) {
	// Temporary, real logic will evolve across MX-3:
	status := &AuthorityStatus{
		EngineHealthy: true,
		FreezeActive:  false,
		Notes:         "MX-3.1: Status migrated to new authority engine.",
	}
	return status, nil
}

// Override applies a temporary override or administrative change to a domain.
// This is a small placeholder to support API delegators; real logic will
// be implemented in MX-3.4+.
func (e *Engine) Override(domainID string, changes map[string]interface{}) (*OverrideResult, error) {
	// Return a minimal acknowledgement
	return &OverrideResult{
		DomainID: domainID,
		Changed:  changes,
		Notes:    "MX-3.5 placeholder: override applied (no-op)",
	}, nil
}

// Branch creates a new branched domain. Placeholder implementation for MX-3.6.
func (e *Engine) Branch(ctx context.Context, db *pgxpool.Pool, req CreateBranchRequest) (uuid.UUID, error) {
	// TODO: implement real branch creation with DB writes and receipts.
	id := uuid.New()
	return id, nil
}

// GetBranchInfo returns minimal branch metadata for a given domain.
func (e *Engine) GetBranchInfo(ctx context.Context, db *pgxpool.Pool, domainID uuid.UUID) (*BranchInfo, error) {
	// Placeholder: return depth 1
	return &BranchInfo{
		DomainID:    domainID,
		ParentID:    uuid.Nil,
		BranchDepth: 1,
		Notes:       "MX-3.6 placeholder branch info",
	}, nil
}

// CanInstantiateSeatFromContract checks DSCI eligibility. Placeholder.
func (e *Engine) CanInstantiateSeatFromContract(ctx context.Context, db *pgxpool.Pool, userID, originalDomainID, branchDomainID uuid.UUID) (bool, string, error) {
	// Default to not eligible — real logic to be migrated from legacy authority.
	return false, "not implemented", nil
}
