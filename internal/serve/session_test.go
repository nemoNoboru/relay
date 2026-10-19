package serve_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nemoNoboru/relay/internal/serve"
)

// doMember issues a request to the session handler with an optional bearer
// token and member cookie, returning the recorder.
func doMember(t *testing.T, h http.Handler, method, path, bearer, body, cookie string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: "relay_member", Value: cookie})
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

type meReply struct {
	Member serve.Member   `json:"member"`
	Node   serve.NodeView `json:"node"`
}

// TestSessionExchangesInviteForCookie pins the invite -> cookie -> /me flow.
func TestSessionExchangesInviteForCookie(t *testing.T) {
	ks, err := serve.NewKeyStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}
	h := serve.NewSession("mynode", ks)

	rec := doMember(t, h, http.MethodPost, "/_relay/v1/session", "", `{"token":"`+token+`"}`, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("session create = %d %s, want 200", rec.Code, rec.Body)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "relay_member" || cookies[0].Value != token {
		t.Fatalf("cookies = %+v, want relay_member=%q", cookies, token)
	}
	cookie := cookies[0]
	if !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode || cookie.Path != "/" {
		t.Fatalf("cookie attributes = %+v, want HttpOnly SameSite=Lax Path=/", cookie)
	}
	var created meReply
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Member.ID == "" || created.Member.DisplayName != "alice" || created.Node.Name != "mynode" || created.Node.Members != 1 {
		t.Fatalf("session reply = %+v", created)
	}

	rec = doMember(t, h, http.MethodGet, "/_relay/v1/me", "", "", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("/me = %d %s, want 200", rec.Code, rec.Body)
	}
	var me meReply
	if err := json.Unmarshal(rec.Body.Bytes(), &me); err != nil {
		t.Fatal(err)
	}
	if me.Member.ID != created.Member.ID || me.Member.DisplayName != "alice" {
		t.Fatalf("/me member = %+v", me.Member)
	}
}

// TestSessionMeRequiresCredential pins that the home page cannot read a
// member without a cookie or bearer token.
func TestSessionMeRequiresCredential(t *testing.T) {
	ks, err := serve.NewKeyStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	h := serve.NewSession("mynode", ks)
	for _, rec := range []*httptest.ResponseRecorder{
		doMember(t, h, http.MethodGet, "/_relay/v1/me", "", "", ""),
		doMember(t, h, http.MethodPost, "/_relay/v1/session", "", `{"token":"nope"}`, ""),
	} {
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d %s, want 401", rec.Code, rec.Body)
		}
	}
}

// TestSessionRejectsRevokedInvite pins immediate revocation.
func TestSessionRejectsRevokedInvite(t *testing.T) {
	ks, err := serve.NewKeyStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	token, key, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := ks.Revoke(key.ID); err != nil {
		t.Fatal(err)
	}
	h := serve.NewSession("mynode", ks)
	rec := doMember(t, h, http.MethodPost, "/_relay/v1/session", "", `{"token":"`+token+`"}`, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("revoked session = %d %s, want 401", rec.Code, rec.Body)
	}
	if len(rec.Result().Cookies()) != 0 {
		t.Fatalf("revoked invite set a cookie")
	}
}

// TestSessionProfileUpdate pins display name and bio editing, including the
// fallback to the invitation label.
func TestSessionProfileUpdate(t *testing.T) {
	ks, err := serve.NewKeyStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}
	h := serve.NewSession("mynode", ks)

	rec := doMember(t, h, http.MethodPatch, "/_relay/v1/me", "", `{"displayName":"Alice","bio":"Runs the reading group."}`, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("profile update = %d %s, want 200", rec.Code, rec.Body)
	}
	var updated meReply
	if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
		t.Fatal(err)
	}
	if updated.Member.DisplayName != "Alice" || updated.Member.Bio != "Runs the reading group." {
		t.Fatalf("updated member = %+v", updated.Member)
	}

	// An empty display name resets to the invitation label.
	rec = doMember(t, h, http.MethodPatch, "/_relay/v1/me", "", `{"displayName":"","bio":""}`, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("profile reset = %d %s, want 200", rec.Code, rec.Body)
	}
	var reset meReply
	if err := json.Unmarshal(rec.Body.Bytes(), &reset); err != nil {
		t.Fatal(err)
	}
	if reset.Member.DisplayName != "alice" || reset.Member.Bio != "" {
		t.Fatalf("reset member = %+v", reset.Member)
	}
}

// TestSessionSignOutClearsCookie pins the clear-cookie response.
func TestSessionSignOutClearsCookie(t *testing.T) {
	ks, err := serve.NewKeyStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	h := serve.NewSession("mynode", ks)
	rec := doMember(t, h, http.MethodDelete, "/_relay/v1/session", "", "", "tok")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("sign out = %d, want 204", rec.Code)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].MaxAge >= 0 {
		t.Fatalf("clear cookie = %+v, want MaxAge < 0", cookies)
	}
}

// TestSessionRejectsCrossOriginWrites pins the Origin check.
func TestSessionRejectsCrossOriginWrites(t *testing.T) {
	ks, err := serve.NewKeyStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	h := serve.NewSession("mynode", ks)
	req := httptest.NewRequest(http.MethodPost, "/_relay/v1/session", strings.NewReader(`{"token":"x"}`))
	req.Host = "relay.example.org"
	req.Header.Set("Origin", "https://evil.example")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("cross-origin session = %d, want 403", rec.Code)
	}
}

// TestUploadAcceptsMemberCookie pins that the browser session can publish
// without pasting the key again.
func TestUploadAcceptsMemberCookie(t *testing.T) {
	data := t.TempDir()
	ks, err := serve.NewKeyStore(data)
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}
	up := serve.NewUploader("mynode", data, ks, serve.NewVersions(data), nil, serve.MaxDeployBytes, nil)
	req := httptest.NewRequest(http.MethodPost, "/_relay/v1/sites/docs/deploy", strings.NewReader(string(tarSite(t, map[string]string{"index.html": "hi"}))))
	req.Header.Set("Content-Type", "application/x-tar")
	req.AddCookie(&http.Cookie{Name: "relay_member", Value: token})
	rec := httptest.NewRecorder()
	up.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("cookie deploy = %d %s, want 200", rec.Code, rec.Body)
	}
}

// TestInvitesAcceptMemberCookie pins that a signed-in member can invite from
// the browser without exposing the token to JavaScript.
func TestInvitesAcceptMemberCookie(t *testing.T) {
	ks, err := serve.NewKeyStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/_relay/v1/invites", strings.NewReader(`{"name":"bob"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "relay_member", Value: token})
	rec := httptest.NewRecorder()
	serve.NewInvites(ks).ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("cookie invite = %d %s, want 201", rec.Code, rec.Body)
	}
}

// TestMutatingSurfacesRejectCrossOrigin pins the Origin gate on the cookie-
// authenticated deploy and invite endpoints.
func TestMutatingSurfacesRejectCrossOrigin(t *testing.T) {
	data := t.TempDir()
	ks, err := serve.NewKeyStore(data)
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}
	up := serve.NewUploader("mynode", data, ks, serve.NewVersions(data), nil, serve.MaxDeployBytes, nil)
	inv := serve.NewInvites(ks)

	for name, h := range map[string]http.Handler{"deploy": up, "invite": inv} {
		var req *http.Request
		if name == "deploy" {
			req = httptest.NewRequest(http.MethodPost, "/_relay/v1/sites/docs/deploy", strings.NewReader(string(tarSite(t, map[string]string{"index.html": "hi"}))))
			req.Header.Set("Content-Type", "application/x-tar")
		} else {
			req = httptest.NewRequest(http.MethodPost, "/_relay/v1/invites", strings.NewReader(`{"name":"bob"}`))
			req.Header.Set("Content-Type", "application/json")
		}
		req.Host = "relay.example.org"
		req.Header.Set("Origin", "https://evil.example")
		req.AddCookie(&http.Cookie{Name: "relay_member", Value: token})
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("%s cross-origin = %d, want 403", name, rec.Code)
		}
	}
}
