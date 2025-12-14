# Architecture Overview

This document provides an in-depth explanation of the chat and notifications system architecture, design decisions, and data flow patterns.

## Table of Contents
- [Requirements](#requirements)
- [The Dual-Write Problem](#the-dual-write-problem)
- [Event-Driven Architecture: Outbox vs CDC](#event-driven-architecture-outbox-vs-cdc)
- [Service Components](#service-components)
- [Chat Message Flow](#chat-message-flow)
- [Stream ID and Consistent Routing](#stream-id-and-consistent-routing)
- [Websocket Service Segregation](#websocket-service-segregation)
- [Idempotency and Event Store](#idempotency-and-event-store)
- [Retry Strategy and Failover](#retry-strategy-and-failover)
- [Notification Delivery Strategies](#notification-delivery-strategies)
- [TL;DR - Complete Flow](#tldr---complete-flow)

---

## Requirements (From my perspective)

The system must support:

1. **Real-time Chat Synchronization**: When a user sends a chat message, both the sender and receiver must see the message in real-time
2. **Chat Persistence**: All chat messages must be reliably stored in MongoDB for history and retrieval
3. **General Notifications**: Users should receive real-time notifications for:
   - Chat messages (when not actively in the chat)
   - Purchase updates
   - Payment reminders
   - Shipping updates
4. **Email Notifications**: Some events should trigger email delivery
5. **High Availability**: The system must handle partial failures gracefully and scale independently

---

## The Dual-Write Problem

In a traditional non-event-driven design, when a user sends a chat message, we need to:
1. Persist the message to MongoDB
2. Send the message to websocket connections in real-time

This creates a **dual-write problem**:

```
User sends message
    ↓
┌─────────────────────────┐
│  Chat Persistence API   │
└─────────────────────────┘
    ↓
    ├──→ Write to MongoDB ✓
    │
    └──→ Send to WebSocket ✗  (What if this fails?)
```

**Failure scenarios:**
- ✅ Persist to DB succeeds, ❌ WebSocket send fails → User doesn't see message in real-time
- ❌ Persist to DB fails, ✅ WebSocket send succeeds → Message lost after reload
- Handling retries and rollbacks becomes complex and error-prone

**Solution:** Move to an event-driven architecture where the database is the single source of truth, and changes are propagated via events.

---

## Event-Driven Architecture: Outbox vs CDC

There are two primary patterns for event-driven architectures:

### 1. Outbox Pattern
- Application writes to both the main table AND an "outbox" table in a single transaction
- A separate poller reads the outbox table and publishes events to Kafka
- Once published, events are deleted from the outbox
- **Pull-based** polling mechanism

### 2. Change Data Capture (CDC) with Debezium
- Application writes ONLY to the main table
- Debezium connector watches the MongoDB oplog/change stream
- Automatically publishes database changes to Kafka
- **Push-based** from the database Write-Ahead Log (WAL)

### Why I Chose CDC

For this use case, **CDC is superior** because:

1. **No Transformation Needed**: Chat messages are persisted exactly as they need to be delivered. I don't need to aggregate or transform data, so CDC's direct capture is perfect.

2. **Push-Based is Faster**: Debezium listens to MongoDB's change stream and immediately pushes events to Kafka. The outbox pattern requires periodic polling, introducing latency.

3. **Less Manual Work**: I don't need to manually watch database changes, maintain outbox tables, or implement polling logic. Debezium handles all of this.

4. **Single Write Path**: Application code only worries about persistence, not event publishing.

```
Outbox Pattern:                CDC Pattern (Our Choice):
                              
Write → [Main Table]           Write → [MongoDB]
     ↓                              ↓ (oplog)
     → [Outbox Table]              [Debezium Connector]
          ↓ (poll)                      ↓ (push)
     [Poller Service]              [Kafka Topic]
          ↓
     [Kafka Topic]
```

**Note:** Debezium Kafka Connector is configured at the infrastructure/deployment level to watch MongoDB collections and publish change events to Kafka topics.

---

## Service Components

### Architecture Diagram

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                           CLIENT / FRONTEND                                  │
│  - Sends chat via HTTP                                                       │
│  - Subscribes to chat websocket (stream_id routing)                          │
│  - Subscribes to general notification websocket (user_id routing)            │
│  - Fetches chat history on connection failure                                │
└──────────────────────────────────────────────────────────────────────────────┘
         │ POST /chat                    │ WS /chat/subscribe    │ WS /notifications/subscribe
         ↓                               ↓                        ↓
┌─────────────────────┐      ┌──────────────────────┐   ┌───────────────────────────┐
│  chatpersistence    │      │ chatwebsockets       │   │ generalnotifications      │
│                     │      │ handler              │   │ handler                   │
│  - HTTP API         │      │                      │   │                           │
│  - Validates msg    │      │ - Manages chat WS    │   │ - Manages notification WS │
│  - Persists to DB   │      │ - Routes by stream_id│   │ - Routes by user_id       │
└─────────────────────┘      │ - Local broadcast    │   │ - Handles chat/purchase/  │
         ↓                   └──────────────────────┘   │   payment/shipping        │
    [MongoDB]                        ↑                  └───────────────────────────┘
         ↓ (oplog/change stream)     │ HTTP POST                   ↑ HTTP POST
    [Debezium Connector]             │                             │
         ↓                            │                             │
    [Kafka Topic: chat.changes]      │                             │
         ↓                            │                             │
┌─────────────────────────────────────────────────────────────────────────────┐
│  chatpersistencechangehandler                                               │
│                                                                             │
│  - Consumes Kafka CDC events                                                │
│  - Idempotency check (RedisEventStore)                                      │
│  - Concurrent HTTP calls:                                                   │
│    1. Forward to chatwebsocketshandler                                      │
│    2. Forward to generalnotificationshandler                                │
│  - Retry on failure (kafkawrapper)                                          │
│  - Close WS on exhausted retries                                            │
└─────────────────────────────────────────────────────────────────────────────┘
         │ (Alternative: separate consumer group)
         ↓
┌─────────────────────┐
│  emailhandler       │
│                     │
│  - Consumes events  │
│  - Sends emails     │
└─────────────────────┘
```

### 1. chatpersistence

**Responsibility**: Chat message persistence API

```
POST /chat/messages
{
  "content": "Hello!",
  "sender_id": "user123",
  "receiver_id": "user456",
  "stream_id": "abc123def456"  // computed on frontend
}
    ↓
Validate request
    ↓
Insert into MongoDB
    ↓
Return 201 Created
```

- Receives chat messages via HTTP POST
- Validates message content and metadata
- Persists to MongoDB (single write, source of truth)
- Returns success/failure to client
- **No websocket logic** - purely persistence

### 2. chatpersistencechangehandler

**Responsibility**: CDC event consumer and message synchronization orchestrator

```
Kafka Consumer (chat.changes topic)
    ↓
Receive CDC event from Debezium
    ↓
Check idempotency (RedisEventStore)
    ├─ Already processed? → Drop
    └─ New event? → Continue
        ↓
    Parse ChatMessagePersistenceChangeEvent
        ↓
    Concurrent HTTP calls:
        ├─→ POST /chat/forward-to-websocket      (chatwebsocketshandler)
        │       ↓
        │   Returns 200 OK → Success
        │   Returns 206 Partial Content → Retry (some WS failed)
        │   Returns 500 Error → Retry
        │
        └─→ POST /notifications/chat             (generalnotificationshandler)
                ↓
            Broadcast to user's general notification websocket
    ↓
Mark as processed in RedisEventStore
    ↓
Commit Kafka offset
```

**Key Features:**
- Consumes from Kafka using `kafkawrapper.ConsumerGroup`
- Uses `RedisEventStore` for idempotency (default 5m TTL, configurable)
- Forwards messages concurrently using `concurrent.NewGroup`
- Retries failed forwards with backoff (configurable via `MessageRetryConfig`)
- On exhausted retries: closes websocket connections, forcing frontend to fetch from persistence

### 3. chatwebsocketshandler

**Responsibility**: Manage websocket connections for active chat participants

```
WebSocket Handshake:
GET /chat/subscribe?stream_id=abc123&sender_id=user123&receiver_id=user456
    ↓
Register connection in WebSocketManager
    ↓
Keep connection alive (heartbeat, ping/pong)

Message Forward:
POST /chat/forward-to-websocket
{
  "message_id": "msg123",
  "content": "Hello!",
  "stream_id": "abc123",
  "sender_id": "user123",
  "receiver_id": "user456"
}
    ↓
Broadcast to all local connections with matching stream_id
    ├─ All delivered → Return 200 OK
    ├─ Some failed → Return 206 Partial Content (triggers retry)
    └─ None delivered → Return 500 Internal Server Error
```

**Key Features:**
- Routes connections by `stream_id` (deterministic hash of sender + receiver)
- Local-only broadcasting (no cross-pod communication needed)
- Returns HTTP 206 on partial delivery to trigger CDC consumer retry
- Gracefully handles connection failures and cleanup

### 4. generalnotificationshandler

**Responsibility**: Manage websocket connections for general notifications

```
WebSocket Handshake:
GET /notifications/subscribe?user_id=user456
    ↓
Register connection in WebSocketManager
    ↓
Keep connection alive

Notification Types:
POST /notifications/chat              (Chat notification for inactive users)
POST /notifications/purchase          (Purchase update)
POST /notifications/payment-reminder  (Payment reminder)
POST /notifications/shipping          (Shipping update)
    ↓
Broadcast to all connections for user_id
```

**Key Features:**
- Routes connections by `user_id`
- Handles multiple notification types (chat, purchase, payment, shipping)
- Independent from chat websocket service (different scaling needs)

> **Note:** Currently, only **chat notifications** are fully implemented with CDC integration. The other notification types (purchase updates, payment reminders, shipping updates) are not yet implemented but can follow the same architecture pattern as `chatpersistencechangehandler`:
> - Set up Debezium CDC connectors for respective MongoDB collections (purchases, payments, shipments)
> - Create dedicated consumer services following the same code pattern
> - Use `RedisEventStore` for idempotency with appropriate TTL configurations
> - Forward to `generalnotificationshandler` via the corresponding endpoints

### 5. emailhandler

**Responsibility**: Send email notifications

```
Kafka Consumer (events topic)
    ↓
Process event (chat, purchase, payment, shipping)
    ↓
Format email template
    ↓
Send via SMTP
```

**Key Features:**
- Consumes from Kafka (same or different topic/consumer group)
- SMTP email sending with template support
- Can be scaled independently based on email volume

---

## Chat Message Flow

### Complete End-to-End Flow

```
 [Frontend]
     │
     │ 1. User types message and clicks send
     │
     ↓
┌─────────────────────┐
│ POST /chat/messages │ 2. HTTP request with message payload
└─────────────────────┘
     │
     ↓
┌─────────────────────────────────────────────────┐
│           chatpersistence                       │
│  - Validate message                             │
│  - Insert into MongoDB (SINGLE WRITE)           │
└─────────────────────────────────────────────────┘
     │
     ↓
┌─────────────────────────────────────────────────┐
│           MongoDB                               │
│  - Stores message                               │
│  - Change stream emits oplog event              │
└─────────────────────────────────────────────────┘
     │
     ↓
┌─────────────────────────────────────────────────┐
│       Debezium Kafka Connector                  │
│  - Watches MongoDB change stream                │
│  - Publishes CDC event to Kafka                 │
└─────────────────────────────────────────────────┘
     │
     ↓
┌─────────────────────────────────────────────────┐
│     Kafka Topic: chat.changes                   │
│  - Partitioned by stream_id                     │
│  - Guarantees ordering per partition            │
└─────────────────────────────────────────────────┘
     │
     ↓
┌─────────────────────────────────────────────────┐
│   chatpersistencechangehandler                  │
│  - Consumer Group reads from partition          │
│  - Check RedisEventStore (idempotency)          │
│  - Parse CDC event → ChatMessage                │
└─────────────────────────────────────────────────┘
     │
     ├────────────────────────┬───────────────────────┐
     ↓                        ↓                       ↓
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│ Forward to Chat  │  │ Forward to Gen.  │  │ (Alternative:    │
│ WebSocket        │  │ Notification     │  │ Separate consumer│
│ Service          │  │ Service          │  │ for email)       │
└──────────────────┘  └──────────────────┘  └──────────────────┘
     │                        │
     ↓                        ↓
┌──────────────────┐  ┌──────────────────┐
│ chatwebsockets   │  │ generalnotif.    │
│ handler          │  │ handler          │
│                  │  │                  │
│ Broadcast to     │  │ Broadcast to     │
│ stream_id        │  │ user_id          │
│ subscribers      │  │ subscribers      │
└──────────────────┘  └──────────────────┘
     │                        │
     ↓                        ↓
┌──────────────────┐  ┌──────────────────┐
│ Active Chat      │  │ Notification     │
│ Users            │  │ Badge/Alert      │
│ (sender +        │  │ (if not in chat) │
│  receiver)       │  │                  │
└──────────────────┘  └──────────────────┘
```

### Step-by-Step Breakdown

1. **User sends message**: Frontend computes `stream_id` from `sender_id` and `receiver_id`, sends POST request to `chatpersistence` API

2. **Persistence**: `chatpersistence` validates and writes to MongoDB in a single transaction

3. **CDC Capture**: Debezium watches MongoDB change stream, captures the insert operation, publishes CDC event to Kafka topic

4. **Event Consumption**: `chatpersistencechangehandler` consumes event from Kafka partition (partitioned by `stream_id` for ordering)

5. **Idempotency Check**: Checks `RedisEventStore` using `message_id:stream_id` as key. If already processed, drops the event.

6. **Concurrent Forwarding**: 
   - Forwards to `chatwebsocketshandler` (for users actively in the chat)
   - Forwards to `generalnotificationshandler` (for notification badge if receiver not in chat)

7. **WebSocket Delivery**: Both websocket services broadcast to their respective connections

8. **Frontend Receives**: Active chat users see message in real-time via chat websocket; inactive users see notification badge via general notification websocket

---

## Stream ID and Consistent Routing

### What is Stream ID?

`stream_id` is a **deterministic identifier** for a chat conversation between two users. It's computed by:

1. Sorting the two user IDs alphabetically
2. Concatenating them
3. Taking a SHA-256 hash
4. Using the first 16 characters

```go
// From chatstream/stream_id.go
func ComputeStreamID(userID1, userID2 string) string {
    ids := []string{userID1, userID2}
    sort.Strings(ids)  // Ensures same result regardless of order
    
    combined := ids[0] + ids[1]
    hash := sha256.Sum256([]byte(combined))
    
    return hex.EncodeToString(hash[:])[:16]
}
```

**Example:**
- User A (ID: `user123`) and User B (ID: `user456`)
- Stream ID: `ComputeStreamID("user123", "user456")` → Always returns same hash
- Both users compute the same `stream_id` on frontend

### Why Stream ID?

Stream ID enables **consistent routing** at multiple levels:

1. **Kafka Partitioning**: Messages with the same `stream_id` always go to the same partition, guaranteeing ordering

2. **L7 Load Balancer Routing**: Using consistent hashing on `stream_id`, the load balancer can route websocket connections for the same conversation to the same pod

3. **Local WebSocket Management**: Since both sender and receiver connect to the same pod (via consistent hash), I can broadcast messages locally without cross-pod communication

```
Frontend (User A & User B)
    │
    │ Both compute same stream_id = "abc123def456"
    │
    ↓
L7 Load Balancer (Consistent Hash on stream_id)
    │
    ├─ stream_id "abc123def456" → Pod 1
    ├─ stream_id "xyz789ghi012" → Pod 2
    └─ stream_id "def456jkl345" → Pod 3
    │
    ↓
┌────────────────────────────────────────┐
│  Pod 1 - chatwebsocketshandler         │
│                                        │
│  WebSocketManager:                     │
│    stream_id "abc123def456":           │
│      - Connection 1 (User A)           │
│      - Connection 2 (User B)           │
│                                        │
│  Broadcast locally to both connections │
└────────────────────────────────────────┘
```

**Benefits:**
- No Redis pub/sub needed for broadcasting
- No cross-pod communication overhead
- Simpler failure handling (connection and message in same pod)

**Tradeoffs:**
- Consistent hashing can create **hotspots** if certain conversations have very high message volume
- Rebalancing on pod scaling can temporarily disrupt routing (acceptable with proper connection retry)

---

## Websocket Service Segregation

### Why Separate CDC Consumer from WebSocket Management?

Initially, you might think: "Why not manage websocket connections directly in the CDC consumer (`chatpersistencechangehandler`)?"

**Problem 1: Kafka Partition Constraints**

Kafka consumers in a consumer group scale up to the number of partitions. If we have:
- 10 Kafka partitions
- Max 10 CDC consumer pods

But we might need:
- 50 websocket handler pods to handle connection volume

If I couple CDC consumption with websocket management, I'm limited to 10 pods.

**Problem 2: Rebalancing Complexity**

When Kafka consumer group rebalances (e.g., pod scaling, failures):
- Partitions are reassigned to different pods
- If websockets are managed in the same pod, I'd need to:
  - Transfer websocket connections during rebalance, OR
  - Broadcast messages across all pods to find the right connection

Both options are complex and error-prone.

**Problem 3: Broadcast Overhead**

Even with Kafka partition ordering guarantees, there's no guarantee that:
- The Kafka message lands in Pod A, AND
- The websocket connection is also in Pod A

Without consistent routing, I'd need Redis pub/sub to broadcast every message to all pods, creating overhead.

**Solution: Separate Services**

```
┌──────────────────────────────────────────────┐
│  chatpersistencechangehandler                │
│  - Fixed scaling (up to Kafka partitions)    │
│  - Consumes CDC events                       │
│  - Forwards to websocket service via HTTP    │
│  - Can retry independently                   │
└──────────────────────────────────────────────┘
         │ HTTP POST
         ↓
┌──────────────────────────────────────────────┐
│  chatwebsocketshandler                       │
│  - Independent scaling (50+ pods if needed)  │
│  - L7 consistent hash routing by stream_id   │
│  - Local websocket management                │
│  - No Kafka partition constraints            │
└──────────────────────────────────────────────┘
```

### Redis Pub/Sub vs Consistent Hash Routing

I evaluated two solutions for websocket delivery:

#### Option 1: Redis Pub/Sub

```
CDC Consumer receives message
    ↓
Publish to Redis channel
    ↓
All websocket pods subscribe to channel
    ↓
Each pod checks if it has the connection
    ↓
Deliver to local connections
```

**Pros:**
- Simple to implement
- No special load balancer config

**Cons:**
- Every message broadcasts to ALL pods
- High Redis load with many messages
- Wasted network bandwidth (pods without the connection still receive)
- For group chats with multiple receivers across pods, overhead multiplies

#### Option 2: Consistent Hash Routing (Our Choice)

```
CDC Consumer receives message with stream_id
    ↓
HTTP POST to websocket service
    ↓
L7 Load Balancer uses consistent hash on stream_id
    ↓
Routes to specific pod that manages this stream_id
    ↓
Pod delivers to local connections only
```

**Pros:**
- Targeted delivery (only relevant pod receives)
- Lower network and Redis overhead
- Scales better with message volume
- Simpler pod-local websocket management

**Cons:**
- Requires L7 load balancer with consistent hashing support (nginx, envoy, etc.)
- Potential hotspots with high-traffic conversations
- Rebalancing on scaling needs connection retry logic

I chose **Option 2** because the overhead reduction and scaling benefits outweigh the load balancer configuration complexity.

---

## Idempotency and Event Store

### Why Idempotency is Critical

Kafka guarantees **at-least-once delivery**, meaning:
- A message may be delivered multiple times (e.g., consumer crashes before committing offset)
- I must ensure duplicate messages don't result in duplicate websocket sends or emails

### RedisEventStore Implementation

```go
type RedisEventStore[MsgValue any] struct {
    RedisClient           redis.Client
    DeduplicationTTL      time.Duration  // Default: 5m
    EventStoreKeyPrefix   string         // e.g., "chat-cdc"
}
```

**How it works:**

1. **Before Processing**:
   ```go
   dedupKey := "eventstore:chat-cdc:message_id:stream_id"
   exists := redis.Exists(dedupKey)
   if exists {
       return DROP  // Already processed
   }
   ```

2. **After Processing**:
   ```go
   redis.Set(dedupKey, "1", 5*time.Minute)
   ```

3. **TTL Cleanup**: Redis automatically expires keys after TTL, no manual cleanup needed

**Key Format**: `eventstore:{prefix}:{message_id}:{stream_id}`

**Configuration**:
- `DEDUPLICATION_TTL`: How long to remember processed events (default 5m, configurable up to 12hr for notification scenarios)
- `EVENT_STORE_KEY_PREFIX`: Prefix for different event types (e.g., `chat-cdc`, `email-events`)

**Why Redis?**
- Fast in-memory lookups (sub-millisecond)
- Built-in TTL expiration
- Shared state across multiple consumer pods
- Simple key-value operations

---

## Retry Strategy and Failover

### Kafka Consumer Retry with kafkawrapper

The `kafkawrapper` library provides configurable retry logic with exponential backoff:

```go
// In connections/kafka_consumer_group.go
wrappedHandler := kafkawrapper.WrapWithRetryBackoffHandler(
    handler,
    kafkaInfo.MessageRetryConfig,  // Configurable retry settings
)
```

**Configuration** (via environment variables):
```
MESSAGE_RETRY_MAX_RETRIES=5              # Max retry attempts
MESSAGE_RETRY_INITIAL_INTERVAL=100ms     # Initial backoff
MESSAGE_RETRY_MAX_INTERVAL=30s           # Max backoff
MESSAGE_RETRY_MULTIPLIER=2.0             # Exponential multiplier
```

**Example retry sequence**:
```
Attempt 1: Failed → Wait 100ms
Attempt 2: Failed → Wait 200ms
Attempt 3: Failed → Wait 400ms
Attempt 4: Failed → Wait 800ms
Attempt 5: Failed → Wait 1.6s
...
Max retries exceeded → Failover
```

### HTTP 206 Partial Content Triggers Retry

In `chatwebsocketshandler`, the service returns different status codes:

```go
// chatwebsocketshandler/handler/chat_web_socket_forwarder.go

deliveredCount, err := BroadcastPayloadToLocalSubscribers(ctx, stream_id, message)

if err == nil {
    return 200 OK  // All connections received message
}

if deliveredCount == 0 {
    return 500 Internal Server Error  // No connections received
}

// Some connections received, some failed
return 206 Partial Content  // TRIGGERS RETRY
```

In `chatpersistencechangehandler`:

```go
// chatpersistencechangehandler/service/chat_message_sync.go

statusCode, err := CallHTTP(POST, "/chat/forward-to-websocket", chatMessage)

if statusCode == http.StatusPartialContent {
    return errors.New("partial content delivered")  // kafkawrapper will retry
}
```

**Flow:**
```
1. CDC consumer calls chatwebsocketshandler
2. chatwebsocketshandler tries to deliver to 3 connections
3. 2 succeed, 1 fails → Return HTTP 206
4. CDC consumer receives 206 → Return error
5. kafkawrapper intercepts error → Retry with backoff
6. Retry #1: All 3 connections succeed → Return HTTP 200
7. CDC consumer commits Kafka offset
```

### Failover: Close WebSocket on Exhausted Retries

If all retry attempts fail, I implement a **graceful degradation** strategy:

**Current behavior:**
- After max retries, the CDC consumer gives up on that message
- The websocket service closes affected connections
- Frontend detects websocket closure
- Frontend falls back to HTTP polling or fetches chat history from `chatpersistence` API

**Why this approach?**
- Prevents infinite retry loops
- Avoids blocking the Kafka consumer (other messages can still be processed)
- Frontend can still provide degraded functionality (polling instead of real-time)
- User experience: slight delay in receiving messages, but no data loss (message is in MongoDB)

**Implementation:**
```go
// After max retries exhausted in kafkawrapper:
// 1. Log error with high severity
// 2. WebSocket manager closes connections for this stream_id
// 3. Frontend receives close frame
// 4. Frontend switches to polling mode and shows reconnection UI
```

**Frontend Fallback Flow:**
```
WebSocket Connection Closed
    ↓
Show "Reconnecting..." UI
    ↓
Poll GET /chat/messages?stream_id=abc123&since=last_message_id
    ↓
Display any missed messages
    ↓
Attempt to reconnect WebSocket (exponential backoff)
```

This ensures the system remains operational even under partial failures, prioritizing availability over strict real-time guarantees.

---

## Notification Delivery Strategies

When a chat message is persisted and captured by CDC, I need to decide how to deliver notifications to users who are NOT actively in the chat. There are three viable strategies, each with different tradeoffs.

### Context: The Problem

```
User A sends message to User B

User B might be:
1. Active in chat → Should receive via chat websocket (chatwebsocketshandler)
2. Not in chat, but app is open → Should receive notification badge (generalnotificationshandler)
3. App closed → Should receive email (emailhandler)
```

For cases 2 and 3, I need to forward the chat message to notification services.

### Option 1: Concurrent HTTP Calls in Same Consumer

**Implementation**: `chatpersistencechangehandler` makes concurrent HTTP calls to both services

```go
// Current implementation in chat_message_sync.go

func ForwardChatMessageToWebsocketServices(ctx, msg) error {
    // Task 1: Forward to chat websocket
    forwardToChat := concurrent.NewTask(func(ctx) error {
        return POST("/chat/forward-to-websocket", msg)
    })
    
    // Task 2: Forward to general notification
    forwardToNotification := concurrent.NewTask(func(ctx) error {
        return POST("/notifications/chat", msg)
    })
    
    // Execute both concurrently
    err := concurrent.NewGroup(ctx).Exec(forwardToChat, forwardToNotification)
    return err.ErrorOrNil()
}
```

**Flow Diagram:**
```
┌─────────────────────────────────────────────┐
│  chatpersistencechangehandler               │
│                                             │
│  Consume CDC event                          │
│     ↓                                       │
│  Concurrent calls:                          │
│     ├─→ chatwebsocketshandler               │
│     └─→ generalnotificationshandler         │
│                                             │
│  Both must succeed for commit               │
└─────────────────────────────────────────────┘
```

**Pros:**
- Simple implementation (already in code)
- Low latency (concurrent calls)
- Single consumer group = lower resource usage

**Cons:**
- Coupled retry logic (if notification fails, chat websocket also retries)
- Cannot scale notification and chat delivery independently
- Both services must be available for message to be marked as processed

**Best for:** Small to medium scale where simplicity is preferred

---

### Option 2: Separate Consumer Group with Long TTL Deduplication

**Implementation**: Create a second consumer group (`chat-notification-consumer`) that also consumes from the same Kafka topic but with different event store configuration

```go
// Second consumer group configuration
type NotificationConsumerConfig struct {
    TopicName:   "chat.changes"           // Same topic
    ConsumerName: "chat-notification-consumer"  // Different group
    EventStoreTTL: 6 * time.Hour          // Long TTL!
}
```

**Flow Diagram:**
```
[Kafka Topic: chat.changes]
    │
    ├────────────────────────┬──────────────────────┐
    ↓                        ↓                      ↓
┌─────────────┐      ┌──────────────┐      ┌────────────────┐
│ Consumer 1  │      │ Consumer 2   │      │ Consumer 3     │
│ (Chat WS)   │      │ (Gen. Notif) │      │ (Email)        │
│             │      │              │      │                │
│ TTL: 5min   │      │ TTL: 6-12hr  │      │ TTL: 24hr      │
└─────────────┘      └──────────────┘      └────────────────┘
```

**Long TTL Strategy:**
- Use `stream_id` as deduplication key
- Set TTL to 6-12 hours
- If User B receives multiple messages from User A within 6 hours, only the FIRST message triggers a notification
- Prevents notification spam for ongoing conversations

**Example:**
```
10:00 AM - User A sends "Hello" → Notification sent ✓
10:05 AM - User A sends "How are you?" → Notification suppressed (same stream, within 6hr)
10:10 AM - User A sends "Are you there?" → Notification suppressed
4:00 PM - User A sends "Let's meet" → Notification sent ✓ (6hr elapsed)
```

**Pros:**
- Independent retry policies for chat vs notifications
- Smart notification deduplication (no spam)
- Can scale notification consumer independently
- User doesn't get bombarded with notifications for active conversations

**Cons:**
- More consumer groups = more Kafka connections and resource usage
- Slightly more complex configuration
- Need to tune TTL based on user behavior

**Best for:** Medium scale with desire for intelligent notification management

---

### Option 3: Fully Segregated Consumer Groups (Email + Notification)

**Implementation**: Three separate consumer groups, each with specialized retry and idempotency policies

```
┌──────────────────────────────────────────────────────────┐
│           Kafka Topic: chat.changes                      │
└──────────────────────────────────────────────────────────┘
    │
    ├────────────────────┬─────────────────────┬──────────────────┐
    ↓                    ↓                     ↓                  ↓
┌─────────────┐  ┌───────────────┐  ┌──────────────┐  ┌──────────────┐
│ Consumer 1  │  │ Consumer 2    │  │ Consumer 3   │  │ Consumer 4   │
│             │  │               │  │              │  │              │
│ Chat WS     │  │ General       │  │ Email        │  │ Push Notif   │
│ Forward     │  │ Notification  │  │              │  │ (Future)     │
│             │  │               │  │              │  │              │
│ TTL: 5m     │  │ TTL: 6hr      │  │ TTL: 24hr    │  │ TTL: 1hr     │
│ Retry: 5x   │  │ Retry: 3x     │  │ Retry: 10x   │  │ Retry: 3x    │
└─────────────┘  └───────────────┘  └──────────────┘  └──────────────┘
      │                  │                 │                 │
      ↓                  ↓                 ↓                 ↓
 chatwebsockets   generalnotif.      emailhandler      pushnotif
 handler          handler                               handler
```

**Consumer Specialization:**

1. **Chat WebSocket Consumer**
   - TTL: 5 minutes (short, expect quick delivery)
   - Retry: 5 attempts (aggressive)
   - Failure: Close websocket, frontend fallback

2. **General Notification Consumer**
   - TTL: 6 hours (dedup ongoing conversations)
   - Retry: 3 attempts (moderate)
   - Failure: Log and skip (non-critical)

3. **Email Consumer**
   - TTL: 24 hours (prevent duplicate emails)
   - Retry: 10 attempts (persistent)
   - Failure: Dead letter queue for manual review

4. **Push Notification Consumer** (Future)
   - TTL: 1 hour
   - Retry: 3 attempts
   - Failure: Log and skip

**Pros:**
- **Maximum flexibility**: Each consumer has tailored retry, TTL, and failure policies
- **Independent scaling**: Scale email consumer during high volume, notification consumer separately
- **Resource allocation**: Allocate more resources to critical services (chat) vs nice-to-have (email)
- **Idempotency per service**: Email handler can be fully idempotent without affecting chat delivery
- **Future-proof**: Easy to add new consumers (push notifications, SMS, etc.)

**Cons:**
- **Most resource-intensive**: Multiple consumer groups consume same topic
- **Configuration complexity**: Need to manage multiple consumer group configs
- **Kafka partition utilization**: Each consumer group maintains its own offset

**Best for:** **Large scale with diverse user base** where resource allocation and independent retry strategies are critical

---

### Comparison Table

| Aspect | Option 1: Concurrent Calls | Option 2: Shared Consumer + Long TTL | Option 3: Segregated Consumers |
|--------|---------------------------|--------------------------------------|--------------------------------|
| **Complexity** | Low ✅ | Medium | High |
| **Resource Usage** | Low ✅ | Medium | High |
| **Retry Independence** | No ❌ | Yes ✅ | Yes ✅ |
| **Scaling Independence** | No ❌ | Partial | Yes ✅ |
| **Notification Dedup** | No ❌ | Yes ✅ | Yes ✅ |
| **Idempotency Flexibility** | Low | Medium | High ✅ |
| **Best Scale** | Small-Medium | Medium | Large ✅ |
| **Current Implementation** | ✅ Yes | No | No |

### Recommendation by Scale

- **Small/Medium (< 100K users)**: Option 1 (current implementation)
  - Simplicity and low resource usage outweigh benefits of segregation
  
- **Medium (100K - 1M users)**: Option 2
  - Smart notification deduplication becomes important
  - Resource usage still manageable
  
- **Large (> 1M users)**: Option 3
  - Independent scaling and retry policies critical
  - Resource allocation optimization pays off
  - Flexibility for future notification channels (push, SMS)

---

## TL;DR - Complete Flow

### Chat Message Journey (60 seconds summary)

```
┌──────────────────────────────────────────────────────────────────────────┐
│ 1. USER SENDS MESSAGE                                                    │
│    Frontend: compute stream_id, POST /chat/messages                      │
└──────────────────────────────────────────────────────────────────────────┘
                              ↓
┌──────────────────────────────────────────────────────────────────────────┐
│ 2. PERSIST (Single Source of Truth)                                     │
│    chatpersistence: Validate → Insert MongoDB → Return 201               │
└──────────────────────────────────────────────────────────────────────────┘
                              ↓
┌──────────────────────────────────────────────────────────────────────────┐
│ 3. CDC CAPTURE                                                           │
│    Debezium watches MongoDB oplog → Publishes to Kafka                   │
└──────────────────────────────────────────────────────────────────────────┘
                              ↓
┌──────────────────────────────────────────────────────────────────────────┐
│ 4. EVENT CONSUMPTION & ORCHESTRATION                                     │
│    chatpersistencechangehandler:                                         │
│      - Consume from Kafka (partitioned by stream_id)                     │
│      - Check RedisEventStore (idempotency, 5m TTL)                       │
│      - Concurrent HTTP calls:                                            │
│          • chatwebsocketshandler (active chat users)                     │
│          • generalnotificationshandler (notification badge)              │
│      - Retry on failure (configurable, max 5x with backoff)              │
│      - On exhausted retries: close websocket, frontend → polling         │
└──────────────────────────────────────────────────────────────────────────┘
          ↓                                           ↓
┌───────────────────────────┐         ┌────────────────────────────────────┐
│ 5a. CHAT WEBSOCKET        │         │ 5b. GENERAL NOTIFICATION           │
│    chatwebsocketshandler: │         │    generalnotificationshandler:    │
│    - L7 consistent hash   │         │    - Route by user_id              │
│      routes by stream_id  │         │    - Broadcast notification badge  │
│    - Broadcast locally    │         │    - For users NOT in chat         │
│    - Return 206 if partial│         │                                    │
└───────────────────────────┘         └────────────────────────────────────┘
          ↓                                           ↓
┌───────────────────────────┐         ┌────────────────────────────────────┐
│ 6a. REAL-TIME DELIVERY    │         │ 6b. NOTIFICATION UI                │
│    User A & B see message │         │    User B sees badge               │
│    instantly in chat      │         │    "1 new message from User A"     │
└───────────────────────────┘         └────────────────────────────────────┘
```

### Key Architectural Points

1. **Single Write**: Only write to MongoDB, CDC propagates changes (no dual-write problem)

2. **Push-based CDC**: Debezium pushes changes from oplog to Kafka (faster than polling)

3. **Idempotency**: Redis-based event store with configurable TTL prevents duplicate processing

4. **Consistent Routing**: `stream_id` enables L7 load balancer to route both users to same pod

5. **Service Segregation**: CDC consumer and websocket managers are separate for independent scaling

6. **Retry + Failover**: Configurable retry with backoff, graceful degradation on exhaustion

7. **Notification Flexibility**: Three strategies (concurrent, shared consumer, segregated) for different scales

### Failure Handling Summary

| Failure Scenario | Handling |
|------------------|----------|
| MongoDB write fails | Return HTTP 500, user retries |
| Debezium down | Messages queued in oplog, catch up when restored |
| Kafka unavailable | Debezium buffers, publishes when Kafka returns |
| CDC consumer crashes | Kafka redelivers from last committed offset |
| Duplicate Kafka message | RedisEventStore drops duplicate |
| Websocket delivery partial fail | Return HTTP 206, trigger retry |
| All retries exhausted | Close websocket, frontend polls persistence API |
| Redis down | EventStore fails open (process anyway, risk duplicates) |

### Performance Characteristics

- **Latency**: ~50-200ms end-to-end (persistence → websocket delivery)
- **Throughput**: Limited by Kafka partitions for CDC consumer, unlimited for websocket pods
- **Scalability**: Websocket services scale independently using consistent hashing
- **Idempotency**: Sub-millisecond Redis lookups, minimal overhead

---

## Conclusion

This architecture balances **reliability, scalability, and complexity**:

- **Reliability**: CDC ensures messages are never lost, idempotency prevents duplicates, retry + failover handle transient failures
- **Scalability**: Service segregation allows independent scaling, consistent routing avoids broadcast overhead
- **Complexity**: Managed through clear service boundaries, well-defined retry policies, and graceful degradation

The system is production-ready for medium scale and can evolve to Option 3 (segregated consumers) as user base grows.

For questions or improvements, refer to the code in `chatpersistencechangehandler/service/chat_message_sync.go` and related service implementations.

