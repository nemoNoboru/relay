package identity_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/nemoNoboru/relay/internal/identity"
)

// newCredentialStore returns a store with one live key for "alice".
func newCredentialStore(t *testing.T) (*identity.KeyStore, string, identity.KeyInfo) {
	t.Helper()
	ks, err := identity.NewKeyStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	token, key, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}
	return ks, token, key
}

// TestKeyStoreSetCredentialsHashVerify pins the happy path: credentials set,
// correct password authenticates as the invitation token, wrong and unknown
// credentials do not, and the admin view never carries the hash or salt.
func TestKeyStoreSetCredentialsHashVerify(t *testing.T) {
	ks, token, key := newCredentialStore(t)

	updated, found, err := ks.SetCredentials(key.ID, "alice", "correct horse battery")
	if err != nil || !found {
		t.Fatalf("SetCredentials = found %v, err %v", found, err)
	}
	if updated.Username != "alice" {
		t.Fatalf("username = %q, want alice", updated.Username)
	}

	tests := []struct {
		name      string
		username  string
		password  string
		wantToken string
		wantOK    bool
	}{
		{"correct", "alice", "correct horse battery", token, true},
		{"wrong password", "alice", "wrong", "", false},
		{"unknown username", "nobody", "correct horse battery", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ks.AuthenticatePassword(tt.username, tt.password)
			if ok != tt.wantOK || (ok && got != tt.wantToken) {
				t.Fatalf("AuthenticatePassword(%q, %q) = %q, %v; want %q, %v", tt.username, tt.password, got, ok, tt.wantToken, tt.wantOK)
			}
		})
	}

	b, err := json.Marshal(updated)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "passwordHash") || strings.Contains(string(b), "passwordSalt") {
		t.Fatalf("admin view leaks credential material: %s", b)
	}
}

// TestKeyStoreUsernameUniqueness pins that a username is node-local and
// unique among live members, and that UsernameTaken reports it.
func TestKeyStoreUsernameUniqueness(t *testing.T) {
	ks, _, alice := newCredentialStore(t)
	_, bob, err := ks.Create("bob")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := ks.SetCredentials(alice.ID, "same", "password123"); err != nil {
		t.Fatal(err)
	}
	if !ks.UsernameTaken("same") {
		t.Fatalf("UsernameTaken(same) = false, want true")
	}
	if _, _, err := ks.SetCredentials(bob.ID, "same", "password123"); err != identity.ErrUsernameTaken {
		t.Fatalf("second SetCredentials err = %v, want ErrUsernameTaken", err)
	}
	if ks.UsernameTaken("other") {
		t.Fatalf("UsernameTaken(other) = true, want false")
	}
}

// TestKeyStoreSetCredentialsRejects pins the username grammar and password
// floor on the credential setter.
func TestKeyStoreSetCredentialsRejects(t *testing.T) {
	ks, _, key := newCredentialStore(t)
	tests := []struct {
		name     string
		username string
		password string
		want     error
	}{
		{"empty username", "", "password123", identity.ErrBadUser},
		{"space in username", "has space", "password123", identity.ErrBadUser},
		{"slash in username", "a/b", "password123", identity.ErrBadUser},
		{"short password", "alice", "short", identity.ErrBadPassword},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, err := ks.SetCredentials(key.ID, tt.username, tt.password); err != tt.want {
				t.Fatalf("SetCredentials(%q, %q) err = %v, want %v", tt.username, tt.password, err, tt.want)
			}
		})
	}
}

// TestKeyStoreRevokedMemberCannotAuthenticateByPassword pins that revocation
// also stops password login, and that a revoked member's username is freed.
func TestKeyStoreRevokedMemberCannotAuthenticateByPassword(t *testing.T) {
	ks, _, key := newCredentialStore(t)
	if _, _, err := ks.SetCredentials(key.ID, "alice", "password123"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ks.Revoke(key.ID); err != nil {
		t.Fatal(err)
	}
	if _, ok := ks.AuthenticatePassword("alice", "password123"); ok {
		t.Fatalf("revoked member still authenticates by password")
	}
	if ks.UsernameTaken("alice") {
		t.Fatalf("revoked member still holds the username")
	}
}

// TestKeyStoreClearCredentials pins that clearing removes only the credential
// fields: the invitation token keeps authenticating.
func TestKeyStoreClearCredentials(t *testing.T) {
	ks, token, key := newCredentialStore(t)
	if _, _, err := ks.SetCredentials(key.ID, "alice", "password123"); err != nil {
		t.Fatal(err)
	}
	cleared, found, err := ks.ClearCredentials(key.ID)
	if err != nil || !found {
		t.Fatalf("ClearCredentials = found %v, err %v", found, err)
	}
	if cleared.Username != "" {
		t.Fatalf("cleared username = %q, want empty", cleared.Username)
	}
	if _, ok := ks.AuthenticatePassword("alice", "password123"); ok {
		t.Fatalf("cleared credentials still authenticate by password")
	}
	if ks.UsernameTaken("alice") {
		t.Fatalf("cleared username is still taken")
	}
	if user, ok := ks.Authenticate(token); !ok || user != "alice" {
		t.Fatalf("invitation token = %q, %v; want alice, true after clear", user, ok)
	}
}

// TestKeyStoreCredentialsPersist pins that credentials survive a reload.
func TestKeyStoreCredentialsPersist(t *testing.T) {
	data := t.TempDir()
	ks, err := identity.NewKeyStore(data)
	if err != nil {
		t.Fatal(err)
	}
	token, key, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := ks.SetCredentials(key.ID, "alice", "password123"); err != nil {
		t.Fatal(err)
	}

	reloaded, err := identity.NewKeyStore(data)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := reloaded.AuthenticatePassword("alice", "password123"); !ok || got != token {
		t.Fatalf("password did not survive reload: %q, %v", got, ok)
	}
	if !reloaded.UsernameTaken("alice") {
		t.Fatalf("username did not survive reload")
	}
}
