// Package session provides session token generation.
package session

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateToken creates cryptographically secure random token.
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
