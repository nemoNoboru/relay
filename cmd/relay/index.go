package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nemoNoboru/relay/internal/domain"
	"github.com/nemoNoboru/relay/internal/serve"
	"github.com/nemoNoboru/relay/internal/social"
)

// runIndex backfills posts.json from the sites already on disk. It is
// idempotent: indexed sites are skipped. Sites without a known owner are
// left unlisted rather than guessed.
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
	entries, err := os.ReadDir(filepath.Join(*data, "sites"))
	if err != nil {
		fmt.Fprintf(stderr, "relay: reading sites: %v\n", err)
		return 1
	}

	indexed, skipped := 0, 0
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") || !entry.IsDir() {
			continue
		}
		if _, err := domain.ParseName(name); err != nil {
			skipped++
			continue
		}
		if _, ok := posts.Get(name); ok {
			continue
		}
		owner := ks.SiteUser(name)
		if owner == "" {
			fmt.Fprintf(stderr, "relay: skipping %s: no recorded owner\n", name)
			skipped++
			continue
		}
		member, ok := ks.MemberByUser(owner)
		if !ok {
			fmt.Fprintf(stderr, "relay: skipping %s: owner %q has no live key\n", name, owner)
			skipped++
			continue
		}
		info, err := entry.Info()
		if err != nil {
			fmt.Fprintf(stderr, "relay: skipping %s: %v\n", name, err)
			skipped++
			continue
		}
		post, err := posts.Upsert(social.Post{
			Site:        name,
			Member:      member.ID,
			Title:       social.TitleFromSite(name),
			Template:    "custom",
			PublishedAt: info.ModTime(),
		})
		if err != nil {
			fmt.Fprintf(stderr, "relay: indexing %s: %v\n", name, err)
			skipped++
			continue
		}
		fmt.Fprintf(stdout, "indexed %s as %s (%s)\n", post.Site, post.Title, member.DisplayName)
		indexed++
	}
	fmt.Fprintf(stdout, "indexed %d, skipped %d\n", indexed, skipped)
	return 0
}
