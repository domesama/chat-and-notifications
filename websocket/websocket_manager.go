package websocket

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/domesama/concurrent"
	"github.com/gorilla/websocket"
)

type WebSocketManager interface {
	// RegisterConnection adds a new WebSocket connection to the manager
	// key: The grouping key (e.g., stream_id, room_id, user_id)
	// metadata: Connection metadata for logging and identification
	RegisterConnection(key string, metadata Metadata, conn *websocket.Conn) *WebSocketConnection

	// UnregisterConnection removes WebSocket connections matching the predicate for the given key
	UnregisterConnection(key string, predicate ConnectionPredicate)

	// BroadcastPayloadToLocalSubscribers sends a payload to all connections under the given key
	// Returns the number of connections the message was delivered to and any errors encountered
	BroadcastPayloadToLocalSubscribers(ctx context.Context, key string, message []byte) (deliveredCount int, err error)

	// CloseAll closes all WebSocket connections with the given close code and reason
	CloseAll(code int, reason string)
}

type webSocketManager struct {
	connections sync.Map // map[key][]WebSocketConnection
	WebSocketConfig
}

func ProvideDefaultWebSocketManager(cfg WebSocketConfig) WebSocketManager {
	return &webSocketManager{
		connections:     sync.Map{},
		WebSocketConfig: cfg,
	}
}

// ConnectionPredicate is a function that returns true if a connection should be removed
type ConnectionPredicate func(*WebSocketConnection) bool

func (m *webSocketManager) RegisterConnection(key string, metadata Metadata,
	conn *websocket.Conn) *WebSocketConnection {
	c := NewWebSocketConnection(conn, metadata, m.WriteWait)

	connsInterface, _ := m.connections.LoadOrStore(key, []*WebSocketConnection{})
	conns := connsInterface.([]*WebSocketConnection)

	conns = append(conns, c)
	m.connections.Store(key, conns)

	slog.Info(
		"WebSocket connection registered",
		"key", key,
		"metadata", metadata,
		"connected_at", c.ConnectedAt,
	)

	go m.handleHeartbeat(c, conn)

	return c
}

func (m *webSocketManager) UnregisterConnection(key string, predicate ConnectionPredicate) {
	connsInterface, ok := m.connections.Load(key)
	if !ok {
		return
	}

	conns := connsInterface.([]*WebSocketConnection)
	newConns := make([]*WebSocketConnection, 0, len(conns))

	for _, c := range conns {
		if predicate(c) {
			_ = c.Close()
			slog.Info(
				"WebSocket connection unregistered",
				"key", key,
				"metadata", c.Metadata,
			)
		} else {
			newConns = append(newConns, c)
		}
	}

	if len(newConns) == 0 {
		m.connections.Delete(key)
	} else {
		m.connections.Store(key, &newConns)
	}
}

func (m *webSocketManager) BroadcastPayloadToLocalSubscribers(ctx context.Context, key string, message []byte) (
	deliveredCount int, err error,
) {
	connsInterface, ok := m.connections.Load(key)
	if !ok {
		slog.Debug("no connections found for key", "key", key)
		return 0, nil
	}

	conns := connsInterface.([]*WebSocketConnection)
	if len(conns) == 0 {
		return 0, nil
	}

	// Setup concurrent websocket writes for each WebSocketConnection this key output
	emptyResult := map[string]struct{}{}

	publishToEachWebSocket := func(ctx context.Context, i int, connection *WebSocketConnection) (struct{}, error) {
		return struct{}{}, connection.send(message)
	}

	useMetadataAsKey := func(i int, from *WebSocketConnection) string {
		return from.Metadata.ToString()
	}

	publishToEachWebSocketTask := concurrent.NewSliceTask(
		conns,
		emptyResult,
		publishToEachWebSocket,
		useMetadataAsKey,
	)

	// Concurrently publish to all WebSocket connections and wait for completion
	multiError := concurrent.NewGroup(ctx).Exec(publishToEachWebSocketTask)
	deliveredCount = len(conns)

	if multiError != nil && multiError.ErrorOrNil() != nil {
		deliveredCount = deliveredCount - multiError.Len()
	}

	return deliveredCount, multiError.ErrorOrNil()
}

func (m *webSocketManager) CloseAll(code int, reason string) {
	slog.Info("closing all WebSocket connections", "code", code, "reason", reason)

	m.connections.Range(
		func(key, value interface{}) bool {
			conns := value.([]*WebSocketConnection)
			for _, c := range conns {
				closeMsg := websocket.FormatCloseMessage(code, reason)
				_ = c.conn.WriteControl(websocket.CloseMessage, closeMsg, time.Now().Add(5*time.Second))
				_ = c.Close()
			}
			return true
		},
	)

	m.connections = sync.Map{}
}

// handleHeartbeat manages ping/pong heartbeat for a connection
func (m *webSocketManager) handleHeartbeat(c *WebSocketConnection, conn *websocket.Conn) {
	ticker := time.NewTicker(m.PingInterval)
	defer ticker.Stop()

	_ = conn.SetReadDeadline(time.Now().Add(m.PongWait))
	conn.SetPongHandler(
		func(string) error {
			_ = conn.SetReadDeadline(time.Now().Add(m.PongWait))
			return nil
		},
	)

	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				slog.Info(
					"WebSocket read error, closing connection",
					"error", err,
					"metadata", c.Metadata,
				)
				return
			}
		}
	}()

	for {
		select {
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(m.WriteWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				slog.Info(
					"failed to send ping, closing connection",
					"error", err,
					"metadata", c.Metadata,
				)
				return
			}
		case <-c.CloseChan:
			return
		}
	}
}
