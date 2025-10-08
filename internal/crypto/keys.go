package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
)

type Signer struct {
	Priv ed25519.PrivateKey
	Pub  ed25519.PublicKey
}

// EnsureDomainKeys loads keys if present; otherwise generates & saves them.
// Keys are stored under versions/v0.6/keys/<domain>.priv / <domain>.pub (base64).
func EnsureDomainKeys(domain string) (*Signer, error) {
	dir := "versions/v0.6/keys"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	privPath := filepath.Join(dir, domain+".priv")
	pubPath := filepath.Join(dir, domain+".pub")

	// Try load
	privBytes, perr := os.ReadFile(privPath)
	pubBytes, uerr := os.ReadFile(pubPath)
	if perr == nil && uerr == nil {
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

	// Generate
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	// Save base64
	if err := os.WriteFile(privPath, []byte(base64.StdEncoding.EncodeToString(priv)), 0600); err != nil {
		return nil, err
	}
	if err := os.WriteFile(pubPath, []byte(base64.StdEncoding.EncodeToString(pub)), 0644); err != nil {
		return nil, err
	}
	return &Signer{Priv: priv, Pub: pub}, nil
}

// Sign returns base64(sig) for the given message bytes.
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
