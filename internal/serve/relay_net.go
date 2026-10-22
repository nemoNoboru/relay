package serve

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
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
		"--max-filesize", strconv.FormatInt(relayMaxPage, 10),
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

// oembedReply is the small slice of an oEmbed response the card needs.
type oembedReply struct {
	Title      string `json:"title"`
	AuthorName string `json:"author_name"`
	Thumbnail  string `json:"thumbnail_url"`
}

// oembedURL returns a provider's public oEmbed endpoint for a page, or ""
// when the host has no known endpoint. It exists for hosts (YouTube, Vimeo)
// that serve a metadata-free shell to anonymous clients.
func oembedURL(source string) string {
	u, err := url.Parse(source)
	if err != nil {
		return ""
	}
	host := strings.ToLower(u.Host)
	switch {
	case host == "youtube.com", host == "www.youtube.com", host == "m.youtube.com", host == "youtu.be":
		return "https://www.youtube.com/oembed?url=" + url.QueryEscape(source) + "&format=json"
	case host == "vimeo.com", strings.HasSuffix(host, ".vimeo.com"):
		return "https://vimeo.com/api/oembed.json?url=" + url.QueryEscape(source)
	}
	return ""
}

func fetchOEmbed(ctx context.Context, raw string) (oembedReply, error) {
	var reply oembedReply
	b, err := httpGet(ctx, raw, 64<<10)
	if err != nil {
		return reply, err
	}
	if err := json.Unmarshal(b, &reply); err != nil {
		return reply, fmt.Errorf("parse oembed: %w", err)
	}
	return reply, nil
}

// relayFetch builds the fallback card: the page's social preview metadata
// plus a downloaded copy of its preview image. A known provider's oEmbed
// endpoint is authoritative for the title (cleaner than og:title, and still
// served when the page is a metadata-free login shell) and fills the
// thumbnail when the page carried none.
func relayFetch(ctx context.Context, source string) (relayCard, error) {
	var card relayCard
	var image string
	if page, err := httpGet(ctx, source, relayMaxHTML); err == nil {
		meta := extractMeta(string(page))
		card.Title = meta.title
		card.Description = meta.description
		image = meta.image
	}
	if oe := oembedURL(source); oe != "" {
		if om, err := fetchOEmbed(ctx, oe); err == nil {
			if om.Title != "" {
				card.Title = om.Title
			}
			if card.Description == "" {
				card.Description = om.AuthorName
			}
			if image == "" {
				image = om.Thumbnail
			}
		}
	}
	if image != "" {
		if imgURL, err := resolveURL(source, image); err == nil {
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

func collectMedia(dir string) ([]relayMedia, error) {
	var out []relayMedia
	err := filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info == nil || info.IsDir() {
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
	return out, err
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
