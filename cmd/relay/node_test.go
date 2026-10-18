package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nemoNoboru/relay/internal/serve"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestResolveAdminTokenExplicit(t *testing.T) {
	data := t.TempDir()
	token, err := resolveAdminToken(data, "supersecret", discardLogger())
	if err != nil {
		t.Fatal(err)
	}
	if token != "supersecret" {
		t.Fatalf("token = %q, want supersecret", token)
	}
	b, err := os.ReadFile(filepath.Join(data, "admin-token"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(b)) != "supersecret" {
		t.Fatalf("token file = %q", b)
	}
	if perm := tokenFilePerm(t, data); perm != 0o600 {
		t.Fatalf("token file mode = %o, want 600", perm)
	}
}

// TestResolveAdminTokenDefault pins the streamlined first run: with nothing
// configured the node comes up on the well-known default token, which the
// console then forces the operator to replace.
func TestResolveAdminTokenDefault(t *testing.T) {
	data := t.TempDir()
	token, err := resolveAdminToken(data, "", discardLogger())
	if err != nil {
		t.Fatal(err)
	}
	if token != serve.DefaultAdminToken {
		t.Fatalf("token = %q, want the default %q", token, serve.DefaultAdminToken)
	}
	if perm := tokenFilePerm(t, data); perm != 0o600 {
		t.Fatalf("token file mode = %o, want 600", perm)
	}
}

func TestResolveAdminTokenExisting(t *testing.T) {
	data := t.TempDir()
	if err := os.WriteFile(filepath.Join(data, "admin-token"), []byte("already\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	token, err := resolveAdminToken(data, "", discardLogger())
	if err != nil {
		t.Fatal(err)
	}
	if token != "already" {
		t.Fatalf("token = %q, want already", token)
	}
}

func tokenFilePerm(t *testing.T, data string) os.FileMode {
	t.Helper()
	fi, err := os.Stat(filepath.Join(data, "admin-token"))
	if err != nil {
		t.Fatal(err)
	}
	return fi.Mode().Perm()
}

// TestHealthz pins the liveness/readiness endpoint that Kubernetes probes.
func TestHealthz(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	healthz(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got, want := rec.Body.String(), "ok\n"; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Fatalf("Content-Type = %q", got)
	}
}

// TestRunNodeTLSFlagsTogether pins that --tls-cert and --tls-key must be
// given as a pair.
func TestRunNodeTLSFlagsTogether(t *testing.T) {
	for _, args := range [][]string{
		{"--name", "n", "--data", "d", "--listen", ":0", "--tls-cert", "c.pem"},
		{"--name", "n", "--data", "d", "--listen", ":0", "--tls-key", "k.pem"},
	} {
		if code := runNode(args, io.Discard, io.Discard); code != 2 {
			t.Fatalf("runNode(%v) = %d, want 2", args, code)
		}
	}
}
