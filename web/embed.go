// Package webui embeds the built Relay owner UI (web/) into the node binary.
//
// The frontend is built by Vite into dist/ (see README.md in this directory).
// The node serves the embedded filesystem under /_relay/ (see
// internal/serve/ui.go); because Vite is configured with a relative base
// (base: "./"), the same dist/ works at any URL depth — Relay's addressing
// is path-based.
//
// Build order: the root Makefile builds the frontend before `go build`.
// dist/.gitkeep keeps the embed pattern satisfiable before the first
// frontend build; a real build replaces its contents (see .gitignore).
package webui

import "embed"

// FS is the built owner UI (Vite output), rooted at dist/.
//
//go:embed all:dist
var FS embed.FS
