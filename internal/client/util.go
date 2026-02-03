package client

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateAPIKey generates a cryptographically secure random API key.
// Returns a 32-character hexadecimal string.
func GenerateAPIKey() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}
