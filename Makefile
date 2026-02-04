SHELL := /bin/sh

BIN := cier
PKG := ./cmd/cier
GOCACHE_DIR := /tmp/go-cache

.PHONY: build tidy fmt vet test lint clean help

help:
	@echo "CIer Makefile"
	@echo "Targets:"
	@echo "  build  - Build the CLI binary"
	@echo "  tidy   - Run go mod tidy"
	@echo "  fmt    - Run gofmt on all Go files"
	@echo "  vet    - Run go vet"
	@echo "  test   - Run tests"
	@echo "  lint   - Run basic static checks"
	@echo "  clean  - Remove build artifacts"

build:
	GOCACHE=$(GOCACHE_DIR) go build -o $(BIN) $(PKG)

tidy:
	GOCACHE=$(GOCACHE_DIR) go mod tidy

fmt:
	gofmt -w $$(rg --files -g '*.go')

vet:
	GOCACHE=$(GOCACHE_DIR) go vet ./...

test:
	GOCACHE=$(GOCACHE_DIR) go test ./...

lint: fmt vet
	@echo "lint: ok"

clean:
	rm -f $(BIN)
