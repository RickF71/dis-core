package revoke

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"dis-core/internal/db"
)

type RevocationEntry struct {
	RevocationID   string `json:"revocation_id"`
	RevokedRef     string `json:"revoked_ref"`
	RevokedType    string `json:"revoked_type"`
	Reason         string `json:"reason"`
	RevokedBy      string `json:"revoked_by"`
	RevocationTime string `json:"revocation_time"`
	ValidUntil     string `json:"valid_until,omitempty"`
	Signature      string `json:"signature,omitempty"`
}

// Handle returns an http.HandlerFunc bound to the given DB.
func Handle(store *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var entry RevocationEntry
			if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			entry.RevocationID = "rev-" + db.NowRFC3339Nano()
			entry.RevocationTime = time.Now().UTC().Format(time.RFC3339)
			if entry.ValidUntil == "" {
				entry.ValidUntil = time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
			}

			_, err := store.Exec(`
				INSERT INTO revocations 
				(revocation_id, revoked_ref, revoked_type, reason, revoked_by, revocation_time, valid_until, signature)
				VALUES ($1,$2,$3,$4,$5,$6::timestamptz,$7::timestamptz,$8);
			`, entry.RevocationID, entry.RevokedRef, entry.RevokedType, entry.Reason,
				entry.RevokedBy, entry.RevocationTime, entry.ValidUntil, entry.Signature)
			if err != nil {
				http.Error(w, "Failed to insert revocation: "+err.Error(), http.StatusInternalServerError)
				return
			}

			content := fmt.Sprintf("Revocation: %s of %s by %s (reason: %s)",
				entry.RevokedType, entry.RevokedRef, entry.RevokedBy, entry.Reason)
			_, err = store.Exec(`
				INSERT INTO receipts (receipt_id, schema_ref, content, timestamp)
				VALUES ($1, $2, $3, NOW());
			`, "rcpt-"+entry.RevocationID, "bridge-receipt-template.v0", content)
			if err != nil {
				log.Printf("⚠️ Failed to emit receipt: %v", err)
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "revoked", "entry": entry})

		case http.MethodGet:
			rows, err := store.Query(`
				SELECT revocation_id, revoked_ref, revoked_type, reason, revoked_by,
				       revocation_time, valid_until, signature
				FROM revocations
				ORDER BY revocation_time DESC
				LIMIT 25;
			`)
			if err != nil {
				http.Error(w, "Failed to query revocations: "+err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var revs []RevocationEntry
			for rows.Next() {
				var r RevocationEntry
				if err := rows.Scan(&r.RevocationID, &r.RevokedRef, &r.RevokedType, &r.Reason,
					&r.RevokedBy, &r.RevocationTime, &r.ValidUntil, &r.Signature); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				revs = append(revs, r)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(revs)

		default:
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		}
	}
}
