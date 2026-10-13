// Command relay runs a relay node. In this slice it only provides static
// hosting: site folders under <data>/sites/ are served at
// /<node>/<site>/... on a plain HTTP listener.
package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"relay/internal/domain"
	"relay/internal/serve"
)

// drainTimeout bounds graceful shutdown (architecture 2.1).
const drainTimeout = 30 * time.Second

func main() {
	os.Exit(run(slog.New(slog.NewJSONHandler(os.Stderr, nil)), os.Args[1:]))
}

// run executes the node and returns the process exit code. Config errors
// exit 2 before any socket is opened.
func run(logger *slog.Logger, args []string) int {
	flags := flag.NewFlagSet("relay node", flag.ContinueOnError)
	name := flags.String("name", "", "node name (required, [a-z0-9-])")
	data := flags.String("data", "./data", "data directory containing sites/")
	listen := flags.String("listen", ":8080", "address to listen on")
	if len(args) > 0 && args[0] == "node" { // `relay node --name ...`
		args = args[1:]
	}
	if err := flags.Parse(args); err != nil {
		return 2 // usage error; the flag package already printed it
	}
	if *name == "" {
		logger.Error("missing required flag --name")
		return 2
	}
	node, err := domain.ParseName(*name)
	if err != nil {
		logger.Error("invalid --name", "err", err)
		return 2
	}
	sites, skipped, err := serve.LoadSites(*data)
	if err != nil {
		logger.Error("loading sites", "err", err)
		return 1
	}
	for _, s := range skipped {
		logger.Warn("skipping invalid site entry", "name", s)
	}
	if len(sites) == 0 {
		logger.Warn("no sites found; every request will 404", "data", *data)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{Addr: *listen, Handler: serve.NodeHandler(node, sites)}
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	logger.Info("relay node listening", "node", string(node), "addr", *listen, "sites", len(sites))
	select {
	case <-ctx.Done():
		stop()
		logger.Info("shutting down", "drain", drainTimeout.String())
		shutdownCtx, cancel := context.WithTimeout(context.Background(), drainTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful shutdown failed", "err", err)
			return 1
		}
		return 0
	case err := <-errCh:
		logger.Error("listener failed", "err", err)
		return 1
	}
}
