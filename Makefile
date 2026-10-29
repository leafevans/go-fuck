# Makefile for go-fuck.
#
# The recipes only invoke the Go toolchain and the small helper programs in
# tools/, so they work on Linux, macOS and Windows regardless of the shell that
# make uses. Helpers replace POSIX shell one-liners that are not portable to
# cmd.exe, for example "make fmt-check" runs tools/checkfmt instead of
# "test -z \"$(gofmt -l .)\"", and "make clean" runs tools/clean instead of
# "rm -rf".

GO            ?= go
BINARY        ?= go-fuck
GOLANGCI_LINT ?= golangci-lint

# Inputs for "make packaging": the release version (without the leading v) and
# the path of the checksums.txt asset of that release.
VERSION   ?=
CHECKSUMS ?= dist/checksums.txt

# Add an .exe suffix on Windows so that "make build" produces a file the
# system can start directly.
ifeq ($(OS),Windows_NT)
EXE := .exe
endif

# Extra linker flags, for example:
#
#	make build LDFLAGS="-X github.com/leafevans/go-fuck/internal/consts.Version=1.2.3"
LDFLAGS ?=

.PHONY: help build test test-race cover bench vet fmt fmt-check lint install clean packaging packaging-check rules-list rules-list-check

.DEFAULT_GOAL := help

help: ## Print this help
	@$(GO) run ./tools/help

build: ## Build the go-fuck binary for the current platform
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY)$(EXE) ./cmd/go-fuck

test: ## Run the test suite
	$(GO) test ./... -count=1

test-race: ## Run the test suite with the race detector enabled
	$(GO) test -race ./... -count=1

cover: ## Run the test suite with coverage and print the total
	$(GO) test ./... -count=1 -coverprofile=coverage.out
	@$(GO) run ./tools/coveragecheck -profile coverage.out

bench: ## Run the benchmarks
	$(GO) test ./internal/... -run '^$$' -bench . -benchmem

vet: ## Run go vet on all packages
	$(GO) vet ./...

fmt: ## Format all Go files with gofmt -s
	gofmt -s -w .

fmt-check: ## Fail when a Go file is not formatted
	@$(GO) run ./tools/checkfmt

lint: ## Run golangci-lint with the repository configuration
	$(GOLANGCI_LINT) run

install: ## Install the go-fuck command with go install
	$(GO) install ./cmd/go-fuck

clean: ## Remove the built binary and the GoReleaser output
	@$(GO) run ./tools/clean $(BINARY) $(BINARY).exe dist coverage.out

packaging: ## Regenerate the distribution manifests for VERSION from CHECKSUMS
	@$(GO) run ./tools/update-packaging -version $(VERSION) -checksums $(CHECKSUMS)

packaging-check: ## Validate the distribution manifests
	@$(GO) test ./packaging -count=1

rules-list: ## Regenerate the rule list in README.md
	@$(GO) run ./tools/gen-rule-list

rules-list-check: ## Fail when the README.md rule list is out of date
	@$(GO) run ./tools/gen-rule-list -check