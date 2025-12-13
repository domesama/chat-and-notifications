package websocket

import (
	"log/slog"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/gorilla/websocket"
)

type WebSocketConnection struct {
	conn        *websocket.Conn
	Metadata    Metadata // Generic metadata for logging and identification
	ConnectedAt time.Time
	CloseChan   chan struct{}
	writeMu     sync.Mutex // Protects concurrent writes to the websocket connection
	writeWait   time.Duration
}

type Metadata map[string][]string

func (m Metadata) ToString() string {
	bytes, err := json.Marshal(m)
	if err != nil {
		// This should never happen
		slog.Error("failed to marshal metadata to string", "error", err)
	}
	return string(bytes)
}

func NewWebSocketConnection(
	conn *websocket.Conn,
	metadata Metadata,
	writeWait time.Duration) *WebSocketConnection {
	c := WebSocketConnection{
		conn:        conn,
		Metadata:    metadata,
		ConnectedAt: time.Now(),
		CloseChan:   make(chan struct{}),
		writeWait:   writeWait,
	}

	return &c
}

// Close closes the WebSocket connection and cleanup resources
func (c *WebSocketConnection) Close() error {
	close(c.CloseChan)
	return c.conn.Close()
}

// send synchronously sends a message to the WebSocket connection
// Returns an error if the connection is closed or if the write fails
func (c *WebSocketConnection) send(data []byte) (err error) {
	select {
	case <-c.CloseChan:
		return
	default:
		c.writeMu.Lock()
		defer c.writeMu.Unlock()

		if err = c.conn.SetWriteDeadline(time.Now().Add(c.writeWait)); err != nil {
			return
		}

		if err = c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return
		}
	}

	return
}
