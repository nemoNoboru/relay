package serve_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nemoNoboru/relay/internal/domain"
	"github.com/nemoNoboru/relay/internal/serve"
)

const adminToken = "admin-secret"

func newAdmin(t *testing.T) (*serve.Admin, *serve.KeyStore) {
	t.Helper()
	data := t.TempDir()
	ks, err := serve.NewKeyStore(data)
	if err != nil {
		t.Fatal(err)
	}
	return serve.NewAdmin(adminToken, ks, serve.NewVersions(data), nil, nil, nil, nil), ks
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

// TestAdminChangeToken pins the first-login flow: a node left on the
// default token says so, and changing the token swaps the credential and
// saves it.
func TestAdminChangeToken(t *testing.T) {
	data := t.TempDir()
	ks, err := serve.NewKeyStore(data)
	if err != nil {
		t.Fatal(err)
	}
	saved := ""
	a := serve.NewAdmin(serve.DefaultAdminToken, ks, serve.NewVersions(data), nil, nil, nil, func(s string) error {
		saved = s
		return nil
	})
	if !mustChange(t, a, serve.DefaultAdminToken) {
		t.Fatal("a node on the default token does not ask for a change")
	}

	rec := doJSON(t, a, http.MethodPost, "/_relay/v1/admin/token", serve.DefaultAdminToken, `{"token":"chosen-one"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("change = %d %s, want 200", rec.Code, rec.Body)
	}
	if saved != "chosen-one" {
		t.Fatalf("saved = %q, want chosen-one", saved)
	}
	if rec := doJSON(t, a, http.MethodGet, "/_relay/v1/admin", serve.DefaultAdminToken, ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("old token still works = %d, want 401", rec.Code)
	}
	if mustChange(t, a, "chosen-one") {
		t.Fatal("still asking for a change after the token was changed")
	}
}

// TestAdminRejectsDefaultAsNewToken pins that the change must be a real
// change, not the default again.
func TestAdminRejectsDefaultAsNewToken(t *testing.T) {
	a, _ := newAdmin(t)
	rec := doJSON(t, a, http.MethodPost, "/_relay/v1/admin/token", adminToken, `{"token":"`+serve.DefaultAdminToken+`"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

// mustChange reports the overview's mustChangeToken flag.
func mustChange(t *testing.T, a *serve.Admin, token string) bool {
	t.Helper()
	rec := doJSON(t, a, http.MethodGet, "/_relay/v1/admin", token, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("overview = %d %s", rec.Code, rec.Body)
	}
	var ov struct {
		MustChangeToken bool `json:"mustChangeToken"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &ov); err != nil {
		t.Fatal(err)
	}
	return ov.MustChangeToken
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

// newAdminNode wires a registry, uploader, and admin over one data dir, and
// returns them with a minted deploy key token for "tester".
func newAdminNode(t *testing.T) (*serve.Admin, *serve.Registry, http.Handler, string) {
	t.Helper()
	data := t.TempDir()
	ks, err := serve.NewKeyStore(data)
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := ks.Create("tester")
	if err != nil {
		t.Fatal(err)
	}
	versions := serve.NewVersions(data)
	reg, _, err := serve.NewRegistry("mynode", data)
	if err != nil {
		t.Fatal(err)
	}
	reload := func() error {
		_, err := reg.Reload()
		return err
	}
	up := serve.NewUploader("mynode", data, ks, versions, nil, serve.MaxDeployBytes, reload)
	return serve.NewAdmin(adminToken, ks, versions, nil, reload, nil, nil), reg, up, token
}

func deploySite(t *testing.T, up http.Handler, token, site, content string) {
	t.Helper()
	rec := doDeploy(t, up, "/_relay/v1/sites/"+site+"/deploy", token, tarSite(t, map[string]string{"index.html": content}))
	if rec.Code != http.StatusOK {
		t.Fatalf("deploy %s: status %d body %s", site, rec.Code, rec.Body.String())
	}
}

func serveSite(t *testing.T, reg *serve.Registry, path string) string {
	t.Helper()
	srv := httptest.NewServer(reg.Handler())
	defer srv.Close()
	res, err := http.Get(srv.URL + path)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestAdminListVersions(t *testing.T) {
	admin, _, up, token := newAdminNode(t)

	rec := doJSON(t, admin, http.MethodGet, "/_relay/v1/admin/sites/blog/versions", adminToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("empty list status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"versions":[]`) {
		t.Fatalf("empty versions body = %s", rec.Body.String())
	}

	deploySite(t, up, token, "blog", "v1")
	deploySite(t, up, token, "blog", "v2")

	rec = doJSON(t, admin, http.MethodGet, "/_relay/v1/admin/sites/blog/versions", adminToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, body %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Site     string `json:"site"`
		Versions []struct {
			ID string `json:"id"`
		} `json:"versions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Site != "blog" || len(body.Versions) != 1 {
		t.Fatalf("list = %+v, want one version for blog", body)
	}
}

func TestAdminRollback(t *testing.T) {
	admin, reg, up, token := newAdminNode(t)
	deploySite(t, up, token, "blog", "v1")
	deploySite(t, up, token, "blog", "v2")

	rec := doJSON(t, admin, http.MethodGet, "/_relay/v1/admin/sites/blog/versions", adminToken, "")
	var body struct {
		Versions []struct {
			ID string `json:"id"`
		} `json:"versions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Versions) != 1 {
		t.Fatalf("versions = %+v, want 1", body.Versions)
	}

	rec = doJSON(t, admin, http.MethodPost, "/_relay/v1/admin/sites/blog/rollback", adminToken, `{"version":"`+body.Versions[0].ID+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("rollback status = %d, body %s", rec.Code, rec.Body.String())
	}
	if got := serveSite(t, reg, "/mynode/blog/"); got != "v1" {
		t.Fatalf("served after rollback = %q, want v1", got)
	}

	if rec := doJSON(t, admin, http.MethodPost, "/_relay/v1/admin/sites/blog/rollback", adminToken, `{"version":"nope"}`); rec.Code != http.StatusNotFound {
		t.Fatalf("unknown version = %d, want 404", rec.Code)
	}
}

func TestAdminRollbackPrevious(t *testing.T) {
	admin, reg, up, token := newAdminNode(t)
	deploySite(t, up, token, "blog", "v1")
	deploySite(t, up, token, "blog", "v2")

	if rec := doJSON(t, admin, http.MethodPost, "/_relay/v1/admin/sites/blog/rollback", adminToken, ""); rec.Code != http.StatusOK {
		t.Fatalf("rollback previous status = %d, body %s", rec.Code, rec.Body.String())
	}
	if got := serveSite(t, reg, "/mynode/blog/"); got != "v1" {
		t.Fatalf("served after rollback = %q, want v1", got)
	}
}

// TestAdminRollbackReindexes pins that a rollback asks the post index to
// refresh, so the feed does not go stale.
func TestAdminRollbackReindexes(t *testing.T) {
	data := t.TempDir()
	ks, err := serve.NewKeyStore(data)
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := ks.Create("tester")
	if err != nil {
		t.Fatal(err)
	}
	versions := serve.NewVersions(data)
	reg, _, err := serve.NewRegistry("mynode", data)
	if err != nil {
		t.Fatal(err)
	}
	reload := func() error {
		_, err := reg.Reload()
		return err
	}
	up := serve.NewUploader("mynode", data, ks, versions, nil, serve.MaxDeployBytes, reload)
	deploySite(t, up, token, "blog", "v1")
	deploySite(t, up, token, "blog", "v2")

	var reindexed string
	admin := serve.NewAdmin(adminToken, ks, versions, nil, reload, func(name domain.Name) error {
		reindexed = string(name)
		return nil
	}, nil)
	if rec := doJSON(t, admin, http.MethodPost, "/_relay/v1/admin/sites/blog/rollback", adminToken, ""); rec.Code != http.StatusOK {
		t.Fatalf("rollback = %d %s", rec.Code, rec.Body.String())
	}
	if reindexed != "blog" {
		t.Fatalf("reindexed = %q, want blog", reindexed)
	}
}
