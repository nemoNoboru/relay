package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/nemoNoboru/relay/internal/domain"
	"github.com/nemoNoboru/relay/internal/serve"
)

// runNode serves this node's sites and accepts site uploads.
func runNode(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("node", flag.ContinueOnError)
	flags.SetOutput(stderr)
	name := flags.String("name", "", "node name (required; [a-z0-9-], max 63 chars)")
	data := flags.String("data", "", "data directory containing sites/ (required)")
	listen := flags.String("listen", "", "address to listen on, e.g. :8080 (required)")
	uploadToken := flags.String("upload-token", "", "deploy token for uploads (default: generated into <data>/upload-token)")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintf(stderr, "relay: usage: relay node --name <n> --data <d> --listen <addr> [--upload-token <t>]\n")
		return 2
	}
	for _, f := range []struct{ flag, value string }{
		{"--name", *name},
		{"--data", *data},
		{"--listen", *listen},
	} {
		if f.value == "" {
			fmt.Fprintf(stderr, "relay: missing required flag %s\n", f.flag)
			return 2
		}
	}
	node, err := domain.ParseName(*name)
	if err != nil {
		fmt.Fprintf(stderr, "relay: invalid --name: %v\n", err)
		return 2
	}

	logger := slog.New(slog.NewJSONHandler(stderr, nil))
	token, err := resolveUploadToken(*data, *uploadToken, logger)
	if err != nil {
		fmt.Fprintf(stderr, "relay: deploy token: %v\n", err)
		return 1
	}

	registry, skipped, err := serve.NewRegistry(node, *data)
	if err != nil {
		fmt.Fprintf(stderr, "relay: loading sites: %v\n", err)
		return 1
	}
	for _, s := range skipped {
		logger.Warn("skipping site", "reason", s)
	}

	uploader := serve.NewUploader(node, *data, token, serve.MaxDeployBytes, func() error {
		_, err := registry.Reload()
		return err
	})
	ui, err := serve.UIHandler()
	if err != nil {
		fmt.Fprintf(stderr, "relay: serving owner UI: %v\n", err)
		return 1
	}

	mux := http.NewServeMux()
	mux.Handle("/_relay/v1/", uploader)
	mux.Handle("/_relay/", ui)
	mux.Handle("/", registry.Handler())

	logger.Info("relay node listening", "node", string(node), "addr", *listen)
	srv := &http.Server{
		Addr:              *listen,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		var ln net.Listener
		if ln, err = net.Listen("tcp", *listen); err != nil {
			errCh <- fmt.Errorf("listen on %s: %w", *listen, err)
			return
		}
		errCh <- srv.Serve(ln)
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "err", err)
			return 1
		}
		return 0
	case <-ctx.Done():
	}
	logger.Info("shutting down", "drain", "30s")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
		return 1
	}
	return 0
}

// resolveUploadToken returns the node's deploy token. An explicit flag
// value wins and is persisted; otherwise the token is read from
// <dataDir>/upload-token, or generated and persisted (0600) if absent.
func resolveUploadToken(dataDir, explicit string, logger *slog.Logger) (string, error) {
	path := filepath.Join(dataDir, "upload-token")
	if explicit != "" {
		return explicit, writeTokenFile(path, explicit)
	}
	b, err := os.ReadFile(path)
	if err == nil {
		return strings.TrimSpace(string(b)), nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	token, err := randomToken()
	if err != nil {
		return "", err
	}
	if err := writeTokenFile(path, token); err != nil {
		return "", err
	}
	logger.Info("generated deploy token (shown once)", "token", token, "file", path)
	return token, nil
}

func writeTokenFile(path, token string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(token+"\n"), 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// randomToken returns 32 random bytes, base64url-encoded.
func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate deploy token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
