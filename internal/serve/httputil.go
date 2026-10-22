package serve

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// errorEnvelope is the API's error shape (docs/api.md).
type errorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// writeError writes the API's error envelope (docs/api.md).
func writeError(w http.ResponseWriter, status int, code, message string) {
	e := errorEnvelope{}
	e.Error.Code = code
	e.Error.Message = message
	writeJSON(w, status, e)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("write json response", "err", err)
	}
}

// sameOrigin rejects a mutating request whose Origin does not match this
// node. Both scheme and host must agree; a missing Origin is allowed because
// SameSite=Lax already stops cross-site cookie writes and non-browser clients
// (curl, assistants) do not send one.
func sameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return false
	}
	if !strings.EqualFold(u.Host, r.Host) {
		return false
	}
	scheme := "http"
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	return strings.EqualFold(u.Scheme, scheme)
}

// rateGuard throttles one key's actions to a minimum interval. It is an
// in-memory guard, not a security boundary; a restart resets it.
type rateGuard struct {
	mu   sync.Mutex
	min  time.Duration
	last map[string]time.Time
}

func newRateGuard(min time.Duration) *rateGuard {
	return &rateGuard{min: min, last: map[string]time.Time{}}
}

// allow reports whether key may act now and records the attempt when it may.
func (g *rateGuard) allow(key string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if last, ok := g.last[key]; ok && time.Since(last) < g.min {
		return false
	}
	g.last[key] = time.Now()
	return true
}
