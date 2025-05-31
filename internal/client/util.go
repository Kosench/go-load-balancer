package client

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateAPIKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
