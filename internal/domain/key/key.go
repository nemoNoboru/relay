// Package key holds Relay's ownership key model: the Ed25519 keypair
// that is the deed to a site address. It is pure domain logic: no I/O,
// no network, no storage — only the rules the key must obey.
package key

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidFile is returned when a key file does not parse or does not
// pass validation.
var ErrInvalidFile = errors.New("invalid key file")

const (
	// FileVersion is the key file format version this build understands.
	FileVersion = 1

	// fileMagic identifies a Relay key file.
	fileMagic = "relay key"

	// fingerprintLen is the number of hex characters in a fingerprint:
	// the first half of the SHA-256 of the public half.
	fingerprintLen = 16
)

// Key is an Ed25519 ownership keypair. The secret half never leaves the
// key holder; only the public half and its fingerprint travel the wire.
type Key struct {
	public ed25519.PublicKey
	secret ed25519.PrivateKey
}

// New generates a fresh ownership key from a cryptographically secure
// random source.
func New() (Key, error) {
	public, secret, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return Key{}, fmt.Errorf("generate ownership key: %w", err)
	}
	return Key{public: public, secret: secret}, nil
}

// fromSeed reconstructs the keypair from an ed25519 seed.
func fromSeed(seed []byte) Key {
	secret := ed25519.NewKeyFromSeed(seed)
	return Key{public: secret.Public().(ed25519.PublicKey), secret: secret}
}

// Public returns the public half of the key.
func (k Key) Public() ed25519.PublicKey {
	return k.public
}

// HasSecret reports whether k holds the secret half. A key loaded from
// public material only does not.
func (k Key) HasSecret() bool {
	return k.secret != nil
}

// Fingerprint is a short stable identifier for logging and display:
// the first 16 hex characters of the SHA-256 of the public half. A
// fingerprint is safe to log; key material is not.
func (k Key) Fingerprint() string {
	sum := sha256.Sum256(k.public)
	return hex.EncodeToString(sum[:])[:fingerprintLen]
}

// Sign signs msg with the secret half. The signature covers exactly the
// bytes of msg; the canonical-string rules for owner requests live in
// the layer that builds them.
func (k Key) Sign(msg []byte) []byte {
	return ed25519.Sign(k.secret, msg)
}

// Verify reports whether sig is a valid signature by public over msg.
func Verify(public ed25519.PublicKey, msg, sig []byte) bool {
	return ed25519.Verify(public, msg, sig)
}

// File renders the key in the Relay key file format (version
// FileVersion). The file holds the secret half; it is the downloadable
// deed to the address and must be protected like a password.
func (k Key) File() ([]byte, error) {
	if !k.HasSecret() {
		return nil, errors.New("key holds no secret half to export")
	}
	return []byte(fmt.Sprintf("%s v%d\nfingerprint %s\nsecret %s\n",
		fileMagic, FileVersion, k.Fingerprint(),
		base64.StdEncoding.EncodeToString(k.secret.Seed()))), nil
}

// Load parses a Relay key file and returns the key it holds. The file
// is self-validating: the fingerprint line is checked against the
// secret, so a corrupted or swapped file is rejected.
func Load(data []byte) (Key, error) {
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 3 {
		return Key{}, fmt.Errorf("%w: want %d lines, got %d", ErrInvalidFile, 3, len(lines))
	}
	header, fingerprintLine, secretLine := lines[0], lines[1], lines[2]
	if header != fmt.Sprintf("%s v%d", fileMagic, FileVersion) {
		return Key{}, fmt.Errorf("%w: unknown or unsupported header %q", ErrInvalidFile, header)
	}
	fingerprint, ok := strings.CutPrefix(fingerprintLine, "fingerprint ")
	if !ok {
		return Key{}, fmt.Errorf("%w: malformed fingerprint line", ErrInvalidFile)
	}
	if err := checkFingerprint(fingerprint); err != nil {
		return Key{}, fmt.Errorf("%w: %v", ErrInvalidFile, err)
	}
	secret, ok := strings.CutPrefix(secretLine, "secret ")
	if !ok {
		return Key{}, fmt.Errorf("%w: malformed secret line", ErrInvalidFile)
	}
	seed, err := base64.StdEncoding.DecodeString(secret)
	if err != nil || len(seed) != ed25519.SeedSize {
		return Key{}, fmt.Errorf("%w: secret is not a base64 %d-byte seed", ErrInvalidFile, ed25519.SeedSize)
	}
	k := fromSeed(seed)
	if k.Fingerprint() != fingerprint {
		return Key{}, fmt.Errorf("%w: fingerprint does not match secret", ErrInvalidFile)
	}
	return k, nil
}

// checkFingerprint validates the shape of a fingerprint: exactly
// fingerprintLen lowercase hex characters.
func checkFingerprint(fp string) error {
	if len(fp) != fingerprintLen {
		return fmt.Errorf("fingerprint must be %d hex characters, got %d", fingerprintLen, len(fp))
	}
	for i := 0; i < len(fp); i++ {
		c := fp[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return fmt.Errorf("fingerprint %q must be lowercase hex", fp)
		}
	}
	return nil
}
