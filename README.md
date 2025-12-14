# Chat and Notifications Service

A microservices-based real-time chat and notification system built with Go, leveraging event-driven architecture with Change Data Capture (CDC) for reliable message delivery and websocket management.

## Overview

This project provides a scalable, event-driven solution for real-time chat messaging, general notifications (purchase, payment, shipping updates), and email notifications. The system uses MongoDB CDC via Debezium to ensure consistency between database persistence and real-time message delivery, eliminating the dual-write problem inherent in traditional approaches.

## Project Structure

```
notifications/
├── cmd/                                    # Containerized service entry points
│   ├── chatpersistence/                   # Chat message persistence API
│   ├── chatpersistencechangehandler/      # CDC consumer for chat synchronization
│   ├── chatwebsocketshandler/             # Chat websocket management service
│   ├── generalnotificationshandler/       # General notification websocket service
│   └── emailhandler/                      # Email notification service
├── chatpersistence/                        # Chat persistence domain logic
├── chatpersistencechangehandler/          # CDC event handler logic
├── chatwebsocketshandler/                 # Chat websocket handler logic
├── chatstream/                            # Stream ID utilities for consistent routing
├── connections/                           # Infrastructure connections (Kafka, MongoDB, Redis)
├── email/                                 # Email sending logic
├── emailhandler/                          # Email handler domain logic
├── event/                                 # Event processing framework
├── eventmodel/                            # Event data models
├── eventstore/                            # Event deduplication store
├── generalnotifications/                  # General notification domain logic
├── httpserverwrapper/                     # HTTP server utilities
├── model/                                 # Domain models
├── outgoinghttp/                          # HTTP client utilities
├── websocket/                             # Websocket management framework
├── ittest/                                # Integration tests
├── docker/                                # Docker Compose for local development
└── utils/                                 # Shared utilities
```

> **📖 For a deeper understanding of the architecture, design decisions, and data flow, please refer to [architecture.md](./architecture.md)**

## Services

This project contains **5 containerized services**, each intended to be deployed independently:

### 1. **chatpersistence**
HTTP API service that receives chat messages and persists them to MongoDB. Acts as the source of truth for chat message storage.

### 2. **chatpersistencechangehandler**
Kafka consumer that processes Change Data Capture (CDC) events from MongoDB via Debezium. Forwards chat messages to both the chat websocket service and general notification service with retry logic and idempotency guarantees.

### 3. **chatwebsocketshandler**
Manages websocket connections for active chat participants. Routes messages using `stream_id` for consistent hashing at the L7 load balancer level, enabling local websocket management without cross-pod broadcasting.

### 4. **generalnotificationshandler**
Manages websocket connections for general notifications (chat notifications for non-active users, purchase updates, payment reminders, shipping updates). Routes connections by `user_id`.

### 5. **emailhandler**
Kafka consumer that sends email notifications based on various events (chat messages, purchases, payments, shipping updates).

## In house dependencies

This project is built around two key libraries developed in-house:

### [wireprovider](https://github.com/domesama/wireprovider)
A code generation tool for dependency injection that extends Google Wire. It scans Go source files for specially annotated structs (`@@wire-struct@@`) and automatically generates provider functions, reducing boilerplate and ensuring type-safe dependency injection across services.

### [doakes](https://github.com/domesama/doakes)
An observability and telemetry framework providing unified monitoring, metrics, health checks, and distributed tracing. All services leverage `doakes` for centralized telemetry server setup, Prometheus metrics exposition, and health check endpoints.

## Installation

### Prerequisites
- Go 1.25.1 or higher
- Docker and Docker Compose (for local development dependencies)
- Make

### Install Dependencies

```bash
# Install Go dependencies
go mod download

# Generate code (wire, wireprovider)
make generate
```

**If `make generate` fails**, manually install the required tools:

```bash
go install github.com/domesama/wireprovider@v1.0.3
go install github.com/wireinject/wire/cmd/wire@v0.7.1
```

Then run `make generate` again.

### Build Services

```bash
# Build all services
make build

# Binaries will be created in .bin/ directory:
# - .bin/chatpersistence
# - .bin/chatpersistencechangehandler
# - .bin/chatwebsocketshandler
# - .bin/generalnotificationshandler
# - .bin/emailhandler
```

## Running Tests

### Start Integration Test Dependencies

```bash
# Start MongoDB, Kafka, and Redis containers
make start_integration_deps

# Stop containers
make stop_integration_deps

# Restart containers
make restart_integration_deps
```

The Docker Compose setup includes:
- **Kafka** (with KRaft mode) on port 9092
- **Kafka UI** on port 8889
- **MongoDB** on port 47017
- **Redis** on port 46379

### Run Tests

```bash
# Run all tests (unit + integration)
make test

# Run only unit tests
make test-unit

# Run only integration tests
make test-integration

# Generate HTML coverage report
make coverage
# Opens coverage.html in browser
```

## Development

### Code Generation

The project uses code generation for dependency injection:

```bash
# Run all code generators (wire, wireprovider)
make generate
```

This will:
1. Install required tools (`wireprovider`, `wire`)
2. Scan for `@@wire-struct@@` annotations
3. Generate provider functions
4. Run Google Wire for dependency graph construction


### Clean Build Artifacts

```bash
make clean
```

## Configuration

All services are configured via environment variables. Each service looks for variables with specific prefixes:

- `CHAT_PERSISTENCE_CHANGE_*` - CDC handler configuration
- `MESSAGE_RETRY_*` - Kafka retry configuration
- `DEDUPLICATION_TTL` - Event store TTL
- `EVENT_STORE_KEY_PREFIX` - Redis key prefix for event deduplication

See individual service configurations in `cmd/*/wire/` directories and `*/config/` packages for details.

## Architecture Highlights

- **Event-Driven CDC**: Uses Debezium MongoDB connector to capture database changes and publish to Kafka
- **Idempotency**: Redis-based event store prevents duplicate processing with configurable TTL
- **Retry Strategy**: Configurable retry with backoff via `kafkawrapper`, falls back to connection closure on exhaustion
- **Consistent Routing**: `stream_id` enables deterministic L7 load balancer routing for websocket affinity
- **Separation of Concerns**: Chat websockets and general notifications are managed by separate services for independent scaling

For detailed architecture discussion, data flow diagrams, and design decision rationale, see [architecture.md](./architecture.md).


