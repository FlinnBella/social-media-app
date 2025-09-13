package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateUniqueID creates a completely random unique ID (no input data needed)
func GenerateUniqueID() string {
	randomBytes := make([]byte, 12)
	rand.Read(randomBytes)
	return hex.EncodeToString(randomBytes)
}
