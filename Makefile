SHELL := /bin/sh

BIN := cier
PKG := ./cmd/cier

.PHONY: build tidy fmt vet test lint clean help

help:
	@echo "CIer Makefile"
	@echo "Targets:"
	@echo "  build    - Build the CLI binary"
	@echo "  install  - Install to \\$GOPATH"
	@echo "  tidy     - Run go mod tidy"
	@echo "  fmt      - Run gofmt on all Go files"
	@echo "  vet      - Run go vet"
	@echo "  test     - Run tests"
	@echo "  lint     - Run basic static checks"
	@echo "  clean    - Remove build artifacts"

build:
	go build -o bin/$(BIN) ./cmd/$(BIN)

tidy:
	go mod tidy

fmt:
	gofmt -w $$(rg --files -g '*.go')

vet:
	go vet ./...

test:
	go test ./...

lint: fmt vet
	@echo "lint: ok"

clean:
	rm -rf bin dist

install:
	go install ./cmd/$(BIN)
