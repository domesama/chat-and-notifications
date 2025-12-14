package chatwebsockethandlerstub

import (
	"net/http"

	"github.com/domesama/chat-and-notifications/ittest/stub"
	"github.com/domesama/chat-and-notifications/model"
	"github.com/gin-gonic/gin"
)

func ChatWebSocketHandlerRouterStub(engine *gin.Engine) *ChatWebSocketForwarderStub {
	chatForwarderHandler := &ChatWebSocketForwarderStub{}

	engine.POST(
		"/chat/forward-to-websocket", chatForwarderHandler.forwardChatMessageToSubscribers,
	)

	return chatForwarderHandler
}

type ChatWebSocketForwarderStub struct {
	chatWebsocketForwarderStub []ChatWebSocketForwardAPIStub
}

func (c *ChatWebSocketForwarderStub) AddForwardChatMessageToWebSocketStub(stub ...ChatWebSocketForwardAPIStub) {
	c.chatWebsocketForwarderStub = append(c.chatWebsocketForwarderStub, stub...)
}

type ChatWebSocketForwardAPIStub struct {
	Predicates     stub.Predicates[model.ChatMessage]
	StubStatusCode int
}

func (c *ChatWebSocketForwarderStub) forwardChatMessageToSubscribers(gctx *gin.Context) {
	var chatMessage model.ChatMessage
	if err := gctx.ShouldBindJSON(&chatMessage); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, s := range c.chatWebsocketForwarderStub {
		if s.Predicates.IsSatisfied(gctx, chatMessage) {
			gctx.JSON(s.StubStatusCode, gin.H{})
		}
	}
}
