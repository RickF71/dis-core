package receipts

import (
	"crypto/ed25519"
	"dis-core/internal/util/crypto"
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

// DecodePublicKey converts a base64 string to an ed25519.PublicKey.
func DecodePublicKey(b64 string) (ed25519.PublicKey, error) {
	bytes, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, err
	}
	return ed25519.PublicKey(bytes), nil
}

// VerifySignature verifies an ed25519 signature against a given message.
func VerifySignature(message []byte, pubKey ed25519.PublicKey, sigB64 string) (bool, error) {
	sig, err := base64.StdEncoding.DecodeString(sigB64)
	if err != nil {
		return false, err
	}
	return ed25519.Verify(pubKey, message, sig), nil
}
