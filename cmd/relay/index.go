package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/nemoNoboru/relay/internal/serve"
	"github.com/nemoNoboru/relay/internal/social"
)

// runIndex backfills posts.json from the sites already on disk, reading each
// site's relay-post.json sidecar and falling back to the site name and the
// ownership map. It is idempotent: indexed sites are skipped and unattributed
// sites are reported, never guessed.
func runIndex(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("index", flag.ContinueOnError)
	flags.SetOutput(stderr)
	data := flags.String("data", "", "data directory containing sites/ (required)")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *data == "" || flags.NArg() != 0 {
		fmt.Fprintln(stderr, "relay: usage: relay index --data <dir>")
		return 2
	}
	ks, err := serve.NewKeyStore(*data)
	if err != nil {
		fmt.Fprintf(stderr, "relay: loading keys: %v\n", err)
		return 1
	}
	posts, err := social.NewStore(*data)
	if err != nil {
		fmt.Fprintf(stderr, "relay: loading post index: %v\n", err)
		return 1
	}

	indexed, skipped, err := serve.BackfillIndex(posts, ks, *data)
	if err != nil {
		fmt.Fprintf(stderr, "relay: backfill index: %v\n", err)
		return 1
	}
	for _, reason := range skipped {
		fmt.Fprintf(stderr, "relay: skipping %s\n", reason)
	}
	fmt.Fprintf(stdout, "indexed %d, skipped %d\n", indexed, len(skipped))
	return 0
}
