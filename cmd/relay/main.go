// Command relay is the Relay node: one binary that runs as either a
// public or a private node. This entry point holds flags, wiring, and
// exit codes only; the rules live in internal.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/nemoNoboru/relay/internal/domain/key"
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

func usage(w io.Writer) {
	fmt.Fprint(w, `Usage:
  relay version          print the version
  relay key new [-o f]   generate a new ownership key

The ownership key is the deed to a Relay address: guard it like a
password. Without -o the key file is written to stdout.
`)
}
