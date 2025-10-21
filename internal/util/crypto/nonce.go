package crypto

import (
	"crypto/rand"
	"encoding/hex"
)

func RandomNonce(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
