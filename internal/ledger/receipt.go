package ledger

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Receipt represents a ci.call.v1 log entry.
type Receipt struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Actor     string          `json:"actor"`
	Target    string          `json:"target"`
	Domain    string          `json:"domain"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
}

// NewReceipt constructs a new in-memory receipt before DB insert.
func NewReceipt(rtype, actor, target, domain string, payload interface{}) (*Receipt, error) {
	var payloadJSON []byte
	if payload != nil {
		var err error
		payloadJSON, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal payload: %w", err)
		}
	} else {
		payloadJSON = []byte(`{}`)
	}

	id, err := generateReceiptID()
	if err != nil {
		return nil, err
	}

	return &Receipt{
		ID:        id,
		Type:      rtype,
		Actor:     actor,
		Target:    target,
		Domain:    domain,
		Payload:   payloadJSON,
		CreatedAt: time.Now().UTC(),
	}, nil
}

// Insert writes the receipt to the receipts table (no filesystem writes).
func (r *Receipt) Insert(ctx context.Context, tx pgx.Tx) error {
	const q = `
		INSERT INTO receipts (id, type, actor, target, domain, payload, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := tx.Exec(ctx, q, r.ID, r.Type, r.Actor, r.Target, r.Domain, r.Payload, r.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert receipt: %w", err)
	}
	return nil
}

// generateReceiptID creates a cryptographically derived ID without filesystem dependency.
func generateReceiptID() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	hash := sha256.Sum256(buf)
	return "rcpt-" + base64.RawURLEncoding.EncodeToString(hash[:8]), nil
}

// VerifyHash creates a short hash summary of the payload for integrity auditing.
func (r *Receipt) VerifyHash() string {
	sum := sha256.Sum256(r.Payload)
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// Save convenience wrapper: begins + commits a tx for standalone inserts.
func (r *Receipt) Save(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := r.Insert(ctx, tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
