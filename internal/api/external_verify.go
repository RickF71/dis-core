package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"dis-core/internal/ledger"
	"dis-core/internal/network"
	"dis-core/internal/receipts"
)

type Metadata struct {
	VerifiedAtPeer string `json:"verified_at_peer,omitempty"`
}

// ExternalVerifyHandler validates incoming verification receipts from peers.
func ExternalVerifyHandler(cfg *network.NetworkConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var incoming receipts.Receipt
		if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		peerName := incoming.By
		peer := findPeer(cfg, peerName)
		if peer == nil {
			http.Error(w, fmt.Sprintf("unknown peer: %s", peerName), http.StatusForbidden)
			return
		}

		pubKey, err := receipts.DecodePublicKey(peer.PublicKeyB64)
		if err != nil {
			http.Error(w, "invalid peer public key", http.StatusInternalServerError)
			return
		}

		data, _ := json.Marshal(incoming)
		ok, err := receipts.VerifySignature(data, pubKey, incoming.Signature)
		if err != nil || !ok {
			trust.Add(ledger.TrustEntry{
				Peer:       peer.Name,
				Action:     "received",
				Status:     "invalid",
				ReceiptID:  incoming.ReceiptID,
				CoreHash:   incoming.FrozenCoreHash,
				VerifiedAt: time.Now().UTC(),
				Notes:      "signature verification failed",
			})
			resp := map[string]any{"status": "invalid", "peer": peerName, "reason": "signature mismatch"}
			writeJSON(w, resp, http.StatusOK)
			return
		}

		// Save verified peer receipt
		savePath := filepath.Join("versions/v0.7/receipts/peers", fmt.Sprintf("%s.json", peerName))
		_ = os.MkdirAll(filepath.Dir(savePath), 0755)
		incoming.Metadata.VerifiedAtPeer = time.Now().UTC().Format(time.RFC3339)
		out, _ := json.MarshalIndent(incoming, "", "  ")
		if err := os.WriteFile(savePath, out, 0644); err != nil {
			http.Error(w, "failed to store verified peer receipt", http.StatusInternalServerError)
			return
		}

		peer.LastSeen = time.Now().UTC().Format(time.RFC3339)
		log.Printf("🤝 Verified receipt from peer %s and saved to %s", peerName, savePath)

		// 🔹 Add this block for Step 4 – record in trust ledger
		trust, err := ledger.LoadTrustLedger("versions/v0.7/ledger/trust.json")
		if err == nil {
			trust.Add(ledger.TrustEntry{
				Peer:       peerName,
				Action:     "received",
				Status:     "ok",
				ReceiptID:  incoming.ReceiptID,
				CoreHash:   incoming.FrozenCoreHash,
				VerifiedAt: time.Now().UTC(),
			})
		}

		resp := map[string]any{
			"status": "valid",
			"peer":   peerName,
			"saved":  savePath,
		}
		writeJSON(w, resp, http.StatusOK)
	}
}

// helper: find a peer in config
func findPeer(cfg *network.NetworkConfig, name string) *network.Peer {
	for _, p := range cfg.Peers {
		if p.Name == name {
			return &p
		}
	}
	return nil
}
