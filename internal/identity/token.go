package identity

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// GenerateToken returns a new random credential token (32 bytes, base64url).
// It is the one place tokens are minted, for both the admin token and the
// per-user deploy keys.
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// tokenHash returns the hex SHA-256 of a token — the form used for
// constant-time comparison. The token itself is also stored (see deployKey)
// so the operator can copy it from the console.
func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func newKeyID() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate key id: %w", err)
	}
	return "dk_" + hex.EncodeToString(b), nil
}
