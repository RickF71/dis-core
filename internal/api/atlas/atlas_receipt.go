package atlas

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

// LocationReceipt represents a verifiable spatial proof within DIS Atlas.
type LocationReceipt struct {
	ID           string    `json:"id"`          // rcpt-loc-{hash}
	EntityID     string    `json:"entity_id"`   // The DIS entity (person, domain, or asset)
	LocationID   string    `json:"location_id"` // Atlas location reference
	IssuedBy     string    `json:"issued_by"`   // Domain that created the receipt
	Method       string    `json:"method"`      // e.g., "manual", "geoip", "device", "domain-cert"
	Confidence   float64   `json:"confidence"`  // 0.0–1.0 confidence level
	Signature    string    `json:"signature"`   // Placeholder for cryptographic proof
	PolicyRef    string    `json:"policy_ref"`  // Link to freeze/regulation policy if applicable
	IssuedAt     time.Time `json:"issued_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	Checksum     string    `json:"checksum"`     // Derived hash of the serialized record
	Version      string    `json:"version"`      // atlas.rcpt.v0.1
	Verification string    `json:"verification"` // Verification result string
}

// GenerateChecksum computes the deterministic SHA256 checksum of the receipt.
func (r *LocationReceipt) GenerateChecksum() error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	hash := sha256.Sum256(data)
	r.Checksum = hex.EncodeToString(hash[:])
	return nil
}

// NewLocationReceipt creates a new receipt and initializes deterministic fields.
func NewLocationReceipt(entityID, locationID, issuedBy, method string, confidence float64, expires time.Time) (*LocationReceipt, error) {
	r := &LocationReceipt{
		EntityID:   entityID,
		LocationID: locationID,
		IssuedBy:   issuedBy,
		Method:     method,
		Confidence: confidence,
		IssuedAt:   time.Now().UTC(),
		ExpiresAt:  expires,
		Version:    "atlas.rcpt.v0.1",
	}
	if err := r.GenerateChecksum(); err != nil {
		return nil, err
	}
	r.ID = "rcpt-loc-" + r.Checksum[:12]
	return r, nil
}

// VerifyReceipt recomputes checksum and validates expiry.
func (r *LocationReceipt) VerifyReceipt() bool {
	tmp := *r
	oldChecksum := r.Checksum
	tmp.Checksum = ""
	_ = tmp.GenerateChecksum()
	if oldChecksum != tmp.Checksum {
		r.Verification = "checksum_mismatch"
		return false
	}
	if time.Now().UTC().After(r.ExpiresAt) {
		r.Verification = "expired"
		return false
	}
	r.Verification = "valid"
	return true
}
