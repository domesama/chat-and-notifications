package handler

import (
	"github.com/domesama/chat-and-notifications/httpserverwrapper"
	"github.com/gin-gonic/gin"
)

func ProvideRouterCustomizer(handler ChatWebSocketHandler) httpserverwrapper.RouterWithWebSocketCustomizer {
	return handler
}

func (c ChatWebSocketHandler) Configure(b *httpserverwrapper.HTTPServerBuilder) error {
	b.WithMiddleware(gin.Recovery())
	return nil
}

func (c ChatWebSocketHandler) RegisterRoutes(engine *gin.Engine) error {
	engine.POST("/chat/forward-to-websocket", c.ForwardChatMessageToSubscribers)
	return nil
}

func (c ChatWebSocketHandler) RegisterWebSocketRoutes() httpserverwrapper.WebSocketRoutes {
	return httpserverwrapper.WebSocketRoutes{
		"/chat/subscribe-websocket": c.SubscribeChatWebSocketByStreamID,
	}
}
