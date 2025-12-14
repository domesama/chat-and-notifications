package generalnotificationhandlerstub

import (
	"net/http"

	"github.com/domesama/chat-and-notifications/ittest/stub"
	"github.com/domesama/chat-and-notifications/model"
	"github.com/gin-gonic/gin"
)

func GeneralNotificationRouterStub(engine *gin.Engine) *ChatNotificationHandlerStub {
	chatForwarderHandler := ChatNotificationHandlerStub{}

	engine.POST(
		"/notifications/chat", chatForwarderHandler.chatNotificationHandler,
	)

	return &chatForwarderHandler
}

type ChatNotificationHandlerStub struct {
	chatNotiStub []ChatNotificationStub
}

func (c *ChatNotificationHandlerStub) AddForwardChatMessageToWebSocketStub(stub ...ChatNotificationStub) {
	c.chatNotiStub = append(c.chatNotiStub, stub...)
}

type ChatNotificationStub struct {
	Predicates     stub.Predicates[model.ChatMessage]
	StubStatusCode int
}

func (c *ChatNotificationHandlerStub) chatNotificationHandler(gctx *gin.Context) {
	var chatMessage model.ChatMessage
	if err := gctx.ShouldBindJSON(&chatMessage); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, s := range c.chatNotiStub {
		if s.Predicates.IsSatisfied(gctx, chatMessage) {
			gctx.JSON(s.StubStatusCode, gin.H{})
		}
	}
}
