package api

import (
	"dis-core/internal/db"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// DISAuthHandshake represents a trust handshake between two domains or agents.
type DISAuthHandshake struct {
	HandshakeID  string `json:"handshake_id"`
	Initiator    string `json:"initiator"`
	Responder    string `json:"responder"`
	Scope        string `json:"scope"`
	ConsentProof string `json:"consent_proof"`
	ResultToken  string `json:"result_token"`
	ExpiresAt    string `json:"expires_at"`
}

// HandleDISAuthHandshake creates or lists handshakes.
func (s *Server) HandleDISAuthHandshake(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodPost:
		var h DISAuthHandshake
		if err := json.NewDecoder(r.Body).Decode(&h); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		h.HandshakeID = "hs-" + db.NowRFC3339Nano()
		h.ResultToken = "tok-" + h.HandshakeID
		h.ExpiresAt = time.Now().Add(1 * time.Hour).Format(time.RFC3339)

		// 1️⃣ Persist to handshakes table
		_, err := s.store.Exec(`
			INSERT INTO handshakes (handshake_id, initiator, responder, scope, consent_proof, result_token, expires_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			h.HandshakeID, h.Initiator, h.Responder, h.Scope, h.ConsentProof, h.ResultToken, h.ExpiresAt,
		)
		if err != nil {
			http.Error(w, "Failed to insert handshake: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 2️⃣ Emit a consent receipt
		receiptContent := fmt.Sprintf(
			"Handshake: %s → %s (scope: %s)",
			h.Initiator, h.Responder, h.Scope,
		)
		_, err = s.store.Exec(`
			INSERT INTO receipts (receipt_id, schema_ref, content, timestamp)
			VALUES (?, ?, ?, datetime('now'))`,
			"rcpt-"+h.HandshakeID, "bridge-receipt-template.v0", receiptContent,
		)
		if err != nil {
			log.Printf("⚠️ Failed to emit consent receipt: %v", err)
		}

		// 3️⃣ Respond to client
		resp := map[string]any{
			"status":    "ok",
			"handshake": h,
		}
		json.NewEncoder(w).Encode(resp)

	case http.MethodGet:
		// 4️⃣ List recent handshakes
		rows, err := s.store.Query(`
			SELECT handshake_id, initiator, responder, scope, consent_proof, result_token, expires_at
			FROM handshakes ORDER BY id DESC LIMIT 25`)
		if err != nil {
			http.Error(w, "Failed to query handshakes: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var list []DISAuthHandshake
		for rows.Next() {
			var h DISAuthHandshake
			rows.Scan(&h.HandshakeID, &h.Initiator, &h.Responder, &h.Scope, &h.ConsentProof, &h.ResultToken, &h.ExpiresAt)
			list = append(list, h)
		}
		json.NewEncoder(w).Encode(list)

	default:
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
	}
}
