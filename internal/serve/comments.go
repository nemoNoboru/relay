package serve

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/nemoNoboru/relay/internal/social"
)

// commentMinInterval throttles one member's comment creation. It is an
// in-memory guard, not a security boundary; the node is a small group.
const commentMinInterval = 2 * time.Second

// CommentView is a comment plus the author and the sanitized HTML body.
type CommentView struct {
	ID        string    `json:"id"`
	Post      string    `json:"post"`
	Member    string    `json:"member"`
	Author    Author    `json:"author"`
	Body      string    `json:"body"`
	HTML      string    `json:"html"`
	CreatedAt time.Time `json:"createdAt"`
	Mine      bool      `json:"mine,omitempty"`
}

// Comments serves the member comment API:
//
//	GET  /_relay/v1/posts/<site>/comments
//	POST /_relay/v1/posts/<site>/comments
//	DELETE /_relay/v1/comments/<id>          (own comment)
type Comments struct {
	ks       *KeyStore
	posts    *social.Store
	comments *social.CommentStore

	mu   sync.Mutex
	last map[string]time.Time
}

// NewComments returns the comment handler for a node.
func NewComments(ks *KeyStore, posts *social.Store, comments *social.CommentStore) *Comments {
	return &Comments{ks: ks, posts: posts, comments: comments, last: map[string]time.Time{}}
}

func (c *Comments) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")
	switch {
	case strings.HasPrefix(path, "/_relay/v1/posts/") && strings.HasSuffix(path, "/comments"):
		site := strings.TrimSuffix(strings.TrimPrefix(path, "/_relay/v1/posts/"), "/comments")
		c.onPost(w, r, site)
	case strings.HasPrefix(path, "/_relay/v1/comments/") && r.Method == http.MethodDelete:
		c.deleteOwn(w, r, strings.TrimPrefix(path, "/_relay/v1/comments/"))
	default:
		http.NotFound(w, r)
	}
}

func (c *Comments) onPost(w http.ResponseWriter, r *http.Request, site string) {
	member, _, ok := memberFromRequest(c.ks, r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "session_required", "open your invitation link to sign in")
		return
	}
	if site == "" || strings.Contains(site, "/") {
		writeError(w, http.StatusNotFound, "not_found", "no post")
		return
	}
	if _, ok := c.posts.Get(site); !ok {
		writeError(w, http.StatusNotFound, "not_found", "no post "+site)
		return
	}
	switch r.Method {
	case http.MethodGet:
		c.list(w, member, site)
	case http.MethodPost:
		if !sameOrigin(r) {
			writeError(w, http.StatusForbidden, "bad_origin", "request origin does not match this node")
			return
		}
		c.add(w, r, member, site)
	default:
		http.NotFound(w, r)
	}
}

func (c *Comments) list(w http.ResponseWriter, member Member, site string) {
	list := c.comments.List(site)
	out := make([]CommentView, 0, len(list))
	for _, comment := range list {
		out = append(out, commentView(c.ks, comment, member.ID))
	}
	writeJSON(w, http.StatusOK, map[string]any{"comments": out})
}

func (c *Comments) add(w http.ResponseWriter, r *http.Request, member Member, site string) {
	r.Body = http.MaxBytesReader(w, r.Body, 8<<10)
	var req struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "body must be JSON: "+err.Error())
		return
	}
	c.mu.Lock()
	if last, ok := c.last[member.ID]; ok && time.Since(last) < commentMinInterval {
		c.mu.Unlock()
		writeError(w, http.StatusTooManyRequests, "rate_limited", "wait a moment before commenting again")
		return
	}
	c.mu.Unlock()
	comment, err := c.comments.Add(social.Comment{Post: site, Member: member.ID, Body: req.Body})
	if errors.Is(err, social.ErrInvalidComment) {
		writeError(w, http.StatusBadRequest, "bad_comment", err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	c.mu.Lock()
	c.last[member.ID] = time.Now()
	c.mu.Unlock()
	writeJSON(w, http.StatusCreated, map[string]any{"comment": commentView(c.ks, comment, member.ID)})
}

func (c *Comments) deleteOwn(w http.ResponseWriter, r *http.Request, id string) {
	member, _, ok := memberFromRequest(c.ks, r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "session_required", "open your invitation link to sign in")
		return
	}
	if !sameOrigin(r) {
		writeError(w, http.StatusForbidden, "bad_origin", "request origin does not match this node")
		return
	}
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "not_found", "no comment")
		return
	}
	comment, ok := c.comments.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "no comment "+id)
		return
	}
	if comment.Member != member.ID {
		writeError(w, http.StatusForbidden, "forbidden", "you can delete only your own comments")
		return
	}
	if _, _, err := c.comments.Delete(id); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// commentView enriches a stored comment with its author and sanitized render.
func commentView(ks *KeyStore, comment social.Comment, currentMember string) CommentView {
	author := Author{ID: comment.Member, DisplayName: "former member"}
	if m, ok := ks.MemberByID(comment.Member); ok {
		author = Author{ID: m.ID, DisplayName: m.DisplayName, Revoked: m.Revoked}
	}
	return CommentView{
		ID: comment.ID, Post: comment.Post, Member: comment.Member,
		Author: author, Body: comment.Body, HTML: social.RenderComment(comment.Body),
		CreatedAt: comment.CreatedAt, Mine: comment.Member == currentMember,
	}
}
