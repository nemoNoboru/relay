package social

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// MaxCommentLength is the maximum comment body length in runes.
const MaxCommentLength = 2000

// Comment is one member's reply under a post.
type Comment struct {
	ID        string    `json:"id"`
	Post      string    `json:"post"`
	Member    string    `json:"member"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}

// ErrInvalidComment is returned for an empty, oversized, or unaddressed
// comment.
var ErrInvalidComment = errors.New("invalid comment: post, member, and 1-2000 characters are required")

// Validate checks the fields a comment needs.
func (c Comment) Validate() error {
	if c.Post == "" || c.Member == "" {
		return ErrInvalidComment
	}
	body := strings.TrimSpace(c.Body)
	if body == "" || utf8.RuneCountInString(body) > MaxCommentLength {
		return ErrInvalidComment
	}
	return nil
}

// NewCommentID returns a fresh comment id.
func NewCommentID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate comment id: %w", err)
	}
	return "cmt_" + hex.EncodeToString(b), nil
}
