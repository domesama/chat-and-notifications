# Go build settings
GO ?= go
GO_TAGS ?= 
GO_FILES = $(shell find . -name \*.go)
GOBUILDFLAGS ?= -v -trimpath$(if $(GO_TAGS), -tags $(GO_TAGS),)

# Main targets
all: generate build

.PHONY: all generate pregenerate build clean test test-unit test-integration coverage start_integration_deps stop_integration_deps restart_integration_deps help

# Help target
help:
	@echo "Available targets:"
	@echo "  make build                    - Build all binaries"
	@echo "  make generate                 - Run code generation"
	@echo "  make clean                    - Clean build artifacts"
	@echo ""
	@echo "Test targets:"
	@echo "  make test                     - Run all tests (unit + integration) with coverage"
	@echo "  make test-unit                - Run unit tests with coverage"
	@echo "  make test-integration         - Run integration tests with coverage"
	@echo "  make coverage                 - Generate and open HTML coverage report"
	@echo ""
	@echo "Integration test dependencies:"
	@echo "  make start_integration_deps   - Start Docker containers for integration tests"
	@echo "  make stop_integration_deps    - Stop Docker containers"
	@echo "  make restart_integration_deps - Restart Docker containers"

# Code generation
generate: pregenerate
	@echo "Running go generate..."
	@$(GO) generate ./...

pregenerate:
	@echo "Running pregenerate script..."
	@./pregenerate

# Build all binaries
build: .bin/chatpersistence .bin/chatpersistencechangehandler .bin/chatwebsocketshandler .bin/generalnotificationshandler

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

.bin/generalnotificationshandler: go.mod go.sum $(GO_FILES)
	@echo "Building generalnotificationshandler..."
	@cd cmd/generalnotificationshandler && $(GO) build $(GOBUILDFLAGS) -o ../../.bin/generalnotificationshandler .

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf .bin/

# Test targets
test: test-unit test-integration
	@echo "All tests completed!"

test-unit:
	@echo "Running unit tests with coverage..."
	@$(GO) test -v -race -coverprofile=coverage-unit.out -covermode=atomic $$(go list ./... | grep -v /ittest/) || true
	@if [ -f coverage-unit.out ]; then \
		printf "Unit test coverage: "; \
		$(GO) tool cover -func=coverage-unit.out | tail -1; \
	else \
		echo "No coverage data generated for unit tests"; \
	fi

test-integration:
	@echo "Running integration tests with coverage..."
	@$(GO) test -v -race -coverprofile=coverage-integration.out -covermode=atomic -coverpkg=$$(go list ./... | grep -v /ittest/ | tr '\n' ',' | sed 's/,$$//') ./ittest/... || true
	@if [ -f coverage-integration.out ]; then \
		grep -v -E '(main\.go|wire_gen\.go):' coverage-integration.out > coverage-integration-filtered.out || cp coverage-integration.out coverage-integration-filtered.out; \
		mv coverage-integration-filtered.out coverage-integration.out; \
		printf "Integration test coverage: "; \
		$(GO) tool cover -func=coverage-integration.out | tail -1 | awk '{print $$NF}'; \
	else \
		echo "No coverage data generated for integration tests"; \
	fi

coverage: test
	@echo "Generating combined coverage report..."
	@echo "mode: atomic" > coverage.out
	@tail -n +2 coverage-unit.out >> coverage.out 2>/dev/null || true
	@tail -n +2 coverage-integration.out >> coverage.out 2>/dev/null || true
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@echo "Opening coverage report in browser..."
	@open coverage.html || xdg-open coverage.html || echo "Please open coverage.html manually"

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