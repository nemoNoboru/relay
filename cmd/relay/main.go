// Command relay is the Relay node: one binary that serves this node's
// sites, accepts site uploads, and hosts the owner UI. This file holds
// the command dispatch; key.go and node.go hold the two subcommands.
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

func usage(w io.Writer) {
	fmt.Fprint(w, `Usage:
  relay version          print the version
  relay key new [-o f]   generate a new ownership key
  relay node --name <n> --data <d> --listen <addr> [--admin-token <t>]
                         serve <d>/sites/ at /<n>/<site>/ and accept site
                         uploads at /_relay/v1/sites/<site>/deploy

The ownership key is the deed to a Relay address: guard it like a password.
The admin token authorizes the operator console (/_relay/admin.html); without
--admin-token it is generated once into <data>/admin-token (0600) and printed.
Users upload with per-user deploy keys created on the operator console.
`)
}
