package serve

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nemoNoboru/relay/internal/domain"
	"github.com/nemoNoboru/relay/internal/social"
	"github.com/nemoNoboru/relay/internal/templates"
)

// TemplateInfo is the member-facing listing of a template.
type TemplateInfo struct {
	social.Template
	Builtin bool `json:"builtin"`
}

// TemplateStore combines the embedded built-ins with <data>/templates.
// An installed template shadows a built-in with the same id.
type TemplateStore struct {
	builtin fs.FS
	dir     string
}

// NewTemplateStore returns the store for a node's data directory.
func NewTemplateStore(dataDir string) *TemplateStore {
	builtin, err := fs.Sub(templates.Builtin, "builtin")
	if err != nil {
		// The embedded tree is fixed at build time; this cannot fail.
		panic("relay: embedded templates: " + err.Error())
	}
	return &TemplateStore{builtin: builtin, dir: filepath.Join(dataDir, "templates")}
}

// List returns every valid template, installed entries overriding built-ins.
func (s *TemplateStore) List() ([]TemplateInfo, []string) {
	byID := map[string]TemplateInfo{}
	var skipped []string

	if entries, err := fs.ReadDir(s.builtin, "."); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			info, err := s.loadBuiltin(entry.Name())
			if err != nil {
				skipped = append(skipped, entry.Name()+": "+err.Error())
				continue
			}
			byID[info.ID] = info
		}
	}
	if entries, err := os.ReadDir(s.dir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			info, err := s.loadInstalled(entry.Name())
			if err != nil {
				skipped = append(skipped, entry.Name()+": "+err.Error())
				continue
			}
			byID[info.ID] = info
		}
	}

	out := make([]TemplateInfo, 0, len(byID))
	for _, info := range byID {
		out = append(out, info)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, skipped
}

// Open returns a template's file system and manifest.
func (s *TemplateStore) Open(id string) (fs.FS, social.Template, bool) {
	if _, err := domain.ParseName(id); err != nil {
		return nil, social.Template{}, false
	}
	if dir := filepath.Join(s.dir, id); isDir(dir) {
		fsys := os.DirFS(dir)
		t, err := social.LoadTemplate(fsys)
		if err != nil || t.ID != id {
			return nil, social.Template{}, false
		}
		return fsys, t, true
	}
	fsys, err := fs.Sub(s.builtin, id)
	if err != nil {
		return nil, social.Template{}, false
	}
	t, err := social.LoadTemplate(fsys)
	if err != nil || t.ID != id {
		return nil, social.Template{}, false
	}
	return fsys, t, true
}

// Name returns a template's display name, or the id itself when the template
// is unknown (e.g. a plain "custom" upload).
func (s *TemplateStore) Name(id string) string {
	if _, t, ok := s.Open(id); ok && t.Name != "" {
		return t.Name
	}
	return id
}

// Install copies a template folder into <data>/templates/<id>. It is a host
// action (the CLI); there is no public marketplace.
func (s *TemplateStore) Install(srcDir string) (string, error) {
	src := os.DirFS(srcDir)
	t, err := social.LoadTemplate(src)
	if err != nil {
		return "", err
	}
	dst := filepath.Join(s.dir, t.ID)
	if err := os.RemoveAll(dst); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", err
	}
	if err := os.CopyFS(dst, src); err != nil {
		os.RemoveAll(dst)
		return "", fmt.Errorf("copy template: %w", err)
	}
	return t.ID, nil
}

// InstallTar installs a template from a tar archive whose root contains
// relay-template.json, prepublish.html, and publish.js. It refuses to shadow a
// built-in or overwrite an installed template silently.
func (s *TemplateStore) InstallTar(r io.Reader) (TemplateInfo, error) {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return TemplateInfo{}, err
	}
	tmp, err := os.MkdirTemp(s.dir, ".staging-template-")
	if err != nil {
		return TemplateInfo{}, err
	}
	defer os.RemoveAll(tmp)
	if _, _, err := extractTar(tmp, r, 4096); err != nil {
		return TemplateInfo{}, err
	}
	t, err := social.LoadTemplate(os.DirFS(tmp))
	if err != nil {
		return TemplateInfo{}, err
	}
	if _, err := fs.Stat(s.builtin, t.ID); err == nil {
		return TemplateInfo{}, errors.New("template id is built-in")
	}
	dst := filepath.Join(s.dir, t.ID)
	if isDir(dst) {
		return TemplateInfo{}, errors.New("template is already installed")
	}
	if err := os.Rename(tmp, dst); err != nil {
		return TemplateInfo{}, fmt.Errorf("install template: %w", err)
	}
	return TemplateInfo{Template: t, Builtin: false}, nil
}

// Remove deletes an installed template. Built-ins cannot be removed; removing
// an override restores the built-in.
func (s *TemplateStore) Remove(id string) error {
	if _, err := domain.ParseName(id); err != nil {
		return err
	}
	dir := filepath.Join(s.dir, id)
	if !isDir(dir) {
		return errors.New("no installed template " + id)
	}
	return os.RemoveAll(dir)
}

func (s *TemplateStore) loadBuiltin(id string) (TemplateInfo, error) {
	fsys, err := fs.Sub(s.builtin, id)
	if err != nil {
		return TemplateInfo{}, err
	}
	t, err := social.LoadTemplate(fsys)
	if err != nil {
		return TemplateInfo{}, err
	}
	if t.ID != id {
		return TemplateInfo{}, fmt.Errorf("manifest id %q does not match folder %q", t.ID, id)
	}
	return TemplateInfo{Template: t, Builtin: true}, nil
}

func (s *TemplateStore) loadInstalled(id string) (TemplateInfo, error) {
	fsys := os.DirFS(filepath.Join(s.dir, id))
	t, err := social.LoadTemplate(fsys)
	if err != nil {
		return TemplateInfo{}, err
	}
	if t.ID != id {
		return TemplateInfo{}, fmt.Errorf("manifest id %q does not match folder %q", t.ID, id)
	}
	return TemplateInfo{Template: t, Builtin: false}, nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// Templates serves the member template surface:
//
//	GET  /_relay/v1/templates            list manifests
//	POST /_relay/v1/templates            install a template tar
//	GET  /_relay/templates/<id>/<file>   prepublish.html, publish.js, assets
type Templates struct {
	store *TemplateStore
	ks    *KeyStore
}

// NewTemplates returns the template handler.
func NewTemplates(store *TemplateStore, ks *KeyStore) *Templates {
	return &Templates{store: store, ks: ks}
}

func (t *Templates) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// The manifest list is public so a public feed can render template names;
	// installing and serving template files are member actions.
	if strings.TrimSuffix(r.URL.Path, "/") == "/_relay/v1/templates" && (r.Method == http.MethodGet || r.Method == http.MethodHead) {
		t.list(w)
		return
	}
	if _, _, ok := memberFromRequest(t.ks, r); !ok {
		if strings.Contains(r.Header.Get("Accept"), "text/html") {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		writeError(w, http.StatusUnauthorized, "session_required", "open your invitation link to sign in")
		return
	}
	if r.URL.Path == "/_relay/v1/templates" && r.Method == http.MethodPost {
		t.install(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.NotFound(w, r)
		return
	}
	if !strings.HasPrefix(r.URL.Path, "/_relay/templates/") {
		http.NotFound(w, r)
		return
	}
	t.serveFile(w, r, strings.TrimPrefix(r.URL.Path, "/_relay/templates/"))
}

func (t *Templates) list(w http.ResponseWriter) {
	list, skipped := t.store.List()
	for _, s := range skipped {
		fmt.Fprintf(os.Stderr, "serve: skipping template: %s\n", s)
	}
	writeJSON(w, http.StatusOK, map[string]any{"templates": list})
}

func (t *Templates) install(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		writeError(w, http.StatusForbidden, "bad_origin", "request origin does not match this node")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<20)
	info, err := t.store.InstallTar(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_template", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"template": info})
}

func (t *Templates) serveFile(w http.ResponseWriter, r *http.Request, rest string) {
	id, file, found := strings.Cut(rest, "/")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	if !found || file == "" {
		http.Redirect(w, r, "/_relay/templates/"+id+"/prepublish.html", http.StatusFound)
		return
	}
	if !fs.ValidPath(file) || path.Clean(file) != file {
		http.NotFound(w, r)
		return
	}
	fsys, _, ok := t.store.Open(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if _, err := fs.Stat(fsys, file); err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-store")
	http.ServeFileFS(w, r, fsys, file)
}
