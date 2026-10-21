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
	"sync"

	"github.com/nemoNoboru/relay/internal/domain"
	"github.com/nemoNoboru/relay/internal/social"
)

// DefaultAdminToken is the credential a node comes up on when the operator
// configured none, so running a node takes no more than one command. The
// console refuses to do anything else until it has been changed.
const DefaultAdminToken = "relay"

// minAdminToken is the length floor on a replacement admin token. It sits
// above len(DefaultAdminToken) so the default cannot be chosen again.
const minAdminToken = 8

// Admin is the operator-facing control surface: list, create, revoke, and
// rotate per-user deploy keys; view sites by owner; and list or roll back a
// site's versions. Every request is authenticated with the admin token (the
// node's master credential).
type Admin struct {
	mu       sync.RWMutex
	token    [sha256.Size]byte
	ks       *KeyStore
	versions *Versions
	posts    *social.Store
	comments *social.CommentStore
	reload   func() error
	reindex  func(domain.Name) error
	save     func(string) error
}

// NewAdmin returns an admin handler that authorizes with token, operates on
// ks, versions, posts, and comments, calls reload and reindex after a
// rollback or delete, and calls save to persist a changed admin token.
// posts, comments, reload, reindex, and save may be nil.
func NewAdmin(token string, ks *KeyStore, versions *Versions, posts *social.Store, comments *social.CommentStore, reload func() error, reindex func(domain.Name) error, save func(string) error) *Admin {
	return &Admin{
		token:    sha256.Sum256([]byte(token)),
		ks:       ks,
		versions: versions,
		posts:    posts,
		comments: comments,
		reload:   reload,
		reindex:  reindex,
		save:     save,
	}
}

// ServeHTTP routes the admin surface under /_relay/v1/admin:
//
//	GET  /_relay/v1/admin                          overview (keys + sites + comments)
//	POST /_relay/v1/admin/token                    change the admin token { "token": "…" }
//	POST /_relay/v1/admin/keys                     create a key   { "user": "…" }
//	POST /_relay/v1/admin/keys/{id}/revoke         revoke a key
//	POST /_relay/v1/admin/keys/{id}/rotate         rotate a key
//	GET  /_relay/v1/admin/sites/{site}/versions    list rollback snapshots
//	POST /_relay/v1/admin/sites/{site}/rollback    roll back  { "version": "…" }
//	DELETE /_relay/v1/admin/comments/{id}          delete any comment
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
		ov := overview{Overview: a.ks.Overview(), Comments: []CommentView{}, MustChangeToken: a.mustChangeToken()}
		if a.comments != nil {
			ov.Comments = a.recentComments()
		}
		writeJSON(w, http.StatusOK, ov)
	case r.Method == http.MethodPost && p == "/_relay/v1/admin/token":
		a.changeToken(w, r)
	case r.Method == http.MethodPost && p == "/_relay/v1/admin/keys":
		a.createKey(w, r)
	case strings.HasPrefix(p, "/_relay/v1/admin/sites/"):
		a.serveSite(w, r, strings.TrimPrefix(p, "/_relay/v1/admin/sites/"))
	case strings.HasPrefix(p, "/_relay/v1/admin/keys/"):
		a.serveKey(w, r, strings.TrimPrefix(p, "/_relay/v1/admin/keys/"))
	case r.Method == http.MethodDelete && strings.HasPrefix(p, "/_relay/v1/admin/comments/"):
		a.deleteComment(w, strings.TrimPrefix(p, "/_relay/v1/admin/comments/"))
	case r.Method == http.MethodDelete && strings.HasPrefix(p, "/_relay/v1/admin/posts/"):
		a.deletePost(w, r, strings.TrimPrefix(p, "/_relay/v1/admin/posts/"))
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

// overview is the admin read view: the key store's plus whether the node is
// still on the default credential.
type overview struct {
	Overview
	Comments        []CommentView `json:"comments"`
	MustChangeToken bool          `json:"mustChangeToken"`
}

// recentComments returns the newest comments for the host console, capped so
// the overview stays small.
func (a *Admin) recentComments() []CommentView {
	list := a.comments.ListAll()
	if len(list) > 50 {
		list = list[:50]
	}
	out := make([]CommentView, 0, len(list))
	for _, c := range list {
		out = append(out, commentView(a.ks, c, ""))
	}
	return out
}

func (a *Admin) deleteComment(w http.ResponseWriter, id string) {
	if a.comments == nil {
		http.NotFound(w, nil)
		return
	}
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "not_found", "no comment")
		return
	}
	if _, ok, err := a.comments.Delete(id); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
	} else if !ok {
		writeError(w, http.StatusNotFound, "not_found", "no comment "+id)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

// deletePost removes a site's files, snapshots, index entry, comments, and
// ownership record, then reloads the serving registry so the site is gone.
func (a *Admin) deletePost(w http.ResponseWriter, r *http.Request, site string) {
	name, err := domain.ParseName(site)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_name", err.Error())
		return
	}
	if err := a.versions.Remove(name); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	if a.posts != nil {
		if _, err := a.posts.Delete(string(name)); err != nil {
			writeError(w, http.StatusInternalServerError, "internal", err.Error())
			return
		}
	}
	if a.comments != nil {
		if _, err := a.comments.DeletePost(string(name)); err != nil {
			// The site and index are already gone; only the comment sweep
			// failed. Report it, do not fail the delete.
			fmt.Fprintf(os.Stderr, "serve: delete comments for %s: %v\n", name, err)
		}
	}
	if err := a.ks.ForgetSite(string(name)); err != nil {
		fmt.Fprintf(os.Stderr, "serve: forget site %s: %v\n", name, err)
	}
	if a.reload != nil {
		if err := a.reload(); err != nil {
			fmt.Fprintf(os.Stderr, "serve: reload after delete %s: %v\n", name, err)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"site": string(name)})
}

// authorized reports whether r carries the admin token as a bearer token.
func (a *Admin) authorized(r *http.Request) bool {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || token == "" {
		return false
	}
	got := sha256.Sum256([]byte(token))
	a.mu.RLock()
	defer a.mu.RUnlock()
	return subtle.ConstantTimeCompare(got[:], a.token[:]) == 1
}

// mustChangeToken reports whether the node is still on DefaultAdminToken.
func (a *Admin) mustChangeToken() bool {
	want := sha256.Sum256([]byte(DefaultAdminToken))
	a.mu.RLock()
	defer a.mu.RUnlock()
	return subtle.ConstantTimeCompare(want[:], a.token[:]) == 1
}

// changeToken replaces the node's admin credential. The caller is already
// authenticated with the current token, so this is the first thing an
// operator does on a node that came up on the default.
func (a *Admin) changeToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "body must be JSON: "+err.Error())
		return
	}
	// The length floor also rules out keeping DefaultAdminToken, which is
	// deliberately shorter than anything acceptable here.
	token := strings.TrimSpace(req.Token)
	if len(token) < minAdminToken {
		writeError(w, http.StatusBadRequest, "bad_request",
			fmt.Sprintf("the new token must be at least %d characters", minAdminToken))
		return
	}
	if a.save != nil {
		if err := a.save(token); err != nil {
			writeError(w, http.StatusInternalServerError, "internal", err.Error())
			return
		}
	}
	a.mu.Lock()
	a.token = sha256.Sum256([]byte(token))
	a.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{"changed": true})
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
	if a.reindex != nil {
		if err := a.reindex(name); err != nil {
			// The rollback is on disk; the index can be repaired with relay index.
			fmt.Fprintf(os.Stderr, "serve: reindex %s after rollback: %v\n", name, err)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"site": string(name), "version": id})
}
