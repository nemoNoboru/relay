package social

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"

	"github.com/nemoNoboru/relay/internal/domain"
)

// TemplateField describes one field a template's prepublish form expects.
type TemplateField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required,omitempty"`
	Max      int    `json:"max,omitempty"`
	Accept   string `json:"accept,omitempty"`
}

// Template is the manifest of a post template. The form and publish logic
// live in the template's own HTML/JS; Relay only validates and serves them.
type Template struct {
	Schema      int             `json:"schema"`
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Version     string          `json:"version"`
	Fields      []TemplateField `json:"fields"`
	Entry       string          `json:"entry"`
	Script      string          `json:"script"`
}

// ErrBadTemplate is returned for a missing or malformed template manifest.
var ErrBadTemplate = errors.New("invalid template")

// LoadTemplate reads relay-template.json from a template folder and checks
// that its entry and script exist.
func LoadTemplate(fsys fs.FS) (Template, error) {
	b, err := fs.ReadFile(fsys, "relay-template.json")
	if err != nil {
		return Template{}, fmt.Errorf("%w: read manifest: %v", ErrBadTemplate, err)
	}
	var t Template
	if err := json.Unmarshal(b, &t); err != nil {
		return Template{}, fmt.Errorf("%w: parse manifest: %v", ErrBadTemplate, err)
	}
	if t.Entry == "" {
		t.Entry = "prepublish.html"
	}
	if t.Script == "" {
		t.Script = "publish.js"
	}
	if err := t.Validate(fsys); err != nil {
		return Template{}, err
	}
	return t, nil
}

// Validate checks a template manifest against its folder.
func (t Template) Validate(fsys fs.FS) error {
	if t.Schema != 1 {
		return fmt.Errorf("%w: schema must be 1", ErrBadTemplate)
	}
	if !domain.ValidName(t.ID) {
		return fmt.Errorf("%w: bad id %q", ErrBadTemplate, t.ID)
	}
	if t.Name == "" {
		return fmt.Errorf("%w: name is required", ErrBadTemplate)
	}
	for _, name := range []string{t.Entry, t.Script} {
		if _, err := fs.Stat(fsys, name); err != nil {
			return fmt.Errorf("%w: missing %s", ErrBadTemplate, name)
		}
	}
	return nil
}
