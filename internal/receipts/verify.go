package receipts

import (
	"crypto/ed25519"
	"dis-core/internal/crypto"
	"encoding/base64"
	"encoding/json"
	"errors"
)

func VerifyReceiptJSON(jsonBytes []byte) (bool, error) {
	var r Receipt
	if err := json.Unmarshal(jsonBytes, &r); err != nil {
		return false, err
	}
	if r.Hash == "" || r.Signature == "" || r.By == "" {
		return false, errors.New("missing required fields for verification")
	}
	// Load domain keys (or you could reconstruct from Metadata.SignerPublicKeyB64)
	signer, err := crypto.EnsureDomainKeys(r.By)
	if err != nil {
		return false, err
	}
	ok := signer.Verify([]byte(r.Hash), r.Signature)
	return ok, nil
}

// VerifyWithEmbeddedPub uses the embedded pubkey if present (no disk access).
func VerifyWithEmbeddedPub(jsonBytes []byte) (bool, error) {
	var r Receipt
	if err := json.Unmarshal(jsonBytes, &r); err != nil {
		return false, err
	}
	if r.Hash == "" || r.Signature == "" || r.Metadata.SignerPublicKeyB64 == "" {
		return false, errors.New("insufficient data")
	}

	pub, err := base64.StdEncoding.DecodeString(r.Metadata.SignerPublicKeyB64)
	if err != nil {
		return false, err
	}

	// ✅ make an addressable value (or use &literal) and cast pub to the right type
	s := &crypto.Signer{Pub: ed25519.PublicKey(pub)}
	return s.Verify([]byte(r.Hash), r.Signature), nil
}
