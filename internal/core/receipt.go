package core

import (
	"database/sql"
	"fmt"
)

func ListReceipts(db *sql.DB) {
	rows, err := db.Query(`SELECT id, identity_id, action, by_domain, scope, nonce, created_at, policy_checksum, signature, created_at
						   FROM receipts ORDER BY created_at DESC`)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n📜 Receipts:")
	for rows.Next() {
		var id int
		var identity, action, by, scope, nonce, ts, polsum, sig, created string
		rows.Scan(&id, &identity, &action, &by, &scope, &nonce, &ts, &polsum, &sig, &created)
		if len(sig) > 16 {
			sig = sig[:16] + "..."
		}
		if len(polsum) > 16 {
			polsum = polsum[:16] + "..."
		}
		fmt.Printf("[%d] %s | %s | %s | %s | nonce=%s | ts=%s | policy=%s | sig=%s | %s\n",
			id, identity, action, by, scope, nonce, ts, polsum, sig, created)
	}
}
