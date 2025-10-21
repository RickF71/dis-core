package crypto

import (
	"crypto/sha256"
	"encoding/hex"
)

// Sign creates a deterministic hex digest over the provided fields.
func Sign(fields ...string) string {
	h := sha256.New()
	for _, f := range fields {
		h.Write([]byte(f))
	}
	return hex.EncodeToString(h.Sum(nil))
}
