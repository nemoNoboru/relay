# Relay — build and check targets.
#
# The Go node binary embeds the built owner UI (web/dist, see web/embed.go),
# so the frontend must be built before `go build`.

.PHONY: web build check site-check web-check go-check

# Build the owner UI (web/) into web/dist/.
# --include=dev: some environments set NODE_ENV=production, which would
# otherwise make npm skip the dev dependencies (vite/tsc/vitest).
web:
	cd web && npm ci --include=dev && npm run build
	# Vite's emptyOutDir removes it; restore the embed placeholder so the
	# Go embed pattern stays satisfiable even from a bare checkout + build.
	touch web/dist/.gitkeep

# Full: frontend, then the Go binary.
build: web
	go build ./...

# Static root site: the generated pages are in sync with their Markdown.
site-check:
	python3 tools/build.py --check

# Owner UI: typecheck (tsc + svelte-check), production build, unit + SSR
# tests, and the bundle smoke test — all from a clean npm install.
web-check:
	cd web && npm ci --include=dev && npm run build && npm test && npm run smoke

# Go gates: gofmt (must be clean), vet, and the race-enabled test suite.
# The race detector needs cgo, so without a C compiler it falls back to a
# plain test run (with a visible note).
go-check:
	@test -z "$$(gofmt -l .)" || { echo "gofmt -l is not clean:"; gofmt -l .; exit 1; }
	go vet ./...
	@if command -v gcc >/dev/null 2>&1 || command -v cc >/dev/null 2>&1 || command -v clang >/dev/null 2>&1; then \
		CGO_ENABLED=1 go test -race ./...; \
	else \
		echo "go-check: no C compiler found, running without -race"; \
		go test ./...; \
	fi

# CI-style: everything that can be checked without a node binary.
check: site-check web-check go-check