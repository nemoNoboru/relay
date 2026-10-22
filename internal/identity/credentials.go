package identity

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// minMemberPassword is the length floor for a member password. It matches
// the admin token floor so both credentials read the same rule.
const minMemberPassword = 8

// argon2id parameters for member passwords. They are stored on each key so a
// hash keeps verifying if the defaults change later.
const (
	passwordTime    = 3
	passwordMemory  = 64 * 1024 // KiB
	passwordThreads = 4
	passwordKeyLen  = 32
	passwordSaltLen = 16
)

// AuthenticatePassword returns the invitation token of the live member whose
// username and password match, or ok=false when the username is unknown, the
// password is wrong, or the member is revoked. The returned token is the
// value stored in the member cookie on login (decision D4).
func (ks *KeyStore) AuthenticatePassword(username, password string) (string, bool) {
	username = strings.TrimSpace(username)
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	for _, k := range ks.file.Keys {
		if k.Revoked || k.Username != username || k.PasswordHash == "" {
			continue
		}
		if passwordMatches(k, password) {
			return k.Token, true
		}
		return "", false
	}
	return "", false
}

// SetCredentials sets a member's username and password. The plaintext
// password is hashed with argon2id and never stored. found=false means no
// such key. It returns ErrBadUser for a username outside the label grammar,
// ErrBadPassword for a short password, and ErrUsernameTaken when another live
// member already holds the username.
func (ks *KeyStore) SetCredentials(id, username, password string) (KeyInfo, bool, error) {
	username = strings.TrimSpace(username)
	if !validUser(username) {
		return KeyInfo{}, false, ErrBadUser
	}
	if len(password) < minMemberPassword {
		return KeyInfo{}, false, ErrBadPassword
	}
	ks.mu.Lock()
	defer ks.mu.Unlock()
	i := ks.indexOf(id)
	if i == -1 {
		return KeyInfo{}, false, nil
	}
	for j := range ks.file.Keys {
		if j == i || ks.file.Keys[j].Revoked {
			continue
		}
		if ks.file.Keys[j].Username == username {
			return KeyInfo{}, false, ErrUsernameTaken
		}
	}
	salt, hash, err := hashPassword(password)
	if err != nil {
		return KeyInfo{}, false, err
	}
	before := ks.file.Keys[i]
	ks.file.Keys[i].Username = username
	ks.file.Keys[i].PasswordSalt = salt
	ks.file.Keys[i].PasswordHash = hash
	ks.file.Keys[i].PasswordTime = passwordTime
	ks.file.Keys[i].PasswordMemory = passwordMemory
	ks.file.Keys[i].PasswordThreads = passwordThreads
	if err := ks.persist(); err != nil {
		ks.file.Keys[i] = before
		return KeyInfo{}, false, err
	}
	return keyInfo(ks.file.Keys[i]), true, nil
}

// UsernameTaken reports whether a live member already holds username.
func (ks *KeyStore) UsernameTaken(username string) bool {
	username = strings.TrimSpace(username)
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	for _, k := range ks.file.Keys {
		if !k.Revoked && k.Username == username {
			return true
		}
	}
	return false
}

// ClearCredentials removes a member's username and password without touching
// the invitation, the invite tree, or the token. found=false means no such
// key.
func (ks *KeyStore) ClearCredentials(id string) (KeyInfo, bool, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	i := ks.indexOf(id)
	if i == -1 {
		return KeyInfo{}, false, nil
	}
	before := ks.file.Keys[i]
	ks.file.Keys[i].Username = ""
	ks.file.Keys[i].PasswordSalt = ""
	ks.file.Keys[i].PasswordHash = ""
	ks.file.Keys[i].PasswordTime = 0
	ks.file.Keys[i].PasswordMemory = 0
	ks.file.Keys[i].PasswordThreads = 0
	if err := ks.persist(); err != nil {
		ks.file.Keys[i] = before
		return KeyInfo{}, false, err
	}
	return keyInfo(ks.file.Keys[i]), true, nil
}

// UsernameByID returns the username held by key id, empty when none. It is
// the session view's credential hint; revocation does not hide it.
func (ks *KeyStore) UsernameByID(id string) string {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	i := ks.indexOf(id)
	if i == -1 {
		return ""
	}
	return ks.file.Keys[i].Username
}

// passwordMatches verifies password against k's stored argon2id hash in
// constant time. Keys written before credentials existed have no hash and
// never match.
func passwordMatches(k deployKey, password string) bool {
	salt, err := hex.DecodeString(k.PasswordSalt)
	if err != nil || len(salt) == 0 {
		return false
	}
	want, err := hex.DecodeString(k.PasswordHash)
	if err != nil || len(want) == 0 {
		return false
	}
	t, m, p := k.PasswordTime, k.PasswordMemory, k.PasswordThreads
	if t == 0 {
		t = passwordTime
	}
	if m == 0 {
		m = passwordMemory
	}
	if p == 0 {
		p = passwordThreads
	}
	got := argon2.IDKey([]byte(password), salt, t, m, p, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1
}

// hashPassword derives a random salt and the argon2id hash for password. The
// salt and hash are hex-encoded for the JSON store.
func hashPassword(password string) (salt, hash string, err error) {
	b := make([]byte, passwordSaltLen)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generate password salt: %w", err)
	}
	sum := argon2.IDKey([]byte(password), b, passwordTime, passwordMemory, passwordThreads, passwordKeyLen)
	return hex.EncodeToString(b), hex.EncodeToString(sum), nil
}
