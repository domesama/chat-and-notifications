package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/domesama/chat-and-notifications/generalnotifications"
	"github.com/domesama/chat-and-notifications/model"
	"github.com/domesama/chat-and-notifications/websocket"
	"github.com/gin-gonic/gin"
)

// @@wire-struct@@
type GeneralNotificationWebSocketHandler struct {
	WebSocketManager websocket.WebSocketManager
}

func (g GeneralNotificationWebSocketHandler) ForwardChatNotification(gctx *gin.Context) {
	g.forwardNotification(
		gctx,
		generalnotifications.ChatNotification,
		&model.ChatMessage{},
	)
}

func (g GeneralNotificationWebSocketHandler) ForwardPurchaseNotification(gctx *gin.Context) {
	g.forwardNotification(
		gctx,
		generalnotifications.PurchaseNotification,
		&model.PurchaseUpdate{},
	)
}

func (g GeneralNotificationWebSocketHandler) ForwardPaymentReminderNotification(gctx *gin.Context) {
	g.forwardNotification(
		gctx,
		generalnotifications.PaymentReminderNotification,
		&model.PaymentReminder{},
	)
}

func (g GeneralNotificationWebSocketHandler) ForwardShippingUpdateNotification(gctx *gin.Context) {
	g.forwardNotification(
		gctx,
		generalnotifications.ShippingUpdateNotification,
		&model.ShippingUpdate{},
	)
}

// forwardNotification is a generic handler that binds JSON payload and forwards to subscribers
func (g GeneralNotificationWebSocketHandler) forwardNotification(
	gctx *gin.Context,
	notificationType generalnotifications.NotificationType,
	payload any,
) {
	if err := gctx.ShouldBindJSON(payload); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var metadata generalnotifications.NotificationMetadata
	if err := gctx.ShouldBindQuery(&metadata); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"error": "user_id query parameter required"})
		return
	}

	envelope := generalnotifications.NewNotificationEnvelope(notificationType, payload)

	g.forwardNotificationToSubscribers(gctx, metadata.UserID, envelope)
}

// forwardNotificationToSubscribers is a shared helper that broadcasts notification envelope to subscribers
func (g GeneralNotificationWebSocketHandler) forwardNotificationToSubscribers(
	gctx *gin.Context,
	userID string,
	envelope generalnotifications.NotificationEnvelope,
) {
	ctx := gctx.Request.Context()

	// Marshal the notification envelope to JSON for broadcasting
	messageData, err := json.Marshal(envelope)
	if err != nil {
		slog.Error("failed to marshal notification envelope", "error", err)
		gctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize notification"})
		return
	}

	deliveredCount, err := g.WebSocketManager.BroadcastPayloadToLocalSubscribers(
		ctx,
		userID,
		messageData,
	)

	if err == nil {
		gctx.JSON(
			http.StatusOK, gin.H{
				"delivered":         deliveredCount,
				"user_id":           userID,
				"notification_type": envelope.NotificationType,
			},
		)
		return
	}

	if deliveredCount == 0 {
		gctx.JSON(
			http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"message": "failed to deliver notification to any connections",
			},
		)
		return
	}

	gctx.JSON(
		http.StatusPartialContent, gin.H{
			"delivered": deliveredCount,
			"error":     err.Error(),
			"message":   "notification delivered to some connections but encountered errors",
		},
	)
}
