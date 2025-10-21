package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
)

// Directory for persistent domain key storage.
const keyDir = "versions/v0.6/keys"

// Signer wraps an Ed25519 keypair for signing and verification.
type Signer struct {
	Priv ed25519.PrivateKey
	Pub  ed25519.PublicKey
}

// EnsureDomainKeys loads keys if present; otherwise generates and saves them.
// Keys are stored in base64 at: versions/v0.6/keys/<domain>.priv / <domain>.pub
func EnsureDomainKeys(domain string) (*Signer, error) {
	if err := os.MkdirAll(keyDir, 0755); err != nil {
		return nil, err
	}

	privPath := filepath.Join(keyDir, domain+".priv")
	pubPath := filepath.Join(keyDir, domain+".pub")

	// Try loading existing keys
	if s, err := LoadDomainKeys(privPath, pubPath); err == nil {
		return s, nil
	}

	// Generate new pair
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	// Save
	if err := SaveDomainKeys(privPath, pubPath, priv, pub); err != nil {
		return nil, err
	}

	return &Signer{Priv: priv, Pub: pub}, nil
}

// LoadDomainKeys reads base64 key files and returns a valid Signer.
func LoadDomainKeys(privPath, pubPath string) (*Signer, error) {
	privBytes, err1 := os.ReadFile(privPath)
	pubBytes, err2 := os.ReadFile(pubPath)
	if err1 != nil || err2 != nil {
		return nil, errors.New("missing key files")
	}

	priv, err := base64.StdEncoding.DecodeString(string(privBytes))
	if err != nil {
		return nil, err
	}
	pub, err := base64.StdEncoding.DecodeString(string(pubBytes))
	if err != nil {
		return nil, err
	}
	if len(priv) != ed25519.PrivateKeySize || len(pub) != ed25519.PublicKeySize {
		return nil, errors.New("invalid key sizes on disk")
	}

	return &Signer{Priv: ed25519.PrivateKey(priv), Pub: ed25519.PublicKey(pub)}, nil
}

// SaveDomainKeys writes keys to disk in base64 form.
func SaveDomainKeys(privPath, pubPath string, priv ed25519.PrivateKey, pub ed25519.PublicKey) error {
	if err := os.WriteFile(privPath, []byte(base64.StdEncoding.EncodeToString(priv)), 0600); err != nil {
		return err
	}
	if err := os.WriteFile(pubPath, []byte(base64.StdEncoding.EncodeToString(pub)), 0644); err != nil {
		return err
	}
	return nil
}

// Sign returns a base64 signature for the given message bytes.
func (s *Signer) Sign(msg []byte) string {
	sig := ed25519.Sign(s.Priv, msg)
	return base64.StdEncoding.EncodeToString(sig)
}

// Verify returns true if base64(sig) verifies against msg.
func (s *Signer) Verify(msg []byte, b64sig string) bool {
	sig, err := base64.StdEncoding.DecodeString(b64sig)
	if err != nil {
		return false
	}
	return ed25519.Verify(s.Pub, msg, sig)
}
