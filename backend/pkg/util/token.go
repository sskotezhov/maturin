package util

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateToken generates a cryptographically random hex string of 32 bytes.
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
