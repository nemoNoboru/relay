package serve

import (
	"archive/tar"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/nemoNoboru/relay/internal/domain"
)

const (
	// MaxDeployBytes is the default cap on one uploaded site archive.
	MaxDeployBytes = 256 << 20
	// maxDeployFiles caps the number of regular files in one deploy.
	maxDeployFiles = 1 << 16
)

// Uploader accepts a site upload and swaps it in as the live site. It is
// the write half of static hosting; Registry is the read half.
type Uploader struct {
	self     domain.Name
	sitesDir string
	token    [sha256.Size]byte // SHA-256 of the deploy token, for constant-time checks
	maxBytes int64
	maxFiles int
	reload   func() error
}

// NewUploader returns an uploader that writes sites under <dataDir>/sites
// and reports them at /<self>/<site>/. token is the deploy secret; only its
// hash is kept. reload is called after every successful deploy so the
// serving registry picks up the new site; it may be nil.
func NewUploader(self domain.Name, dataDir, token string, maxBytes int64, reload func() error) *Uploader {
	return &Uploader{
		self:     self,
		sitesDir: filepath.Join(dataDir, "sites"),
		token:    sha256.Sum256([]byte(token)),
		maxBytes: maxBytes,
		maxFiles: maxDeployFiles,
		reload:   reload,
	}
}

// ServeHTTP handles POST /_relay/v1/sites/<name>/deploy with an
// application/x-tar body. Everything else is 404.
func (u *Uploader) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const base = "/_relay/v1/sites/"
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
	if !u.authorized(r) {
		writeError(w, http.StatusUnauthorized, "bad_token", "invalid or missing deploy token")
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

	if err := u.commit(name, dir); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	if u.reload != nil {
		if err := u.reload(); err != nil {
			// The site is live on disk; only the in-memory view failed to
			// refresh. Report it, do not fail the deploy.
			fmt.Fprintf(os.Stderr, "serve: reload sites after deploy: %v\n", err)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"site":  string(name),
		"files": files,
		"bytes": bytes,
		"url":   "/" + string(u.self) + "/" + string(name) + "/",
	})
}

// authorized reports whether r carries the deploy token as a bearer token.
func (u *Uploader) authorized(r *http.Request) bool {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || token == "" {
		return false
	}
	got := sha256.Sum256([]byte(token))
	return subtle.ConstantTimeCompare(got[:], u.token[:]) == 1
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

// commit swaps the staged site into place. The current site, if any, is
// moved aside first and removed afterwards; on failure it is restored.
func (u *Uploader) commit(name domain.Name, staging string) error {
	target := filepath.Join(u.sitesDir, string(name))
	backup := target + ".old-" + strconv.FormatInt(time.Now().UnixNano(), 10)

	if err := os.Rename(target, backup); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("move current site aside: %w", err)
	}
	if err := os.Rename(staging, target); err != nil {
		_ = os.Rename(backup, target) // best-effort rollback
		return fmt.Errorf("move staging into place: %w", err)
	}
	if err := os.RemoveAll(backup); err != nil && !os.IsNotExist(err) {
		// Non-fatal: the old site remains as an orphan backup directory.
		fmt.Fprintf(os.Stderr, "serve: remove old site %s: %v\n", backup, err)
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

// errorEnvelope is the contract's error shape (docs/api-contract.md §2).
type errorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// writeError writes the contract's error envelope (docs/api-contract.md §2).
func writeError(w http.ResponseWriter, status int, code, message string) {
	e := errorEnvelope{}
	e.Error.Code = code
	e.Error.Message = message
	writeJSON(w, status, e)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
