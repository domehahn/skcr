.PHONY: build test test-race test-fuzz test-interop vet lint check clean install release-snapshot

BINARY   := skcr
DIST_DIR := dist
MODULE   := github.com/domehahn/skcr/v2
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS  := -ldflags "-X $(MODULE)/internal/cli.Version=$(VERSION) -X $(MODULE)/internal/cli.Commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo none) -X $(MODULE)/internal/cli.Date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ) -s -w"

build:
	mkdir -p $(DIST_DIR)
	go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY) ./cmd/skcr

test:
	go test -mod=mod ./...

test-race:
	go test -mod=mod -race ./internal/compiler/... ./internal/skillmeta/... ./internal/validator/... ./internal/platforms/... ./internal/renderer/...

test-fuzz:
	go test -mod=mod -fuzz=FuzzCompileSkill -fuzztime=5s ./internal/compiler/...
	go test -mod=mod -fuzz=FuzzParseContract -fuzztime=5s ./internal/skillmeta/...
	go test -mod=mod -fuzz=FuzzPathNormalization -fuzztime=5s ./internal/validator/...

test-interop: build
	./$(DIST_DIR)/$(BINARY) version
	go test -mod=mod ./internal/compiler/... -run TestCompileSkillProducesNativeSkilArtifacts

vet:
	go vet ./...

lint:
	gofmt -s -l .

check: vet test test-race test-interop

clean:
	rm -rf $(DIST_DIR)

install: build
	BIN_DIR=$$(go env GOBIN); \
	if [ -z "$$BIN_DIR" ]; then BIN_DIR="$$(go env GOPATH)/bin"; fi; \
	mkdir -p "$$BIN_DIR"; \
	cp $(DIST_DIR)/$(BINARY) "$$BIN_DIR/$(BINARY)"; \
	echo "Installed $(BINARY) to $$BIN_DIR/$(BINARY)"

release-snapshot:
	goreleaser release --snapshot --clean

