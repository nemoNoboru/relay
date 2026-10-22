package serve

import (
	"crypto/tls"
	"net/http/httptest"
	"testing"
)

func TestSameOrigin(t *testing.T) {
	tests := []struct {
		name   string
		origin string
		host   string
		tls    bool
		proto  string
		want   bool
	}{
		{name: "missing origin", want: true},
		{name: "matching http", origin: "http://relay.example.org", host: "relay.example.org", want: true},
		{name: "matching https", origin: "https://relay.example.org", host: "relay.example.org", tls: true, want: true},
		{name: "host mismatch", origin: "https://evil.example", host: "relay.example.org", want: false},
		{name: "scheme mismatch", origin: "https://relay.example.org", host: "relay.example.org", want: false},
		{name: "forwarded https", origin: "https://relay.example.org", host: "relay.example.org", proto: "https", want: true},
		{name: "garbage origin", origin: "://", host: "relay.example.org", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", nil)
			req.Host = tt.host
			req.TLS = nil
			if tt.tls {
				req.TLS = &tls.ConnectionState{}
			}
			if tt.proto != "" {
				req.Header.Set("X-Forwarded-Proto", tt.proto)
			}
			req.Header.Set("Origin", tt.origin)
			if got := sameOrigin(req); got != tt.want {
				t.Fatalf("sameOrigin(%q) = %v, want %v", tt.origin, got, tt.want)
			}
		})
	}
}
