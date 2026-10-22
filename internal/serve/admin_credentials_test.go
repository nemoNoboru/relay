package serve_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/nemoNoboru/relay/internal/serve"
)

// TestAdminSetAndClearCredentials pins the host password reset/clear surface:
// it mutates only the credential fields and leaves the invitation live.
func TestAdminSetAndClearCredentials(t *testing.T) {
	a, ks := newAdmin(t)
	token, key, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}

	rec := doJSON(t, a, http.MethodPost, "/_relay/v1/admin/keys/"+key.ID+"/credentials", adminToken, `{"username":"alice","password":"password123"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("set credentials = %d %s, want 200", rec.Code, rec.Body)
	}
	var set struct {
		Key serve.KeyInfo `json:"key"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &set); err != nil {
		t.Fatal(err)
	}
	if set.Key.Username != "alice" {
		t.Fatalf("set username = %q, want alice", set.Key.Username)
	}
	if set.Key.ID != key.ID || set.Key.Label != "alice" || set.Key.Token != token || set.Key.Revoked {
		t.Fatalf("set mutated non-credential fields: %+v", set.Key)
	}

	rec = doJSON(t, a, http.MethodGet, "/_relay/v1/admin", adminToken, "")
	var ov struct {
		Keys []serve.KeyInfo `json:"keys"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &ov); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, k := range ov.Keys {
		if k.ID == key.ID {
			found = true
			if k.Username != "alice" {
				t.Fatalf("overview username = %q, want alice", k.Username)
			}
		}
	}
	if !found {
		t.Fatalf("key %s missing from overview", key.ID)
	}

	if user, ok := ks.Authenticate(token); !ok || user != "alice" {
		t.Fatalf("invitation after set = %q, %v; want alice, true", user, ok)
	}

	rec = doJSON(t, a, http.MethodDelete, "/_relay/v1/admin/keys/"+key.ID+"/credentials", adminToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("clear credentials = %d %s, want 200", rec.Code, rec.Body)
	}
	var cleared struct {
		Key serve.KeyInfo `json:"key"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &cleared); err != nil {
		t.Fatal(err)
	}
	if cleared.Key.Username != "" {
		t.Fatalf("cleared username = %q, want empty", cleared.Key.Username)
	}
	if _, ok := ks.AuthenticatePassword("alice", "password123"); ok {
		t.Fatalf("cleared credentials still authenticate by password")
	}
	if user, ok := ks.Authenticate(token); !ok || user != "alice" {
		t.Fatalf("invitation after clear = %q, %v; want alice, true", user, ok)
	}
}

// TestAdminSetCredentialsDuplicate pins the distinct username_taken error.
func TestAdminSetCredentialsDuplicate(t *testing.T) {
	a, ks := newAdmin(t)
	_, alice, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := ks.SetCredentials(alice.ID, "alice", "password123"); err != nil {
		t.Fatal(err)
	}
	_, bob, err := ks.Create("bob")
	if err != nil {
		t.Fatal(err)
	}

	rec := doJSON(t, a, http.MethodPost, "/_relay/v1/admin/keys/"+bob.ID+"/credentials", adminToken, `{"username":"alice","password":"password123"}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("duplicate = %d %s, want 409", rec.Code, rec.Body)
	}
}

// TestAdminCredentialsUnknownKey pins 404s for a missing key.
func TestAdminCredentialsUnknownKey(t *testing.T) {
	a, _ := newAdmin(t)
	if rec := doJSON(t, a, http.MethodPost, "/_relay/v1/admin/keys/dk_nope/credentials", adminToken, `{"username":"alice","password":"password123"}`); rec.Code != http.StatusNotFound {
		t.Fatalf("set unknown = %d, want 404", rec.Code)
	}
	if rec := doJSON(t, a, http.MethodDelete, "/_relay/v1/admin/keys/dk_nope/credentials", adminToken, ""); rec.Code != http.StatusNotFound {
		t.Fatalf("clear unknown = %d, want 404", rec.Code)
	}
}
