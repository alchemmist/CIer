.DEFAULT_GOAL := help

SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

BINARY := cier
PKG := ./cmd/cier
GOBIN := $(shell go env GOPATH)/bin
GOFMT_PATHS := ./cmd ./internal
STATICCHECK := $(GOBIN)/staticcheck
GO_PACKAGES := ./...
GOLANGCI_LINT := $(GOBIN)/golangci-lint

.PHONY: help check build test test-race test-cov fmt fmt-check vet staticcheck golangci-lint lint tidy install clean

help:
	@printf "Available targets:\n\n"
	@printf "  check         - run fmt-check, lint, test, build\n"
	@printf "  build         - build binary into ./bin\n"
	@printf "  test          - run all tests\n"
	@printf "  test-race     - run tests with race detector\n"
	@printf "  test-cov      - run tests with coverage profile\n"
	@printf "  fmt           - format Go sources with gofmt\n"
	@printf "  fmt-check     - verify Go sources are formatted\n"
	@printf "  vet           - run go vet\n"
	@printf "  staticcheck   - run staticcheck\n"
	@printf "  golangci-lint - run golangci-lint\n"
	@printf "  lint          - run vet + staticcheck + golangci-lint\n"
	@printf "  tidy          - run go mod tidy\n"
	@printf "  install       - install CLI with go install\n"
	@printf "  clean         - remove build artifacts\n"

check: fmt-check lint test build

build:
	go build -o bin/$(BINARY) ./cmd/$(BINARY)

tidy:
	go mod tidy

fmt:
	gofmt -w $(GOFMT_PATHS)

fmt-check:
	@unformatted="$$(gofmt -l $(GOFMT_PATHS))"; \
	if [ -n "$$unformatted" ]; then \
		echo "gofmt required for:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

vet:
	go vet ./...

test:
	go test $(GO_PACKAGES)

test-race:
	go test -race $(GO_PACKAGES)

test-cov:
	go test ./... -coverprofile=./coverage.out -covermode=atomic -coverpkg=./...

golangci-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOLANGCI_LINT) run ./...

lint: fmt-check vet golangci-lint

clean:
	rm -rf bin dist coverage.out

install: build
	go install ./cmd/$(BINARY)

staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	$(STATICCHECK) $(GO_PACKAGES)
