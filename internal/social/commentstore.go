package social

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/nemoNoboru/relay/internal/persist"
)

// CommentStore persists comments in <dataDir>/comments.json. Writes are
// serialized and atomic; reads take a shared lock.
type CommentStore struct {
	path string
	mu   sync.RWMutex
	file commentFile
}

type commentFile struct {
	Comments []Comment `json:"comments"`
}

// NewCommentStore loads the comment index, creating an empty one if absent.
func NewCommentStore(dataDir string) (*CommentStore, error) {
	s := &CommentStore{path: filepath.Join(dataDir, "comments.json")}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *CommentStore) load() error {
	b, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return s.persist()
	}
	if err != nil {
		return fmt.Errorf("read comments: %w", err)
	}
	var f commentFile
	if err := json.Unmarshal(b, &f); err != nil {
		return fmt.Errorf("parse comments: %w", err)
	}
	s.file = f
	return nil
}

func (s *CommentStore) persist() error {
	return persist.WriteJSON(s.path, s.file, 0o644)
}

// Add appends a comment, assigning an id and timestamp.
func (s *CommentStore) Add(c Comment) (Comment, error) {
	c.Body = trimComment(c.Body)
	if err := c.Validate(); err != nil {
		return Comment{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if c.ID == "" {
		id, err := NewCommentID()
		if err != nil {
			return Comment{}, err
		}
		c.ID = id
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now().UTC()
	}
	s.file.Comments = append(s.file.Comments, c)
	if err := s.persist(); err != nil {
		s.file.Comments = s.file.Comments[:len(s.file.Comments)-1]
		return Comment{}, err
	}
	return c, nil
}

// List returns one post's comments oldest first.
func (s *CommentStore) List(post string) []Comment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Comment, 0)
	for _, c := range s.file.Comments {
		if c.Post == post {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

// ListAll returns every comment newest first.
func (s *CommentStore) ListAll() []Comment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Comment, len(s.file.Comments))
	copy(out, s.file.Comments)
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

// Get returns a comment by id.
func (s *CommentStore) Get(id string) (Comment, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.file.Comments {
		if c.ID == id {
			return c, true
		}
	}
	return Comment{}, false
}

// Delete removes a comment by id.
func (s *CommentStore) Delete(id string) (Comment, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.file.Comments {
		if s.file.Comments[i].ID != id {
			continue
		}
		removed := s.file.Comments[i]
		before := make([]Comment, len(s.file.Comments))
		copy(before, s.file.Comments)
		s.file.Comments = append(s.file.Comments[:i], s.file.Comments[i+1:]...)
		if err := s.persist(); err != nil {
			s.file.Comments = before
			return Comment{}, false, err
		}
		return removed, true, nil
	}
	return Comment{}, false, nil
}

// DeletePost removes every comment on a post, returning the number removed.
// A post with no comments is a no-op.
func (s *CommentStore) DeletePost(post string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	kept := make([]Comment, 0, len(s.file.Comments))
	removed := 0
	for _, c := range s.file.Comments {
		if c.Post == post {
			removed++
			continue
		}
		kept = append(kept, c)
	}
	if removed == 0 {
		return 0, nil
	}
	before := s.file.Comments
	s.file.Comments = kept
	if err := s.persist(); err != nil {
		s.file.Comments = before
		return 0, err
	}
	return removed, nil
}

func trimComment(body string) string {
	return strings.TrimSpace(body)
}
