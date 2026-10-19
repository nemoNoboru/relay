// Package social holds the small social layer around Relay's static sites:
// posts, the post index, and (later) comments and templates. It is pure Go
// model and file persistence; it does not import net/http.
package social

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Post is one published microwebsite in the node's index.
type Post struct {
	ID          string    `json:"id"`
	Site        string    `json:"site"`
	Member      string    `json:"member"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary,omitempty"`
	Template    string    `json:"template"`
	Image       string    `json:"image,omitempty"`
	PublishedAt time.Time `json:"publishedAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Hidden      bool      `json:"hidden,omitempty"`
}

// ErrInvalidPost is returned when a post cannot be indexed.
var ErrInvalidPost = errors.New("invalid post: site and member are required")

// Normalize trims the post's text fields and fills defaults.
func (p *Post) Normalize() {
	p.Site = strings.TrimSpace(p.Site)
	p.Member = strings.TrimSpace(p.Member)
	p.Title = strings.TrimSpace(p.Title)
	p.Summary = strings.TrimSpace(p.Summary)
	p.Template = strings.TrimSpace(p.Template)
	if p.Title == "" {
		p.Title = TitleFromSite(p.Site)
	}
	p.Title = truncate(p.Title, 160)
	p.Summary = truncate(p.Summary, 400)
	p.Image = strings.TrimPrefix(strings.TrimSpace(p.Image), "/")
	if p.Image != "" {
		clean := path.Clean(p.Image)
		if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
			p.Image = ""
		} else {
			p.Image = clean
		}
	}
	if p.Template == "" {
		p.Template = "custom"
	}
}

// Validate checks the fields a post needs to be indexed.
func (p Post) Validate() error {
	if p.Site == "" || p.Member == "" {
		return ErrInvalidPost
	}
	return nil
}

// TitleFromSite turns "sunday-walk" into "Sunday walk".
func TitleFromSite(site string) string {
	words := strings.Fields(strings.ReplaceAll(site, "-", " "))
	for i, w := range words {
		r, size := utf8.DecodeRuneInString(w)
		if r == utf8.RuneError {
			continue
		}
		words[i] = string(unicode.ToUpper(r)) + w[size:]
	}
	return strings.Join(words, " ")
}

// NewID returns a fresh post id.
func NewID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate post id: %w", err)
	}
	return "post_" + hex.EncodeToString(b), nil
}

func truncate(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	r := []rune(s)
	return string(r[:max])
}
