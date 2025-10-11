package ledger

import (
	"crypto/rand"
	"fmt"
)

// GenerateUUID creates a random RFC4122-style identifier.
// This avoids heavy imports while keeping stable uniqueness.
func GenerateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	// Set version (4) and variant bits per RFC 4122
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
