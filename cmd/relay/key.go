package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/nemoNoboru/relay/internal/domain/key"
)

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
