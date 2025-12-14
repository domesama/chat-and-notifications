# Running Locally

This guide walks you through setting up and running all services locally for development.

## Prerequisites

Before you start, ensure you have completed the [Installation](README.md#installation) steps:
- Go 1.25.1 or higher installed
- Docker and Docker Compose installed
- All dependencies installed (`go mod download`)
- Code generated (`make generate`)
- Services built (`make build`)

## 1. Start Dependencies

Start the required infrastructure services (Kafka, MongoDB, Redis) using Docker Compose:

```bash
# Navigate to docker directory
cd docker

# Start all dependencies
docker-compose up -d

# Check that all containers are running
docker-compose ps
```

This will start:
- **Kafka** on `localhost:9092` (with KRaft mode, no Zookeeper needed)
- **Kafka UI** on `http://localhost:8889` (for monitoring topics and messages)
- **MongoDB** on `localhost:47017` (username: `test_user`, password: `test_password`)
- **Redis** on `localhost:46379`

To stop the dependencies:
```bash
docker-compose down
```

## 2. Configure Services

Each service has its own environment configuration file. Use the provided `.env.local.*` files as templates:

```bash
# Navigate back to project root
cd ..

# Copy example environment file for reference
cp .env.example .env

# Service-specific configuration files already exist:
# - .env.local.chatpersistence
# - .env.local.chatpersistencechangehandler
# - .env.local.chatwebsocketshandler
# - .env.local.generalnotificationshandler
# - .env.local.emailhandler
```

### Key Configuration Values

Below are the essential configuration values for each service in local development:

#### chatpersistence

Edit `.env.local.chatpersistence`:

```bash
LISTEN_ADDR=:8080
MONGO_URI=mongodb://test_user:test_password@localhost:47017
MONGO_DATABASE=notifications
```

#### chatpersistencechangehandler

Edit `.env.local.chatpersistencechangehandler`:

```bash
CHAT_PERSISTENCE_CHANGE_KAFKA_BOOTSTRAP_SERVERS=localhost:9092
CHAT_PERSISTENCE_CHANGE_KAFKA_CONSUMER_INFO_TOPIC_NAME=chat-persistence-change
REDIS_ADDR=localhost:46379
DEDUPLICATION_TTL=5m
GENERAL_NOTIFICATION_OUTGOING_CONFIG_CLIENT_HOST=http://localhost:8082
CHAT_MESSAGE_SOCKET_TRANSFER_OUTGOING_CONFIG_CLIENT_HOST=http://localhost:8081
```

#### chatwebsocketshandler

Edit `.env.local.chatwebsocketshandler`:

```bash
LISTEN_ADDR=:8081
MONGO_URI=mongodb://test_user:test_password@localhost:47017
MONGO_DATABASE=notifications
PING_INTERVAL=30s
PONG_WAIT=40s
```

#### generalnotificationshandler

Edit `.env.local.generalnotificationshandler`:

```bash
LISTEN_ADDR=:8082
PING_INTERVAL=30s
PONG_WAIT=40s
```

#### emailhandler

Edit `.env.local.emailhandler`:

```bash
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=your-email@example.com
SMTP_PASSWORD=your-smtp-password
EMAIL_FROM_ADDRESS=noreply@example.com
MONGO_URI=mongodb://test_user:test_password@localhost:47017
MONGO_DATABASE=notifications
```

> **💡 Tip:** See `.env.example` for a complete list of all available configuration options with detailed explanations.

## 3. Set Up Debezium (CDC Connector)

**Note:** Debezium is configured at the infrastructure/deployment level and is typically set up separately from the application services. For local development, you have two options:

### Option A: Run without Debezium (Manual Testing)

You can test individual services without CDC by manually publishing events to Kafka or directly calling service endpoints.

This is the recommended approach for most local development scenarios where you want to test specific service functionality.

### Option B: Set Up Debezium Locally

For full end-to-end CDC flow, you need to set up the Debezium MongoDB connector:

1. **Add Debezium to docker-compose.yaml** (not included by default):
   - Kafka Connect container with Debezium MongoDB connector
   - Configure MongoDB replica set (CDC requires replica set mode)

2. **Configure Debezium connector** to watch MongoDB collections:
   - Topic: `chat-persistence-change`
   - Database: `notifications`
   - Collections: `chat_messages` (and others as needed)

3. **Register connector via Kafka Connect REST API**

Example Debezium connector configuration:
```json
{
  "name": "mongodb-chat-connector",
  "config": {
    "connector.class": "io.debezium.connector.mongodb.MongoDbConnector",
    "mongodb.connection.string": "mongodb://test_user:test_password@mongodb0:47017/?replicaSet=rs0",
    "topic.prefix": "notifications",
    "collection.include.list": "notifications.chat_messages"
  }
}
```

> **📖 For production deployments**, Debezium connectors are typically managed by your infrastructure/platform team and configured to publish change events to the appropriate Kafka topics.

## 4. Start Services

Each service can be started independently. Run them in separate terminal windows/tabs:

### Terminal 1 - Chat Persistence API
```bash
source .env.local.chatpersistence
.bin/chatpersistence
```

### Terminal 2 - Chat WebSocket Handler
```bash
source .env.local.chatwebsocketshandler
.bin/chatwebsocketshandler
```

### Terminal 3 - General Notifications Handler
```bash
source .env.local.generalnotificationshandler
.bin/generalnotificationshandler
```

### Terminal 4 - Chat Persistence Change Handler (CDC Consumer)
```bash
source .env.local.chatpersistencechangehandler
.bin/chatpersistencechangehandler
```

### Terminal 5 - Email Handler
```bash
source .env.local.emailhandler
.bin/emailhandler
```

## 5. Verify Services

Each service exposes health check and metrics endpoints (via the `doakes` library):

```bash
# Check chatpersistence health
curl http://localhost:8080/health

# Check chatwebsocketshandler health
curl http://localhost:8081/health

# Check generalnotificationshandler health
curl http://localhost:8082/health

# View Prometheus metrics (example from chatpersistence)
curl http://localhost:8080/metrics
```

### Monitor Infrastructure

You can also monitor the infrastructure components:

- **Kafka UI**: `http://localhost:8889` - View topics, consumer groups, and messages
- **MongoDB**: Connect via any MongoDB client to `localhost:47017`
- **Redis**: Connect via redis-cli: `redis-cli -p 46379`

## Service Ports Reference

| Service                          | Port  | Description                        |
|----------------------------------|-------|------------------------------------|
| chatpersistence                  | 8080  | Chat persistence API               |
| chatwebsocketshandler            | 8081  | Chat websocket connections         |
| generalnotificationshandler      | 8082  | General notification websockets    |
| emailhandler                     | N/A   | Kafka consumer (no HTTP endpoint)  |
| chatpersistencechangehandler     | N/A   | Kafka consumer (no HTTP endpoint)  |
| Kafka                            | 9092  | Kafka broker                       |
| Kafka UI                         | 8889  | Kafka monitoring UI                |
| MongoDB                          | 47017 | MongoDB database                   |
| Redis                            | 46379 | Redis cache/event store            |

## Troubleshooting

### Services won't start

1. **Check dependencies are running**: `cd docker && docker-compose ps`
2. **Check port conflicts**: Ensure ports 8080-8082, 9092, 8889, 47017, and 46379 are available
3. **Check environment variables**: Verify `.env.local.*` files are properly configured

### Can't connect to MongoDB

1. Verify MongoDB is running: `docker ps | grep mongo-notifications`
2. Test connection: `mongosh mongodb://test_user:test_password@localhost:47017`
3. Check MongoDB logs: `docker logs mongo-notifications`

### Kafka consumer not receiving messages

1. Check Kafka is running: `docker ps | grep kafka-notifications`
2. Verify topic exists in Kafka UI: `http://localhost:8889`
3. Check consumer group status in Kafka UI
4. Verify Debezium connector is configured (if using CDC)

### WebSocket connection issues

1. Ensure the websocket handler service is running
2. Check service health: `curl http://localhost:8081/health`
3. Verify MongoDB connection (websocket handlers store connection state)

## Next Steps

- Read [architecture.md](./architecture.md) for detailed system design
- Check [Running Tests](README.md#running-tests) to run integration tests
- Review `.env.example` for advanced configuration options

