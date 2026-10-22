package serve

import "github.com/nemoNoboru/relay/internal/identity"

// The identity model lives in internal/identity so the HTTP layer stays a
// thin translator. These aliases keep the existing serve surface unchanged.
type (
	KeyStore  = identity.KeyStore
	KeyInfo   = identity.KeyInfo
	Member    = identity.Member
	SiteOwner = identity.SiteOwner
	Overview  = identity.Overview
)

var (
	ErrBadUser       = identity.ErrBadUser
	ErrBadInviter    = identity.ErrBadInviter
	ErrBadProfile    = identity.ErrBadProfile
	ErrUsernameTaken = identity.ErrUsernameTaken
	ErrBadPassword   = identity.ErrBadPassword
)

// GenerateToken returns a new random credential token (32 bytes, base64url).
func GenerateToken() (string, error) {
	return identity.GenerateToken()
}

// NewKeyStore loads the deploy-key store from <dataDir>/deploy-keys.json.
func NewKeyStore(dataDir string) (*KeyStore, error) {
	return identity.NewKeyStore(dataDir)
}
