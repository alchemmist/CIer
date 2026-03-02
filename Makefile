SHELL := /bin/sh

BINARY := cier
PKG := ./cmd/cier
GOBIN := $(shell go env GOPATH)/bin
GOFMT_PATHS := ./cmd ./internal
STATICCHECK := $(GOBIN)/staticcheck
GO_PACKAGES := ./...
GOLANGCI_LINT := $(GOBIN)/golangci-lint

.PHONY: build tidy fmt vet test lint clean help staticcheck

help:
	@echo "CIer Makefile"
	@echo "Targets:"
	@echo "  build           - Build the CLI binary"
	@echo "  install         - Install to \\$GOPATH"
	@echo "  tidy            - Run go mod tidy"
	@echo "  fmt             - Run gofmt on all Go files"
	@echo "  vet             - Run go vet"
	@echo "  test            - Run tests"
	@echo "  lint            - Run basic static checks"
	@echo "  staticcheck     - Run basic static checks"
	@echo "  clean           - Remove build artifacts"

build:
	go build -o bin/$(BINARY) ./cmd/$(BINARY)

tidy:
	go mod tidy

fmt:
	gofmt -w $$(rg --files -g '*.go')

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

golangci-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOLANGCI_LINT) run ./...

lint: fmt-check vet golangci-lint

clean:
	rm -rf bin dist

install:
	go install ./cmd/$(BIN)

staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	$(STATICCHECK) $(GO_PACKAGES)
