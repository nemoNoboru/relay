package serve_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nemoNoboru/relay/internal/serve"
)

// TestKeyStoreCreateAuthenticate pins the happy path: mint a key, the token
// authenticates as its user, and a wrong token does not.
func TestKeyStoreCreateAuthenticate(t *testing.T) {
	ks, err := serve.NewKeyStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	token, key, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}
	if token == "" || key.ID == "" || key.User != "alice" || key.Revoked {
		t.Fatalf("key = %+v, token = %q", key, token)
	}
	if user, ok := ks.Authenticate(token); !ok || user != "alice" {
		t.Fatalf("Authenticate(token) = %q, %v; want alice, true", user, ok)
	}
	if user, ok := ks.Authenticate("wrong"); ok || user != "" {
		t.Fatalf("Authenticate(wrong) = %q, %v; want false", user, ok)
	}
}

// TestKeyStoreRejectsBadUser pins the user-label grammar.
func TestKeyStoreRejectsBadUser(t *testing.T) {
	ks, err := serve.NewKeyStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, user := range []string{"", "  ", "has space", "a/b", ">64" + strings.Repeat("x", 64)} {
		if _, _, err := ks.Create(user); err != serve.ErrBadUser {
			t.Fatalf("Create(%q) err = %v, want ErrBadUser", user, err)
		}
	}
}

// TestKeyStoreRevoke pins that revocation is immediate, idempotent, and
// stops authentication.
func TestKeyStoreRevoke(t *testing.T) {
	ks, err := serve.NewKeyStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	token, key, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}
	if _, found, err := ks.Revoke(key.ID); err != nil || !found {
		t.Fatalf("first revoke = found %v, err %v", found, err)
	}
	if _, ok := ks.Authenticate(token); ok {
		t.Fatalf("revoked key still authenticates")
	}
	if k, found, err := ks.Revoke(key.ID); err != nil || !found || !k.Revoked {
		t.Fatalf("second revoke = %+v found %v err %v, want idempotent", k, found, err)
	}
	if _, found, _ := ks.Revoke("dk_nope"); found {
		t.Fatalf("revoking an unknown id reported found")
	}
}

// TestKeyStoreRotate pins that rotation mints a new live token and kills the
// old one, keeping the same user.
func TestKeyStoreRotate(t *testing.T) {
	ks, err := serve.NewKeyStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	oldToken, oldKey, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}
	newToken, newKey, found, err := ks.Rotate(oldKey.ID)
	if err != nil || !found {
		t.Fatalf("rotate = found %v, err %v", found, err)
	}
	if newToken == "" || newToken == oldToken || newKey.ID == oldKey.ID || newKey.User != "alice" {
		t.Fatalf("new = %q / %+v, old = %q / %+v", newToken, newKey, oldToken, oldKey)
	}
	if _, ok := ks.Authenticate(oldToken); ok {
		t.Fatalf("old token still authenticates after rotate")
	}
	if user, ok := ks.Authenticate(newToken); !ok || user != "alice" {
		t.Fatalf("new token = %q, %v; want alice, true", user, ok)
	}
	if _, _, found, _ := ks.Rotate("dk_nope"); found {
		t.Fatalf("rotating an unknown id reported found")
	}
}

// TestKeyStoreRecordSite pins site attribution and its persistence.
func TestKeyStoreRecordSite(t *testing.T) {
	ks, err := serve.NewKeyStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := ks.RecordSite("blog", "alice"); err != nil {
		t.Fatal(err)
	}
	if err := ks.RecordSite("zine", "alice"); err != nil {
		t.Fatal(err)
	}
	if err := ks.RecordSite("blog", "bob"); err != nil {
		t.Fatal(err)
	}
	ov := ks.Overview()
	if len(ov.Sites) != 2 {
		t.Fatalf("sites = %+v, want 2 entries", ov.Sites)
	}
	for _, s := range ov.Sites {
		switch s.Site {
		case "blog":
			if s.User != "bob" {
				t.Fatalf("blog owner = %q, want bob", s.User)
			}
		case "zine":
			if s.User != "alice" {
				t.Fatalf("zine owner = %q, want alice", s.User)
			}
		default:
			t.Fatalf("unexpected site %q", s.Site)
		}
	}
}

// TestKeyStorePersists pins that state survives a reload (new instance reads
// the same file) and that plaintext tokens are never written to disk.
func TestKeyStorePersists(t *testing.T) {
	data := t.TempDir()
	ks, err := serve.NewKeyStore(data)
	if err != nil {
		t.Fatal(err)
	}
	token, key, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}
	if err := ks.RecordSite("blog", "alice"); err != nil {
		t.Fatal(err)
	}

	reloaded, err := serve.NewKeyStore(data)
	if err != nil {
		t.Fatal(err)
	}
	if user, ok := reloaded.Authenticate(token); !ok || user != "alice" {
		t.Fatalf("token did not survive reload: %q, %v", user, ok)
	}
	if k, found, err := reloaded.Revoke(key.ID); err != nil || !found || !k.Revoked {
		t.Fatalf("revoke after reload = found %v, err %v", found, err)
	}

	raw, err := os.ReadFile(filepath.Join(data, "deploy-keys.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), token) {
		t.Fatalf("plaintext token leaked to disk")
	}
}

// TestKeyStoreTokensAreUnique pins that minted tokens do not collide (smoke,
// not a guarantee).
func TestKeyStoreTokensAreUnique(t *testing.T) {
	ks, err := serve.NewKeyStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		token, _, err := ks.Create("user")
		if err != nil {
			t.Fatal(err)
		}
		if seen[token] {
			t.Fatalf("token collision after %d mints", i)
		}
		seen[token] = true
	}
}
