package model

import (
	"time"

	"github.com/domesama/chat-and-notifications/websocket"
)

type ChatMessage struct {
	MessageID string    `json:"message_id" bson:"_id,omitempty"`
	Content   string    `json:"content" bson:"content" binding:"required"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`

	ChatMetadata `json:",inline" bson:",inline"`
}

// ChatMetadata contains the routing information for chat messages, we also reuse these in our websocket subscriptions.
// The `form` tags are required for WebSocket subscription endpoints because:
// - WebSocket upgrade requests (GET) cannot reliably include a request body
// - Standard WebSocket clients (like gorilla/websocket) don't support sending bodies during handshake
// - Query parameters are the idiomatic way to pass metadata during WebSocket connections
type ChatMetadata struct {
	StreamID   string `json:"stream_id" bson:"stream_id" form:"stream_id" binding:"required"`
	SenderID   string `json:"sender_id" bson:"sender_id" form:"sender_id" binding:"required"`
	ReceiverID string `json:"receiver_id" bson:"receiver_id" form:"receiver_id" binding:"required"`
}

func (c ChatMetadata) ToWebSocketMetadata() websocket.Metadata {
	return websocket.Metadata{
		"stream_id":   {c.StreamID},
		"sender_id":   {c.SenderID},
		"receiver_id": {c.ReceiverID},
	}
}
