package serve

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/nemoNoboru/relay/internal/domain"
)

// Admin is the operator-facing control surface: list, create, revoke, and
// rotate per-user deploy keys; view sites by owner; and list or roll back a
// site's versions. Every request is authenticated with the admin token (the
// node's master credential).
type Admin struct {
	token    [sha256.Size]byte
	ks       *KeyStore
	versions *Versions
	reload   func() error
}

// NewAdmin returns an admin handler that authorizes with token, operates on
// ks and versions, and calls reload after a rollback.
func NewAdmin(token string, ks *KeyStore, versions *Versions, reload func() error) *Admin {
	return &Admin{
		token:    sha256.Sum256([]byte(token)),
		ks:       ks,
		versions: versions,
		reload:   reload,
	}
}

// ServeHTTP routes the admin surface under /_relay/v1/admin:
//
//	GET  /_relay/v1/admin                          overview (keys + sites)
//	POST /_relay/v1/admin/keys                     create a key   { "user": "…" }
//	POST /_relay/v1/admin/keys/{id}/revoke         revoke a key
//	POST /_relay/v1/admin/keys/{id}/rotate         rotate a key
//	GET  /_relay/v1/admin/sites/{site}/versions    list rollback snapshots
//	POST /_relay/v1/admin/sites/{site}/rollback    roll back  { "version": "…" }
//
// Everything else is 404; missing/wrong credentials are 401 bad_token.
func (a *Admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !a.authorized(r) {
		writeError(w, http.StatusUnauthorized, "bad_token", "invalid or missing admin token")
		return
	}
	p := strings.TrimSuffix(r.URL.Path, "/")
	switch {
	case r.Method == http.MethodGet && p == "/_relay/v1/admin":
		writeJSON(w, http.StatusOK, a.ks.Overview())
	case r.Method == http.MethodPost && p == "/_relay/v1/admin/keys":
		a.createKey(w, r)
	case strings.HasPrefix(p, "/_relay/v1/admin/sites/"):
		a.serveSite(w, r, strings.TrimPrefix(p, "/_relay/v1/admin/sites/"))
	case strings.HasPrefix(p, "/_relay/v1/admin/keys/"):
		a.serveKey(w, r, strings.TrimPrefix(p, "/_relay/v1/admin/keys/"))
	default:
		http.NotFound(w, r)
	}
}

func (a *Admin) serveKey(w http.ResponseWriter, r *http.Request, rest string) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	id, action, found := strings.Cut(rest, "/")
	if !found {
		http.NotFound(w, r)
		return
	}
	switch action {
	case "revoke":
		a.revokeKey(w, r, id)
	case "rotate":
		a.rotateKey(w, r, id)
	default:
		http.NotFound(w, r)
	}
}

func (a *Admin) serveSite(w http.ResponseWriter, r *http.Request, rest string) {
	site, action, found := strings.Cut(rest, "/")
	name, err := domain.ParseName(site)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_name", err.Error())
		return
	}
	switch {
	case !found && r.Method == http.MethodGet:
		a.listVersions(w, r, name)
	case action == "versions" && r.Method == http.MethodGet:
		a.listVersions(w, r, name)
	case action == "rollback" && r.Method == http.MethodPost:
		a.rollbackSite(w, r, name)
	default:
		http.NotFound(w, r)
	}
}

// authorized reports whether r carries the admin token as a bearer token.
func (a *Admin) authorized(r *http.Request) bool {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || token == "" {
		return false
	}
	got := sha256.Sum256([]byte(token))
	return subtle.ConstantTimeCompare(got[:], a.token[:]) == 1
}

func (a *Admin) createKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		User string `json:"user"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "body must be JSON: "+err.Error())
		return
	}
	token, key, err := a.ks.Create(strings.TrimSpace(req.User))
	if errors.Is(err, ErrBadUser) {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"key": key, "token": token})
}

func (a *Admin) revokeKey(w http.ResponseWriter, r *http.Request, id string) {
	key, found, err := a.ks.Revoke(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "not_found", "no deploy key "+id)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"key": key})
}

func (a *Admin) rotateKey(w http.ResponseWriter, r *http.Request, id string) {
	token, key, found, err := a.ks.Rotate(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "not_found", "no deploy key "+id)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"key": key, "token": token})
}

func (a *Admin) listVersions(w http.ResponseWriter, r *http.Request, name domain.Name) {
	vs, err := a.versions.List(name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"site": string(name), "versions": vs})
}

func (a *Admin) rollbackSite(w http.ResponseWriter, r *http.Request, name domain.Name) {
	var req struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "bad_request", "body must be JSON: "+err.Error())
		return
	}
	id := req.Version
	if id == "" {
		vs, err := a.versions.List(name)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal", err.Error())
			return
		}
		if len(vs) == 0 {
			writeError(w, http.StatusNotFound, "not_found", "no versions to roll back to")
			return
		}
		id = vs[0].ID
	}
	if _, err := a.versions.Rollback(name, id); err != nil {
		if errors.Is(err, ErrUnknownVersion) {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	if err := a.versions.Prune(name); err != nil {
		fmt.Fprintf(os.Stderr, "serve: prune versions of %s: %v\n", name, err)
	}
	if a.reload != nil {
		if err := a.reload(); err != nil {
			// The rollback is on disk; only the in-memory view failed to
			// refresh. Report it, do not fail the rollback.
			fmt.Fprintf(os.Stderr, "serve: reload sites after rollback: %v\n", err)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"site": string(name), "version": id})
}
