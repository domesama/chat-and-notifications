package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/domesama/chat-and-notifications/chat"
	"github.com/domesama/chat-and-notifications/websocket"
	"github.com/gin-gonic/gin"
)

// @@wire-struct@@
type ChatWebSocketHandler struct {
	WebSocketManager websocket.WebSocketManager
}

func (c ChatWebSocketHandler) ForwardChatMessageToSubscribers(gctx *gin.Context) {
	ctx := gctx.Request.Context()

	var chatMessage chat.ChatMessage
	if err := gctx.ShouldBindJSON(&chatMessage); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Marshal the chat message to JSON for broadcasting
	messageData, err := json.Marshal(chatMessage)
	if err != nil {
		slog.Error("failed to marshal chat message", "error", err)
		gctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize message"})
		return
	}

	// Broadcast to all subscribers
	deliveredCount, err := c.WebSocketManager.BroadcastPayloadToLocalSubscribers(
		ctx,
		chatMessage.StreamID,
		messageData,
	)

	if err == nil {
		gctx.JSON(
			http.StatusOK, gin.H{
				"delivered": deliveredCount,
				"stream_id": chatMessage.StreamID,
			},
		)
		return
	}

	if deliveredCount == 0 {
		gctx.JSON(
			http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"message": "failed to deliver message to any connections",
			},
		)
		return
	}

	gctx.JSON(
		http.StatusPartialContent, gin.H{
			"delivered": deliveredCount,
			"error":     err.Error(),
			"message":   "message delivered to some connections but encountered errors",
		},
	)
	return
}
