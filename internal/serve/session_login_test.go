package serve_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nemoNoboru/relay/internal/serve"
)

type sessionLoginReply struct {
	Username string `json:"username"`
}

func newCredentialStore(t *testing.T) (*serve.KeyStore, string, serve.KeyInfo) {
	t.Helper()
	ks, err := serve.NewKeyStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	token, key, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}
	return ks, token, key
}

// TestLoginSetsMemberCookie pins that a username + password login sets the
// same relay_member cookie the invitation exchange sets.
func TestLoginSetsMemberCookie(t *testing.T) {
	ks, token, key := newCredentialStore(t)
	if _, _, err := ks.SetCredentials(key.ID, "alice", "password123"); err != nil {
		t.Fatal(err)
	}
	h := serve.NewSession("mynode", ks)

	rec := doMember(t, h, http.MethodPost, "/_relay/v1/login", "", `{"username":"alice","password":"password123"}`, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("login = %d %s, want 200", rec.Code, rec.Body)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "relay_member" || cookies[0].Value != token {
		t.Fatalf("cookies = %+v, want relay_member=%q", cookies, token)
	}
	var reply sessionLoginReply
	if err := json.Unmarshal(rec.Body.Bytes(), &reply); err != nil {
		t.Fatal(err)
	}
	if reply.Username != "alice" {
		t.Fatalf("reply username = %q, want alice", reply.Username)
	}
}

// TestLoginBadCredentialsIdentical pins that an unknown username and a wrong
// password return the exact same error, so login does not enumerate members.
func TestLoginBadCredentialsIdentical(t *testing.T) {
	ks, _, key := newCredentialStore(t)
	if _, _, err := ks.SetCredentials(key.ID, "alice", "password123"); err != nil {
		t.Fatal(err)
	}
	var bodies []string
	for _, body := range []string{
		`{"username":"alice","password":"wrong"}`,
		`{"username":"nobody","password":"wrong"}`,
	} {
		h := serve.NewSession("mynode", ks)
		rec := doMember(t, h, http.MethodPost, "/_relay/v1/login", "", body, "")
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("login %s = %d, want 401", body, rec.Code)
		}
		var e struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil {
			t.Fatal(err)
		}
		if e.Error.Code != "bad_credentials" {
			t.Fatalf("error code = %q, want bad_credentials", e.Error.Code)
		}
		bodies = append(bodies, rec.Body.String())
	}
	if bodies[0] != bodies[1] {
		t.Fatalf("bad-credential bodies differ: %q vs %q", bodies[0], bodies[1])
	}
}

// TestLoginRateLimit pins the per-username rate guard.
func TestLoginRateLimit(t *testing.T) {
	ks, _, key := newCredentialStore(t)
	if _, _, err := ks.SetCredentials(key.ID, "alice", "password123"); err != nil {
		t.Fatal(err)
	}
	h := serve.NewSession("mynode", ks)

	if rec := doMember(t, h, http.MethodPost, "/_relay/v1/login", "", `{"username":"alice","password":"wrong"}`, ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("first attempt = %d, want 401", rec.Code)
	}
	rec := doMember(t, h, http.MethodPost, "/_relay/v1/login", "", `{"username":"alice","password":"wrong"}`, "")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("second attempt = %d, want 429", rec.Code)
	}
}

// TestLoginRejectsCrossOrigin pins the same-origin gate on login.
func TestLoginRejectsCrossOrigin(t *testing.T) {
	ks, _, _ := newCredentialStore(t)
	h := serve.NewSession("mynode", ks)
	req := httptest.NewRequest(http.MethodPost, "/_relay/v1/login", strings.NewReader(`{"username":"alice","password":"password123"}`))
	req.Host = "relay.example.org"
	req.Header.Set("Origin", "https://evil.example")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("cross-origin login = %d, want 403", rec.Code)
	}
}

// TestSessionRedemptionSetsCredentials pins that opening an invite with
// credentials sets them and keeps the same token in the cookie.
func TestSessionRedemptionSetsCredentials(t *testing.T) {
	ks, token, _ := newCredentialStore(t)
	h := serve.NewSession("mynode", ks)

	rec := doMember(t, h, http.MethodPost, "/_relay/v1/session", "", `{"token":"`+token+`","username":"alice","password":"password123"}`, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("redemption = %d %s, want 200", rec.Code, rec.Body)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Value != token {
		t.Fatalf("cookies = %+v, want relay_member=%q", cookies, token)
	}
	var reply sessionLoginReply
	if err := json.Unmarshal(rec.Body.Bytes(), &reply); err != nil {
		t.Fatal(err)
	}
	if reply.Username != "alice" {
		t.Fatalf("reply username = %q, want alice", reply.Username)
	}

	rec = doMember(t, h, http.MethodPost, "/_relay/v1/login", "", `{"username":"alice","password":"password123"}`, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("login after redemption = %d %s, want 200", rec.Code, rec.Body)
	}
}

// TestSessionRedemptionSkipLeavesCredentialsEmpty pins the skip path.
func TestSessionRedemptionSkipLeavesCredentialsEmpty(t *testing.T) {
	ks, token, _ := newCredentialStore(t)
	h := serve.NewSession("mynode", ks)

	rec := doMember(t, h, http.MethodPost, "/_relay/v1/session", "", `{"token":"`+token+`"}`, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("redemption = %d %s, want 200", rec.Code, rec.Body)
	}
	var reply sessionLoginReply
	if err := json.Unmarshal(rec.Body.Bytes(), &reply); err != nil {
		t.Fatal(err)
	}
	if reply.Username != "" {
		t.Fatalf("reply username = %q, want empty", reply.Username)
	}
	if _, ok := ks.AuthenticatePassword("alice", "password123"); ok {
		t.Fatalf("skip still set credentials")
	}
}

// TestSessionRedemptionDuplicateUsernameRefused pins that a duplicate is a
// distinct 409 and does not set a cookie.
func TestSessionRedemptionDuplicateUsernameRefused(t *testing.T) {
	ks, _, alice := newCredentialStore(t)
	if _, _, err := ks.SetCredentials(alice.ID, "alice", "password123"); err != nil {
		t.Fatal(err)
	}
	bobToken, _, err := ks.Create("bob")
	if err != nil {
		t.Fatal(err)
	}
	h := serve.NewSession("mynode", ks)

	rec := doMember(t, h, http.MethodPost, "/_relay/v1/session", "", `{"token":"`+bobToken+`","username":"alice","password":"password123"}`, "")
	if rec.Code != http.StatusConflict {
		t.Fatalf("duplicate redemption = %d %s, want 409", rec.Code, rec.Body)
	}
	if len(rec.Result().Cookies()) != 0 {
		t.Fatalf("duplicate redemption set a cookie")
	}
}
