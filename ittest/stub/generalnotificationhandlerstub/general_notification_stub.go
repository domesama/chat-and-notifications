package generalnotificationhandlerstub

import (
	"net/http"

	"github.com/domesama/chat-and-notifications/chat"
	"github.com/domesama/chat-and-notifications/ittest/stub"
	"github.com/gin-gonic/gin"
)

func GeneralNotificationRouterStub(engine *gin.Engine) *ChatNotificationHandlerStub {
	chatForwarderHandler := ChatNotificationHandlerStub{}

	engine.POST(
		"/noti/chat", chatForwarderHandler.chatNotificationHandler,
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
	Predicates     stub.Predicates[chat.ChatMessage]
	StubStatusCode int
}

func (c *ChatNotificationHandlerStub) chatNotificationHandler(gctx *gin.Context) {
	var chatMessage chat.ChatMessage
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
