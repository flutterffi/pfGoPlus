SHELL := /bin/zsh

GO ?= go
GOCACHE_DIR := /Users/platojobs/Desktop/Github/flutterffi/pfGoPlus/.gocache
GOMODCACHE_DIR := /Users/platojobs/Desktop/Github/flutterffi/pfGoPlus/.gomodcache
BUF_CACHE_DIR := /Users/platojobs/Desktop/Github/flutterffi/pfGoPlus/.bufcache
GO_RUN = GOCACHE=$(GOCACHE_DIR) GOMODCACHE=$(GOMODCACHE_DIR) $(GO)

.PHONY: run run-grpc test tidy proto wire wire-install

run:
	$(GO_RUN) run ./cmd/server

run-grpc:
	$(GO_RUN) run ./cmd/grpcserver

test:
	$(GO_RUN) test ./...

tidy:
	$(GO_RUN) mod tidy

proto:
	BUF_CACHE_DIR=$(BUF_CACHE_DIR) ./bin/buf generate

wire-install:
	GOBIN=/Users/platojobs/Desktop/Github/flutterffi/pfGoPlus/bin $(GO_RUN) install github.com/google/wire/cmd/wire@v0.7.0

wire:
	GOCACHE=$(GOCACHE_DIR) GOMODCACHE=$(GOMODCACHE_DIR) ./bin/wire ./internal/bootstrap
