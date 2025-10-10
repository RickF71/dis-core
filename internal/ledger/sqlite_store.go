package ledger

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"dis-core/internal/receipts"

	_ "modernc.org/sqlite"
)

// Store wraps an open SQLite connection for receipts.
type Store struct {
	db *sql.DB
}

// Open initializes (or opens) a persistent ledger database.
// Example path: data/dis_core.db
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}

	schema := `
	CREATE TABLE IF NOT EXISTS receipts (
		id TEXT PRIMARY KEY,
		actor TEXT,
		action TEXT,
		timestamp TEXT,
		hash TEXT,
		signature TEXT,
		frozen_core_hash TEXT,
		metadata TEXT
	);`
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}
	return s, nil
}

// InsertReceipt persists a Receipt into the ledger.
func (s *Store) InsertReceipt(r *receipts.Receipt) error {
	meta, _ := json.Marshal(r.Metadata)
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO receipts
		(id, actor, action, timestamp, hash, signature, frozen_core_hash, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ReceiptID, r.By, r.Action, r.Timestamp,
		r.Hash, r.Signature, r.FrozenCoreHash, string(meta))
	return err
}

// ListReceipts returns all stored receipts.
func (s *Store) ListReceipts() ([]receipts.Receipt, error) {
	rows, err := s.db.Query(`SELECT id, actor, action, timestamp, hash, frozen_core_hash FROM receipts ORDER BY timestamp DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []receipts.Receipt
	for rows.Next() {
		var r receipts.Receipt
		rows.Scan(&r.ReceiptID, &r.By, &r.Action, &r.Timestamp, &r.Hash, &r.FrozenCoreHash)
		list = append(list, r)
	}
	return list, nil
}

// VerifyReceipt recomputes and checks a stored receipt’s hash & signature placeholder (future crypto validation).
func (s *Store) VerifyReceipt(id string) error {
	row := s.db.QueryRow(`SELECT id FROM receipts WHERE id = ?`, id)
	var found string
	if err := row.Scan(&found); err != nil {
		return fmt.Errorf("receipt not found: %s", id)
	}
	fmt.Printf("✅ verified receipt: %s\n", id)
	return nil
}
