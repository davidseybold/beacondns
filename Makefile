# Makefile for Beacon DNS

APP_NAME := beacon-dns
BIN_DIR := bin

CONTROLLER_DIR := ./cmd/controller
AGENT_DIR := ./cmd/agent
CLI_DIR := ./cmd/cli

CONTROLLER_BIN := $(BIN_DIR)/controller
AGENT_BIN := $(BIN_DIR)/agent
CLI_BIN := $(BIN_DIR)/beacon

GO := go

GOLANGCI_LINT := $(shell which golangci-lint 2>/dev/null)

.PHONY: all build-controller build-agent build-cli build run-controller run-agent run-cli test lint fmt clean install-tools

install-tools:
ifndef GOLANGCI_LINT
	@echo ">> Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
else
	@echo ">> golangci-lint is already installed: $(GOLANGCI_LINT)"
endif

all: build

## Build Targets
build: build-controller build-agent build-cli

build-controller:
	@echo ">> Building controller..."
	$(GO) build -o $(CONTROLLER_BIN) $(CONTROLLER_DIR)

build-agent:
	@echo ">> Building agent..."
	$(GO) build -o $(AGENT_BIN) $(AGENT_DIR)

build-cli:
	@echo ">> Building CLI..."
	$(GO) build -o $(CLI_BIN) $(CLI_DIR)

## Run Targets
run-controller: build-controller
	@echo ">> Running controller..."
	BEACON_ENV=local $(CONTROLLER_BIN)

run-agent: build-agent
	@echo ">> Running agent..."
	$(AGENT_BIN)

run-cli: build-cli
	@echo ">> Running CLI..."
	$(CLI_BIN)

## Utility Targets
test:
	@echo ">> Running tests..."
	$(GO) test -v ./...

lint: install-tools
	@echo ">> Linting..."
	golangci-lint run

fmt:
	@echo ">> Formatting..."
	$(GO) fmt ./...

clean:
	@echo ">> Cleaning binaries..."
	rm -rf $(BIN_DIR)
