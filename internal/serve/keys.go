package serve

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
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

// tokenHash returns the hex SHA-256 of a token — the only form ever stored
// on disk or held for comparison.
func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// ErrBadUser is returned for an invalid user label. User labels are display
// names, not path segments, so the grammar is wider than a node/site name.
var ErrBadUser = errors.New("invalid user label: must match [A-Za-z0-9._-]{1,63}")

// deployKey is one issued per-user deploy credential, as persisted. The
// plaintext token is never stored: only its SHA-256 hash.
type deployKey struct {
	ID      string    `json:"id"`
	User    string    `json:"user"`
	Hash    string    `json:"hash"`
	Created time.Time `json:"created"`
	Revoked bool      `json:"revoked,omitempty"`
}

// keyFile is the on-disk shape of the deploy-key store.
type keyFile struct {
	Keys  []deployKey       `json:"keys"`
	Sites map[string]string `json:"sites"` // site name -> user label
}

// PublicKey is the non-secret view of a deploy key, returned by the admin
// API. It never includes the hash.
type PublicKey struct {
	ID      string    `json:"id"`
	User    string    `json:"user"`
	Created time.Time `json:"created"`
	Revoked bool      `json:"revoked"`
}

// SiteOwner pairs a site with the user who last deployed it.
type SiteOwner struct {
	Site string `json:"site"`
	User string `json:"user"`
}

// Overview is the admin API's read view: all keys plus site ownership.
type Overview struct {
	Keys  []PublicKey `json:"keys"`
	Sites []SiteOwner `json:"sites"`
}

// KeyStore persists per-user deploy keys and site ownership under
// <dataDir>/deploy-keys.json. Mutations are serialized and written
// atomically; reads take a shared lock. Tokens exist in plaintext only in
// the Create/Rotate response.
type KeyStore struct {
	path string
	mu   sync.RWMutex
	file keyFile
}

// NewKeyStore loads the store from disk, creating an empty one if absent.
func NewKeyStore(dataDir string) (*KeyStore, error) {
	ks := &KeyStore{
		path: filepath.Join(dataDir, "deploy-keys.json"),
		file: keyFile{Sites: map[string]string{}},
	}
	if err := ks.load(); err != nil {
		return nil, err
	}
	return ks, nil
}

func (ks *KeyStore) load() error {
	b, err := os.ReadFile(ks.path)
	if errors.Is(err, os.ErrNotExist) {
		return ks.persist()
	}
	if err != nil {
		return fmt.Errorf("read deploy keys: %w", err)
	}
	var f keyFile
	if err := json.Unmarshal(b, &f); err != nil {
		return fmt.Errorf("parse deploy keys: %w", err)
	}
	if f.Sites == nil {
		f.Sites = map[string]string{}
	}
	ks.file = f
	return nil
}

// persist writes the store to a temp file and renames it into place, 0600.
// Callers must hold the write lock.
func (ks *KeyStore) persist() error {
	b, err := json.MarshalIndent(ks.file, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(ks.path), 0o755); err != nil {
		return err
	}
	tmp := ks.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, ks.path); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

// Authenticate returns the user a token belongs to, or ok=false if the token
// is unknown or its key is revoked. The comparison runs over every key so
// the timing does not reveal which key matched.
func (ks *KeyStore) Authenticate(token string) (user string, ok bool) {
	want := sha256.Sum256([]byte(token))
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	match := -1
	for i := range ks.file.Keys {
		sum, err := hex.DecodeString(ks.file.Keys[i].Hash)
		if err != nil || len(sum) != len(want) {
			continue
		}
		if subtle.ConstantTimeCompare(sum, want[:]) == 1 {
			match = i
		}
	}
	if match == -1 || ks.file.Keys[match].Revoked {
		return "", false
	}
	return ks.file.Keys[match].User, true
}

// Create mints a deploy key for user and returns the plaintext token (shown
// once) alongside its public view. The token itself is not stored.
func (ks *KeyStore) Create(user string) (string, PublicKey, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	if !validUser(user) {
		return "", PublicKey{}, ErrBadUser
	}
	token, err := GenerateToken()
	if err != nil {
		return "", PublicKey{}, err
	}
	id, err := newKeyID()
	if err != nil {
		return "", PublicKey{}, err
	}
	k := deployKey{ID: id, User: user, Hash: tokenHash(token), Created: time.Now().UTC()}
	ks.file.Keys = append(ks.file.Keys, k)
	if err := ks.persist(); err != nil {
		ks.file.Keys = ks.file.Keys[:len(ks.file.Keys)-1]
		return "", PublicKey{}, err
	}
	return token, publicKey(k), nil
}

// Revoke marks a key as revoked. It is idempotent: revoking an already
// revoked key succeeds. found=false means no such key.
func (ks *KeyStore) Revoke(id string) (PublicKey, bool, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	i := ks.indexOf(id)
	if i == -1 {
		return PublicKey{}, false, nil
	}
	if ks.file.Keys[i].Revoked {
		return publicKey(ks.file.Keys[i]), true, nil
	}
	ks.file.Keys[i].Revoked = true
	if err := ks.persist(); err != nil {
		ks.file.Keys[i].Revoked = false
		return PublicKey{}, false, err
	}
	return publicKey(ks.file.Keys[i]), true, nil
}

// Rotate revokes the old key and mints a fresh one for the same user,
// returning the new plaintext token. found=false means no such key.
func (ks *KeyStore) Rotate(id string) (string, PublicKey, bool, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	i := ks.indexOf(id)
	if i == -1 {
		return "", PublicKey{}, false, nil
	}
	token, err := GenerateToken()
	if err != nil {
		return "", PublicKey{}, false, err
	}
	newID, err := newKeyID()
	if err != nil {
		return "", PublicKey{}, false, err
	}
	user := ks.file.Keys[i].User
	ks.file.Keys[i].Revoked = true
	k := deployKey{ID: newID, User: user, Hash: tokenHash(token), Created: time.Now().UTC()}
	ks.file.Keys = append(ks.file.Keys, k)
	if err := ks.persist(); err != nil {
		ks.file.Keys[i].Revoked = false
		ks.file.Keys = ks.file.Keys[:len(ks.file.Keys)-1]
		return "", PublicKey{}, false, err
	}
	return token, publicKey(k), true, nil
}

// RecordSite attributes a site to the user who last deployed it. A no-op
// when the attribution is unchanged.
func (ks *KeyStore) RecordSite(site, user string) error {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	if ks.file.Sites[site] == user {
		return nil
	}
	ks.file.Sites[site] = user
	return ks.persist()
}

// Overview returns a sorted read view of the store.
func (ks *KeyStore) Overview() Overview {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	keys := make([]PublicKey, 0, len(ks.file.Keys))
	for _, k := range ks.file.Keys {
		keys = append(keys, publicKey(k))
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].Created.Before(keys[j].Created) })
	sites := make([]SiteOwner, 0, len(ks.file.Sites))
	for s, u := range ks.file.Sites {
		sites = append(sites, SiteOwner{Site: s, User: u})
	}
	sort.Slice(sites, func(i, j int) bool { return sites[i].Site < sites[j].Site })
	return Overview{Keys: keys, Sites: sites}
}

// indexOf returns the position of id in Keys, or -1. Callers hold the lock.
func (ks *KeyStore) indexOf(id string) int {
	for i := range ks.file.Keys {
		if ks.file.Keys[i].ID == id {
			return i
		}
	}
	return -1
}

func publicKey(k deployKey) PublicKey {
	return PublicKey{ID: k.ID, User: k.User, Created: k.Created, Revoked: k.Revoked}
}

func validUser(s string) bool {
	if len(s) == 0 || len(s) > 63 {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case r == '.', r == '_', r == '-':
		default:
			return false
		}
	}
	return true
}

func newKeyID() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate key id: %w", err)
	}
	return "dk_" + hex.EncodeToString(b), nil
}
