package util

import (
	"crypto/sha256"
	"encoding/hex"
)

func ChecksumHex(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}
