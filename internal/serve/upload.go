package serve

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nemoNoboru/relay/internal/domain"
	"github.com/nemoNoboru/relay/internal/social"
)

const (
	// MaxDeployBytes is the default cap on one uploaded site archive.
	MaxDeployBytes = 256 << 20
	// maxDeployFiles caps the number of regular files in one deploy.
	maxDeployFiles = 1 << 16
)

// Uploader accepts a site upload and swaps it in as the live site. It is
// the write half of static hosting; Registry is the read half. Uploads are
// authenticated against the per-user deploy keys in a KeyStore, and each
// site is attributed to the user who last deployed it.
type Uploader struct {
	self     domain.Name
	sitesDir string
	ks       *KeyStore
	versions *Versions
	posts    *social.Store
	maxBytes int64
	maxFiles int
	reload   func() error
}

// NewUploader returns an uploader that writes sites under <dataDir>/sites
// and reports them at /<self>/<site>/. ks authorizes uploads and records
// site ownership; versions keeps rollback snapshots; posts indexes the
// publish (it may be nil). reload is called after every successful deploy so
// the serving registry picks up the new site; it may be nil.
func NewUploader(self domain.Name, dataDir string, ks *KeyStore, versions *Versions, posts *social.Store, maxBytes int64, reload func() error) *Uploader {
	return &Uploader{
		self:     self,
		sitesDir: filepath.Join(dataDir, "sites"),
		ks:       ks,
		versions: versions,
		posts:    posts,
		maxBytes: maxBytes,
		maxFiles: maxDeployFiles,
		reload:   reload,
	}
}

// ServeHTTP handles POST /_relay/v1/sites/<name>/deploy with an
// application/x-tar body. Everything else is 404.
func (u *Uploader) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const base = "/_relay/v1/sites/"
	if !sameOrigin(r) {
		writeError(w, http.StatusForbidden, "bad_origin", "request origin does not match this node")
		return
	}
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	rest, ok := strings.CutPrefix(r.URL.Path, base)
	if !ok {
		http.NotFound(w, r)
		return
	}
	segs := strings.Split(rest, "/")
	if len(segs) != 2 || segs[1] != "deploy" {
		http.NotFound(w, r)
		return
	}
	name, err := domain.ParseName(segs[0])
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_name", err.Error())
		return
	}
	u.deploy(w, r, name)
}

func (u *Uploader) deploy(w http.ResponseWriter, r *http.Request, name domain.Name) {
	member, _, ok := memberFromRequest(u.ks, r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "bad_token", "invalid or missing invitation")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, u.maxBytes)

	dir, files, bytes, err := u.stage(name, r.Body)
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeError(w, http.StatusRequestEntityTooLarge, "too_large", err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, "bad_archive", err.Error())
		return
	}
	defer os.RemoveAll(dir) // no-op once renamed into place

	meta := readPostMeta(dir)
	_, postURL, err := u.savePublish(name, dir, member, meta)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"site":    string(name),
		"files":   files,
		"bytes":   bytes,
		"url":     "/" + string(u.self) + "/" + string(name) + "/",
		"postUrl": postURL,
	})
}

// savePublish commits a staged site, records ownership, writes the portable
// sidecar, and upserts the post index. The site bytes are the source of
// truth; index and bookkeeping failures are logged, not fatal.
func (u *Uploader) savePublish(name domain.Name, staging string, member Member, meta postMeta) (social.Post, string, error) {
	post := social.Post{
		Site: string(name), Member: member.ID,
		Title: meta.Title, Summary: meta.Summary, Template: meta.Template, Image: meta.Image,
	}
	post.Normalize()
	if err := writePostMeta(staging, postMeta{
		Title: post.Title, Summary: post.Summary, Template: post.Template, Image: post.Image,
		Author: member.Label, Member: member.ID,
	}); err != nil {
		slog.Warn("write post metadata", "site", name, "err", err)
	}
	if err := u.commit(name, staging); err != nil {
		return social.Post{}, "", err
	}
	if u.reload != nil {
		if err := u.reload(); err != nil {
			slog.Warn("reload sites after publish", "err", err)
		}
	}
	if err := u.ks.RecordSite(string(name), member.Label); err != nil {
		slog.Warn("record site owner", "site", name, "err", err)
	}
	if u.posts != nil {
		indexed, err := u.posts.Upsert(post)
		if err != nil {
			slog.Warn("index post", "site", name, "err", err)
		} else {
			post = indexed
		}
	}
	return post, "/_relay/posts/" + string(name), nil
}

// stage extracts the archive in body into a fresh staging directory next to
// the live sites, returning the directory path and the archive's totals.
// On error the staging directory is removed.
func (u *Uploader) stage(name domain.Name, body io.Reader) (string, int, int64, error) {
	if err := os.MkdirAll(u.sitesDir, 0o755); err != nil {
		return "", 0, 0, fmt.Errorf("create sites directory: %w", err)
	}
	dir, err := os.MkdirTemp(u.sitesDir, ".staging-"+string(name)+"-")
	if err != nil {
		return "", 0, 0, fmt.Errorf("create staging directory: %w", err)
	}
	files, bytes, err := extractTar(dir, body, u.maxFiles)
	if err != nil {
		os.RemoveAll(dir)
		return "", 0, 0, err
	}
	return dir, files, bytes, nil
}

// commit swaps the staged site into place, first snapshotting the current
// live site (if any) so a rollback can return to it.
func (u *Uploader) commit(name domain.Name, staging string) error {
	snapshot, err := u.versions.Snapshot(name)
	if err != nil {
		return fmt.Errorf("snapshot current site: %w", err)
	}
	if err := os.Rename(staging, filepath.Join(u.sitesDir, string(name))); err != nil {
		if snapshot != "" {
			if undoErr := u.versions.Restore(name, snapshot); undoErr != nil {
				slog.Error("restore snapshot after failed deploy", "site", name, "err", undoErr)
			}
		}
		return fmt.Errorf("move staging into place: %w", err)
	}
	if err := u.versions.Prune(name); err != nil {
		// Non-fatal: only the retention cap is unmet.
		slog.Warn("prune versions", "site", name, "err", err)
	}
	return nil
}

// extractTar unpacks a tar archive into dest. It accepts only directories
// and regular files, rejects any entry that would escape dest, and enforces
// the file-count cap. The byte cap is enforced by the caller's reader.
func extractTar(dest string, r io.Reader, maxFiles int) (int, int64, error) {
	tr := tar.NewReader(r)
	var files int
	var bytes int64
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return files, bytes, nil
		}
		if err != nil {
			return files, bytes, fmt.Errorf("not a tar archive: %w", err)
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			// The archive root itself ("tar -c ." emits a "." entry) is not a
			// directory to create inside dest; skip it.
			if path.Clean(hdr.Name) == "." {
				continue
			}
			full, err := safeJoin(dest, hdr.Name)
			if err != nil {
				return files, bytes, err
			}
			if err := os.MkdirAll(full, 0o755); err != nil {
				return files, bytes, fmt.Errorf("create directory %q: %w", hdr.Name, err)
			}
		case tar.TypeReg:
			if files >= maxFiles {
				return files, bytes, fmt.Errorf("too many files (limit %d)", maxFiles)
			}
			full, err := safeJoin(dest, hdr.Name)
			if err != nil {
				return files, bytes, err
			}
			if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
				return files, bytes, fmt.Errorf("create parent of %q: %w", hdr.Name, err)
			}
			out, err := os.OpenFile(full, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
			if err != nil {
				return files, bytes, fmt.Errorf("create file %q: %w", hdr.Name, err)
			}
			n, err := io.Copy(out, tr)
			if closeErr := out.Close(); err == nil {
				err = closeErr
			}
			if err != nil {
				return files, bytes, fmt.Errorf("write file %q: %w", hdr.Name, err)
			}
			bytes += n
			files++
		default:
			return files, bytes, fmt.Errorf("unsupported archive entry %q (type %q)", hdr.Name, hdr.Typeflag)
		}
	}
}

// safeJoin joins a tar entry name onto dest and rejects any path that would
// escape dest: absolute paths, ".." segments, and entries resolving to dest
// itself.
func safeJoin(dest, name string) (string, error) {
	if name == "" || strings.ContainsRune(name, 0) {
		return "", errors.New("empty or invalid path in archive")
	}
	if strings.Contains(name, "\\") {
		return "", fmt.Errorf("path escapes the site: %q", name)
	}
	clean := path.Clean(name)
	if path.IsAbs(clean) || clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("path escapes the site: %q", name)
	}
	full := filepath.Join(dest, filepath.FromSlash(clean))
	// Definitive containment check: after joining, the result must still be
	// dest itself or something under it. This also closes Windows-specific
	// escapes (drive letters, backslashes) that the slash-only checks miss.
	if full != dest && !strings.HasPrefix(full, dest+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes the site: %q", name)
	}
	return full, nil
}

// uniqueSite derives a free site name from a slug. It is shared by the
// template publisher and the url relayer; both need a name that is not
// already taken by an indexed post or a site folder.
func (u *Uploader) uniqueSite(base string) domain.Name {
	base = strings.Trim(strings.ToLower(base), "-")
	if base == "" {
		base = "post"
	}
	for i := 1; i < 1000; i++ {
		candidate := base
		if i > 1 {
			suffix := "-" + strconv.Itoa(i)
			if len(candidate)+len(suffix) > 63 {
				candidate = candidate[:63-len(suffix)]
			}
			candidate += suffix
		}
		if u.posts != nil {
			if _, ok := u.posts.Get(candidate); ok {
				continue
			}
		}
		if _, err := os.Stat(filepath.Join(u.sitesDir, candidate)); err == nil {
			continue
		}
		return domain.Name(candidate)
	}
	return domain.Name(base)
}
