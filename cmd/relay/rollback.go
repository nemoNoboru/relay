package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nemoNoboru/relay/internal/domain"
)

// runRollback handles the rollback subcommand: re-point a site at one of its
// previous versions, using the node's admin token.
func runRollback(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("rollback", flag.ContinueOnError)
	flags.SetOutput(stderr)
	origin := flags.String("origin", "", "node origin, e.g. https://relay.example.org (required)")
	site := flags.String("site", "", "site name to roll back (required)")
	adminToken := flags.String("admin-token", os.Getenv("RELAY_ADMIN_TOKEN"), "admin token (default: $RELAY_ADMIN_TOKEN)")
	to := flags.String("to", "", "version id to roll back to (default: previous version)")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintf(stderr, "relay: usage: relay rollback --origin <url> --site <name> [--admin-token <t>] [--to <version>]\n")
		return 2
	}
	for _, f := range []struct{ flag, value string }{
		{"--origin", *origin},
		{"--site", *site},
		{"--admin-token", *adminToken},
	} {
		if f.value == "" {
			fmt.Fprintf(stderr, "relay: missing required flag %s\n", f.flag)
			return 2
		}
	}
	if _, err := domain.ParseName(*site); err != nil {
		fmt.Fprintf(stderr, "relay: invalid --site: %v\n", err)
		return 2
	}

	var body io.Reader
	if *to != "" {
		b, err := json.Marshal(map[string]string{"version": *to})
		if err != nil {
			fmt.Fprintf(stderr, "relay: %v\n", err)
			return 1
		}
		body = bytes.NewReader(b)
	} else {
		body = strings.NewReader("")
	}

	url := strings.TrimSuffix(*origin, "/") + "/_relay/v1/admin/sites/" + *site + "/rollback"
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		fmt.Fprintf(stderr, "relay: %v\n", err)
		return 1
	}
	req.Header.Set("Authorization", "Bearer "+*adminToken)
	if *to != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := (&http.Client{Timeout: time.Minute}).Do(req)
	if err != nil {
		fmt.Fprintf(stderr, "relay: rollback: %v\n", err)
		return 1
	}
	defer res.Body.Close()
	respBody, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		fmt.Fprintf(stderr, "relay: rollback failed: %s: %s\n", res.Status, strings.TrimSpace(string(respBody)))
		return 1
	}
	fmt.Fprintln(stdout, strings.TrimSpace(string(respBody)))
	return 0
}
