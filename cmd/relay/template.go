package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/nemoNoboru/relay/internal/serve"
)

// runTemplate manages installed post templates. Built-ins are embedded in the
// binary; an installed template with the same id shadows one.
func runTemplate(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		templateUsage(stderr)
		return 2
	}
	switch args[0] {
	case "list":
		fs := flag.NewFlagSet("template list", flag.ContinueOnError)
		fs.SetOutput(stderr)
		data := fs.String("data", "", "data directory (required)")
		if err := fs.Parse(args[1:]); err != nil || *data == "" || fs.NArg() != 0 {
			templateUsage(stderr)
			return 2
		}
		list, skipped := serve.NewTemplateStore(*data).List()
		for _, s := range skipped {
			fmt.Fprintf(stderr, "relay: skipping template: %s\n", s)
		}
		for _, t := range list {
			kind := "installed"
			if t.Builtin {
				kind = "builtin"
			}
			fmt.Fprintf(stdout, "%-10s %-8s %s\n", t.ID, kind, t.Name)
		}
		return 0
	case "install":
		fs := flag.NewFlagSet("template install", flag.ContinueOnError)
		fs.SetOutput(stderr)
		data := fs.String("data", "", "data directory (required)")
		if err := fs.Parse(args[1:]); err != nil || *data == "" || fs.NArg() != 1 {
			templateUsage(stderr)
			return 2
		}
		id, err := serve.NewTemplateStore(*data).Install(fs.Arg(0))
		if err != nil {
			fmt.Fprintf(stderr, "relay: install template: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "installed template %s\n", id)
		return 0
	case "remove":
		fs := flag.NewFlagSet("template remove", flag.ContinueOnError)
		fs.SetOutput(stderr)
		data := fs.String("data", "", "data directory (required)")
		if err := fs.Parse(args[1:]); err != nil || *data == "" || fs.NArg() != 1 {
			templateUsage(stderr)
			return 2
		}
		if err := serve.NewTemplateStore(*data).Remove(fs.Arg(0)); err != nil {
			fmt.Fprintf(stderr, "relay: remove template: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "removed template %s\n", fs.Arg(0))
		return 0
	default:
		templateUsage(stderr)
		return 2
	}
}

func templateUsage(w io.Writer) {
	fmt.Fprint(w, `Usage:
  relay template list --data <dir>
  relay template install --data <dir> <folder>
  relay template remove --data <dir> <id>

A template folder needs relay-template.json, prepublish.html, and publish.js.
Built-in templates (note, link, gallery, article, songs, videos, carousel,
jetset, longform) live in the binary; an
installed template with the same id shadows one.
`)
}
