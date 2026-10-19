package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nemoNoboru/relay/internal/serve"
	"github.com/nemoNoboru/relay/internal/social"
)

func TestIndexBackfillsOwnedSites(t *testing.T) {
	data := t.TempDir()
	ks, err := serve.NewKeyStore(data)
	if err != nil {
		t.Fatal(err)
	}
	_, key, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}
	if err := ks.RecordSite("sunday-walk", "alice"); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(data, "sites", "sunday-walk"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(data, "sites", "sunday-walk", "index.html"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(data, "sites", "unowned"), 0o755); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	if code := runIndex([]string{"--data", data}, &out, &errOut); code != 0 {
		t.Fatalf("runIndex = %d, stderr = %s", code, errOut.String())
	}
	posts, err := social.NewStore(data)
	if err != nil {
		t.Fatal(err)
	}
	post, ok := posts.Get("sunday-walk")
	if !ok || post.Member != key.ID || post.Title != "Sunday Walk" || post.Template != "custom" {
		t.Fatalf("indexed post = %+v %v", post, ok)
	}
	if _, ok := posts.Get("unowned"); ok {
		t.Fatalf("unowned site was indexed")
	}
	if !strings.Contains(errOut.String(), "unowned") {
		t.Fatalf("stderr = %q, want the unowned skip reported", errOut.String())
	}

	// Idempotent: a second run does not duplicate or overwrite.
	out.Reset()
	errOut.Reset()
	if code := runIndex([]string{"--data", data}, &out, &errOut); code != 0 {
		t.Fatalf("second runIndex = %d", code)
	}
	if !strings.Contains(out.String(), "indexed 0") {
		t.Fatalf("second run = %q, want indexed 0", out.String())
	}
}
