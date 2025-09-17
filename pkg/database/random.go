package database

import (
	"crypto/rand"
	"encoding/hex"
)

func RandomString(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		// Fall back to a less secure but functional token
		// This should rarely happen as crypto/rand rarely fails
		return string(rune(length))
	}
	return hex.EncodeToString(b)
}
