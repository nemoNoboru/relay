package social

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Store persists the post index in <dataDir>/posts.json. Writes are
// serialized and atomic; reads take a shared lock.
type Store struct {
	path string
	mu   sync.RWMutex
	file storeFile
}

type storeFile struct {
	Posts []Post `json:"posts"`
}

// NewStore loads the post index, creating an empty one if absent.
func NewStore(dataDir string) (*Store, error) {
	s := &Store{path: filepath.Join(dataDir, "posts.json")}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) load() error {
	b, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return s.persist()
	}
	if err != nil {
		return fmt.Errorf("read post index: %w", err)
	}
	var f storeFile
	if err := json.Unmarshal(b, &f); err != nil {
		return fmt.Errorf("parse post index: %w", err)
	}
	s.file = f
	return nil
}

// persist writes the index to a temp file and renames it into place.
// Callers hold the write lock.
func (s *Store) persist() error {
	b, err := json.MarshalIndent(s.file, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, s.path); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

// List returns a copy of the index, newest first.
func (s *Store) List() []Post {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Post, len(s.file.Posts))
	copy(out, s.file.Posts)
	sort.Slice(out, func(i, j int) bool {
		if out[i].PublishedAt.Equal(out[j].PublishedAt) {
			return out[i].Site < out[j].Site
		}
		return out[i].PublishedAt.After(out[j].PublishedAt)
	})
	return out
}

// Get returns the post for a site, if indexed.
func (s *Store) Get(site string) (Post, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.file.Posts {
		if p.Site == site {
			return p, true
		}
	}
	return Post{}, false
}

// Upsert inserts or updates a post by site. An existing post keeps its id and
// published time; a new post gets a fresh id and defaults to now when no
// published time is supplied.
func (s *Store) Upsert(p Post) (Post, error) {
	p.Normalize()
	if err := p.Validate(); err != nil {
		return Post{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	before := make([]Post, len(s.file.Posts))
	copy(before, s.file.Posts)

	for i := range s.file.Posts {
		if s.file.Posts[i].Site != p.Site {
			continue
		}
		p.ID = s.file.Posts[i].ID
		p.PublishedAt = s.file.Posts[i].PublishedAt
		p.UpdatedAt = now
		p.Hidden = s.file.Posts[i].Hidden
		s.file.Posts[i] = p
		if err := s.persist(); err != nil {
			s.file.Posts = before
			return Post{}, err
		}
		return p, nil
	}

	if p.ID == "" {
		id, err := NewID()
		if err != nil {
			return Post{}, err
		}
		p.ID = id
	}
	if p.PublishedAt.IsZero() {
		p.PublishedAt = now
	}
	p.UpdatedAt = now
	s.file.Posts = append(s.file.Posts, p)
	if err := s.persist(); err != nil {
		s.file.Posts = before
		return Post{}, err
	}
	return p, nil
}

// SetHidden marks a post hidden from feeds and profiles. The static site
// keeps serving; hidden is an index-level moderation flag.
func (s *Store) SetHidden(site string, hidden bool) (Post, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.file.Posts {
		if s.file.Posts[i].Site != site {
			continue
		}
		old := s.file.Posts[i].Hidden
		s.file.Posts[i].Hidden = hidden
		if err := s.persist(); err != nil {
			s.file.Posts[i].Hidden = old
			return Post{}, false, err
		}
		return s.file.Posts[i], true, nil
	}
	return Post{}, false, nil
}

// Delete removes a post from the index. The caller decides whether to delete
// the site files; the index never owns the bytes.
func (s *Store) Delete(site string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.file.Posts {
		if s.file.Posts[i].Site != site {
			continue
		}
		before := make([]Post, len(s.file.Posts))
		copy(before, s.file.Posts)
		s.file.Posts = append(s.file.Posts[:i], s.file.Posts[i+1:]...)
		if err := s.persist(); err != nil {
			s.file.Posts = before
			return false, err
		}
		return true, nil
	}
	return false, nil
}
