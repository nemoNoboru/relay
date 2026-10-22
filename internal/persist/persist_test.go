package persist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileCreatesParentAndMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "token")
	if err := WriteFile(path, []byte("secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "secret\n" {
		t.Fatalf("content = %q", b)
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o, want 600", fi.Mode().Perm())
	}
}

func TestWriteFileReplacesAtomically(t *testing.T) {
	path := filepath.Join(t.TempDir(), "posts.json")
	if err := WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(path, []byte("new-content"), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "new-content" {
		t.Fatalf("content = %q", b)
	}
	// The temp files must be gone after the rename.
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() != "posts.json" {
			t.Fatalf("leftover file %q", e.Name())
		}
	}
}

func TestWriteJSONRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	in := struct {
		Name string `json:"name"`
	}{Name: "relay"}
	if err := WriteJSON(path, in, 0o600); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var out struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.Name != "relay" {
		t.Fatalf("name = %q", out.Name)
	}
}
