package main

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestResolveAdminTokenGenerated(t *testing.T) {
	data := t.TempDir()
	token, err := resolveAdminToken(data, "", discardLogger())
	if err != nil {
		t.Fatal(err)
	}
	if len(token) < 32 {
		t.Fatalf("generated token too short: %q", token)
	}
	again, err := resolveAdminToken(data, "", discardLogger())
	if err != nil {
		t.Fatal(err)
	}
	if token != again {
		t.Fatalf("generated token not stable: %q vs %q", token, again)
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
