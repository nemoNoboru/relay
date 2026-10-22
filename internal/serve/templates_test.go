package serve_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/nemoNoboru/relay/internal/serve"
)

func mustTemplateStore(t *testing.T, data string) *serve.TemplateStore {
	t.Helper()
	store, err := serve.NewTemplateStore(data)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func TestBuiltinTemplatesAreListed(t *testing.T) {
	store := mustTemplateStore(t, t.TempDir())
	list, skipped := store.List()
	if len(skipped) != 0 {
		t.Fatalf("skipped = %v", skipped)
	}
	ids := map[string]bool{}
	for _, info := range list {
		ids[info.ID] = true
		if !info.Builtin {
			t.Fatalf("%s is not marked builtin", info.ID)
		}
		if info.Entry != "prepublish.html" || info.Script != "publish.js" {
			t.Fatalf("%s manifest = %+v", info.ID, info.Template)
		}
	}
	for _, id := range []string{"note", "link", "gallery", "article", "songs", "videos", "carousel", "jetset", "longform", "reply", "event"} {
		if !ids[id] {
			t.Fatalf("builtin template %q missing", id)
		}
	}
}

func TestTemplateOpenRejectsTraversalAndUnknown(t *testing.T) {
	store := mustTemplateStore(t, t.TempDir())
	for _, id := range []string{"", "../note", "nope", "_relay"} {
		if _, _, ok := store.Open(id); ok {
			t.Fatalf("Open(%q) succeeded", id)
		}
	}
	fsys, tmpl, ok := store.Open("note")
	if !ok || tmpl.ID != "note" {
		t.Fatalf("Open(note) = %+v %v", tmpl, ok)
	}
	if _, err := fsys.Open("prepublish.html"); err != nil {
		t.Fatalf("prepublish.html: %v", err)
	}
}

func TestTemplateInstallAndRemove(t *testing.T) {
	data := t.TempDir()
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "relay-template.json"), `{"schema":1,"id":"zine","name":"Zine","description":"A test template.","version":"1.0.0","entry":"prepublish.html","script":"publish.js"}`)
	writeFile(t, filepath.Join(src, "prepublish.html"), "<!doctype html><title>Zine</title>")
	writeFile(t, filepath.Join(src, "publish.js"), "export const x = 1;")
	store := mustTemplateStore(t, data)
	id, err := store.Install(src)
	if err != nil || id != "zine" {
		t.Fatalf("Install = %q %v", id, err)
	}
	if _, err := os.Stat(filepath.Join(data, "templates", "zine", "prepublish.html")); err != nil {
		t.Fatal(err)
	}
	if _, _, ok := store.Open("zine"); !ok {
		t.Fatalf("installed template not opened")
	}
	if err := store.Remove("zine"); err != nil {
		t.Fatal(err)
	}
	if _, _, ok := store.Open("zine"); ok {
		t.Fatalf("removed template still opens")
	}
}

func TestInstallTemplateTar(t *testing.T) {
	store := mustTemplateStore(t, t.TempDir())
	body := tarSite(t, map[string]string{
		"relay-template.json": `{"schema":1,"id":"zine","name":"Zine","description":"A test template.","version":"1.0.0","entry":"prepublish.html","script":"publish.js"}`,
		"prepublish.html":     "<!doctype html><title>Zine</title>",
		"publish.js":          "export const x = 1;",
	})
	info, err := store.InstallTar(bytes.NewReader(body))
	if err != nil || info.ID != "zine" || info.Builtin {
		t.Fatalf("InstallTar = %+v %v", info, err)
	}
	if _, _, ok := store.Open("zine"); !ok {
		t.Fatalf("installed template not found")
	}
	if _, err := store.InstallTar(bytes.NewReader(body)); err == nil {
		t.Fatalf("second install overwrote the template")
	}
	builtin := tarSite(t, map[string]string{
		"relay-template.json": `{"schema":1,"id":"note","name":"Bad","version":"1.0.0","entry":"prepublish.html","script":"publish.js"}`,
		"prepublish.html":     "<!doctype html>",
		"publish.js":          "x",
	})
	if _, err := store.InstallTar(bytes.NewReader(builtin)); err == nil {
		t.Fatalf("built-in id was shadowed")
	}
}

func TestMemberCanInstallTemplate(t *testing.T) {
	data := t.TempDir()
	store := mustTemplateStore(t, data)
	ks, err := serve.NewKeyStore(data)
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}
	body := tarSite(t, map[string]string{
		"relay-template.json": `{"schema":1,"id":"poster","name":"Poster","description":"A test template.","version":"1.0.0","entry":"prepublish.html","script":"publish.js"}`,
		"prepublish.html":     "<!doctype html><title>Poster</title>",
		"publish.js":          "export const x = 1;",
	})
	req := httptest.NewRequest(http.MethodPost, "/_relay/v1/templates", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/x-tar")
	rec := httptest.NewRecorder()
	serve.NewTemplates(store, ks).ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("install = %d %s", rec.Code, rec.Body)
	}
	if _, _, ok := store.Open("poster"); !ok {
		t.Fatalf("installed template not found")
	}
}
