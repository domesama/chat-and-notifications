# Go build settings
GO ?= go
GO_TAGS ?= 
GO_FILES = $(shell find . -name \*.go)
GOBUILDFLAGS ?= -v -trimpath$(if $(GO_TAGS), -tags $(GO_TAGS),)

# Main targets
all: generate build

.PHONY: all generate pregenerate build clean

# Code generation
generate: pregenerate
	@echo "Running go generate..."
	@$(GO) generate ./...

pregenerate:
	@echo "Running pregenerate script..."
	@./pregenerate

# Build all binaries
build: .bin/chatpersistence .bin/chatpersistencechangehandler .bin/chatwebsocketshandler

# Tidy go modules
go.sum: go.mod
	@echo "Running go mod tidy..."
	@$(GO) mod tidy

# Build individual binaries
.bin/chatpersistence: go.mod go.sum $(GO_FILES)
	@echo "Building chatpersistence..."
	@cd cmd/chatpersistence && $(GO) build $(GOBUILDFLAGS) -o ../../.bin/chatpersistence .

.bin/chatpersistencechangehandler: go.mod go.sum $(GO_FILES)
	@echo "Building chatpersistencechangehandler..."
	@cd cmd/chatpersistencechangehandler && $(GO) build $(GOBUILDFLAGS) -o ../../.bin/chatpersistencechangehandler .

.bin/chatwebsocketshandler: go.mod go.sum $(GO_FILES)
	@echo "Building chatwebsocketshandler..."
	@cd cmd/chatwebsocketshandler && $(GO) build $(GOBUILDFLAGS) -o ../../.bin/chatwebsocketshandler .

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf .bin/

# Integration test dependencies
COMPOSE_PROJECT=chat-and-notifications
COMPOSE_FILE=docker/docker-compose.yaml

start_integration_deps:
	docker-compose -p $(COMPOSE_PROJECT) -f $(COMPOSE_FILE) up -d

stop_integration_deps:
	docker-compose -p $(COMPOSE_PROJECT) -f $(COMPOSE_FILE) down

restart_integration_deps:
	docker-compose -p $(COMPOSE_PROJECT) -f $(COMPOSE_FILE) down
	docker-compose -p $(COMPOSE_PROJECT) -f $(COMPOSE_FILE) up -d