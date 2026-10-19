package serve

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// postMeta is the portable sidecar written into a post's folder. The
// authoritative index is posts.json; this file keeps a post readable and
// importable without the node's index.
type postMeta struct {
	Title    string `json:"title,omitempty"`
	Summary  string `json:"summary,omitempty"`
	Template string `json:"template,omitempty"`
	Image    string `json:"image,omitempty"`
	Author   string `json:"author,omitempty"`
	Member   string `json:"member,omitempty"`
}

func readPostMeta(dir string) postMeta {
	var m postMeta
	b, err := os.ReadFile(filepath.Join(dir, "relay-post.json"))
	if err != nil {
		return m
	}
	_ = json.Unmarshal(b, &m) // a malformed sidecar falls back to defaults
	return m
}

func writePostMeta(dir string, m postMeta) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "relay-post.json"), append(b, '\n'), 0o644)
}
