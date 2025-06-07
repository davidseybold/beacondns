# Makefile for Beacon DNS

APP_NAME := beacon-dns
BIN_DIR := bin

CONTROLLER_DIR := ./cmd/controller
RESOLVER_DIR := ./cmd/resolver
DUG_DIR := ./cmd/dug

CONTROLLER_BIN := $(BIN_DIR)/controller
RESOLVER_BIN := $(BIN_DIR)/resolver
DUG_BIN := $(BIN_DIR)/dug

GO := go

GOLANGCI_LINT := $(shell which golangci-lint 2>/dev/null)


.PHONY: all build-controller build-agent build-cli build run-controller run-agent run-cli test lint fmt clean install-tools 

# Install tools (golangci-lint, buf, and protoc-gen-go tools)
install-tools:
ifndef GOLANGCI_LINT
	@echo ">> Installing golangci-lint..."
	brew install golangci-lint
endif

# Build Targets
build: build-controller build-agent build-cli

build-controller:
	@echo ">> Building controller..."
	$(GO) build -o $(CONTROLLER_BIN) $(CONTROLLER_DIR)

build-resolver:
	@echo ">> Building resolver..."
	$(GO) build -o $(RESOLVER_BIN) $(RESOLVER_DIR)

build-dug:
	@echo ">> Building dug..."
	$(GO) build -o $(DUG_BIN) $(DUG_DIR)

# Run Targets
run-controller: build-controller
	@echo ">> Running controller..."
	BEACON_ENV=local $(CONTROLLER_BIN)

run-resolver: build-resolver
	@echo ">> Running resolver..."
	BEACON_ENV=local $(RESOLVER_BIN)

run-local:
	@echo ">> Running local..."
	docker compose -f local/docker-compose.yaml up --build

run-e2e-test:
	@echo ">> Running e2e test..."
	$(GO) test -v ./tests/e2e_test.go

# Utility Targets
test:
	@echo ">> Running tests..."
	$(GO) test -v ./...

lint:
	@echo ">> Linting..."
	golangci-lint run

fmt:
	@echo ">> Formatting..."
	golangci-lint fmt

clean:
	@echo ">> Cleaning binaries..."
	rm -rf $(BIN_DIR)
