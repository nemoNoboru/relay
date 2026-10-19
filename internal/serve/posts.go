package serve

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nemoNoboru/relay/internal/domain"
	"github.com/nemoNoboru/relay/internal/social"
)

// Author is the member summary attached to a post in the API.
type Author struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Revoked     bool   `json:"revoked,omitempty"`
}

// PostView is a post plus the author and the two URLs a reader needs.
type PostView struct {
	social.Post
	Author    Author        `json:"author"`
	URL       string        `json:"url"`
	PostURL   string        `json:"postUrl"`
	AuthorURL string        `json:"authorUrl"`
	ImageURL  string        `json:"imageUrl,omitempty"`
	Comments  []CommentView `json:"comments"`
}

// Posts serves the member-only post index surface:
//
//	GET /_relay/v1/feed
//	GET /_relay/v1/posts/<site>
//	GET /_relay/v1/people/<member-id>
//
// Comments and template publishing are later phases.
type Posts struct {
	node     domain.Name
	ks       *KeyStore
	posts    *social.Store
	comments *social.CommentStore
}

// NewPosts returns the post index handler for a node.
func NewPosts(node domain.Name, ks *KeyStore, posts *social.Store, comments *social.CommentStore) *Posts {
	return &Posts{node: node, ks: ks, posts: posts, comments: comments}
}

// ServeHTTP routes the post index surface. Everything is member-only.
func (p *Posts) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	member, _, ok := memberFromRequest(p.ks, r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "session_required", "open your invitation link to sign in")
		return
	}
	path := strings.TrimSuffix(r.URL.Path, "/")
	switch {
	case path == "/_relay/v1/feed":
		p.feed(w, r, member)
	case strings.HasPrefix(path, "/_relay/v1/posts/"):
		p.post(w, member, strings.TrimPrefix(path, "/_relay/v1/posts/"))
	case strings.HasPrefix(path, "/_relay/v1/people/"):
		p.profile(w, r, member, strings.TrimPrefix(path, "/_relay/v1/people/"))
	default:
		http.NotFound(w, r)
	}
}

func (p *Posts) feed(w http.ResponseWriter, r *http.Request, member Member) {
	page, next := p.page(r, p.posts.List(), func(social.Post) bool { return true })
	out := p.views(page, member.ID)
	reply := map[string]any{"posts": out}
	if next != "" {
		reply["next"] = next
	}
	writeJSON(w, http.StatusOK, reply)
}

func (p *Posts) post(w http.ResponseWriter, member Member, site string) {
	if site == "" || strings.Contains(site, "/") {
		writeError(w, http.StatusNotFound, "not_found", "no post")
		return
	}
	post, ok := p.posts.Get(site)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "no post "+site)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"post": p.views([]social.Post{post}, member.ID)[0]})
}

func (p *Posts) profile(w http.ResponseWriter, r *http.Request, viewer Member, memberID string) {
	if memberID == "" || strings.Contains(memberID, "/") {
		writeError(w, http.StatusNotFound, "not_found", "no member")
		return
	}
	member, ok := p.ks.MemberByID(memberID)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "no member "+memberID)
		return
	}
	all := p.posts.List()
	total := 0
	for _, post := range all {
		if post.Member == memberID && !post.Hidden {
			total++
		}
	}
	page, next := p.page(r, all, func(post social.Post) bool { return post.Member == memberID })
	out := p.views(page, viewer.ID)
	reply := map[string]any{"member": member, "posts": out, "total": total}
	if next != "" {
		reply["next"] = next
	}
	writeJSON(w, http.StatusOK, reply)
}

func (p *Posts) views(posts []social.Post, currentMember string) []PostView {
	out := make([]PostView, 0, len(posts))
	for _, post := range posts {
		out = append(out, postView(p.node, p.ks, p.comments, post, currentMember))
	}
	return out
}

func postView(node domain.Name, ks *KeyStore, comments *social.CommentStore, post social.Post, currentMember string) PostView {
	author := Author{ID: post.Member, DisplayName: "former member"}
	if m, ok := ks.MemberByID(post.Member); ok {
		author = Author{ID: m.ID, DisplayName: m.DisplayName, Revoked: m.Revoked}
	}
	imageURL := ""
	if post.Image != "" {
		imageURL = "/" + string(node) + "/" + post.Site + "/" + post.Image
	}
	views := make([]CommentView, 0)
	if comments != nil {
		for _, c := range comments.List(post.Site) {
			views = append(views, commentView(ks, c, currentMember))
		}
	}
	return PostView{
		Post:      post,
		Author:    author,
		URL:       "/" + string(node) + "/" + post.Site + "/",
		PostURL:   "/_relay/posts/" + post.Site,
		AuthorURL: "/_relay/people/" + post.Member,
		ImageURL:  imageURL,
		Comments:  views,
	}
}

// pageCursor is the position after which the next page starts. It mirrors the
// store's ordering: publishedAt descending, then site ascending.
type pageCursor struct {
	At   time.Time
	Site string
}

func (p *Posts) page(r *http.Request, all []social.Post, keep func(social.Post) bool) ([]social.Post, string) {
	limit := 10
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = min(n, 50)
		}
	}
	cursor, hasCursor := decodeCursor(r.URL.Query().Get("cursor"))
	out := make([]social.Post, 0, limit+1)
	for _, post := range all {
		if post.Hidden || !keep(post) {
			continue
		}
		if hasCursor && !afterCursor(post, cursor) {
			continue
		}
		out = append(out, post)
	}
	next := ""
	if len(out) > limit {
		last := out[limit-1]
		next = encodeCursor(pageCursor{At: last.PublishedAt, Site: last.Site})
		out = out[:limit]
	}
	return out, next
}

func afterCursor(post social.Post, c pageCursor) bool {
	if post.PublishedAt.Before(c.At) {
		return true
	}
	return post.PublishedAt.Equal(c.At) && post.Site > c.Site
}

func encodeCursor(c pageCursor) string {
	return base64.RawURLEncoding.EncodeToString([]byte(c.At.UTC().Format(time.RFC3339Nano) + "\x00" + c.Site))
}

func decodeCursor(raw string) (pageCursor, bool) {
	if raw == "" {
		return pageCursor{}, false
	}
	b, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return pageCursor{}, false
	}
	parts := strings.SplitN(string(b), "\x00", 2)
	if len(parts) != 2 {
		return pageCursor{}, false
	}
	at, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil || parts[1] == "" {
		return pageCursor{}, false
	}
	return pageCursor{At: at, Site: parts[1]}, true
}
