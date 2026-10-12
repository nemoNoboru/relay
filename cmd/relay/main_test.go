package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nemoNoboru/relay/internal/domain/key"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantCode int
		wantOut  string
		wantErr  string
	}{
		{"version", []string{"version"}, 0, "relay " + version + "\n", ""},
		{"no args", nil, 2, "", "Usage:"},
		{"unknown command", []string{"bogus"}, 2, "", "unknown command"},
		{"key without subcommand", []string{"key"}, 2, "", "usage: relay key new"},
		{"key with wrong subcommand", []string{"key", "old"}, 2, "", "usage: relay key new"},
		{"help", []string{"help"}, 0, "Usage:", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var stdout, stderr bytes.Buffer
			code := run(tt.args, &stdout, &stderr)
			if code != tt.wantCode {
				t.Errorf("run(%v) exit code = %d; want %d", tt.args, code, tt.wantCode)
			}
			if !strings.Contains(stdout.String(), tt.wantOut) {
				t.Errorf("run(%v) stdout = %q; want it to contain %q", tt.args, stdout.String(), tt.wantOut)
			}
			if !strings.Contains(stderr.String(), tt.wantErr) {
				t.Errorf("run(%v) stderr = %q; want it to contain %q", tt.args, stderr.String(), tt.wantErr)
			}
		})
	}
}

func TestRunKeyNew(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"key", "new"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run(key new) exit code = %d; want 0 (stderr: %s)", code, stderr.String())
	}
	k, err := key.Load(stdout.Bytes())
	if err != nil {
		t.Fatalf("key new output is not a valid key file: %v\n%s", err, stdout.String())
	}
	if !k.HasSecret() {
		t.Error("generated key has no secret half")
	}
	if strings.Contains(stdout.String(), k.Fingerprint()) != true {
		t.Error("key file does not contain its fingerprint")
	}
}

func TestRunKeyNewToFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "deed")
	var stdout, stderr bytes.Buffer
	code := run([]string{"key", "new", "-o", path}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run(key new -o) exit code = %d; want 0 (stderr: %s)", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Errorf("key file leaked to stdout: %q", stdout.String())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	k, err := key.Load(data)
	if err != nil {
		t.Fatalf("key file on disk is not valid: %v", err)
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := fi.Mode().Perm(); perm != 0o600 {
		t.Errorf("key file mode = %o; want 0600", perm)
	}
	if !strings.Contains(stderr.String(), k.Fingerprint()) {
		t.Errorf("stderr = %q; want it to mention the fingerprint", stderr.String())
	}
}

func TestRunKeyNewBadFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"key", "new", "--bogus"}, &stdout, &stderr)
	if code != 2 {
		t.Errorf("run(key new --bogus) exit code = %d; want 2", code)
	}
	if stdout.Len() != 0 {
		t.Errorf("key file leaked to stdout on bad flag: %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "flag provided but not defined") {
		t.Errorf("stderr = %q; want flag error message", stderr.String())
	}
}
