package identity

import (
	"errors"
	"sync"
	"time"
)

// ErrBadUser is returned for an invalid user label. User labels are display
// names, not path segments, so the grammar is wider than a node/site name.
var ErrBadUser = errors.New("invalid user label: must match [A-Za-z0-9._-]{1,63}")

// ErrBadInviter is returned when an invite is offered by a token that is
// unknown or already revoked.
var ErrBadInviter = errors.New("unknown or revoked inviter key")

// ErrBadProfile is returned when a member's display name or bio is invalid.
var ErrBadProfile = errors.New("invalid profile: display name up to 63 characters and bio up to 280")

// ErrUsernameTaken is returned when a username is already held by a live
// member. Usernames are node-local and unique.
var ErrUsernameTaken = errors.New("username is already taken")

// ErrBadPassword is returned when a member password is shorter than the
// node's password floor.
var ErrBadPassword = errors.New("invalid password: at least 8 characters")

// deployKey is one issued per-user deploy credential, as persisted. The
// plaintext token is kept so the operator can copy it from the console; the
// hash is what Authenticate compares against, in constant time.
type deployKey struct {
	ID      string    `json:"id"`
	Label   string    `json:"user"`
	Token   string    `json:"token"`
	Hash    string    `json:"hash"`
	Created time.Time `json:"created"`
	Revoked bool      `json:"revoked,omitempty"`
	// InvitedBy is the ID of the key that invited this one, empty for keys
	// the operator issued. Revoking a key revokes everything below it.
	InvitedBy string `json:"invitedBy,omitempty"`
	// DisplayName and Bio are optional member profile fields. An empty
	// DisplayName falls back to the last segment of Label.
	DisplayName string `json:"displayName,omitempty"`
	Bio         string `json:"bio,omitempty"`
	// Username and password are the return credentials. The plaintext
	// password is never stored: only the salt, the argon2id hash, and the
	// parameters that produced it.
	Username        string `json:"username,omitempty"`
	PasswordSalt    string `json:"passwordSalt,omitempty"`
	PasswordHash    string `json:"passwordHash,omitempty"`
	PasswordTime    uint32 `json:"passwordTime,omitempty"`
	PasswordMemory  uint32 `json:"passwordMemory,omitempty"`
	PasswordThreads uint8  `json:"passwordThreads,omitempty"`
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
	Label       string    `json:"user"`
	Token       string    `json:"token"`
	Created     time.Time `json:"created"`
	Revoked     bool      `json:"revoked"`
	InvitedBy   string    `json:"invitedBy,omitempty"`
	DisplayName string    `json:"displayName,omitempty"`
	Bio         string    `json:"bio,omitempty"`
	Username    string    `json:"username,omitempty"`
}

// Member is the member-facing view of a live invitation key. It is the
// identity the member cookie resolves to.
type Member struct {
	ID          string    `json:"id"`
	Label       string    `json:"user"`
	DisplayName string    `json:"displayName"`
	Bio         string    `json:"bio,omitempty"`
	Created     time.Time `json:"created"`
	InvitedBy   string    `json:"invitedBy,omitempty"`
	Revoked     bool      `json:"revoked,omitempty"`
}

// SiteOwner pairs a site with the user who last deployed it.
type SiteOwner struct {
	Site  string `json:"site"`
	Label string `json:"user"`
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
