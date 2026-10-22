package serve

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/nemoNoboru/relay/internal/persist"
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
	path := filepath.Join(dir, "relay-post.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return m
	}
	if err := json.Unmarshal(b, &m); err != nil {
		slog.Warn("parse post metadata", "path", path, "err", err)
	}
	return m
}

func writePostMeta(dir string, m postMeta) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return persist.WriteFile(filepath.Join(dir, "relay-post.json"), append(b, '\n'), 0o644)
}
