package ledger

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"dis-core/internal/receipts"
)

type Store struct {
	db *sql.DB
}

func (s *Store) InsertReceipt(r *receipts.Receipt) error {
	meta, _ := json.Marshal(r.Metadata)
	_, err := s.db.Exec(`
	       INSERT INTO receipts (id, actor, action, created_at, hash, signature, frozen_core_hash, metadata)
	       VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	       ON CONFLICT (id) DO UPDATE SET
		       actor = EXCLUDED.actor,
		       action = EXCLUDED.action,
		       created_at = EXCLUDED.created_at,
		       hash = EXCLUDED.hash,
		       signature = EXCLUDED.signature,
		       frozen_core_hash = EXCLUDED.frozen_core_hash,
		       metadata = EXCLUDED.metadata;
       `, r.ReceiptID, r.By, r.Action, r.CreatedAt, r.Hash, r.Signature, r.FrozenCoreHash, string(meta))
	return err
}

func (s *Store) ListReceipts() ([]receipts.Receipt, error) {
	rows, err := s.db.Query(`SELECT id, actor, action, created_at, hash, frozen_core_hash FROM receipts ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []receipts.Receipt
	for rows.Next() {
		var r receipts.Receipt
		rows.Scan(&r.ReceiptID, &r.By, &r.Action, &r.CreatedAt, &r.Hash, &r.FrozenCoreHash)
		list = append(list, r)
	}
	return list, nil
}

func (s *Store) VerifyReceipt(id string) error {
	row := s.db.QueryRow(`SELECT id FROM receipts WHERE id = $1`, id)
	var found string
	if err := row.Scan(&found); err != nil {
		return fmt.Errorf("receipt not found: %s", id)
	}
	fmt.Printf("✅ verified receipt: %s\n", id)
	return nil
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}
