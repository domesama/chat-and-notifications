package handler

import (
	"net/http"

	"github.com/domesama/chat-and-notifications/chat"
	"github.com/domesama/chat-and-notifications/websocket"
	"github.com/gin-gonic/gin"
)

func (c ChatWebSocketHandler) SubscribeChatWebSocketByStreamID(gctx *gin.Context) (
	key string, metadata websocket.Metadata,
) {
	var req chat.ChatMetadata

	// WebSocket connections pass metadata via query parameters
	// (WebSocket upgrade requests don't support request bodies)
	if err := gctx.ShouldBindQuery(&req); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key = req.StreamID
	metadata = req.ToWebSocketMetadata()
	return
}
