package main

import (
	"archive/tar"
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nemoNoboru/relay/internal/domain"
)

// runDeploy handles the deploy subcommand: archive a folder's contents and
// upload them to a node, replacing the site atomically.
func runDeploy(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("deploy", flag.ContinueOnError)
	flags.SetOutput(stderr)
	origin := flags.String("origin", "", "node origin, e.g. https://relay.example.org (required)")
	site := flags.String("site", "", "site name to deploy under (required)")
	token := flags.String("token", os.Getenv("RELAY_DEPLOY_TOKEN"), "deploy key token (default: $RELAY_DEPLOY_TOKEN)")
	insecure := flags.Bool("insecure", false, "skip TLS certificate verification (self-signed certs only)")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 1 {
		fmt.Fprintf(stderr, "relay: usage: relay deploy --origin <url> --site <name> [--token <t>] <folder>\n")
		return 2
	}
	for _, f := range []struct{ flag, value string }{
		{"--origin", *origin},
		{"--site", *site},
		{"--token", *token},
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

	body, err := tarDir(flags.Arg(0))
	if err != nil {
		fmt.Fprintf(stderr, "relay: %v\n", err)
		return 1
	}

	url := strings.TrimSuffix(*origin, "/") + "/_relay/v1/sites/" + *site + "/deploy"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(stderr, "relay: %v\n", err)
		return 1
	}
	req.Header.Set("Authorization", "Bearer "+*token)
	req.Header.Set("Content-Type", "application/x-tar")

	res, err := newHTTPClient(*insecure, 5*time.Minute).Do(req)
	if err != nil {
		fmt.Fprintf(stderr, "relay: deploy: %v\n", err)
		return 1
	}
	defer res.Body.Close()
	respBody, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		fmt.Fprintf(stderr, "relay: deploy failed: %s: %s\n", res.Status, strings.TrimSpace(string(respBody)))
		return 1
	}
	fmt.Fprintln(stdout, strings.TrimSpace(string(respBody)))
	return 0
}

// newHTTPClient returns a client with the given timeout, optionally skipping
// TLS verification (for talking to a node with a self-signed certificate).
func newHTTPClient(insecure bool, timeout time.Duration) *http.Client {
	if !insecure {
		return &http.Client{Timeout: timeout}
	}
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return &http.Client{Timeout: timeout, Transport: tr}
}

// tarDir returns a tar archive of the contents of root: regular files and
// directories only, with site-relative paths, no symlinks.
func tarDir(root string) ([]byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlinks are not supported: %s", p)
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil // the archive root itself is not an entry
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		var hdr *tar.Header
		switch {
		case d.IsDir():
			hdr = &tar.Header{Name: rel, Mode: 0o755, Typeflag: tar.TypeDir}
		case info.Mode().IsRegular():
			hdr = &tar.Header{Name: rel, Mode: 0o644, Size: info.Size(), Typeflag: tar.TypeReg}
		default:
			return fmt.Errorf("unsupported file type: %s", p)
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if hdr.Typeflag == tar.TypeReg {
			f, err := os.Open(p)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, f); err != nil {
				f.Close()
				return err
			}
			if err := f.Close(); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
