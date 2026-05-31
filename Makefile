SHELL := /bin/zsh

GO ?= go
GOCACHE_DIR := /Users/platojobs/Desktop/Github/flutterffi/pfGoPlus/.gocache
GOMODCACHE_DIR := /Users/platojobs/Desktop/Github/flutterffi/pfGoPlus/.gomodcache
GO_RUN = GOCACHE=$(GOCACHE_DIR) GOMODCACHE=$(GOMODCACHE_DIR) $(GO)

.PHONY: run test tidy

run:
	$(GO_RUN) run ./cmd/server

test:
	$(GO_RUN) test ./...

tidy:
	$(GO_RUN) mod tidy
