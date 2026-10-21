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
	"slices"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
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

// ErrBadUser is returned for an invalid user label. User labels are display
// names, not path segments, so the grammar is wider than a node/site name.
var ErrBadUser = errors.New("invalid user label: must match [A-Za-z0-9._-]{1,63}")

// ErrBadInviter is returned when an invite is offered by a token that is
// unknown or already revoked.
var ErrBadInviter = errors.New("unknown or revoked inviter key")

// ErrBadProfile is returned when a member's display name or bio is invalid.
var ErrBadProfile = errors.New("invalid profile: display name up to 63 characters and bio up to 280")

// deployKey is one issued per-user deploy credential, as persisted. The
// plaintext token is kept so the operator can copy it from the console; the
// hash is what Authenticate compares against, in constant time.
type deployKey struct {
	ID      string    `json:"id"`
	User    string    `json:"user"`
	Token   string    `json:"token"`
	Hash    string    `json:"hash"`
	Created time.Time `json:"created"`
	Revoked bool      `json:"revoked,omitempty"`
	// InvitedBy is the ID of the key that invited this one, empty for keys
	// the operator issued. Revoking a key revokes everything below it.
	InvitedBy string `json:"invitedBy,omitempty"`
	// DisplayName and Bio are optional member profile fields. An empty
	// DisplayName falls back to the last segment of User.
	DisplayName string `json:"displayName,omitempty"`
	Bio         string `json:"bio,omitempty"`
}

// keyFile is the on-disk shape of the deploy-key store.
type keyFile struct {
	Keys  []deployKey       `json:"keys"`
	Sites map[string]string `json:"sites"` // site name -> user label
}

// KeyInfo is the view of a deploy key returned by the admin API. It carries
// the plaintext token (for the operator console) but never the hash.
type KeyInfo struct {
	ID          string    `json:"id"`
	User        string    `json:"user"`
	Token       string    `json:"token"`
	Created     time.Time `json:"created"`
	Revoked     bool      `json:"revoked"`
	InvitedBy   string    `json:"invitedBy,omitempty"`
	DisplayName string    `json:"displayName,omitempty"`
	Bio         string    `json:"bio,omitempty"`
}

// Member is the member-facing view of a live invitation key. It is the
// identity the member cookie resolves to.
type Member struct {
	ID          string    `json:"id"`
	User        string    `json:"user"`
	DisplayName string    `json:"displayName"`
	Bio         string    `json:"bio,omitempty"`
	Created     time.Time `json:"created"`
	InvitedBy   string    `json:"invitedBy,omitempty"`
	Revoked     bool      `json:"revoked,omitempty"`
}

// SiteOwner pairs a site with the user who last deployed it.
type SiteOwner struct {
	Site string `json:"site"`
	User string `json:"user"`
}

// Overview is the admin API's read view: all keys plus site ownership.
type Overview struct {
	Keys  []KeyInfo   `json:"keys"`
	Sites []SiteOwner `json:"sites"`
}

// KeyStore persists per-user deploy keys and site ownership under
// <dataDir>/deploy-keys.json (0600). Mutations are serialized and written
// atomically; reads take a shared lock.
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

// MemberByToken returns the live member a token belongs to. It is the
// session/cookie lookup; revocation takes effect immediately.
func (ks *KeyStore) MemberByToken(token string) (Member, bool) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	i := ks.indexOfToken(token)
	if i == -1 || ks.file.Keys[i].Revoked {
		return Member{}, false
	}
	return memberOf(ks.file.Keys[i]), true
}

// MemberByID returns a member by key id, including a revoked member. It is
// used to show the author of posts that outlive the member's key.
func (ks *KeyStore) MemberByID(id string) (Member, bool) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	i := ks.indexOf(id)
	if i == -1 {
		return Member{}, false
	}
	return memberOf(ks.file.Keys[i]), true
}

// MemberByUser returns a live member by the stable user label. If several
// keys share a label (a rotation), the first live one wins.
func (ks *KeyStore) MemberByUser(user string) (Member, bool) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	for _, k := range ks.file.Keys {
		if k.User == user && !k.Revoked {
			return memberOf(k), true
		}
	}
	return Member{}, false
}

// SiteUser returns the user label a site is attributed to, if any.
func (ks *KeyStore) SiteUser(site string) string {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	return ks.file.Sites[site]
}

// LiveCount returns the number of members with a live key.
func (ks *KeyStore) LiveCount() int {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	n := 0
	for _, k := range ks.file.Keys {
		if !k.Revoked {
			n++
		}
	}
	return n
}

// LiveMembers returns the node's live members, sorted by display name. It is
// the public directory view: no tokens, no revoked members.
func (ks *KeyStore) LiveMembers() []Member {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	out := make([]Member, 0, len(ks.file.Keys))
	for _, k := range ks.file.Keys {
		if !k.Revoked {
			out = append(out, memberOf(k))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DisplayName < out[j].DisplayName })
	return out
}

// UpdateProfile changes a member's display name and bio. An empty display
// name resets the view to the fallback (the last segment of User).
func (ks *KeyStore) UpdateProfile(token, displayName, bio string) (Member, bool, error) {
	displayName, okName := cleanText(displayName, 63, false)
	bio, okBio := cleanText(bio, 280, true)
	if !okName || !okBio {
		return Member{}, false, ErrBadProfile
	}

	ks.mu.Lock()
	defer ks.mu.Unlock()
	i := ks.indexOfToken(token)
	if i == -1 || ks.file.Keys[i].Revoked {
		return Member{}, false, nil
	}
	oldName, oldBio := ks.file.Keys[i].DisplayName, ks.file.Keys[i].Bio
	ks.file.Keys[i].DisplayName = displayName
	ks.file.Keys[i].Bio = bio
	if err := ks.persist(); err != nil {
		ks.file.Keys[i].DisplayName, ks.file.Keys[i].Bio = oldName, oldBio
		return Member{}, false, err
	}
	return memberOf(ks.file.Keys[i]), true, nil
}

// Authenticate returns the user a token belongs to, or ok=false if the token
// is unknown or its key is revoked. The comparison runs over every key so
// the timing does not reveal which key matched.
func (ks *KeyStore) Authenticate(token string) (user string, ok bool) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	i := ks.indexOfToken(token)
	if i == -1 || ks.file.Keys[i].Revoked {
		return "", false
	}
	return ks.file.Keys[i].User, true
}

// Create mints a deploy key for user and returns the plaintext token
// alongside its view (the token is also stored, for the operator console).
func (ks *KeyStore) Create(user string) (string, KeyInfo, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	if !validUser(user) {
		return "", KeyInfo{}, ErrBadUser
	}
	return ks.mint(user, "")
}

// Invite mints a key on behalf of a key holder: anyone with a live key can
// pass an invite on. The new label carries the chain ("alice/bob") and the
// new key records its inviter, so revoking someone also revokes everyone
// they brought in.
func (ks *KeyStore) Invite(inviterToken, name string) (string, KeyInfo, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	if !validUser(name) {
		return "", KeyInfo{}, ErrBadUser
	}
	i := ks.indexOfToken(inviterToken)
	if i == -1 || ks.file.Keys[i].Revoked {
		return "", KeyInfo{}, ErrBadInviter
	}
	return ks.mint(ks.file.Keys[i].User+"/"+name, ks.file.Keys[i].ID)
}

// mint appends a fresh key and persists it. Callers hold the write lock and
// have already validated the label.
func (ks *KeyStore) mint(user, invitedBy string) (string, KeyInfo, error) {
	token, err := GenerateToken()
	if err != nil {
		return "", KeyInfo{}, err
	}
	id, err := newKeyID()
	if err != nil {
		return "", KeyInfo{}, err
	}
	k := deployKey{
		ID: id, User: user, Token: token, Hash: tokenHash(token),
		Created: time.Now().UTC(), InvitedBy: invitedBy,
	}
	ks.file.Keys = append(ks.file.Keys, k)
	if err := ks.persist(); err != nil {
		ks.file.Keys = ks.file.Keys[:len(ks.file.Keys)-1]
		return "", KeyInfo{}, err
	}
	return token, keyInfo(k), nil
}

// Revoke marks a key, and everyone it invited, as revoked. It is
// idempotent: revoking an already revoked key succeeds. found=false means
// no such key.
func (ks *KeyStore) Revoke(id string) (KeyInfo, bool, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	i := ks.indexOf(id)
	if i == -1 {
		return KeyInfo{}, false, nil
	}
	before := slices.Clone(ks.file.Keys)
	for _, j := range ks.descendants(id) {
		ks.file.Keys[j].Revoked = true
	}
	if err := ks.persist(); err != nil {
		ks.file.Keys = before
		return KeyInfo{}, false, err
	}
	return keyInfo(ks.file.Keys[i]), true, nil
}

// Rotate revokes the old key and mints a fresh one for the same user,
// returning the new plaintext token. found=false means no such key.
func (ks *KeyStore) Rotate(id string) (string, KeyInfo, bool, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	i := ks.indexOf(id)
	if i == -1 {
		return "", KeyInfo{}, false, nil
	}
	token, err := GenerateToken()
	if err != nil {
		return "", KeyInfo{}, false, err
	}
	newID, err := newKeyID()
	if err != nil {
		return "", KeyInfo{}, false, err
	}
	before := slices.Clone(ks.file.Keys)
	old := ks.file.Keys[i]
	ks.file.Keys[i].Revoked = true
	// Invitees belong to the person, not the token, so they follow the key
	// across a rotation and a later revoke still reaches them.
	for j := range ks.file.Keys {
		if ks.file.Keys[j].InvitedBy == id {
			ks.file.Keys[j].InvitedBy = newID
		}
	}
	k := deployKey{
		ID: newID, User: old.User, Token: token, Hash: tokenHash(token),
		Created: time.Now().UTC(), InvitedBy: old.InvitedBy,
	}
	ks.file.Keys = append(ks.file.Keys, k)
	if err := ks.persist(); err != nil {
		ks.file.Keys = before
		return "", KeyInfo{}, false, err
	}
	return token, keyInfo(k), true, nil
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

// ForgetSite drops a site's ownership record. A no-op when the site is not
// recorded.
func (ks *KeyStore) ForgetSite(site string) error {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	if _, ok := ks.file.Sites[site]; !ok {
		return nil
	}
	delete(ks.file.Sites, site)
	return ks.persist()
}

// Overview returns a sorted read view of the store.
func (ks *KeyStore) Overview() Overview {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	keys := make([]KeyInfo, 0, len(ks.file.Keys))
	for _, k := range ks.file.Keys {
		keys = append(keys, keyInfo(k))
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

// indexOfToken returns the position of the key a token belongs to, or -1.
// The scan covers every key so the timing does not reveal which matched.
// Callers hold the lock.
func (ks *KeyStore) indexOfToken(token string) int {
	want := sha256.Sum256([]byte(token))
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
	return match
}

// descendants returns the positions of id and every key invited below it.
// Callers hold the lock.
func (ks *KeyStore) descendants(id string) []int {
	children := map[string][]string{}
	for _, k := range ks.file.Keys {
		children[k.InvitedBy] = append(children[k.InvitedBy], k.ID)
	}
	found := map[string]bool{}
	var walk func(string)
	walk = func(id string) {
		if found[id] {
			return // also keeps a corrupted cycle from spinning
		}
		found[id] = true
		for _, c := range children[id] {
			walk(c)
		}
	}
	walk(id)
	var out []int
	for i := range ks.file.Keys {
		if found[ks.file.Keys[i].ID] {
			out = append(out, i)
		}
	}
	return out
}

func keyInfo(k deployKey) KeyInfo {
	return KeyInfo{
		ID: k.ID, User: k.User, Token: k.Token,
		Created: k.Created, Revoked: k.Revoked, InvitedBy: k.InvitedBy,
		DisplayName: k.DisplayName, Bio: k.Bio,
	}
}

// memberOf projects a stored key into the member-facing view.
func memberOf(k deployKey) Member {
	name := k.DisplayName
	if name == "" {
		name = k.User
		if i := strings.LastIndexByte(name, '/'); i >= 0 {
			name = name[i+1:]
		}
	}
	return Member{
		ID: k.ID, User: k.User, DisplayName: name, Bio: k.Bio,
		Created: k.Created, InvitedBy: k.InvitedBy, Revoked: k.Revoked,
	}
}

// cleanText trims and validates a profile field. Control characters are
// rejected (except newline and tab in a multiline bio).
func cleanText(s string, max int, multiline bool) (string, bool) {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) > max {
		return "", false
	}
	for _, r := range s {
		if r == utf8.RuneError || r == 0x7f {
			return "", false
		}
		if r < 0x20 && !(multiline && (r == '\n' || r == '\t')) {
			return "", false
		}
	}
	return s, true
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
