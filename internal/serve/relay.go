// Package serve is the HTTP layer of the Relay node: static site hosting,
// the owner UI, the member session and profile APIs, the post and comment
// surfaces, template publishing, and the relay-a-url re-hosting surface.
// Identity rules live in internal/identity and the social model lives in
// internal/social; this package translates HTTP into those features.
package serve

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
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

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu   sync.Mutex
	jobs map[string]*relayJob

	download func(ctx context.Context, source, dir string) (string, error)
	fetch    func(ctx context.Context, source string) (relayCard, error)
}

// NewRelayer returns the relay-a-url handler for a node. The handler owns its
// background jobs: Shutdown cancels them and waits for them to finish.
func NewRelayer(node domain.Name, ks *KeyStore, uploader *Uploader) *Relayer {
	ctx, cancel := context.WithCancel(context.Background())
	return &Relayer{
		node:     node,
		ks:       ks,
		uploader: uploader,
		ctx:      ctx,
		cancel:   cancel,
		jobs:     map[string]*relayJob{},
		download: relayDownload,
		fetch:    relayFetch,
	}
}

// Shutdown cancels in-flight relay jobs and waits for them to exit. It returns
// ctx.Err() when the wait outlives the caller's context.
func (rl *Relayer) Shutdown(ctx context.Context) error {
	rl.cancel()
	done := make(chan struct{})
	go func() {
		rl.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
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
	rl.wg.Add(1)
	go func() {
		defer rl.wg.Done()
		rl.run(job, member)
	}()
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
	ctx, cancel := context.WithTimeout(rl.ctx, relayJobTimeout)
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
	media, walkErr := collectMedia(dlDir)
	if walkErr != nil {
		slog.Warn("collect relay media", "err", walkErr)
	}
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
			slog.Warn("relay copy", "file", m.Path, "err", err)
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
			slog.Warn("relay write preview image", "err", err)
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
		slog.Warn("relay webhook marshal", "err", err)
		return
	}
	dctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(dctx, http.MethodPost, job.Webhook, strings.NewReader(string(body)))
	if err != nil {
		slog.Warn("relay webhook request", "err", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Relay/1.0 (+webhook)")
	resp, err := relayHTTP.Do(req)
	if err != nil {
		slog.Warn("relay webhook delivery", "url", job.Webhook, "err", err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode > 299 {
		slog.Warn("relay webhook status", "url", job.Webhook, "status", resp.Status)
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
