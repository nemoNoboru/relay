# Relay — build and quality gates.
#
# The Go node binary embeds the built owner UI (web/dist, see web/embed.go),
# so the frontend must be built before `go build`.
#
# Note: staticcheck v0.8.1 reads standard-library export data, so run it
# with the Go toolchain pinned in go.mod (1.26), not a newer major.

GO ?= go
NODE ?= my-group
DATA ?= ./data
LISTEN ?= :8080
RELAY ?= ./relay

.PHONY: check web web-deps build dev run web-check go-check test fmt vet staticcheck clean

# Build the owner UI (web/) into web/dist/.
# --include=dev: some environments set NODE_ENV=production, which would
# otherwise make npm skip the dev dependencies (vite/tsc/vitest).
web:
	@if [ ! -d web/node_modules ]; then \
		echo "installing web dependencies"; \
		(cd web && npm ci --include=dev); \
	fi
	cd web && npm run build
	# Vite's emptyOutDir removes it; restore the embed placeholder so the
	# Go embed pattern stays satisfiable even from a bare checkout + build.
	touch web/dist/.gitkeep

# Install the web dependencies explicitly. `make web` does this on demand.
web-deps:
	cd web && npm ci --include=dev

# Full: frontend, then the Go binary.
build: web
	$(GO) build -o relay ./cmd/relay

# Build the web UI and binary, then run a local node.
#   make dev
#   make dev NODE=my-group DATA=./data LISTEN=:8080
dev: build
	@mkdir -p $(DATA)
	$(RELAY) node --name $(NODE) --data $(DATA) --listen $(LISTEN)

# Run the already-built binary without rebuilding.
run:
	@test -x $(RELAY) || { echo "no $(RELAY); run 'make build' first"; exit 1; }
	@mkdir -p $(DATA)
	$(RELAY) node --name $(NODE) --data $(DATA) --listen $(LISTEN)

# Owner UI: typecheck (tsc + svelte-check), production build, unit + SSR
# tests, and the bundle smoke test — all from a clean npm install.
web-check:
	cd web && npm ci --include=dev && npm run build && npm test && npm run smoke

# Go gates: gofmt + vet + staticcheck + test -race. The full gate runs in
# CI (.github/workflows/test.yml); pre-commit runs a fast subset before
# each commit (.pre-commit-config.yaml).
go-check:
	@test -z "$$(gofmt -l .)" || { echo "gofmt: unformatted files:"; gofmt -l .; exit 1; }
	$(GO) vet ./...
	$(GO) run honnef.co/go/tools/cmd/staticcheck@v0.8.1 ./...
	$(GO) test -race ./...

# CI-style: everything that can be checked without a node binary.
check: web-check go-check

test: ## go test -race ./...
	$(GO) test -race ./...

fmt: ## gofmt -w .
	gofmt -w .

vet: ## go vet ./...
	$(GO) vet ./...

staticcheck: ## staticcheck ./...
	$(GO) run honnef.co/go/tools/cmd/staticcheck@v0.8.1 ./...

clean:
	rm -f relay
