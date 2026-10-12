# Quality gates for the Relay Go module. The full gate runs in CI
# (.github/workflows/test.yml); pre-commit runs a fast subset before
# each commit (.pre-commit-config.yaml).
#
# Note: staticcheck v0.8.1 reads standard-library export data, so run it
# with the Go toolchain pinned in go.mod (1.26), not a newer major.

GO ?= go

.PHONY: check test fmt vet staticcheck build clean

check: ## the full gate: gofmt + vet + staticcheck + test -race
	@test -z "$$(gofmt -l .)" || { echo "gofmt: unformatted files:"; gofmt -l .; exit 1; }
	$(GO) vet ./...
	$(GO) run honnef.co/go/tools/cmd/staticcheck@v0.8.1 ./...
	$(GO) test -race ./...

test: ## go test -race ./...
	$(GO) test -race ./...

fmt: ## gofmt -w .
	gofmt -w .

vet: ## go vet ./...
	$(GO) vet ./...

staticcheck: ## staticcheck ./...
	$(GO) run honnef.co/go/tools/cmd/staticcheck@v0.8.1 ./...

build: ## build the relay binary
	$(GO) build -o relay ./cmd/relay

clean:
	rm -f relay
