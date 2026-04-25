.DEFAULT_GOAL := help

GO        ?= go
LDFLAGS   ?= -s -w -X main.version=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
BIN_DIR   := bin

.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make <target>\n\nTargets:\n"} \
		/^[a-zA-Z0-9_.-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build the costify binary
	@mkdir -p $(BIN_DIR)
	$(GO) build -trimpath -ldflags='$(LDFLAGS)' -o $(BIN_DIR)/costify ./cmd/costify

.PHONY: install
install: ## go install
	$(GO) install -trimpath -ldflags='$(LDFLAGS)' ./cmd/costify

.PHONY: test
test: ## Run unit + golden tests
	$(GO) test -race -count=1 ./...

.PHONY: lint
lint: ## golangci-lint
	golangci-lint run

.PHONY: fmt
fmt: ## gofmt -w
	$(GO) fmt ./...

.PHONY: vet
vet: ## go vet
	$(GO) vet ./...

.PHONY: release-dryrun
release-dryrun: ## GoReleaser snapshot build (no publish)
	goreleaser release --snapshot --clean

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BIN_DIR) dist/
