// Command relay is the Relay node: one binary that runs as either a
// public or a private node. This entry point holds flags, wiring, and
// exit codes only; the rules live in internal.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nemoNoboru/relay/internal/domain"
	"github.com/nemoNoboru/relay/internal/domain/key"
	"github.com/nemoNoboru/relay/internal/serve"
)

// version is set at build time:
//
//	go build -ldflags "-X main.version=0.1.0"
var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes one relay invocation and returns its exit code.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stderr)
		return 2
	}
	switch args[0] {
	case "version":
		fmt.Fprintln(stdout, "relay", version)
		return 0
	case "key":
		return runKey(args[1:], stdout, stderr)
	case "node":
		return runNode(args[1:], stdout, stderr)
	case "help", "--help", "-h":
		usage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "relay: unknown command %q\n\n", args[0])
		usage(stderr)
		return 2
	}
}

// runKey handles the key subcommand.
func runKey(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 || args[0] != "new" {
		fmt.Fprintf(stderr, "relay: usage: relay key new [-o file]\n")
		return 2
	}
	fs := flag.NewFlagSet("key new", flag.ContinueOnError)
	fs.SetOutput(stderr)
	out := fs.String("o", "", "write the key file to this path instead of stdout")
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintf(stderr, "relay: usage: relay key new [-o file]\n")
		return 2
	}
	k, err := key.New()
	if err != nil {
		fmt.Fprintf(stderr, "relay: %v\n", err)
		return 1
	}
	file, err := k.File()
	if err != nil {
		fmt.Fprintf(stderr, "relay: %v\n", err)
		return 1
	}
	if *out == "" {
		if _, err := stdout.Write(file); err != nil {
			fmt.Fprintf(stderr, "relay: write key file: %v\n", err)
			return 1
		}
		return 0
	}
	if err := os.WriteFile(*out, file, 0o600); err != nil {
		fmt.Fprintf(stderr, "relay: write key file %q: %v\n", *out, err)
		return 1
	}
	fmt.Fprintf(stderr, "relay: ownership key written to %s (fingerprint %s)\n", *out, k.Fingerprint())
	return 0
}

// runNode handles the node subcommand: serve the site folders under
// <data>/sites/ at /<name>/<site>/ on a plain HTTP listener.
func runNode(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("node", flag.ContinueOnError)
	flags.SetOutput(stderr)
	name := flags.String("name", "", "node name (required; [a-z0-9-], max 63 chars)")
	data := flags.String("data", "", "data directory containing sites/ (required)")
	listen := flags.String("listen", "", "address to listen on, e.g. :8080 (required)")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintf(stderr, "relay: usage: relay node --name <n> --data <d> --listen <addr>\n")
		return 2
	}
	for _, f := range []struct {
		flag  string
		value string
	}{
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
	sites, warnings, err := serve.LoadSites(*data)
	if err != nil {
		fmt.Fprintf(stderr, "relay: loading sites: %v\n", err)
		return 1
	}
	logger := slog.New(slog.NewJSONHandler(stderr, nil))
	for _, w := range warnings {
		logger.Warn("skipping site", "reason", w)
	}
	logger.Info("relay node listening", "node", string(node), "addr", *listen, "sites", len(sites))
	srv := &http.Server{
		Addr:              *listen,
		Handler:           serve.NodeHandler(node, sites),
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

func usage(w io.Writer) {
	fmt.Fprint(w, `Usage:
  relay version          print the version
  relay key new [-o f]   generate a new ownership key
  relay node --name <n> --data <d> --listen <addr>
                         serve the site folders under <d>/sites/ at
                         /<n>/<site>/ on a plain HTTP listener

The ownership key is the deed to a Relay address: guard it like a
password. Without -o the key file is written to stdout.
`)
}
