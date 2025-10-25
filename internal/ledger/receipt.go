package ledger

import (
	"crypto/rand"
	"crypto/sha256"
	"dis-core/internal/util/crypto"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Receipt defines the DIS ci.call.v1 structure
type Receipt struct {
	ReceiptID      string       `json:"receipt_id"`
	By             string       `json:"by"`
	Action         string       `json:"action"`
	CreatedAt      string       `json:"created_at"`
	Hash           string       `json:"hash"`
	Provenance     []Provenance `json:"provenance"`
	Signature      string       `json:"signature"`
	FrozenCoreHash string       `json:"frozen_core_hash"`
	Metadata       Metadata     `json:"metadata"`
}

type Provenance struct {
	Type           string   `json:"type"`
	Ref            string   `json:"ref"`
	Status         string   `json:"status"`
	RedactedFields []string `json:"redacted_fields,omitempty"`
}

type Metadata struct {
	IssuedFromConsole  string `json:"issued_from_console"`
	IssuerSeat         string `json:"issuer_seat"`
	VerifiedAt         string `json:"verified_at,omitempty"`
	VerificationMethod string `json:"verification_method,omitempty"`
	SignerPublicKeyB64 string `json:"signer_public_key_b64,omitempty"`
}

// NewReceipt creates a signed ci.call.v1 receipt for an action.
func NewReceipt(by, action, frozenCoreHash, consoleID, issuerSeat string) *Receipt {
	createdAt := time.Now().Format(time.RFC3339Nano)

	// Payload to hash/sign (stable ordering!)
	payload := fmt.Sprintf("%s|%s|%s|%s|%s|%s",
		by, action, createdAt, frozenCoreHash, consoleID, issuerSeat)

	// Hash payload
	hash := sha256.Sum256([]byte(payload))
	hashHex := hex.EncodeToString(hash[:])

	// Ensure domain keys & sign hashHex
	signer, _ := crypto.EnsureDomainKeys(by) // domain-scoped keys (e.g., "domain.terra")
	sigB64 := signer.Sign([]byte(hashHex))

	return &Receipt{
		ReceiptID:      generateReceiptID(),
		By:             by,
		Action:         action,
		CreatedAt:      createdAt,
		Hash:           hashHex,
		Signature:      sigB64,
		FrozenCoreHash: frozenCoreHash,
		Metadata: Metadata{
			IssuedFromConsole: consoleID,
			IssuerSeat:        issuerSeat,
			SignerPublicKeyB64: base64.StdEncoding.EncodeToString(
				signer.Pub), // assuming crypto.Signer has Pub []byte or ed25519.PublicKey
		},
	}
}

// generateReceiptID returns a random SHA-256-based identifier.
func generateReceiptID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	sum := sha256.Sum256(buf)
	return hex.EncodeToString(sum[:8]) // 16 hex chars, enough for uniqueness
}

// Save writes the receipt JSON into the receipts directory.
func (r *Receipt) Save() error {
	dir := "receipts"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}

	filename := filepath.Join(dir, fmt.Sprintf("%s.json", r.ReceiptID))
	return os.WriteFile(filename, data, 0644)
}
