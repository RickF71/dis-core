package db

import (
	"context"
	"database/sql"
	"time"
)

type Receipt struct {
	ID     int64  `json:"id"`
	Ref    string `json:"ref"`
	By     string `json:"by"`
	Scope  string `json:"scope"`
	Result string `json:"result"`
	Sig    string `json:"sig"`
	Nonce  string `json:"nonce"`
	TS     string `json:"timestamp"` // RFC3339
}

// EnsureReceiptsSchema creates the receipts table if missing.
func EnsureReceiptsSchema(db *sql.DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS receipts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  ref TEXT,
  by TEXT NOT NULL,
  scope TEXT NOT NULL,
  result TEXT NOT NULL,
  sig TEXT NOT NULL,
  nonce TEXT NOT NULL,
  ts TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_receipts_ts ON receipts(ts);
CREATE INDEX IF NOT EXISTS idx_receipts_scope ON receipts(scope);
`
	_, err := db.Exec(schema)
	return err
}

func InsertReceipt(db *sql.DB, r *Receipt) (int64, error) {
	if r.TS == "" {
		r.TS = time.Now().UTC().Format(time.RFC3339Nano)
	}
	q := `INSERT INTO receipts(ref,by,scope,result,sig,nonce,ts) VALUES(?,?,?,?,?,?,?)`
	res, err := db.Exec(q, r.Ref, r.By, r.Scope, r.Result, r.Sig, r.Nonce, r.TS)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

type ListOpts struct {
	Limit  int
	Offset int
	Scope  string // optional filter
}

func ListReceipts(db *sql.DB, opts ListOpts) ([]Receipt, error) {
	if opts.Limit <= 0 || opts.Limit > 500 {
		opts.Limit = 100
	}
	if opts.Offset < 0 {
		opts.Offset = 0
	}

	var (
		rows *sql.Rows
		err  error
	)
	if opts.Scope != "" {
		rows, err = db.QueryContext(context.Background(),
			`SELECT id,ref,by,scope,result,sig,nonce,ts
			 FROM receipts
			 WHERE scope = ?
			 ORDER BY id DESC
			 LIMIT ? OFFSET ?`, opts.Scope, opts.Limit, opts.Offset)
	} else {
		rows, err = db.QueryContext(context.Background(),
			`SELECT id,ref,by,scope,result,sig,nonce,ts
			 FROM receipts
			 ORDER BY id DESC
			 LIMIT ? OFFSET ?`, opts.Limit, opts.Offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Receipt
	for rows.Next() {
		var r Receipt
		if err := rows.Scan(&r.ID, &r.Ref, &r.By, &r.Scope, &r.Result, &r.Sig, &r.Nonce, &r.TS); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
