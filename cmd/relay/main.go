// Command relay is the Relay node: one binary that serves this node's
// sites, accepts site uploads, and hosts the owner UI. This file holds the
// command dispatch and usage text; each other file holds one subcommand.
package main

import (
	"fmt"
	"io"
	"os"
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
	case "deploy":
		return runDeploy(args[1:], stdout, stderr)
	case "rollback":
		return runRollback(args[1:], stdout, stderr)
	case "index":
		return runIndex(args[1:], stdout, stderr)
	case "template":
		return runTemplate(args[1:], stdout, stderr)
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

func usage(w io.Writer) {
	fmt.Fprint(w, `Usage:
  relay version          print the version
  relay deploy --origin <url> --site <n> [--token <t>] [--insecure] <folder>
                         upload a folder as page <n> (replaces it atomically)
  relay rollback --origin <url> --site <n> [--admin-token <t>] [--to <v>] [--insecure]
                         roll site <n> back to a previous version
  relay index --data <d> backfill the post index from existing sites
  relay template list|install|remove --data <d> [folder|id]
                         manage installed post templates
  relay node --name <n> --data <d> --listen <addr> [--admin-token <t>]
                         [--tls-cert <f> --tls-key <f>]
                         serve <d>/sites/ at /<n>/<site>/ and accept site
                         uploads at /_relay/v1/sites/<site>/deploy

An invitation link is the member credential. The host makes the first ones on
the host console (/_relay/admin.html); members can invite people they trust.
The admin password authorizes the host console only. Without --admin-token it
defaults to "relay" and is written to <data>/admin-token (0600).
`)
}
