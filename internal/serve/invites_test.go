package serve_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/nemoNoboru/relay/internal/serve"
)

// TestInvitesPassesOn pins that a key holder can invite someone else with
// nothing but their own deploy token.
func TestInvitesPassesOn(t *testing.T) {
	ks, err := serve.NewKeyStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	alice, _, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}
	rec := doJSON(t, serve.NewInvites(ks), http.MethodPost, "/_relay/v1/invites", alice, `{"name":"bob"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("invite = %d %s, want 201", rec.Code, rec.Body)
	}
	var got struct {
		Key   serve.KeyInfo `json:"key"`
		Token string        `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Key.User != "alice/bob" {
		t.Fatalf("invited user = %q, want alice/bob", got.Key.User)
	}
	if user, ok := ks.Authenticate(got.Token); !ok || user != "alice/bob" {
		t.Fatalf("invited token = %q, %v; want alice/bob, true", user, ok)
	}
}

// TestInvitesRejectsStrangers pins that only a live key may invite.
func TestInvitesRejectsStrangers(t *testing.T) {
	ks, err := serve.NewKeyStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{"", "wrong"} {
		rec := doJSON(t, serve.NewInvites(ks), http.MethodPost, "/_relay/v1/invites", token, `{"name":"bob"}`)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("invite with %q = %d, want 401", token, rec.Code)
		}
	}
}
