package serve

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// Admin is the operator-facing control surface: list, create, revoke, and
// rotate per-user deploy keys, and view sites by owner. Every request is
// authenticated with the admin token (the node's master credential).
type Admin struct {
	token [sha256.Size]byte
	ks    *KeyStore
}

// NewAdmin returns an admin handler that authorizes with token and operates
// on ks.
func NewAdmin(token string, ks *KeyStore) *Admin {
	return &Admin{token: sha256.Sum256([]byte(token)), ks: ks}
}

// ServeHTTP routes the admin surface under /_relay/v1/admin:
//
//	GET  /_relay/v1/admin                        overview (keys + sites)
//	POST /_relay/v1/admin/keys                   create a key   { "user": "…" }
//	POST /_relay/v1/admin/keys/{id}/revoke       revoke a key
//	POST /_relay/v1/admin/keys/{id}/rotate       rotate a key
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
	default:
		rest, ok := strings.CutPrefix(p, "/_relay/v1/admin/keys/")
		if !ok || r.Method != http.MethodPost {
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
