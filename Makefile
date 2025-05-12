# Makefile for Beacon DNS

APP_NAME := beacon-dns
BIN_DIR := bin

CONTROLLER_DIR := ./cmd/controller
RESOLVER_DIR := ./cmd/resolver

CONTROLLER_BIN := $(BIN_DIR)/controller
RESOLVER_BIN := $(BIN_DIR)/resolver

GO := go

GOLANGCI_LINT := $(shell which golangci-lint 2>/dev/null)

# Proto related
PROTO_DIR := proto
GEN_DIR := internal/proto
PROTO_FILES := $(wildcard $(PROTO_DIR)/**/*.proto)

# Buf-related
BUF := buf
BUF_CONFIG := buf.gen.yaml
BUF_LINT := buf lint
BUF_FORMAT := buf format

.PHONY: all build-controller build-agent build-cli build run-controller run-agent run-cli test lint fmt clean install-tools generate-grpc buf-lint buf-format

# Install tools (golangci-lint, buf, and protoc-gen-go tools)
install-tools:
ifndef GOLANGCI_LINT
	@echo ">> Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
endif
	@echo ">> Installing Buf..."
	@brew install bufbuild/buf/buf || true # Try to install Buf via brew
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate gRPC Go code from Proto files using Buf
generate-grpc: $(PROTO_FILES)
	@echo ">> Generating gRPC Go code using Buf..."
	@buf generate $(PROTO_DIR) --template $(BUF_CONFIG)

# Lint the Protos using Buf
buf-lint:
	@echo ">> Linting proto files using Buf..."
	@buf lint $(PROTO_DIR)

# Format the Protos using Buf
buf-format:
	@echo ">> Formatting proto files using Buf..."
	@buf format -w $(PROTO_DIR)

# Build Targets
build: build-controller build-agent build-cli

build-controller: generate-grpc
	@echo ">> Building controller..."
	$(GO) build -o $(CONTROLLER_BIN) $(CONTROLLER_DIR)

build-resolver: generate-grpc
	@echo ">> Building resolver..."
	$(GO) build -o $(RESOLVER_BIN) $(RESOLVER_DIR)

# Run Targets
run-controller: build-controller
	@echo ">> Running controller..."
	BEACON_ENV=local $(CONTROLLER_BIN)

run-resolver: build-resolver
	@echo ">> Running resolver..."
	BEACON_ENV=local $(RESOLVER_BIN)

# Utility Targets
test:
	@echo ">> Running tests..."
	$(GO) test -v ./...

lint:
	@echo ">> Linting..."
	golangci-lint run

fmt:
	@echo ">> Formatting..."
	$(GO) fmt ./...

clean:
	@echo ">> Cleaning binaries..."
	rm -rf $(BIN_DIR)
	rm -rf $(GEN_DIR)/grpc
