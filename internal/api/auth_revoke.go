package api

import (
	"dis-core/internal/db"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// RevocationEntry represents a revoked handshake, credential, or session.
type RevocationEntry struct {
	RevocationID   string `json:"revocation_id"`
	RevokedRef     string `json:"revoked_ref"`
	RevokedType    string `json:"revoked_type"` // credential | handshake | session
	Reason         string `json:"reason"`
	RevokedBy      string `json:"revoked_by"`
	RevocationTime string `json:"revocation_time"`
	ValidUntil     string `json:"valid_until,omitempty"`
	Signature      string `json:"signature,omitempty"`
}

func (s *Server) HandleAuthRevoke(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var entry RevocationEntry
		if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		entry.RevocationID = "rev-" + db.NowRFC3339Nano()
		entry.RevocationTime = db.NowRFC3339Nano()
		if entry.ValidUntil == "" {
			entry.ValidUntil = time.Now().Add(24 * time.Hour).Format(time.RFC3339)
		}

		// 1️⃣ Persist to revocations table
		_, err := s.store.Exec(`
			INSERT INTO revocations 
			(revocation_id, revoked_ref, revoked_type, reason, revoked_by, revocation_time, valid_until, signature)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			entry.RevocationID, entry.RevokedRef, entry.RevokedType, entry.Reason,
			entry.RevokedBy, entry.RevocationTime, entry.ValidUntil, entry.Signature,
		)
		if err != nil {
			http.Error(w, "Failed to insert revocation: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 2️⃣ Emit a receipt
		receiptContent := fmt.Sprintf(
			"Revocation: %s of %s by %s (reason: %s)",
			entry.RevokedType, entry.RevokedRef, entry.RevokedBy, entry.Reason,
		)
		_, err = s.store.Exec(`
			INSERT INTO receipts (receipt_id, schema_ref, content, timestamp)
			VALUES (?, ?, ?, datetime('now'))`,
			"rcpt-"+entry.RevocationID, "bridge-receipt-template.v0", receiptContent,
		)
		if err != nil {
			log.Printf("⚠️ Failed to emit receipt: %v", err)
		}

		// 3️⃣ Return JSON response
		resp := map[string]any{
			"status": "revoked",
			"entry":  entry,
		}
		json.NewEncoder(w).Encode(resp)

	case http.MethodGet:
		// Optional: list recent revocations
		rows, err := s.store.Query(`
			SELECT revocation_id, revoked_ref, revoked_type, reason, revoked_by, revocation_time, valid_until 
			FROM revocations ORDER BY revocation_time DESC LIMIT 25`)
		if err != nil {
			http.Error(w, "Failed to query revocations: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var revs []RevocationEntry
		for rows.Next() {
			var r RevocationEntry
			rows.Scan(&r.RevocationID, &r.RevokedRef, &r.RevokedType, &r.Reason,
				&r.RevokedBy, &r.RevocationTime, &r.ValidUntil)
			revs = append(revs, r)
		}
		json.NewEncoder(w).Encode(revs)

	default:
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
	}
}
