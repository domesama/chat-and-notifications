package websocket

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// SubscribeToWebSocket establishes a WebSocket connection and returns a channel for receiving messages.
// It spawns a goroutine to continuously read messages from the WebSocket and unmarshal them into type T.
func SubscribeToWebSocket[T any](ctx context.Context, wsURL string) (chan T, func(), error) {
	dialer := &websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	headers := http.Header{}
	headers.Set("Content-Type", "application/json")

	conn, resp, err := dialer.DialContext(ctx, wsURL, headers)
	if err != nil {
		return nil, nil, err
	}
	if resp != nil {
		_ = resp.Body.Close()
	}

	// Create buffered channel for messages
	msgChan := make(chan T, 10)

	// Start goroutine to read messages from WebSocket
	go func() {
		defer close(msgChan)

		for {
			select {
			case <-ctx.Done():
				slog.Info("WebSocket context cancelled, closing connection")
				return
			default:
				_, message, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
						slog.Error("WebSocket read error", "error", err)
					}
					return
				}

				var msg T
				if err := json.Unmarshal(message, &msg); err != nil {
					slog.Error("Failed to unmarshal WebSocket message", "error", err, "message", string(message))
					continue
				}

				msgChan <- msg
			}
		}
	}()

	// Cleanup function
	cleanup := func() {
		_ = conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		)
		_ = conn.Close()
	}

	return msgChan, cleanup, nil
}
