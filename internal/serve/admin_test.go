package serve_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nemoNoboru/relay/internal/serve"
)

const adminToken = "admin-secret"

func newAdmin(t *testing.T) (*serve.Admin, *serve.KeyStore) {
	t.Helper()
	ks, err := serve.NewKeyStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return serve.NewAdmin(adminToken, ks), ks
}

// doJSON issues a request to the admin handler with an optional bearer token
// and JSON body, returning the recorder.
func doJSON(t *testing.T, h http.Handler, method, path, token string, body string) *httptest.ResponseRecorder {
	t.Helper()
	var rdr *strings.Reader
	if body == "" {
		rdr = strings.NewReader("")
	} else {
		rdr = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, rdr)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// TestAdminRequiresToken pins that the whole admin surface is 401 without the
// admin token, and that the admin token does not authorize uploads' admin
// surface when wrong.
func TestAdminRequiresToken(t *testing.T) {
	a, _ := newAdmin(t)
	for _, req := range []struct{ method, path string }{
		{http.MethodGet, "/_relay/v1/admin"},
		{http.MethodPost, "/_relay/v1/admin/keys"},
		{http.MethodPost, "/_relay/v1/admin/keys/dk_1/revoke"},
		{http.MethodPost, "/_relay/v1/admin/keys/dk_1/rotate"},
	} {
		rec := doJSON(t, a, req.method, req.path, "", "")
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s without token = %d, want 401", req.method, req.path, rec.Code)
		}
		if rec := doJSON(t, a, req.method, req.path, "wrong", ""); rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s with wrong token = %d, want 401", req.method, req.path, rec.Code)
		}
	}
}

// TestAdminOverview pins the read view: keys (no hashes) and site ownership.
func TestAdminOverview(t *testing.T) {
	a, ks := newAdmin(t)
	if _, _, err := ks.Create("alice"); err != nil {
		t.Fatal(err)
	}
	token, _, err := ks.Create("bob")
	if err != nil {
		t.Fatal(err)
	}
	if err := ks.RecordSite("blog", "alice"); err != nil {
		t.Fatal(err)
	}
	_ = token

	rec := doJSON(t, a, http.MethodGet, "/_relay/v1/admin", adminToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body.String())
	}
	var ov struct {
		Keys  []map[string]any `json:"keys"`
		Sites []struct {
			Site string `json:"site"`
			User string `json:"user"`
		} `json:"sites"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &ov); err != nil {
		t.Fatal(err)
	}
	if len(ov.Keys) != 2 {
		t.Fatalf("keys = %d, want 2", len(ov.Keys))
	}
	for _, k := range ov.Keys {
		if _, hasHash := k["hash"]; hasHash {
			t.Fatalf("key %+v leaks its hash", k)
		}
		if tok, ok := k["token"].(string); !ok || tok == "" {
			t.Fatalf("key %+v is missing its token", k)
		}
	}
	if len(ov.Sites) != 1 || ov.Sites[0].Site != "blog" || ov.Sites[0].User != "alice" {
		t.Fatalf("sites = %+v, want blog->alice", ov.Sites)
	}
}

// TestAdminCreateKey pins the create endpoint: 201 with a one-time token,
// 400 on a bad user or bad JSON.
func TestAdminCreateKey(t *testing.T) {
	a, _ := newAdmin(t)

	rec := doJSON(t, a, http.MethodPost, "/_relay/v1/admin/keys", adminToken, `{"user":"alice"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body.String())
	}
	var created struct {
		Key struct {
			ID      string `json:"id"`
			User    string `json:"user"`
			Revoked bool   `json:"revoked"`
		} `json:"key"`
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Key.User != "alice" || created.Token == "" || created.Key.ID == "" {
		t.Fatalf("created = %+v", created)
	}

	if rec := doJSON(t, a, http.MethodPost, "/_relay/v1/admin/keys", adminToken, `{"user":"has space"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad user status = %d, want 400", rec.Code)
	}
	if rec := doJSON(t, a, http.MethodPost, "/_relay/v1/admin/keys", adminToken, `not json`); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad json status = %d, want 400", rec.Code)
	}
}

// TestAdminRevoke pins revoke (idempotent, 200) and unknown id (404).
func TestAdminRevoke(t *testing.T) {
	a, ks := newAdmin(t)
	_, key, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}

	rec := doJSON(t, a, http.MethodPost, "/_relay/v1/admin/keys/"+key.ID+"/revoke", adminToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body.String())
	}
	if rec := doJSON(t, a, http.MethodPost, "/_relay/v1/admin/keys/"+key.ID+"/revoke", adminToken, ""); rec.Code != http.StatusOK {
		t.Fatalf("idempotent revoke = %d, want 200", rec.Code)
	}
	if rec := doJSON(t, a, http.MethodPost, "/_relay/v1/admin/keys/dk_nope/revoke", adminToken, ""); rec.Code != http.StatusNotFound {
		t.Fatalf("unknown id = %d, want 404", rec.Code)
	}
}

// TestAdminRotate pins rotate: 201 with a fresh token, and the old key stops
// authenticating.
func TestAdminRotate(t *testing.T) {
	a, ks := newAdmin(t)
	oldToken, key, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}

	rec := doJSON(t, a, http.MethodPost, "/_relay/v1/admin/keys/"+key.ID+"/rotate", adminToken, "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body.String())
	}
	var rotated struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &rotated); err != nil {
		t.Fatal(err)
	}
	if rotated.Token == "" || rotated.Token == oldToken {
		t.Fatalf("rotated token = %q", rotated.Token)
	}
	if _, ok := ks.Authenticate(oldToken); ok {
		t.Fatalf("old token still authenticates after rotate")
	}
	if _, ok := ks.Authenticate(rotated.Token); !ok {
		t.Fatalf("new token does not authenticate")
	}
	if rec := doJSON(t, a, http.MethodPost, "/_relay/v1/admin/keys/dk_nope/rotate", adminToken, ""); rec.Code != http.StatusNotFound {
		t.Fatalf("unknown id = %d, want 404", rec.Code)
	}
}

// TestAdminUnknownRoute pins 404 outside the defined surface.
func TestAdminUnknownRoute(t *testing.T) {
	a, _ := newAdmin(t)
	for _, req := range []struct{ method, path string }{
		{http.MethodGet, "/_relay/v1/admin/keys"},
		{http.MethodPost, "/_relay/v1/admin"},
		{http.MethodPost, "/_relay/v1/admin/keys/dk_1/delete"},
		{http.MethodGet, "/_relay/v1/other"},
	} {
		if rec := doJSON(t, a, req.method, req.path, adminToken, ""); rec.Code != http.StatusNotFound {
			t.Fatalf("%s %s = %d, want 404", req.method, req.path, rec.Code)
		}
	}
}
