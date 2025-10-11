package receipts

import (
	"crypto/rand"
	"crypto/sha256"
	"dis-core/internal/crypto"
	"dis-core/internal/db"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Receipt defines the DIS ci.call.v1 structure
type Receipt struct {
	ReceiptID      string       `json:"receipt_id"`
	By             string       `json:"by"`
	Action         string       `json:"action"`
	Timestamp      string       `json:"timestamp"`
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
	timestamp := db.NowRFC3339Nano()

	// Payload to hash/sign (stable ordering!)
	payload := fmt.Sprintf("%s|%s|%s|%s|%s|%s",
		by, action, timestamp, frozenCoreHash, consoleID, issuerSeat)

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
		Timestamp:      timestamp,
		Hash:           hashHex,
		Signature:      sigB64,
		FrozenCoreHash: frozenCoreHash,
		Provenance: []Provenance{
			{Type: "SAT", Ref: "sat-demo-001", Status: "valid"},
			{Type: "domain", Ref: by, Status: "valid"},
			{Type: "policy", Ref: "policy.freeze.rego", Status: "valid"},
		},
		Metadata: Metadata{
			IssuedFromConsole:  consoleID,
			IssuerSeat:         issuerSeat,
			VerifiedAt:         timestamp,
			VerificationMethod: "SAT-check",
			SignerPublicKeyB64: base64.StdEncoding.EncodeToString(signer.Pub),
		},
	}
}

func generateReceiptID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("r-%s", hex.EncodeToString(b))
}

// ToJSON returns the receipt as formatted JSON for saving or transmission.
func (r *Receipt) ToJSON() (string, error) {
	bytes, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Save writes the receipt JSON to a specified folder.
// It names the file automatically using the receipt_id.
func (r *Receipt) Save(dir string) error {
	jsonOut, err := r.ToJSON()
	if err != nil {
		return err
	}

	// Ensure directory exists
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	// Construct full path (e.g., versions/v0.6/receipts/r-xxxx.json)
	filename := filepath.Join(dir, fmt.Sprintf("%s.json", r.ReceiptID))
	return os.WriteFile(filename, []byte(jsonOut), 0644)
}
