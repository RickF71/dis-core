package dis

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"dis-core/internal/db"
)

type DISAuthHandshake struct {
	HandshakeID  string `json:"handshake_id"`
	Initiator    string `json:"initiator"`
	Responder    string `json:"responder"`
	Scope        string `json:"scope"`
	ConsentProof string `json:"consent_proof"`
	ResultToken  string `json:"result_token"`
	ExpiresAt    string `json:"expires_at"`
}

// Handle returns an http.HandlerFunc bound to a specific DB.
func Handle(store *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var h DISAuthHandshake
			if err := json.NewDecoder(r.Body).Decode(&h); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			h.HandshakeID = "hs-" + db.NowRFC3339Nano()
			h.ResultToken = "tok-" + h.HandshakeID
			h.ExpiresAt = time.Now().UTC().Add(1 * time.Hour).Format(time.RFC3339)

			_, err := store.Exec(`
				INSERT INTO handshakes
					(handshake_id, initiator, responder, scope, consent_proof, result_token, expires_at)
				VALUES
					($1, $2, $3, $4, $5, $6, $7::timestamptz);
			`, h.HandshakeID, h.Initiator, h.Responder, h.Scope, h.ConsentProof, h.ResultToken, h.ExpiresAt)
			if err != nil {
				http.Error(w, "Failed to insert handshake: "+err.Error(), http.StatusInternalServerError)
				return
			}

			content := fmt.Sprintf("Handshake: %s → %s (scope: %s)", h.Initiator, h.Responder, h.Scope)
			_, err = store.Exec(`
				INSERT INTO receipts (receipt_id, schema_ref, content, timestamp)
				VALUES ($1, $2, $3, NOW());
			`, "rcpt-"+h.HandshakeID, "bridge-receipt-template.v0", content)
			if err != nil {
				log.Printf("⚠️ Failed to emit consent receipt: %v", err)
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "handshake": h})

		case http.MethodGet:
			rows, err := store.Query(`
				SELECT handshake_id, initiator, responder, scope,
				       consent_proof, result_token, expires_at
				FROM handshakes
				ORDER BY id DESC
				LIMIT 25;
			`)
			if err != nil {
				http.Error(w, "Failed to query handshakes: "+err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var list []DISAuthHandshake
			for rows.Next() {
				var h DISAuthHandshake
				if err := rows.Scan(&h.HandshakeID, &h.Initiator, &h.Responder, &h.Scope,
					&h.ConsentProof, &h.ResultToken, &h.ExpiresAt); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				list = append(list, h)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(list)

		default:
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		}
	}
}
