package handler

import (
	"net/http"

	"github.com/domesama/chat-and-notifications/generalnotifications"
	"github.com/domesama/chat-and-notifications/websocket"
	"github.com/gin-gonic/gin"
)

// SubscribeNotificationWebSocketByUserID handles WebSocket subscription by user ID
func (g GeneralNotificationWebSocketHandler) SubscribeNotificationWebSocketByUserID(gctx *gin.Context) (
	key string, metadata websocket.Metadata,
) {
	var req generalnotifications.NotificationMetadata

	if err := gctx.ShouldBindQuery(&req); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key = req.UserID
	metadata = req.ToWebSocketMetadata()
	return
}
