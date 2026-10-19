package main

import (
	"context"
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
	"github.com/nemoNoboru/relay/internal/social"
)

// runNode serves this node's sites and accepts site uploads.
func runNode(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("node", flag.ContinueOnError)
	flags.SetOutput(stderr)
	name := flags.String("name", "", "node name (required; [a-z0-9-], max 63 chars)")
	data := flags.String("data", "", "data directory containing sites/ (required)")
	listen := flags.String("listen", "", "address to listen on, e.g. :8080 (required)")
	adminToken := flags.String("admin-token", "", "admin token for the operator console (default: generated into <data>/admin-token)")
	tlsCert := flags.String("tls-cert", "", "TLS certificate file (serve HTTPS; requires --tls-key)")
	tlsKey := flags.String("tls-key", "", "TLS private key file (requires --tls-cert)")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintf(stderr, "relay: usage: relay node --name <n> --data <d> --listen <addr> [--admin-token <t>]\n")
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
	if (*tlsCert == "") != (*tlsKey == "") {
		fmt.Fprintf(stderr, "relay: --tls-cert and --tls-key must be set together\n")
		return 2
	}
	node, err := domain.ParseName(*name)
	if err != nil {
		fmt.Fprintf(stderr, "relay: invalid --name: %v\n", err)
		return 2
	}

	logger := slog.New(slog.NewJSONHandler(stderr, nil))
	token, err := resolveAdminToken(*data, *adminToken, logger)
	if err != nil {
		fmt.Fprintf(stderr, "relay: admin token: %v\n", err)
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

	keys, err := serve.NewKeyStore(*data)
	if err != nil {
		fmt.Fprintf(stderr, "relay: deploy keys: %v\n", err)
		return 1
	}

	versions := serve.NewVersions(*data)
	posts, err := social.NewStore(*data)
	if err != nil {
		fmt.Fprintf(stderr, "relay: post index: %v\n", err)
		return 1
	}
	comments, err := social.NewCommentStore(*data)
	if err != nil {
		fmt.Fprintf(stderr, "relay: comments: %v\n", err)
		return 1
	}
	reload := func() error {
		_, err := registry.Reload()
		return err
	}
	uploader := serve.NewUploader(node, *data, keys, versions, posts, serve.MaxDeployBytes, reload)
	tokenPath := filepath.Join(*data, "admin-token")
	reindex := func(name domain.Name) error {
		return serve.IndexSite(posts, *data, name)
	}
	admin := serve.NewAdmin(token, keys, versions, comments, reload, reindex, func(t string) error {
		return writeTokenFile(tokenPath, t)
	})
	ui, err := serve.UIHandler()
	if err != nil {
		fmt.Fprintf(stderr, "relay: serving owner UI: %v\n", err)
		return 1
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthz)
	mux.Handle("/_relay/v1/admin", admin)
	mux.Handle("/_relay/v1/admin/", admin)
	mux.Handle("/_relay/v1/invites", serve.NewInvites(keys))
	session := serve.NewSession(node, keys)
	mux.Handle("/_relay/v1/session", session)
	mux.Handle("/_relay/v1/me", session)
	postIndex := serve.NewPosts(node, keys, posts, comments)
	commentAPI := serve.NewComments(keys, posts, comments)
	templateStore := serve.NewTemplateStore(*data)
	templateAPI := serve.NewTemplates(templateStore, keys)
	publisher := serve.NewPublisher(node, uploader, templateStore, keys)
	mux.Handle("/_relay/v1/feed", postIndex)
	mux.Handle("/_relay/v1/people/", postIndex)
	mux.Handle("/_relay/v1/posts/", serve.NewPostSurface(postIndex, commentAPI))
	mux.Handle("/_relay/v1/comments/", commentAPI)
	mux.Handle("/_relay/v1/templates", templateAPI)
	mux.Handle("/_relay/v1/publish", publisher)
	mux.Handle("/_relay/templates/", templateAPI)
	mux.Handle("/_relay/v1/", uploader)
	mux.Handle("/_relay/", ui)
	// The node root serves the embedded landing page and its companions;
	// everything else (sites at /<node>/<site>/...) goes to the registry.
	mux.Handle("/", serve.Landing(registry.Handler()))

	logger.Info("relay node listening", "node", string(node), "addr", *listen, "tls", *tlsCert != "")
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
		if *tlsCert != "" {
			errCh <- srv.ServeTLS(ln, *tlsCert, *tlsKey)
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

// healthz reports liveness and readiness for Kubernetes probes: 200 "ok"
// whenever the node process is serving. It depends on nothing but the
// listener and mux, so a successful probe means the node is up.
func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "ok\n")
}

// resolveAdminToken returns the node's admin token — the operator's master
// credential. An explicit flag value wins and is persisted; otherwise the
// token is read from <dataDir>/admin-token, falling back to
// serve.DefaultAdminToken so a node needs no setup to sign in to. The
// console then refuses to go further until the default is replaced.
func resolveAdminToken(dataDir, explicit string, logger *slog.Logger) (string, error) {
	path := filepath.Join(dataDir, "admin-token")
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
	if err := writeTokenFile(path, serve.DefaultAdminToken); err != nil {
		return "", err
	}
	logger.Warn("no admin token set; using the default — change it at /_relay/admin.html",
		"token", serve.DefaultAdminToken, "file", path)
	return serve.DefaultAdminToken, nil
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
