package main

import (
	"bytes"
	"strings"
	"testing"
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
