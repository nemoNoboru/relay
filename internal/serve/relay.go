// Package serve hosts a "relay a url" surface: a member pastes a link and
// the node re-hosts what it points at. The download runs out of process with
// yt-dlp; when yt-dlp cannot fetch anything (or is not installed), the node
// falls back to a share card built from the page's social preview metadata.
package serve

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/nemoNoboru/relay/internal/domain"
)

const (
	// relayJobTimeout bounds one relay job from start to finish.
	relayJobTimeout = 10 * time.Minute
	// relayMaxJobs caps in-flight jobs so a busy node cannot fork unbounded
	// yt-dlp processes.
	relayMaxJobs = 4
	// relayJobKeep is how many finished job records are remembered for
	// status polling before older ones are dropped.
	relayJobKeep = 128
	// relayMaxMedia caps the number of downloaded files a job keeps.
	relayMaxMedia = 48
	// relayMaxImage caps a downloaded social preview image.
	relayMaxImage = 16 << 20
	// relayMaxHTML caps the page fetched for metadata extraction.
	relayMaxHTML = 2 << 20
	// relayMaxPage caps a single downloaded media file.
	relayMaxPage = 500 << 20
)

// relayHTTP is the client for the fallback card: fetching the page and its
// preview image. The per-request timeout keeps a dead host from pinning the
// job past its own context deadline.
var relayHTTP = &http.Client{Timeout: 30 * time.Second}

var relayImageExt = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
	".gif": true, ".avif": true, ".svg": true, ".bmp": true,
}

var relayVideoExt = map[string]bool{
	".mp4": true, ".webm": true, ".mkv": true, ".mov": true,
	".m4v": true, ".avi": true, ".ts": true,
}

var relayAudioExt = map[string]bool{
	".m4a": true, ".mp3": true, ".opus": true, ".ogg": true,
	".oga": true, ".flac": true, ".wav": true, ".aac": true, ".weba": true,
}

type relayKind int

const (
	relayKindImage relayKind = iota
	relayKindVideo
	relayKindAudio
)

// relayMedia is one media file collected from a yt-dlp download.
type relayMedia struct {
	Path string
	Kind relayKind
}

// relayCard is the fallback share card: metadata and a preview image pulled
// from the target page itself.
type relayCard struct {
	Title       string
	Description string
	Image       []byte
	ImageExt    string
}

// relayJob is one re-hosting request. Its fields are guarded by Relayer.mu:
// the HTTP surface reads them while run writes them.
type relayJob struct {
	ID      string    `json:"id"`
	Source  string    `json:"source"`
	Status  string    `json:"status"` // running, done, failed
	Title   string    `json:"title,omitempty"`
	Kind    string    `json:"kind,omitempty"` // media, card
	Site    string    `json:"site,omitempty"`
	URL     string    `json:"url,omitempty"`
	PostURL string    `json:"postUrl,omitempty"`
	Error   string    `json:"error,omitempty"`
	Created time.Time `json:"created"`

	// Webhook is an optional callback url the node posts the finished job to.
	// It is a delivery detail for bots, not part of the job's public JSON.
	Webhook string `json:"-"`
}

// Relayer accepts a URL from a member, downloads its content with yt-dlp in a
// background job, and publishes the result as a normal site. download and
// fetch are the two network boundaries; tests replace them with stubs.
type Relayer struct {
	node     domain.Name
	ks       *KeyStore
	uploader *Uploader

	mu   sync.Mutex
	jobs map[string]*relayJob

	download func(ctx context.Context, source, dir string) (string, error)
	fetch    func(ctx context.Context, source string) (relayCard, error)
}

// NewRelayer returns the relay-a-url handler for a node.
func NewRelayer(node domain.Name, ks *KeyStore, uploader *Uploader) *Relayer {
	return &Relayer{
		node:     node,
		ks:       ks,
		uploader: uploader,
		jobs:     map[string]*relayJob{},
		download: relayDownload,
		fetch:    relayFetch,
	}
}

// ServeHTTP routes the relay surface:
//
//	POST /_relay/v1/relay       start a job  { "url": "…" }
//	GET  /_relay/v1/relay/<id>  poll a job
//
// Everything else is 404.
func (rl *Relayer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimSuffix(r.URL.Path, "/")
	switch {
	case p == "/_relay/v1/relay" && r.Method == http.MethodPost:
		rl.start(w, r)
	case strings.HasPrefix(p, "/_relay/v1/relay/") && r.Method == http.MethodGet:
		rl.status(w, r, strings.TrimPrefix(p, "/_relay/v1/relay/"))
	default:
		http.NotFound(w, r)
	}
}

func (rl *Relayer) start(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		writeError(w, http.StatusForbidden, "bad_origin", "request origin does not match this node")
		return
	}
	member, _, ok := memberFromRequest(rl.ks, r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "session_required", "open your invitation link to sign in")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	var req struct {
		URL     string `json:"url"`
		Webhook string `json:"webhook"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "body must be JSON: "+err.Error())
		return
	}
	source, err := validateRelayURL(strings.TrimSpace(req.URL))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_url", err.Error())
		return
	}
	webhook := strings.TrimSpace(req.Webhook)
	if webhook != "" {
		if _, err := validateRelayURL(webhook); err != nil {
			writeError(w, http.StatusBadRequest, "bad_webhook", "webhook must be a full http or https url")
			return
		}
	}
	if rl.countRunning() >= relayMaxJobs {
		writeError(w, http.StatusTooManyRequests, "too_many", "too many relays running; try again in a moment")
		return
	}
	job := &relayJob{ID: newRelayID(), Source: source, Status: "running", Created: time.Now().UTC(), Webhook: webhook}
	rl.add(job)
	go rl.run(job, member)
	writeJSON(w, http.StatusAccepted, job)
}

func (rl *Relayer) status(w http.ResponseWriter, r *http.Request, id string) {
	if id == "" || strings.Contains(id, "/") {
		http.NotFound(w, r)
		return
	}
	if _, _, ok := memberFromRequest(rl.ks, r); !ok {
		writeError(w, http.StatusUnauthorized, "session_required", "open your invitation link to sign in")
		return
	}
	rl.mu.Lock()
	job, ok := rl.jobs[id]
	var cp relayJob
	if ok {
		cp = *job
	}
	rl.mu.Unlock()
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "no relay job "+id)
		return
	}
	writeJSON(w, http.StatusOK, cp)
}

// run owns one job's lifetime. It is a top-level worker: the background
// context is bounded by relayJobTimeout, so the goroutine always exits and
// the in-flight yt-dlp process is killed by its own context.
func (rl *Relayer) run(job *relayJob, member Member) {
	ctx, cancel := context.WithTimeout(context.Background(), relayJobTimeout)
	defer cancel()
	defer rl.deliver(ctx, job)

	if err := os.MkdirAll(rl.uploader.sitesDir, 0o755); err != nil {
		rl.setFailed(job, fmt.Errorf("create sites directory: %w", err))
		return
	}
	dlDir, err := os.MkdirTemp(rl.uploader.sitesDir, ".relay-dl-"+job.ID+"-")
	if err != nil {
		rl.setFailed(job, fmt.Errorf("create download directory: %w", err))
		return
	}
	defer os.RemoveAll(dlDir)

	staging, err := os.MkdirTemp(rl.uploader.sitesDir, ".staging-relay-"+job.ID+"-")
	if err != nil {
		rl.setFailed(job, fmt.Errorf("create staging directory: %w", err))
		return
	}
	// A successful savePublish renames staging into place, after which this
	// remove is a no-op; every other path cleans the staging files up.
	defer os.RemoveAll(staging)

	title, err := rl.download(ctx, job.Source, dlDir)
	media := collectMedia(dlDir)
	if err == nil && len(media) > 0 {
		rl.publishMedia(job, member, staging, title, media)
		return
	}

	card, err := rl.fetch(ctx, job.Source)
	if err != nil {
		// The card is a best-effort fallback; a failed fetch still yields a
		// page that carries the original url.
		card = relayCard{}
	}
	rl.publishCard(job, member, staging, card)
}

// publishMedia copies the downloaded files into the staging directory and
// writes an index.html that plays them, then commits the site.
func (rl *Relayer) publishMedia(job *relayJob, member Member, staging, title string, media []relayMedia) {
	if title == "" {
		title = displayHost(job.Source)
	}
	copied := make([]relayMedia, 0, len(media))
	for i, m := range media {
		name := relayMediaName(i, len(media), strings.ToLower(filepath.Ext(m.Path)))
		if err := copyFile(filepath.Join(staging, name), m.Path); err != nil {
			fmt.Fprintf(os.Stderr, "serve: relay copy %s: %v\n", m.Path, err)
			continue
		}
		copied = append(copied, relayMedia{Path: name, Kind: m.Kind})
	}
	if len(copied) == 0 {
		rl.publishCard(job, member, staging, relayCard{Title: title})
		return
	}
	sort.Slice(copied, func(i, j int) bool {
		if relayKindOrder(copied[i].Kind) != relayKindOrder(copied[j].Kind) {
			return relayKindOrder(copied[i].Kind) < relayKindOrder(copied[j].Kind)
		}
		return copied[i].Path < copied[j].Path
	})
	cover := ""
	for _, m := range copied {
		if m.Kind == relayKindImage {
			cover = m.Path
			break
		}
	}
	if err := os.WriteFile(filepath.Join(staging, "index.html"), []byte(relayMediaPage(title, job.Source, copied)), 0o644); err != nil {
		rl.setFailed(job, fmt.Errorf("write index.html: %w", err))
		return
	}
	site := rl.uploader.uniqueSite(slugify(title))
	if _, _, err := rl.uploader.savePublish(site, staging, member, postMeta{
		Title: title, Summary: "Relayed from " + displayHost(job.Source), Template: "relay", Image: cover,
	}); err != nil {
		rl.setFailed(job, fmt.Errorf("publish relayed media: %w", err))
		return
	}
	rl.setDone(job, title, "media", string(site))
}

// publishCard writes a share card for a page that produced no media, then
// commits the site.
func (rl *Relayer) publishCard(job *relayJob, member Member, staging string, card relayCard) {
	title := strings.TrimSpace(card.Title)
	if title == "" {
		title = displayHost(job.Source)
	}
	image := ""
	if len(card.Image) > 0 && card.ImageExt != "" {
		name := "preview" + card.ImageExt
		if err := os.WriteFile(filepath.Join(staging, name), card.Image, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "serve: relay write preview image: %v\n", err)
		} else {
			image = name
		}
	}
	if err := os.WriteFile(filepath.Join(staging, "index.html"), []byte(relayCardPage(title, card.Description, image, job.Source)), 0o644); err != nil {
		rl.setFailed(job, fmt.Errorf("write index.html: %w", err))
		return
	}
	site := rl.uploader.uniqueSite(slugify(title))
	if _, _, err := rl.uploader.savePublish(site, staging, member, postMeta{
		Title: title, Summary: card.Description, Template: "relay", Image: image,
	}); err != nil {
		rl.setFailed(job, fmt.Errorf("publish relayed card: %w", err))
		return
	}
	rl.setDone(job, title, "card", string(site))
}

func (rl *Relayer) setDone(job *relayJob, title, kind, site string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	job.Status = "done"
	job.Title = title
	job.Kind = kind
	job.Site = site
	job.URL = "/" + string(rl.node) + "/" + site + "/"
	job.PostURL = "/_relay/posts/" + site
}

func (rl *Relayer) setFailed(job *relayJob, err error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if job.Status == "done" {
		return
	}
	job.Status = "failed"
	job.Error = err.Error()
}

// deliver posts the finished job to its webhook url, if one was given. It is
// best-effort: a failed delivery is logged, never part of the job result.
func (rl *Relayer) deliver(ctx context.Context, job *relayJob) {
	if job.Webhook == "" {
		return
	}
	rl.mu.Lock()
	payload := *job
	rl.mu.Unlock()

	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "serve: relay webhook marshal: %v\n", err)
		return
	}
	dctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(dctx, http.MethodPost, job.Webhook, strings.NewReader(string(body)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "serve: relay webhook request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Relay/1.0 (+webhook)")
	resp, err := relayHTTP.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "serve: relay webhook to %s: %v\n", job.Webhook, err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode > 299 {
		fmt.Fprintf(os.Stderr, "serve: relay webhook to %s: %s\n", job.Webhook, resp.Status)
	}
}

func (rl *Relayer) add(job *relayJob) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if len(rl.jobs) >= relayJobKeep {
		for id, j := range rl.jobs {
			if j.Status != "running" {
				delete(rl.jobs, id)
			}
			if len(rl.jobs) < relayJobKeep {
				break
			}
		}
	}
	rl.jobs[job.ID] = job
}

func (rl *Relayer) countRunning() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	n := 0
	for _, j := range rl.jobs {
		if j.Status == "running" {
			n++
		}
	}
	return n
}

// relayDownload runs yt-dlp against source, writing files into dir, and
// returns the title of the first downloaded entry. A missing yt-dlp binary is
// an ordinary error: the caller falls back to the share card.
func relayDownload(ctx context.Context, source, dir string) (string, error) {
	exe, err := exec.LookPath("yt-dlp")
	if err != nil {
		return "", errors.New("yt-dlp is not installed")
	}
	args := []string{
		"--no-progress",
		"--no-warnings",
		"--no-overwrites",
		"--playlist-items", "1-24",
		"--max-downloads", "24",
		"--max-filesize", fmt.Sprintf("%dB", relayMaxPage),
		"--write-thumbnail",
		"--print", "%(title)s",
		"-o", filepath.Join(dir, "%(id)s.%(ext)s"),
		source,
	}
	cmd := exec.CommandContext(ctx, exe, args...)
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return "", err
	}
	for _, line := range strings.Split(out.String(), "\n") {
		if s := strings.TrimSpace(line); s != "" {
			return s, nil
		}
	}
	return "", nil
}

// relayFetch builds the fallback card: the page's social preview metadata
// plus a downloaded copy of its preview image.
func relayFetch(ctx context.Context, source string) (relayCard, error) {
	var card relayCard
	page, err := httpGet(ctx, source, relayMaxHTML)
	if err != nil {
		return card, err
	}
	meta := extractMeta(string(page))
	card.Title = meta.title
	card.Description = meta.description
	if meta.image != "" {
		imgURL, err := resolveURL(source, meta.image)
		if err == nil {
			if b, ext, err := httpGetImage(ctx, imgURL); err == nil {
				card.Image, card.ImageExt = b, ext
			}
		}
	}
	return card, nil
}

func httpGet(ctx context.Context, raw string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Relay/1.0 (+rehost)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,image/*;q=0.9,*/*;q=0.5")
	resp, err := relayHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: %s", raw, resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, limit))
}

// httpGetImage downloads a social preview image, refusing anything that is
// not an image or over the size cap.
func httpGetImage(ctx context.Context, raw string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "Relay/1.0 (+rehost)")
	req.Header.Set("Accept", "image/*")
	resp, err := relayHTTP.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("fetch image %s: %s", raw, resp.Status)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "" && !strings.HasPrefix(ct, "image/") {
		return nil, "", errors.New("not an image")
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, relayMaxImage))
	if err != nil {
		return nil, "", err
	}
	ext := imageExt(resp.Header.Get("Content-Type"), raw)
	if ext == "" {
		return nil, "", errors.New("unknown image type")
	}
	return b, ext, nil
}

func imageExt(contentType, raw string) string {
	if mt, _, err := mime.ParseMediaType(contentType); err == nil {
		switch mt {
		case "image/jpeg":
			return ".jpg"
		case "image/png":
			return ".png"
		case "image/webp":
			return ".webp"
		case "image/gif":
			return ".gif"
		case "image/avif":
			return ".avif"
		case "image/svg+xml":
			return ".svg"
		case "image/bmp":
			return ".bmp"
		}
	}
	if u, err := url.Parse(raw); err == nil {
		if e := strings.ToLower(filepath.Ext(u.Path)); relayImageExt[e] {
			return e
		}
	}
	return ""
}

func validateRelayURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return "", errors.New("enter a full http or https url")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", errors.New("enter a full http or https url")
	}
	return u.String(), nil
}

func displayHost(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	return u.Host
}

func resolveURL(base, ref string) (string, error) {
	b, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	r, err := url.Parse(ref)
	if err != nil {
		return "", err
	}
	u := b.ResolveReference(r)
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", errors.New("resolved url is not http")
	}
	return u.String(), nil
}

func newRelayID() string {
	id, err := GenerateToken()
	if err != nil {
		return fmt.Sprintf("rj_%d", time.Now().UnixNano())
	}
	return "rj_" + id
}

func collectMedia(dir string) []relayMedia {
	var out []relayMedia
	_ = filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		if kind, ok := relayClassify(info.Name()); ok {
			out = append(out, relayMedia{Path: p, Kind: kind})
		}
		return nil
	})
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	if len(out) > relayMaxMedia {
		out = out[:relayMaxMedia]
	}
	return out
}

func relayClassify(name string) (relayKind, bool) {
	ext := strings.ToLower(filepath.Ext(name))
	switch {
	case relayImageExt[ext]:
		return relayKindImage, true
	case relayVideoExt[ext]:
		return relayKindVideo, true
	case relayAudioExt[ext]:
		return relayKindAudio, true
	}
	return 0, false
}

func relayKindOrder(k relayKind) int {
	switch k {
	case relayKindVideo:
		return 0
	case relayKindAudio:
		return 1
	default:
		return 2
	}
}

func relayMediaName(i, total int, ext string) string {
	if total == 1 {
		return "media" + ext
	}
	return fmt.Sprintf("media-%d%s", i+1, ext)
}

func copyFile(dst, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// --- social preview extraction -------------------------------------------

type relayMeta struct {
	title, description, image string
}

var (
	metaTagRe  = regexp.MustCompile(`(?is)<meta\b[^>]*>`)
	metaAttrRe = regexp.MustCompile(`(?is)([a-zA-Z_:][-a-zA-Z0-9_:.]*)\s*=\s*("([^"]*)"|'([^']*)'|([^\s>]+))`)
	titleTagRe = regexp.MustCompile(`(?is)<title\b[^>]*>(.*?)</title>`)
)

// extractMeta reads the social preview fields from a page. It prefers
// Open Graph over Twitter cards over plain <title>/<meta name=description>.
func extractMeta(page string) relayMeta {
	var ogTitle, twTitle, ogDesc, twDesc, metaDesc, ogImage, twImage string
	for _, tag := range metaTagRe.FindAllString(page, -1) {
		attrs := map[string]string{}
		for _, m := range metaAttrRe.FindAllStringSubmatch(tag, -1) {
			name := strings.ToLower(m[1])
			val := m[3]
			if val == "" {
				val = m[4]
			}
			if val == "" {
				val = m[5]
			}
			attrs[name] = html.UnescapeString(val)
		}
		prop := strings.ToLower(attrs["property"])
		if prop == "" {
			prop = strings.ToLower(attrs["name"])
		}
		content := attrs["content"]
		switch prop {
		case "og:title":
			if ogTitle == "" {
				ogTitle = content
			}
		case "twitter:title":
			if twTitle == "" {
				twTitle = content
			}
		case "og:description":
			if ogDesc == "" {
				ogDesc = content
			}
		case "twitter:description":
			if twDesc == "" {
				twDesc = content
			}
		case "description":
			if metaDesc == "" {
				metaDesc = content
			}
		case "og:image", "og:image:url", "og:image:secure_url":
			if ogImage == "" {
				ogImage = content
			}
		case "twitter:image", "twitter:image:src":
			if twImage == "" {
				twImage = content
			}
		}
	}
	htmlTitle := ""
	if m := titleTagRe.FindStringSubmatch(page); m != nil {
		htmlTitle = strings.TrimSpace(html.UnescapeString(m[1]))
	}
	return relayMeta{
		title:       firstNonEmpty(ogTitle, twTitle, htmlTitle),
		description: firstNonEmpty(ogDesc, twDesc, metaDesc),
		image:       firstNonEmpty(ogImage, twImage),
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

// --- generated pages ------------------------------------------------------

const relayPageStyle = `:root{--ink:#141210;--paper:#f0e6d2;--surface:#fbf6e9;--red:#ce1f2b;--hair:rgba(20,18,16,.16)}
*{box-sizing:border-box;margin:0;padding:0}
body{background:var(--paper);color:var(--ink);font-family:'Figtree','Nunito',system-ui,sans-serif;line-height:1.55;-webkit-font-smoothing:antialiased}
.bar{display:flex;align-items:center;justify-content:space-between;gap:12px;border-bottom:1px solid var(--hair);padding:14px 20px;background:rgba(240,230,210,.92)}
.brand{display:flex;align-items:center;gap:10px;font-family:'Nunito',system-ui,sans-serif;font-weight:800;font-size:.8rem;letter-spacing:.08em;text-decoration:none;color:var(--ink)}
.brand .mark{display:grid;place-items:center;width:28px;height:28px;background:var(--red);color:var(--paper);border-radius:8px}
.stamp{font-family:'Figtree',system-ui,sans-serif;font-size:.6rem;letter-spacing:.1em;text-transform:uppercase;border:1px solid var(--hair);border-radius:999px;padding:4px 9px;opacity:.7}
main{max-width:820px;margin:0 auto;padding:34px 22px 90px}
h1{font-family:'Nunito',system-ui,sans-serif;font-weight:700;line-height:1.05;font-size:2rem;margin:0 0 10px}
.source{font-family:'Figtree',system-ui,sans-serif;font-size:.72rem;letter-spacing:.04em;opacity:.6;margin:0 0 20px}
.source a{color:var(--ink);border-bottom:1px solid var(--red);text-decoration:none}
.media{display:flex;flex-direction:column;gap:16px}
.media img{border:1px solid var(--hair);border-radius:12px;max-width:100%;height:auto;background:#fff}
.media video{width:100%;max-height:70vh;background:#000;border-radius:12px}
.media audio{width:100%}
.card{background:var(--surface);border:1px solid var(--hair);border-radius:12px;padding:20px;max-width:640px}
.card .cover{width:100%;border-radius:9px;border:1px solid var(--hair);margin-bottom:16px}
.card h1{font-size:1.6rem}
.card .desc{opacity:.85;margin:0 0 18px;max-width:58ch}
.card .go{display:inline-block;font-size:.78rem;font-family:'Figtree',system-ui,sans-serif;color:var(--paper);background:var(--red);border-radius:8px;padding:9px 14px;text-decoration:none;word-break:break-all}`

func relayPage(title, body string) string {
	return `<!doctype html><html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>` + html.EscapeString(title) + `</title><style>` + relayPageStyle + `</style></head>
<body><header class="bar"><a class="brand" href="/"><span class="mark">R</span> RELAY</a><span class="stamp">relayed</span></header>
<main>` + body + `</main></body></html>`
}

func relayMediaPage(title, source string, media []relayMedia) string {
	var b strings.Builder
	b.WriteString(`<h1>` + html.EscapeString(title) + `</h1>`)
	b.WriteString(`<p class="source">Relayed from <a href="` + html.EscapeString(source) + `">` + html.EscapeString(displayHost(source)) + `</a></p>`)
	b.WriteString(`<div class="media">`)
	for _, m := range media {
		switch m.Kind {
		case relayKindVideo:
			b.WriteString(`<video src="` + m.Path + `" controls playsinline preload="metadata"></video>`)
		case relayKindAudio:
			b.WriteString(`<audio src="` + m.Path + `" controls preload="metadata"></audio>`)
		default:
			b.WriteString(`<img src="` + m.Path + `" alt="">`)
		}
	}
	b.WriteString(`</div>`)
	return relayPage(title, b.String())
}

func relayCardPage(title, description, image, source string) string {
	var b strings.Builder
	b.WriteString(`<div class="card">`)
	if image != "" {
		b.WriteString(`<img class="cover" src="` + html.EscapeString(image) + `" alt="">`)
	}
	b.WriteString(`<h1>` + html.EscapeString(title) + `</h1>`)
	if description != "" {
		b.WriteString(`<p class="desc">` + html.EscapeString(description) + `</p>`)
	}
	b.WriteString(`<a class="go" href="` + html.EscapeString(source) + `">` + html.EscapeString(source) + ` ↗</a>`)
	b.WriteString(`</div>`)
	return relayPage(title, b.String())
}
