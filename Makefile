# Loy - Build-time Go Developer Platform
# Makefile for local development and CI pipelines

SHELL := /bin/bash
GO ?= go
BINARY_NAME := bin/loy
MODULE := github.com/loy-go/loy

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -s -w \
	-X $(MODULE)/internal/version.Version=$(VERSION) \
	-X $(MODULE)/internal/version.Commit=$(COMMIT) \
	-X $(MODULE)/internal/version.Date=$(DATE)

.PHONY: all build test test-all test-cover lint vet check check-docs clean golden help

all: check build

## build: Compile stripped, trimpathed binary to bin/loy
build:
	@mkdir -p bin
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/loy

## test: Run fast unit tests with race detector (< 5s)
test:
	$(GO) test -v -short -race ./...

## test-all: Run all tests including slow integration tests
test-all:
	$(GO) test -v -race ./...

## test-cover: Run tests with coverage profile output
test-cover:
	@mkdir -p coverage
	$(GO) test -v -race -coverprofile=coverage/coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "Coverage report saved to coverage/coverage.html"

## lint: Run golangci-lint across the codebase
lint:
	golangci-lint run ./...

## vet: Run go vet across all packages
vet:
	$(GO) vet ./...

## check-docs: Verify CLI documentation synchronization and link integrity
check-docs:
	@echo "Verifying CLI documentation synchronization..."
	@$(GO) test -v ./cmd/docgen -run TestDocumentationSyncGate
	@echo "Verifying documentation link integrity..."
	@python3 scripts/check_docs_links.py
	@echo "Documentation gates verified cleanly."

## check: Run full pre-commit verification gate (vet, lint, short tests, doc sync)
check: vet lint test check-docs

## golden: Update generator golden fixtures after intentional template modifications
golden:
	$(GO) test -v ./internal/generator/... -update

## docgen: Generate markdown reference pages from Cobra CLI into website/
docgen:
	$(GO) run ./cmd/docgen --out website/src/content/docs/reference/cli

## docs-dev: Start local Astro Starlight documentation development server
docs-dev: docgen
	cd website && pnpm dev

## docs-build: Build static documentation site with Pagefind search index
docs-build: docgen
	cd website && pnpm build

## clean: Remove built binaries and test coverage files
clean:
	rm -rf bin/ coverage/

## help: Display available targets
help:
	@echo "Loy Developer Platform Makefile"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed -e 's/## //' | awk 'BEGIN {FS = ": "}; {printf "  %-12s %s\n", $$1, $$2}'
