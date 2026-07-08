package util

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashToken returns the lowercase hex SHA-256 of the input (64 chars).
// Used for refresh + password-reset tokens — we never store the raw value.
func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
