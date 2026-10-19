package serve

import (
	"net/http"
	"strings"
)

// PostSurface routes /_relay/v1/posts/… between the post wrapper metadata
// and the comment subresource. Keeping the dispatch here lets node.go mount
// one prefix without duplicating path rules.
type PostSurface struct {
	posts    http.Handler
	comments http.Handler
}

// NewPostSurface returns the dispatcher for the posts prefix.
func NewPostSurface(posts, comments http.Handler) *PostSurface {
	return &PostSurface{posts: posts, comments: comments}
}

func (s *PostSurface) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")
	rest := strings.TrimPrefix(path, "/_relay/v1/posts/")
	if rest != "comments" && strings.HasSuffix(rest, "/comments") {
		s.comments.ServeHTTP(w, r)
		return
	}
	s.posts.ServeHTTP(w, r)
}
