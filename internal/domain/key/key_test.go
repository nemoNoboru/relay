package key

import (
	"crypto/ed25519"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testSeed is a fixed seed so fingerprint and key-file tests are
// deterministic.
var testSeed = []byte("relay-test-seed-0123456789abcdef")

// wantFingerprint is the golden fingerprint of testSeed's public half,
// kept as a regression vector.
const wantFingerprint = "612375dfc54fb507"

func TestFingerprint(t *testing.T) {
	tests := []struct {
		name string
		seed []byte
		want string
	}{
		{"golden vector", testSeed, wantFingerprint},
		{"deterministic", testSeed, wantFingerprint},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			k := fromSeed(tt.seed)
			if got := k.Fingerprint(); got != tt.want {
				t.Errorf("Fingerprint() = %q; want %q", got, tt.want)
			}
		})
	}

	// Different secrets give different fingerprints.
	other := make([]byte, ed25519.SeedSize)
	other[0] = testSeed[0] ^ 1
	if fromSeed(other).Fingerprint() == fromSeed(testSeed).Fingerprint() {
		t.Error("distinct keys produced the same fingerprint")
	}
}

func TestSignVerify(t *testing.T) {
	k := fromSeed(testSeed)
	msg := []byte("hello, relay")
	sig := k.Sign(msg)

	tests := []struct {
		name string
		pub  ed25519.PublicKey
		msg  []byte
		sig  []byte
		want bool
	}{
		{"valid", k.Public(), msg, sig, true},
		{"tampered message", k.Public(), []byte("hello, relaly"), sig, false},
		{"wrong key", fromSeed(otherSeed()).Public(), msg, sig, false},
		{"empty message", k.Public(), []byte{}, k.Sign([]byte{}), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Verify(tt.pub, tt.msg, tt.sig); got != tt.want {
				t.Errorf("Verify() = %v; want %v", got, tt.want)
			}
		})
	}
}

func otherSeed() []byte {
	seed := make([]byte, ed25519.SeedSize)
	seed[31] = 1
	return seed
}

func TestLoad(t *testing.T) {
	k := fromSeed(testSeed)
	file, err := k.File()
	if err != nil {
		t.Fatalf("File(): %v", err)
	}
	valid := string(file)

	corruptFingerprint := strings.Replace(valid, "fingerprint "+wantFingerprint, "fingerprint 0000000000000000", 1)
	corruptSecret := strings.Replace(valid, "secret ", "secret AAAA", 1) // valid base64, wrong length

	tests := []struct {
		name    string
		data    string
		wantErr error
	}{
		{"valid", valid, nil},
		{"trailing newline tolerated", valid + "\n", nil},
		{"unknown version", strings.Replace(valid, "relay key v1", "relay key v2", 1), ErrInvalidFile},
		{"wrong magic", strings.Replace(valid, "relay key", "relay token", 1), ErrInvalidFile},
		{"tampered fingerprint", corruptFingerprint, ErrInvalidFile},
		{"uppercase fingerprint", strings.Replace(valid, "fingerprint "+wantFingerprint, "fingerprint "+strings.ToUpper(wantFingerprint), 1), ErrInvalidFile},
		{"bad base64 secret", corruptSecret, ErrInvalidFile},
		{"short seed", strings.Replace(valid, "secret ", "secret YQ==", 1), ErrInvalidFile},
		{"missing line", "relay key v1\nfingerprint " + wantFingerprint, ErrInvalidFile},
		{"extra line", valid + "bogus\n", ErrInvalidFile},
		{"empty", "", ErrInvalidFile},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Load([]byte(tt.data))
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Load() error = %v; want nil", err)
				}
				if got.Fingerprint() != wantFingerprint {
					t.Errorf("Load() fingerprint = %q; want %q", got.Fingerprint(), wantFingerprint)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Load() error = %v; want %v", err, tt.wantErr)
			}
		})
	}
}

// TestKeyFileRoundTrip is the boundary test: the key file crosses a
// real file, so write it to a temp dir and read it back.
func TestKeyFileRoundTrip(t *testing.T) {
	k, err := New()
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	path := filepath.Join(t.TempDir(), "deed")
	file, err := k.File()
	if err != nil {
		t.Fatalf("File(): %v", err)
	}
	if err := os.WriteFile(path, file, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	loaded, err := Load(data)
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	if loaded.Fingerprint() != k.Fingerprint() {
		t.Errorf("round trip fingerprint = %q; want %q", loaded.Fingerprint(), k.Fingerprint())
	}
	if !loaded.HasSecret() {
		t.Error("loaded key has no secret half")
	}
}

func TestNew(t *testing.T) {
	a, err := New()
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	b, err := New()
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if !a.HasSecret() || !b.HasSecret() {
		t.Fatal("New() keys must hold the secret half")
	}
	if len(a.Public()) != ed25519.PublicKeySize {
		t.Errorf("Public() length = %d; want %d", len(a.Public()), ed25519.PublicKeySize)
	}
	if a.Fingerprint() == b.Fingerprint() {
		t.Error("two fresh keys share a fingerprint")
	}
}

func TestFileWithoutSecret(t *testing.T) {
	k := Key{}
	if _, err := k.File(); err == nil {
		t.Error("File() on a key without a secret half succeeded")
	}
}

// ExampleNew shows the intended usage sequence: generate, display the
// fingerprint, sign, verify.
func ExampleNew() {
	k, err := New()
	if err != nil {
		panic(err)
	}
	_ = k.Fingerprint()
	sig := k.Sign([]byte("hello, relay"))
	_ = Verify(k.Public(), []byte("hello, relay"), sig)
}
