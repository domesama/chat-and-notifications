package handler

import (
	"net/http"

	"github.com/domesama/chat-and-notifications/chatpersistence/service"
	"github.com/domesama/chat-and-notifications/model"
	"github.com/gin-gonic/gin"
)

// @@wire-struct@@
type ChatPersistenceHandler struct {
	ChatPersistenceService service.ChatPersistenceService
}

func (c ChatPersistenceHandler) HandleChatPersistence(gctx *gin.Context) {
	var chatMessage model.ChatMessage
	if err := gctx.ShouldBindJSON(&chatMessage); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.ChatPersistenceService.PersistChatMessage(gctx.Request.Context(), chatMessage)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save message"})
		return
	}

	gctx.JSON(http.StatusCreated, chatMessage)
}
