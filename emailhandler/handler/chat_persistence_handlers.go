package handler

import (
	"net/http"

	"github.com/domesama/chat-and-notifications/emailhandler/service"
	"github.com/domesama/chat-and-notifications/model"
	"github.com/gin-gonic/gin"
)

// @@wire-struct@@
type EmailHandler struct {
	ChatMailingService     service.ChatMailingService
	PurchaseMailingService service.PurchaseMailingService
}

func (c EmailHandler) HandleChatMailing(gctx *gin.Context) {
	var chatMessage model.ChatMessage
	if err := gctx.ShouldBindJSON(&chatMessage); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.ChatMailingService.NotifyChatMessageToReceiverEmail(gctx.Request.Context(), chatMessage)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send email"})
		return
	}

	gctx.JSON(http.StatusCreated, chatMessage)
}

func (c EmailHandler) HandlePurchaseMailing(gctx *gin.Context) {
	var purchaseUpdate model.PurchaseUpdate
	if err := gctx.ShouldBindJSON(&purchaseUpdate); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.PurchaseMailingService.NotifyPurchaseToShopOwnerEmail(gctx.Request.Context(), purchaseUpdate)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send email"})
		return
	}

	gctx.JSON(http.StatusCreated, purchaseUpdate)
}
