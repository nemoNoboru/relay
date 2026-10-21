package serve

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/nemoNoboru/relay/internal/domain"
	"github.com/nemoNoboru/relay/internal/social"
)

const (
	// maxPublishBytes caps the decoded files of a template-published post.
	// The tar endpoint keeps its larger 256 MiB cap.
	maxPublishBytes = 64 << 20
	// maxPublishJSON leaves room for base64 expansion on top of the bytes cap.
	maxPublishJSON = maxPublishBytes*4/3 + 1<<20
)

// PublishFile is one file produced by a template's browser-side publish.js.
type PublishFile struct {
	Path    string `json:"path"`
	Content string `json:"content,omitempty"`
	Base64  string `json:"base64,omitempty"`
}

type publishRequest struct {
	Template string        `json:"template"`
	Title    string        `json:"title"`
	Summary  string        `json:"summary"`
	Site     string        `json:"site"`
	Image    string        `json:"image"`
	Files    []PublishFile `json:"files"`
}

// Publisher accepts template output from a member and writes it as a normal
// post. Template JavaScript runs in the browser; the node only validates and
// stores the finished files.
type Publisher struct {
	node      domain.Name
	uploader  *Uploader
	templates *TemplateStore
	ks        *KeyStore
}

// NewPublisher returns the template publish endpoint.
func NewPublisher(node domain.Name, uploader *Uploader, templates *TemplateStore, ks *KeyStore) *Publisher {
	return &Publisher{node: node, uploader: uploader, templates: templates, ks: ks}
}

// ServeHTTP handles POST /_relay/v1/publish. Everything else is 404.
func (p *Publisher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	if !sameOrigin(r) {
		writeError(w, http.StatusForbidden, "bad_origin", "request origin does not match this node")
		return
	}
	member, _, ok := memberFromRequest(p.ks, r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "session_required", "open your invitation link to sign in")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxPublishJSON)
	var req publishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeError(w, http.StatusRequestEntityTooLarge, "too_large", "template publish is over the size cap")
			return
		}
		writeError(w, http.StatusBadRequest, "bad_request", "body must be JSON: "+err.Error())
		return
	}
	if _, _, ok := p.templates.Open(req.Template); !ok {
		writeError(w, http.StatusBadRequest, "bad_template", "unknown template "+req.Template)
		return
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = "Untitled"
	}
	author := social.Post{Title: title}
	author.Normalize()
	title = author.Title

	var name domain.Name
	if strings.TrimSpace(req.Site) == "" {
		name = p.uploader.uniqueSite(slugify(title))
	} else {
		parsed, err := domain.ParseName(strings.TrimSpace(req.Site))
		if err != nil {
			writeError(w, http.StatusBadRequest, "bad_name", err.Error())
			return
		}
		name = parsed
	}

	if len(req.Files) == 0 {
		writeError(w, http.StatusBadRequest, "bad_file", "the template produced no files")
		return
	}
	if len(req.Files) > p.uploader.maxFiles {
		writeError(w, http.StatusRequestEntityTooLarge, "too_large", "too many files")
		return
	}

	image := social.Post{Image: req.Image}
	image.Normalize()
	req.Image = image.Image

	staging, files, bytes, err := p.stage(name, req.Files)
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeError(w, http.StatusRequestEntityTooLarge, "too_large", err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, "bad_file", err.Error())
		return
	}
	defer os.RemoveAll(staging)
	if req.Image != "" {
		if _, err := os.Stat(filepath.Join(staging, filepath.FromSlash(req.Image))); err != nil {
			writeError(w, http.StatusBadRequest, "bad_image", "cover image not found in the generated files")
			return
		}
	}

	post, postURL, err := p.uploader.savePublish(name, staging, member, postMeta{
		Title: title, Summary: strings.TrimSpace(req.Summary), Template: req.Template, Image: req.Image,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	view := postView(p.node, p.ks, nil, post, member.ID)
	view.PostURL = postURL
	writeJSON(w, http.StatusOK, map[string]any{
		"post":  view,
		"files": files,
		"bytes": bytes,
	})
}

// stage writes the JSON files into a fresh site directory next to the live
// sites. It enforces path containment, duplicate paths, the file count, and
// the decoded byte cap.
func (p *Publisher) stage(name domain.Name, files []PublishFile) (string, int, int64, error) {
	if err := os.MkdirAll(p.uploader.sitesDir, 0o755); err != nil {
		return "", 0, 0, fmt.Errorf("create sites directory: %w", err)
	}
	dir, err := os.MkdirTemp(p.uploader.sitesDir, ".staging-publish-"+string(name)+"-")
	if err != nil {
		return "", 0, 0, fmt.Errorf("create staging directory: %w", err)
	}
	cleanup := func() { os.RemoveAll(dir) }

	seen := map[string]bool{}
	var total int64
	hasIndex := false
	for _, file := range files {
		raw := strings.TrimSpace(file.Path)
		if raw == "" || strings.ContainsRune(raw, 0) || strings.Contains(raw, "\\") {
			cleanup()
			return "", 0, 0, fmt.Errorf("invalid file path %q", file.Path)
		}
		clean := path.Clean(raw)
		if seen[clean] {
			cleanup()
			return "", 0, 0, fmt.Errorf("duplicate file path %q", raw)
		}
		seen[clean] = true
		full, err := safeJoin(dir, raw)
		if err != nil {
			cleanup()
			return "", 0, 0, err
		}
		data, err := decodePublishFile(file)
		if err != nil {
			cleanup()
			return "", 0, 0, err
		}
		total += int64(len(data))
		if total > maxPublishBytes {
			cleanup()
			return "", 0, 0, fmt.Errorf("post is over the %d MiB cap", maxPublishBytes>>20)
		}
		if clean == "index.html" {
			hasIndex = true
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			cleanup()
			return "", 0, 0, fmt.Errorf("create parent of %q: %w", raw, err)
		}
		if err := os.WriteFile(full, data, 0o644); err != nil {
			cleanup()
			return "", 0, 0, fmt.Errorf("write %q: %w", raw, err)
		}
	}
	if !hasIndex {
		cleanup()
		return "", 0, 0, errors.New("the template must produce index.html")
	}
	return dir, len(files), total, nil
}

func decodePublishFile(file PublishFile) ([]byte, error) {
	if file.Base64 != "" {
		if file.Content != "" {
			return nil, fmt.Errorf("%s: use content or base64, not both", file.Path)
		}
		b, err := base64.StdEncoding.DecodeString(file.Base64)
		if err != nil {
			b, err = base64.RawStdEncoding.DecodeString(file.Base64)
		}
		if err != nil {
			return nil, fmt.Errorf("%s: invalid base64", file.Path)
		}
		return b, nil
	}
	return []byte(file.Content), nil
}

// slugify turns a title into a candidate site name. Site uniqueness is the
// uploader's rule (Uploader.uniqueSite), shared by the template publisher
// and the url relayer.
func slugify(s string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if b.Len() > 0 && !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
		if b.Len() >= 63 {
			break
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "post"
	}
	return out
}
